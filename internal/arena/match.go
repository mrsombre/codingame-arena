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
	BlueBotBin  string
	RedBotBin   string
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
	bluePlaysRight := !runner.Options.Debug && !runner.Options.NoSwap && seed%2 != 0

	referee, players := runner.Factory.NewGame(seed, runner.Options.GameOptions)
	referee.Init(players)

	sideOptions := runner.Options
	if bluePlaysRight {
		sideOptions.BlueBotBin, sideOptions.RedBotBin = sideOptions.RedBotBin, sideOptions.BlueBotBin
	}

	controllers, cleanup, err := attachCommandPlayers(sideOptions, players)
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
	deactivationTurns := [2]int{-1, -1}
	tracing := runner.Options.TraceWriter != nil
	var traceTurns []TraceTurn

	for turn = 0; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		// When fewer than two players are active, the engine is in its
		// game-over frame: skip subprocess polling and command parsing and
		// just drive the game forward. Mirrors Java's gameTurn else-branch,
		// where the referee runs only resetGameTurnData / performGameUpdate /
		// performGameOver / endGame on the trailing frame.
		liveTurn := referee.ActivePlayers(players) >= 2

		wasDeactivated := [2]bool{players[0].IsDeactivated(), players[1].IsDeactivated()}
		playerOutputs := [2]string{}
		var turnInput []string
		// Blue is our bot identity from --blue; after seed-driven swap, blue
		// plays the right side. Capturing only blue's view keeps traces compact;
		// for symmetric-input games either side would yield the same lines.
		blueSideIndex := 0
		if bluePlaysRight {
			blueSideIndex = 1
		}

		if liveTurn {
			for _, controller := range controllers {
				controller.BeginTurn()
			}

			for _, player := range players {
				if player.IsDeactivated() || referee.ShouldSkipPlayerTurn(player) {
					continue
				}
				lines := referee.FrameInfoFor(player)
				for _, line := range lines {
					player.SendInputLine(line)
				}
				if tracing && player.GetIndex() == blueSideIndex {
					turnInput = append([]string(nil), lines...)
				}
				if runner.Options.Debug {
					side := "left"
					if player.GetIndex() == 1 {
						side = "right"
					}
					fmt.Fprintf(os.Stderr, "--- turn %d %s input ---\n", turn, side)
					for _, line := range lines {
						fmt.Fprintln(os.Stderr, line)
					}
				}
				_ = player.Execute()
				if outs := player.GetOutputs(); len(outs) > 0 {
					playerOutputs[player.GetIndex()] = strings.Join(outs, "\n")
				}
				if runner.Options.Debug {
					side := "left"
					if player.GetIndex() == 1 {
						side = "right"
					}
					fmt.Fprintf(os.Stderr, "turn %d %s output: %s\n", turn, side, strings.Join(player.GetOutputs(), " | "))
				}
			}
		}

		// Trace snapshot (pre-parse). Traces are attached after PerformGameUpdate,
		// since they describe outcomes of this turn's moves.
		if tracing {
			turnTiming := &TraceTurnTiming{}
			if liveTurn {
				turnTiming.Response = [2]float64{
					durationMillis(controllers[0].LastOutputDuration()),
					durationMillis(controllers[1].LastOutputDuration()),
				}
			}
			traceTurns = append(traceTurns, TraceTurn{
				Turn:      turn,
				GameInput: turnInput,
				Output:    playerOutputs,
				Timing:    turnTiming,
			})
		}

		if liveTurn {
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
		}

		referee.PerformGameUpdate(turn)

		// Record turn-of-deactivation for any side that became deactivated
		// this turn (covers both bad-command parses earlier in the iteration
		// and engine-driven deactivations during PerformGameUpdate).
		for i, player := range players {
			if deactivationTurns[i] == -1 && player.IsDeactivated() {
				deactivationTurns[i] = turn
			}
		}

		if tracing {
			if ttp, ok := referee.(TurnTraceProvider); ok {
				traceTurns[len(traceTurns)-1].Traces = ttp.TurnTraces(turn, players)
			}
		}
	}

	if !referee.Ended() {
		referee.EndGame()
	}

	// Capture raw in-match scores before OnEnd so the trace records intrinsic
	// game scores rather than tie-break-adjusted values Player.GetScore may
	// return afterward.
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

	if bluePlaysRight {
		result = swapMatchSides(result)
		result.Swapped = true
	}

	if tracing {
		// Trace uses in-match side perspective: index 0 is left side, index 1
		// is right side. result fields are blue/red perspective after the
		// potential swap-back; un-swap to restore the in-match view.
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
		deactivated := [2]bool{deactivationTurns[0] != -1, deactivationTurns[1] != -1}
		// Derive winner from raw scores so traces stay self-consistent even
		// when tie-break adjustments made result.Winner differ from the raw
		// outcome. Deactivation overrides a raw-score tie so a timed-out side
		// is never recorded as drawing against the survivor.
		if haveRawScores {
			traceWinner = TraceWinnerFromScores(traceScores, deactivated)
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
		league := 0
		if lr, ok := runner.Factory.(LeagueResolver); ok {
			league = lr.ResolveLeague(runner.Options.GameOptions)
		}
		var endReason string
		if erp, ok := referee.(EndReasonProvider); ok {
			endReason = erp.EndReason(turn, players, deactivationTurns)
		}
		traceMatch := TraceMatch{
			MatchID:     simulationID,
			GameID:      runner.Factory.Name(),
			PuzzleID:    runner.Factory.PuzzleID(),
			Seed:        seed,
			Blue:        filepath.Base(runner.Options.BlueBotBin),
			League:      league,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			EndReason:   endReason,
			Deactivated: deactivated,
			Scores:      [2]TraceScore{TraceScore(traceScores[0]), TraceScore(traceScores[1])},
			Ranks:       RanksFromWinner(traceWinner),
			Players:     [2]string{filepath.Base(sideOptions.BlueBotBin), filepath.Base(sideOptions.RedBotBin)},
			Timing:      traceTiming,
			Turns:       traceTurns,
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
	sideBotBins := []string{options.BlueBotBin, options.RedBotBin}

	for sideIndex, path := range sideBotBins {
		cp, err := newCommandPlayer(players[sideIndex], path)
		if err != nil {
			for _, controller := range controllers {
				_ = controller.Close()
			}
			return nil, nil, fmt.Errorf("failed to start side %d session: %w", sideIndex, err)
		}
		// Spawn eagerly so subprocess startup runs in parallel with referee
		// setup and global-info dispatch instead of being charged to the
		// first-turn timer.
		if err := cp.start(); err != nil {
			_ = cp.Close()
			for _, controller := range controllers {
				_ = controller.Close()
			}
			return nil, nil, fmt.Errorf("failed to start side %d session: %w", sideIndex, err)
		}
		cp.playerIdx = sideIndex
		players[sideIndex].SetExecuteFunc(cp.Execute)
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

	// Build common side metrics: win/loss/draw + scores + timing.
	leftWins, rightWins, draws := 0.0, 0.0, 0.0
	switch winner {
	case 0:
		leftWins = 1.0
	case 1:
		rightWins = 1.0
	default:
		draws = 1.0
	}

	metrics := []Metric{
		{Label: "turns", Value: float64(turns)},
		{Label: "wins_blue", Value: leftWins},
		{Label: "wins_red", Value: rightWins},
		{Label: "loses_blue", Value: rightWins},
		{Label: "loses_red", Value: leftWins},
		{Label: "draws", Value: draws},
		{Label: "score_blue", Value: float64(players[0].GetScore())},
		{Label: "score_red", Value: float64(players[1].GetScore())},
		{Label: "ttfo_blue", Value: durationMillis(ttfo[0])},
		{Label: "ttfo_red", Value: durationMillis(ttfo[1])},
		{Label: "aot_blue", Value: durationMillis(aot[0])},
		{Label: "aot_red", Value: durationMillis(aot[1])},
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
	// Swap compatibility metric labels from left/right to blue/red perspective.
	labelIdx := make(map[string]int, len(r.Metrics))
	for i, m := range r.Metrics {
		labelIdx[m.Label] = i
	}
	swapped := make(map[int]bool)
	for i, m := range r.Metrics {
		if swapped[i] {
			continue
		}
		if strings.HasSuffix(m.Label, "_blue") {
			otherSideLabel := strings.TrimSuffix(m.Label, "_blue") + "_red"
			if j, ok := labelIdx[otherSideLabel]; ok {
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
