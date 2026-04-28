package arena

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// MatchOptions configures a single match execution.
type MatchOptions struct {
	MaxTurns    int
	P0Bin       string
	P1Bin       string
	Debug       bool
	NoSwap      bool
	TraceWriter *TraceWriter
	GameOptions *viper.Viper
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

	// Flush global info to subprocess stdin immediately so the bot can begin
	// reading while it warms up; turn 0 timing then reflects post-startup
	// response time rather than interpreter init.
	for _, controller := range controllers {
		if err := controller.FlushInput(); err != nil {
			panic(err)
		}
	}

	maxTurns := runner.Options.MaxTurns
	turn := 0
	var badCommands []BadCommandInfo
	tracing := runner.Options.TraceWriter != nil
	var traceTurns []TraceTurn

	for turn = 0; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		for _, controller := range controllers {
			controller.BeginTurn()
		}

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

		// Trace snapshot (pre-parse). Traces are attached after PerformGameUpdate,
		// since they describe outcomes of this turn's moves.
		if tracing {
			traceTurns = append(traceTurns, TraceTurn{
				Turn:      turn,
				GameInput: turnInput,
				P0Output:  playerOutputs[0],
				P1Output:  playerOutputs[1],
				Timing: &TraceTurnTiming{Response: [2]float64{
					durationMillis(controllers[0].LastOutputDuration()),
					durationMillis(controllers[1].LastOutputDuration()),
				}},
			})
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

		if tracing {
			if ttp, ok := referee.(TurnTraceProvider); ok {
				traceTurns[len(traceTurns)-1].Traces = ttp.TurnTraces(turn, players)
			}
		}
	}

	if !referee.Ended() {
		referee.EndGame()
	}

	// Capture raw in-match scores before OnEnd so the trace records the
	// intrinsic sum (e.g. alive bird segments) rather than the tie-break-
	// adjusted values Player.GetScore returns afterward.
	var rawScores [2]int
	var haveRawScores bool
	if rsp, ok := referee.(RawScoresProvider); ok {
		rawScores = rsp.RawScores()
		haveRawScores = true
	}

	referee.OnEnd()

	result := buildMatchResult(simulationID, seed, turn, players, controllers)
	result.BadCommands = badCommands
	if haveRawScores {
		result.RawScores = rawScores
		result.HaveRawScores = true
	}

	// Collect game-specific metrics.
	if mp, ok := referee.(MetricsProvider); ok {
		result.Metrics = append(result.Metrics, mp.Metrics()...)
	}

	if swapSides {
		result = swapMatchSides(result)
		result.Swapped = true
	}

	if tracing {
		// Trace uses in-match side perspective (p0 = left of the map).
		// result fields are user-perspective after the potential swap-back;
		// un-swap to restore the in-match view.
		traceScores := result.Scores
		if haveRawScores {
			traceScores = result.RawScores
		}
		traceWinner := result.Winner
		if result.Swapped {
			traceScores[0], traceScores[1] = traceScores[1], traceScores[0]
			if traceWinner != -1 {
				traceWinner = 1 - traceWinner
			}
		}
		// Derive winner from raw scores so traces stay self-consistent even
		// when tie-break adjustments made result.Winner differ from the raw
		// outcome.
		if haveRawScores {
			switch {
			case traceScores[0] > traceScores[1]:
				traceWinner = 0
			case traceScores[1] > traceScores[0]:
				traceWinner = 1
			default:
				traceWinner = -1
			}
		}
		stats := [2]playerTimingStats{controllers[0].TimingStats(), controllers[1].TimingStats()}
		traceTiming := &TraceTiming{
			FirstResponse: [2]float64{
				durationMillis(stats[0].TimeToFirstOutput),
				durationMillis(stats[1].TimeToFirstOutput),
			},
			ResponseAverage: [2]float64{
				durationMillis(stats[0].AverageOutputTime),
				durationMillis(stats[1].AverageOutputTime),
			},
			ResponseMedian: [2]float64{
				durationMillis(stats[0].MedianOutputTime),
				durationMillis(stats[1].MedianOutputTime),
			},
		}
		traceMatch := TraceMatch{
			MatchID:  simulationID,
			GameID:   runner.Factory.Name(),
			PuzzleID: runner.Factory.PuzzleID(),
			Seed:     seed,
			Scores:   [2]TraceScore{TraceScore(traceScores[0]), TraceScore(traceScores[1])},
			Ranks:    RanksFromWinner(traceWinner),
			Players:  [2]string{filepath.Base(matchOptions.P0Bin), filepath.Base(matchOptions.P1Bin)},
			Timing:   traceTiming,
			Turns:    traceTurns,
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
		var timeout hardTimeoutError
		if errors.As(err, &timeout) {
			player.SetTimedOut(true)
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
		// Spawn eagerly so subprocess startup runs in parallel with referee
		// setup and global-info dispatch instead of being charged to the
		// first-turn timer.
		if err := cp.start(); err != nil {
			_ = cp.Close()
			for _, controller := range controllers {
				_ = controller.Close()
			}
			return nil, nil, fmt.Errorf("failed to start player %d session: %w", i, err)
		}
		cp.playerIdx = i
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

	ttfo := [2]time.Duration{}
	aot := [2]time.Duration{}
	for i, controller := range controllers {
		stats := controller.TimingStats()
		ttfo[i] = stats.TimeToFirstOutput
		aot[i] = stats.AverageOutputTime
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
		{Label: "ttfo_p0", Value: durationMillis(ttfo[0])},
		{Label: "ttfo_p1", Value: durationMillis(ttfo[1])},
		{Label: "aot_p0", Value: durationMillis(aot[0])},
		{Label: "aot_p1", Value: durationMillis(aot[1])},
	}

	return MatchResult{
		ID:                simulationID,
		Seed:              seed,
		Turns:             turns,
		Scores:            [2]int{players[0].GetScore(), players[1].GetScore()},
		Winner:            winner,
		LossReasons:       [2]LossReason{lossReasonFor(players[0], winner, 0), lossReasonFor(players[1], winner, 1)},
		TimeToFirstOutput: ttfo,
		AverageOutputTime: aot,
		Metrics:           metrics,
	}
}

func (r MatchResult) TTFO() [2]float64 {
	return [2]float64{durationMillis(r.TimeToFirstOutput[0]), durationMillis(r.TimeToFirstOutput[1])}
}

func (r MatchResult) AOT() [2]float64 {
	return [2]float64{durationMillis(r.AverageOutputTime[0]), durationMillis(r.AverageOutputTime[1])}
}

func swapMatchSides(r MatchResult) MatchResult {
	r.Scores[0], r.Scores[1] = r.Scores[1], r.Scores[0]
	r.RawScores[0], r.RawScores[1] = r.RawScores[1], r.RawScores[0]
	r.LossReasons[0], r.LossReasons[1] = r.LossReasons[1], r.LossReasons[0]
	r.TimeToFirstOutput[0], r.TimeToFirstOutput[1] = r.TimeToFirstOutput[1], r.TimeToFirstOutput[0]
	r.AverageOutputTime[0], r.AverageOutputTime[1] = r.AverageOutputTime[1], r.AverageOutputTime[0]
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
