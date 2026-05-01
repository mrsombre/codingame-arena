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
// The second return value is the post-OnEnd score reported by Player.GetScore
// for each side; convert verification compares these against the replay's
// recorded scores (which are already post-OnEnd, with -1 for deactivated
// players and tie-break adjustments applied) since TraceMatch.Scores carries
// raw bird-segment counts that diverge whenever OnEnd touched the value.
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

		// When fewer than two players are active, the engine is running its
		// game-over frame: don't poll players for outputs and don't re-parse
		// their commands (the surviving side has likely exhausted its replay
		// move list, and parsing an empty line would deactivate it and skip
		// PerformGameOver's remaining-pellets transfer). Mirror Java's
		// gameTurn else-branch, which only drives the game forward.
		liveTurn := referee.ActivePlayers(players) >= 2

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

		// CG's MultiplayerGameManager auto-ends the match when fewer than
		// two players are active. Without this the surviving side keeps
		// drifting on inertial Facing() moves until apples run out or a
		// wall takes them, producing scores that don't match the replay.
		if !referee.Ended() && referee.ActivePlayers(players) < 2 {
			referee.EndGame()
		}
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
	scores := finalScores
	winner := -1
	if haveRawScores {
		scores = rawScores
	}
	switch {
	case scores[0] > scores[1]:
		winner = 0
	case scores[1] > scores[0]:
		winner = 1
	}

	var endReason string
	if erp, ok := referee.(EndReasonProvider); ok {
		endReason = erp.EndReason(turn, players, deactivationTurns)
	}

	deactivated := [2]bool{deactivationTurns[0] != -1, deactivationTurns[1] != -1}

	return TraceMatch{
		MatchID:     0,
		GameID:      factory.Name(),
		PuzzleID:    factory.PuzzleID(),
		Seed:        seed,
		EndReason:   endReason,
		Deactivated: deactivated,
		Scores:      [2]TraceScore{TraceScore(scores[0]), TraceScore(scores[1])},
		Ranks:       RanksFromWinner(winner),
		Players:     [2]string{filepath.Base(botNames[0]), filepath.Base(botNames[1])},
		Timing:      &TraceTiming{},
		Turns:       traceTurns,
	}, finalScores
}
