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
	P0 []string
	P1 []string
}

// RunReplay re-simulates a match by feeding pre-recorded player outputs into
// the engine instead of spawning external bot processes. The returned
// TraceMatch has the same shape as TraceWriter.WriteMatch produces, so viewers
// that consume /api/matches can render replays without any format translation.
//
// botNames are copied into TraceMatch.Players (basename applied). maxTurns of
// 0 defaults to factory.MaxTurns().
func RunReplay(
	factory GameFactory,
	seed int64,
	gameOptions *viper.Viper,
	moves ReplayMoves,
	botNames [2]string,
	maxTurns int,
) TraceMatch {
	referee, players := factory.NewGame(seed, gameOptions)
	referee.Init(players)

	if maxTurns <= 0 {
		maxTurns = factory.MaxTurns()
	}

	moveLists := [2][]string{moves.P0, moves.P1}
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
	turn := 0
	for ; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		playerOutputs := [2]string{}
		var turnInput traceTurnInput

		for _, player := range players {
			if player.IsDeactivated() || referee.ShouldSkipPlayerTurn(player) {
				continue
			}
			lines := referee.FrameInfoFor(player)
			for _, line := range lines {
				player.SendInputLine(line)
			}
			if player.GetIndex() == 0 {
				turnInput.P0 = append([]string(nil), lines...)
			} else {
				turnInput.P1 = append([]string(nil), lines...)
			}
			_ = player.Execute()
			if outs := player.GetOutputs(); len(outs) > 0 {
				playerOutputs[player.GetIndex()] = strings.Join(outs, "\n")
			}
		}

		tt := TraceTurn{
			Turn:      turn,
			GameInput: turnInput,
			P0Output:  playerOutputs[0],
			P1Output:  playerOutputs[1],
			Timing:    &TraceTurnTiming{Response: [2]float64{}},
		}
		if tp, ok := referee.(TraceProvider); ok {
			tt.GameState = tp.SnapshotTurn(turn, players)
		}

		handlePlayerCommands(players, referee)

		if referee.ActivePlayers(players) < 2 {
			referee.EndGame()
			traceTurns = append(traceTurns, tt)
			break
		}

		referee.PerformGameUpdate(turn)

		if tep, ok := referee.(TurnEventProvider); ok {
			tt.Events = tep.TurnEvents(turn, players)
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

	scores := [2]int{players[0].GetScore(), players[1].GetScore()}
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

	return TraceMatch{
		MatchID:  0,
		GameID:   factory.Name(),
		PuzzleID: factory.PuzzleID(),
		Seed:     seed,
		Scores:   [2]TraceScore{TraceScore(scores[0]), TraceScore(scores[1])},
		Ranks:    RanksFromWinner(winner),
		Players:  [2]string{filepath.Base(botNames[0]), filepath.Base(botNames[1])},
		Timing:   &TraceTiming{},
		Turns:    traceTurns,
	}
}
