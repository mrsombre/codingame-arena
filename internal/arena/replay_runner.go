package arena

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ReplayMoves holds per-turn outputs for each side during a replay.
// Each string is one turn's complete output (typically a single line with
// semicolon-separated commands). Strings are 0-indexed by turn.
type ReplayMoves struct {
	Left  []string
	Right []string
}

// RunReplay re-simulates a match by feeding pre-recorded player outputs into
// the engine instead of spawning external bot processes. The returned
// TraceMatch has the same shape as TraceWriter.WriteMatch produces, so viewers
// that consume /api/matches can render replays without any format translation.
// TraceMatch.Scores carries the raw pre-OnEnd value (intrinsic in-game count)
// and TraceMatch.FinalScores carries the post-OnEnd value matching CG's
// gameResult.scores; both have -1 substituted for any deactivated side.
//
// botNames are copied into TraceMatch.Players (basename applied). blueSide
// (0 or 1) is the in-match side whose FrameInfoFor lines are recorded as
// each turn's GameInput; values outside [0,1] fall back to side 0. maxTurns
// of 0 defaults to factory.MaxTurns().
func RunReplay(
	factory GameFactory,
	seed int64,
	gameOptions *viper.Viper,
	moves ReplayMoves,
	botNames [2]string,
	blueSide int,
	maxTurns int,
) (TraceMatch, [2]int) {
	if blueSide != 0 && blueSide != 1 {
		blueSide = 0
	}
	referee, players := factory.NewGame(seed, gameOptions)
	referee.Init(players)

	if maxTurns <= 0 {
		maxTurns = factory.MaxTurns()
	}

	moveLists := [2][]string{moves.Left, moves.Right}
	turnCounts := [2]int{0, 0}
	for i, player := range players {
		idx := i
		pl := player
		pl.SetExecuteFunc(func() error {
			_ = pl.ConsumeInputLines()
			list := moveLists[idx]
			turnIdx := turnCounts[idx]
			turnCounts[idx]++
			if turnIdx >= len(list) {
				pl.SetOutputs(nil)
				return nil
			}
			line := strings.TrimRight(list[turnIdx], "\r\n")
			if line == "" {
				pl.SetOutputs(nil)
				return nil
			}
			pl.SetOutputs([]string{line})
			return nil
		})
	}

	for _, player := range players {
		for _, line := range referee.GlobalInfoFor(player) {
			player.SendInputLine(line)
		}
	}

	var traceTurns []TraceTurn
	deactivationTurns := [2]int{-1, -1}
	turn := 0
	for ; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		// Skip polling and command parsing when the iteration is a no-input
		// drain step: either fewer than two players are active (a side just
		// got deactivated and the engine is wrapping up) or the engine has
		// flagged its game-over frame (Spring 2020's post-end gameTurn that
		// runs PerformGameOver and ends the match). In both cases the
		// outcome is decided; re-polling would deactivate exhausted replay
		// bots on a Timeout that doesn't exist in the recorded match.
		liveTurn := referee.ActivePlayers(players) >= 2
		if reporter, ok := referee.(GameOverFrameReporter); ok && reporter.InGameOverFrame() {
			liveTurn = false
		}

		playerOutputs := [2]string{}
		var turnInput []string

		if liveTurn {
			for _, player := range players {
				if player.IsDeactivated() || referee.ShouldSkipPlayerTurn(player) {
					continue
				}
				lines := referee.FrameInfoFor(player)
				for _, line := range lines {
					player.SendInputLine(line)
				}
				if player.GetIndex() == blueSide {
					turnInput = append([]string(nil), lines...)
				}
				_ = player.Execute()
				if outs := player.GetOutputs(); len(outs) > 0 {
					playerOutputs[player.GetIndex()] = strings.Join(outs, "\n")
				}
			}
		}

		tt := TraceTurn{
			Turn:      turn,
			GameInput: turnInput,
			Output:    playerOutputs,
			Timing:    &TraceTurnTiming{Response: [2]float64{}},
		}

		if liveTurn {
			handlePlayerCommands(players, referee)
		}

		referee.PerformGameUpdate(turn)

		for i, player := range players {
			if deactivationTurns[i] == -1 && player.IsDeactivated() {
				deactivationTurns[i] = turn
			}
		}

		if ttp, ok := referee.(TurnTraceProvider); ok {
			tt.Traces = ttp.TurnTraces(turn, players)
		}
		traceTurns = append(traceTurns, tt)
	}

	if !referee.Ended() {
		referee.EndGame()
	}

	var rawScores [2]int
	var haveRawScores bool
	if rsp, ok := referee.(RawScoresProvider); ok {
		rawScores = rsp.RawScores()
		haveRawScores = true
	}

	referee.OnEnd()

	finalScores := [2]int{players[0].GetScore(), players[1].GetScore()}
	rawTraceScores := finalScores
	if haveRawScores {
		rawTraceScores = rawScores
	}
	finalTraceScores := finalScores
	deactivated := [2]bool{deactivationTurns[0] != -1, deactivationTurns[1] != -1}
	// Match CG's gameResult.scores convention: a deactivated side's score is
	// reported as -1 even when its raw in-game count was higher. Without this
	// the trace would show a live count for a player who never finished the
	// match, disagreeing with the replay JSON for the same field.
	for i := range deactivated {
		if deactivated[i] {
			rawTraceScores[i] = -1
			finalTraceScores[i] = -1
		}
	}
	// Ranks derive from finalScores (post-OnEnd) so a deactivated side never
	// records as drawing against the survivor and tie-break adjustments the
	// engine performs in OnEnd are reflected. convert overrides this with the
	// replay's gameResult.ranks for replay traces, which carries CG-side
	// tiebreakers the engine doesn't model.
	winner := TraceWinnerFromScores(finalScores, deactivated)

	var endReason string
	if erp, ok := referee.(EndReasonProvider); ok {
		endReason = erp.EndReason(turn, players, deactivationTurns)
	}

	return TraceMatch{
		MatchID:     0,
		GameID:      factory.Name(),
		PuzzleID:    factory.PuzzleID(),
		Seed:        seed,
		EndReason:   endReason,
		Deactivated: deactivated,
		Scores:      [2]TraceScore{TraceScore(rawTraceScores[0]), TraceScore(rawTraceScores[1])},
		FinalScores: [2]TraceScore{TraceScore(finalTraceScores[0]), TraceScore(finalTraceScores[1])},
		Ranks:       RanksFromWinner(winner),
		Players:     [2]string{filepath.Base(botNames[0]), filepath.Base(botNames[1])},
		Timing:      &TraceTiming{},
		Turns:       traceTurns,
		MainTurns:   countMainTurns(traceTurns),
	}, finalScores
}

// countMainTurns returns the number of trace turns where at least one side
// produced output — i.e. player-decision turns. Phase turns (spring2021)
// and post-end frames (spring2020) carry empty Output on both sides and
// are excluded.
func countMainTurns(turns []TraceTurn) int {
	n := 0
	for _, t := range turns {
		if t.Output[0] != "" || t.Output[1] != "" {
			n++
		}
	}
	return n
}
