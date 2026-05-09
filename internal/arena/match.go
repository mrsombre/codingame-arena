package arena

import (
	"errors"
	"fmt"
	"io"
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
	TraceSink   TraceSink
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

	tracing := runner.Options.TraceSink != nil

	// Send global info to each side's stdin (each receives its own
	// per-side view — fog-of-war / player-index headers respected).
	for _, player := range players {
		lines := referee.GlobalInfoFor(player)
		for _, line := range lines {
			player.SendInputLine(line)
		}
	}

	// Capture the trace's setup lines from a canonical side-agnostic view
	// (god-mode if the referee implements TraceGlobalInfoProducer, otherwise
	// players[0]'s perspective). Independent of what bots actually received
	// so analyzers see full state regardless of fog-of-war.
	var traceSetup []string
	if tracing {
		traceSetup = captureTraceGlobalInfo(referee, players)
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
	// firstOutputTurns[i] is the loop turn index of the first turn side i was
	// prompted for output (-1 if never prompted). Lets EndReason classify
	// TIMEOUT_START as "deactivated on first output turn" without depending on
	// game-specific frame numbering — Spring 2021's first output turn is loop
	// turn 1 (after GATHERING), Spring 2020's is loop turn 0.
	firstOutputTurns := [2]int{-1, -1}
	var traceTurns []TraceTurn

	for turn = 0; !referee.Ended() && turn < maxTurns; turn++ {
		referee.ResetGameTurnData()

		// When fewer than two players are active, the engine is in its
		// game-over frame: skip subprocess polling and command parsing and
		// just drive the game forward. Mirrors Java's gameTurn else-branch,
		// where the referee runs only resetGameTurnData / performGameUpdate /
		// performGameOver / endGame on the trailing frame. The same applies
		// when the engine has flagged its post-end frame explicitly (Spring
		// 2020's gameOverFrame branch) — both sides may still be active, but
		// the outcome is decided.
		liveTurn := referee.ActivePlayers(players) >= 2
		if reporter, ok := referee.(GameOverFrameReporter); ok && reporter.InGameOverFrame() {
			liveTurn = false
		}

		wasDeactivated := [2]bool{players[0].IsDeactivated(), players[1].IsDeactivated()}
		// outputTurn[i] mirrors the gating used in the inner Execute loop below:
		// true iff side i would be prompted this turn. Captured up front so the
		// trace records the prompt regardless of whether Execute later
		// deactivates the side (an empty-output Timeout still counts as an
		// output turn — the bot was asked, just failed to answer).
		outputTurn := [2]bool{}
		if liveTurn {
			for i, player := range players {
				outputTurn[i] = !wasDeactivated[i] && !referee.ShouldSkipPlayerTurn(player)
			}
		}
		for i := range outputTurn {
			if outputTurn[i] && firstOutputTurns[i] == -1 {
				firstOutputTurns[i] = turn
			}
		}
		playerOutputs := [2]string{}
		var turnInput []string

		if liveTurn {
			for _, controller := range controllers {
				controller.BeginTurn()
			}

			// Capture the trace's gameInput once per live turn — independent of
			// which sides will be prompted below, and side-agnostic (god-mode
			// when the referee implements TraceFrameInfoProducer; otherwise
			// players[0]'s perspective). Done before any FrameInfoFor send so
			// the snapshot reflects the state bots are about to act on.
			if tracing && (outputTurn[0] || outputTurn[1]) {
				turnInput = captureTraceFrameInfo(referee, players)
			}

			for _, player := range players {
				if player.IsDeactivated() || referee.ShouldSkipPlayerTurn(player) {
					continue
				}
				lines := referee.FrameInfoFor(player)
				for _, line := range lines {
					player.SendInputLine(line)
				}
				_ = player.Execute()
				if outs := player.GetOutputs(); len(outs) > 0 {
					playerOutputs[player.GetIndex()] = strings.Join(outs, "\n")
				}
				if runner.Options.Debug {
					writeDebugStderr(os.Stderr, turn, player.GetIndex(), controllers[player.GetIndex()].TakeStderr())
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
				Turn:         turn,
				GameInput:    turnInput,
				Output:       playerOutputs,
				IsOutputTurn: outputTurn,
				Timing:       turnTiming,
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

		if tracing {
			if rsp, ok := referee.(RawScoresProvider); ok {
				traceTurns[len(traceTurns)-1].Score = rsp.RawScores()
			}
			if decorator, ok := referee.(TraceTurnDecorator); ok {
				if state := decorator.DecorateTraceTurn(turn, players); len(state) > 0 {
					traceTurns[len(traceTurns)-1].State = state
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
				per := ttp.TurnTraces(turn, players)
				dst := &traceTurns[len(traceTurns)-1].Traces
				for i := range per {
					dst[i] = append(dst[i], per[i]...)
				}
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
		// Trace uses in-match side perspective: index 0 is the left side, index 1
		// is the right side. result fields are blue/red perspective after the
		// potential swap-back; un-swap to restore the in-match view.
		//
		// Scores stores the raw pre-OnEnd value when the engine exposes one
		// (intrinsic in-game count, no tie-break adjustments), and FinalScores
		// stores the post-OnEnd value matching CG's gameResult.scores
		// convention. Ranks follow result.Winner — i.e., the post-OnEnd
		// outcome — so a tie-broken match never records as a draw, and a
		// deactivated side never ties against the survivor.
		rawTraceScores := result.Scores
		if haveRawScores {
			rawTraceScores = result.RawScores
		}
		finalTraceScores := result.Scores
		traceWinner := result.Winner
		if result.Swapped {
			rawTraceScores[0], rawTraceScores[1] = rawTraceScores[1], rawTraceScores[0]
			finalTraceScores[0], finalTraceScores[1] = finalTraceScores[1], finalTraceScores[0]
			if traceWinner != -1 {
				traceWinner = 1 - traceWinner
			}
		}
		disqualified := [2]bool{deactivationTurns[0] != -1, deactivationTurns[1] != -1}
		// Self-play traces stay in engine-truth units: Scores/FinalScores
		// carry the engine's actual accumulated values for both sides, and
		// Disqualified[i] flags whether the engine deactivated that side.
		// Replay-converted traces (see convert.go) further overwrite
		// FinalScores with CG's gameResult.scores when the match was
		// disqualified, since CG's record is the authoritative outcome
		// there.
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
			endReason = erp.EndReason(turn, players, deactivationTurns, firstOutputTurns)
		}
		traceMatch := TraceMatch{
			MatchID:     simulationID,
			PuzzleName:  runner.Factory.Name(),
			PuzzleID:    runner.Factory.PuzzleID(),
			Seed:        seed,
			Blue:        filepath.Base(runner.Options.BlueBotBin),
			League:      league,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			EndReason:   endReason,
			Disqualified: disqualified,
			Scores:      [2]TraceScore{TraceScore(rawTraceScores[0]), TraceScore(rawTraceScores[1])},
			FinalScores: [2]TraceScore{TraceScore(finalTraceScores[0]), TraceScore(finalTraceScores[1])},
			Ranks:       RanksFromWinner(traceWinner),
			Setup:       traceSetup,
			Players:     [2]string{filepath.Base(sideOptions.BlueBotBin), filepath.Base(sideOptions.RedBotBin)},
			Timing:      traceTiming,
			Turns:       traceTurns,
			MainTurns:   countMainTurns(traceTurns),
		}
		if err := runner.Options.TraceSink.WriteMatch(traceMatch); err != nil {
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
		if _, ok := errors.AsType[hardTimeoutError](err); ok {
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
		cp, err := newCommandPlayer(players[sideIndex], path, options.Debug)
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

// writeDebugStderr writes the lines a bot produced on its own stderr during
// the just-completed turn under a `--- turn N <side> stderr ---` header.
// Side is "left" for player index 0, "right" for index 1. No-op when the
// bot stayed silent.
func writeDebugStderr(w io.Writer, turn, playerIdx int, lines []string) {
	if len(lines) == 0 {
		return
	}
	side := "left"
	if playerIdx == 1 {
		side = "right"
	}
	_, _ = fmt.Fprintf(w, "--- turn %d %s stderr ---\n", turn, side)
	for _, line := range lines {
		_, _ = fmt.Fprintln(w, line)
	}
}
