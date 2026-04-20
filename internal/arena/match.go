package arena

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// MatchOptions configures a single match execution.
type MatchOptions struct {
	MaxTurns    int
	P0Bin       string
	P1Bin       string
	Debug       bool
	Timing      bool
	NoSwap      bool
	TraceWriter *TraceWriter
	GameOptions map[string]string
}

// Runner executes matches using a GameFactory.
type Runner struct {
	Options MatchOptions
	Factory GameFactory
}

// NewRunner creates a Runner with the given factory and options.
func NewRunner(factory GameFactory, options MatchOptions) *Runner {
	if options.MaxTurns == 0 {
		options.MaxTurns = factory.MaxTurns()
	}
	return &Runner{Factory: factory, Options: options}
}

// RunMatch executes a single match simulation.
func (runner *Runner) RunMatch(simulationID int, seed int64) MatchResult {
	swapSides := !runner.Options.Debug && !runner.Options.NoSwap && seed%2 != 0

	referee, players := runner.Factory.NewGame(seed, runner.Options.GameOptions)
	referee.Init(players)

	matchOptions := runner.Options
	if swapSides {
		matchOptions.P0Bin, matchOptions.P1Bin = matchOptions.P1Bin, matchOptions.P0Bin
	}

	controllers, cleanup, err := attachCommandPlayers(matchOptions, players)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// Send global info.
	for _, player := range players {
		lines := referee.GlobalInfoFor(player)
		for _, line := range lines {
			player.SendInputLine(line)
		}
		if runner.Options.Debug {
			fmt.Fprintf(os.Stderr, "--- p%d global input ---\n", player.GetIndex())
			for _, line := range lines {
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}

	maxTurns := runner.Options.MaxTurns
	turn := 0
	var badCommands []BadCommandInfo
	tracing := runner.Options.TraceWriter != nil
	var traceTurns []TraceTurn

	for turn = 0; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		wasDeactivated := [2]bool{players[0].IsDeactivated(), players[1].IsDeactivated()}
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
			if tracing {
				if player.GetIndex() == 0 {
					turnInput.P0 = append([]string(nil), lines...)
				} else {
					turnInput.P1 = append([]string(nil), lines...)
				}
			}
			if runner.Options.Debug {
				fmt.Fprintf(os.Stderr, "--- turn %d p%d input ---\n", turn, player.GetIndex())
				for _, line := range lines {
					fmt.Fprintln(os.Stderr, line)
				}
			}
			_ = player.Execute()
			if outs := player.GetOutputs(); len(outs) > 0 {
				playerOutputs[player.GetIndex()] = strings.Join(outs, "\n")
			}
			if runner.Options.Debug {
				fmt.Fprintf(os.Stderr, "turn %d p%d output: %s\n", turn, player.GetIndex(), strings.Join(player.GetOutputs(), " | "))
			}
		}

		// Trace snapshot (pre-parse).
		if tracing {
			tt := TraceTurn{
				Turn:      turn,
				GameInput: turnInput,
				P0Output:  playerOutputs[0],
				P1Output:  playerOutputs[1],
			}
			if tep, ok := referee.(TurnEventProvider); ok {
				tt.Events = tep.TurnEvents(turn, players)
			}
			if tp, ok := referee.(TraceProvider); ok {
				tt.GameState = tp.SnapshotTurn(turn, players)
			}
			traceTurns = append(traceTurns, tt)
		}

		handlePlayerCommands(players, referee)

		// Detect newly deactivated players after output and command parsing.
		for i, player := range players {
			if !wasDeactivated[i] && player.IsDeactivated() {
				badCommands = append(badCommands, BadCommandInfo{
					Seed:    seed,
					Player:  i,
					Turn:    turn,
					Command: playerOutputs[i],
					Reason:  player.DeactivationReason(),
				})
			}
		}

		if referee.ActivePlayers(players) < 2 {
			referee.EndGame()
			break
		}

		referee.PerformGameUpdate(turn)
	}

	if !referee.Ended() {
		referee.EndGame()
	}
	referee.OnEnd()

	result := buildMatchResult(simulationID, seed, turn, players, controllers)
	result.BadCommands = badCommands

	// Collect game-specific metrics.
	if mp, ok := referee.(MetricsProvider); ok {
		result.Metrics = append(result.Metrics, mp.Metrics()...)
	}

	if swapSides {
		result = swapMatchSides(result)
		result.Swapped = true
	}

	if tracing {
		traceMatch := TraceMatch{
			MatchID: simulationID,
			Seed:    seed,
			Winner:  result.Winner,
			Scores:  result.Scores,
			Turns:   traceTurns,
		}
		if err := runner.Options.TraceWriter.WriteMatch(traceMatch); err != nil {
			panic(err)
		}
	}

	return result
}

func handlePlayerCommands(players []Player, referee Referee) {
	for _, player := range players {
		if player.IsDeactivated() {
			continue
		}
		err := player.GetOutputError()
		if err == nil {
			continue
		}
		player.Deactivate(err.Error())
	}
	referee.ParsePlayerOutputs(players)
}

func attachCommandPlayers(options MatchOptions, players []Player) ([]*commandPlayer, func(), error) {
	controllers := make([]*commandPlayer, 0, len(players))
	bins := []string{options.P0Bin, options.P1Bin}

	for i, path := range bins {
		cp, err := newCommandPlayer(players[i], path)
		if err != nil {
			for _, controller := range controllers {
				_ = controller.Close()
			}
			return nil, nil, fmt.Errorf("failed to start player %d session: %w", i, err)
		}
		cp.playerIdx = i
		cp.timing = options.Timing
		players[i].SetExecuteFunc(cp.Execute)
		controllers = append(controllers, cp)
	}

	return controllers, func() {
		for _, controller := range controllers {
			_ = controller.Close()
		}
	}, nil
}

func buildMatchResult(simulationID int, seed int64, turns int, players []Player, controllers []*commandPlayer) MatchResult {
	winner := -1
	if players[0].GetScore() > players[1].GetScore() {
		winner = 0
	} else if players[1].GetScore() > players[0].GetScore() {
		winner = 1
	}

	firstAnswer := [2]time.Duration{}
	turnP99 := [2]time.Duration{}
	turnMax := [2]time.Duration{}
	for i, controller := range controllers {
		stats := controller.TimingStats()
		firstAnswer[i] = stats.FirstAnswer
		turnP99[i] = stats.TurnP99
		turnMax[i] = stats.TurnMax
	}

	// Build common metrics: win/loss/draw + scores + timing.
	winsP0, winsP1, draws := 0.0, 0.0, 0.0
	switch winner {
	case 0:
		winsP0 = 1.0
	case 1:
		winsP1 = 1.0
	default:
		draws = 1.0
	}

	metrics := []Metric{
		{Label: "turns", Value: float64(turns)},
		{Label: "wins_p0", Value: winsP0},
		{Label: "wins_p1", Value: winsP1},
		{Label: "loses_p0", Value: winsP1},
		{Label: "loses_p1", Value: winsP0},
		{Label: "draws", Value: draws},
		{Label: "score_p0", Value: float64(players[0].GetScore())},
		{Label: "score_p1", Value: float64(players[1].GetScore())},
		{Label: "time_to_first_answer_p0", Value: durationMillis(firstAnswer[0])},
		{Label: "time_to_first_answer_p1", Value: durationMillis(firstAnswer[1])},
		{Label: "time_to_turn_p99_p0", Value: durationMillis(turnP99[0])},
		{Label: "time_to_turn_p99_p1", Value: durationMillis(turnP99[1])},
		{Label: "time_to_turn_max_p0", Value: durationMillis(turnMax[0])},
		{Label: "time_to_turn_max_p1", Value: durationMillis(turnMax[1])},
	}

	return MatchResult{
		ID:                simulationID,
		Seed:              seed,
		Turns:             turns,
		Scores:            [2]int{players[0].GetScore(), players[1].GetScore()},
		Winner:            winner,
		LossReasons:       [2]LossReason{lossReasonFor(players[0], winner, 0), lossReasonFor(players[1], winner, 1)},
		TimeToFirstAnswer: firstAnswer,
		TimeToTurnP99:     turnP99,
		TimeToTurnMax:     turnMax,
		Metrics:           metrics,
	}
}

func swapMatchSides(r MatchResult) MatchResult {
	r.Scores[0], r.Scores[1] = r.Scores[1], r.Scores[0]
	r.LossReasons[0], r.LossReasons[1] = r.LossReasons[1], r.LossReasons[0]
	r.TimeToFirstAnswer[0], r.TimeToFirstAnswer[1] = r.TimeToFirstAnswer[1], r.TimeToFirstAnswer[0]
	r.TimeToTurnP99[0], r.TimeToTurnP99[1] = r.TimeToTurnP99[1], r.TimeToTurnP99[0]
	r.TimeToTurnMax[0], r.TimeToTurnMax[1] = r.TimeToTurnMax[1], r.TimeToTurnMax[0]
	for i := range r.BadCommands {
		r.BadCommands[i].Player = 1 - r.BadCommands[i].Player
	}
	// Swap metric labels: build index, then swap _p0 <-> _p1 values in one pass.
	labelIdx := make(map[string]int, len(r.Metrics))
	for i, m := range r.Metrics {
		labelIdx[m.Label] = i
	}
	swapped := make(map[int]bool)
	for i, m := range r.Metrics {
		if swapped[i] {
			continue
		}
		if strings.HasSuffix(m.Label, "_p0") {
			p1Label := strings.TrimSuffix(m.Label, "_p0") + "_p1"
			if j, ok := labelIdx[p1Label]; ok {
				r.Metrics[i].Value, r.Metrics[j].Value = r.Metrics[j].Value, r.Metrics[i].Value
				swapped[i] = true
				swapped[j] = true
			}
		}
	}
	switch r.Winner {
	case 0:
		r.Winner = 1
	case 1:
		r.Winner = 0
	}
	return r
}

func lossReasonFor(player Player, winner, playerIndex int) LossReason {
	if player.IsTimedOut() {
		return LossReasonTimeout
	}
	if player.IsDeactivated() {
		return LossReasonBadCommand
	}
	if winner >= 0 && winner != playerIndex {
		return LossReasonScore
	}
	return LossReasonNone
}
