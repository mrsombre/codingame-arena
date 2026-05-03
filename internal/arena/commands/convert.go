package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// errReplayMismatch flags a verification failure (engine output disagrees with
// the replay) so the caller can skip writing the trace and move on instead of
// aborting the whole batch.
var errReplayMismatch = errors.New("replay mismatch")

type convertOutcome int

const (
	convertOutcomeSaved convertOutcome = iota
	convertOutcomeSkippedExisting
	convertOutcomeSkippedPuzzle
	convertOutcomeSkippedMismatch
	convertOutcomeFailed
)

type convertReplayTarget struct {
	ID   int64
	Path string
}

type convertResult struct {
	Target  convertReplayTarget
	Outcome convertOutcome
	Detail  string
}

type convertSummary struct {
	Total           int
	Saved           int
	SkippedExisting int
	SkippedPuzzle   int
	SkippedMismatch int
	Failed          int
}

// ConvertOptions bundles the knobs convertReplay needs.
type ConvertOptions struct {
	TraceDir string
	Force    bool
	League   string
}

// convertReplay processes a single target. Replay-content failures (existing
// trace, wrong puzzle, engine mismatch) become a non-error result so the
// batch keeps moving; only genuine I/O failures return an error.
func convertReplay(factory arena.GameFactory, opts ConvertOptions, target convertReplayTarget) (convertResult, error) {
	res := convertResult{Target: target}

	tracePath := filepath.Join(opts.TraceDir, arena.TraceFileName(arena.TraceTypeReplay, target.ID, 0))
	if !opts.Force {
		switch _, err := os.Stat(tracePath); {
		case err == nil:
			res.Outcome = convertOutcomeSkippedExisting
			res.Detail = "trace exists"
			return res, nil
		case !os.IsNotExist(err):
			return convertResult{}, fmt.Errorf("stat %s: %w", tracePath, err)
		}
	}

	replay, err := readConvertReplay(target.Path)
	if err != nil {
		return convertResult{}, err
	}
	if replay.PuzzleID != factory.PuzzleID() {
		res.Outcome = convertOutcomeSkippedPuzzle
		res.Detail = fmt.Sprintf("puzzleId %d != %d", replay.PuzzleID, factory.PuzzleID())
		return res, nil
	}

	trace, league, err := convertReplayTrace(factory, replay, opts.League)
	if err != nil {
		if errors.Is(err, errReplayMismatch) {
			res.Outcome = convertOutcomeSkippedMismatch
			res.Detail = err.Error()
			return res, nil
		}
		return convertResult{}, fmt.Errorf("convert replay %d: %w", target.ID, err)
	}
	applyReplayMetadata(&trace, replay)

	if err := arena.NewTraceWriter(opts.TraceDir, target.ID).WriteMatch(trace); err != nil {
		return convertResult{}, fmt.Errorf("write trace for replay %d: %w", target.ID, err)
	}

	res.Outcome = convertOutcomeSaved
	res.Detail = fmt.Sprintf("league=%d turns=%d scores=%.1f:%.1f", league, len(trace.Turns), trace.Scores[0], trace.Scores[1])
	return res, nil
}

func readConvertReplay(path string) (arena.CodinGameReplay[arena.CodinGameReplayFrame], error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return arena.CodinGameReplay[arena.CodinGameReplayFrame]{}, fmt.Errorf("read %s: %w", path, err)
	}

	var replay arena.CodinGameReplay[arena.CodinGameReplayFrame]
	if err := json.Unmarshal(data, &replay); err != nil {
		return arena.CodinGameReplay[arena.CodinGameReplayFrame]{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return replay, nil
}

func applyReplayMetadata(trace *arena.TraceMatch, replay arena.CodinGameReplay[arena.CodinGameReplayFrame]) {
	trace.MatchID = 0
	trace.Type = arena.TraceTypeReplay
	trace.Blue = replay.Blue
	trace.League = replay.League
	trace.CreatedAt = replay.FetchedAt
}

func summarizeConvertResults(results []convertResult) convertSummary {
	s := convertSummary{Total: len(results)}
	for _, r := range results {
		switch r.Outcome {
		case convertOutcomeSaved:
			s.Saved++
		case convertOutcomeSkippedExisting:
			s.SkippedExisting++
		case convertOutcomeSkippedPuzzle:
			s.SkippedPuzzle++
		case convertOutcomeSkippedMismatch:
			s.SkippedMismatch++
		case convertOutcomeFailed:
			s.Failed++
		}
	}
	return s
}

func convertReplayTrace(factory arena.GameFactory, replay arena.CodinGameReplay[arena.CodinGameReplayFrame], leagueOverride string) (arena.TraceMatch, int, error) {
	seed, ok := arena.ResolveReplaySeed(replay)
	if !ok {
		return arena.TraceMatch{}, 0, fmt.Errorf("replay missing seed")
	}

	league := arena.ParseReplayLeague(replay.QuestionTitle)
	if leagueOverride != "" {
		n, err := strconv.Atoi(leagueOverride)
		if err != nil {
			return arena.TraceMatch{}, 0, fmt.Errorf("invalid league override %q: %w", leagueOverride, err)
		}
		league = n
	}

	gameOptions := viper.New()
	if league > 0 {
		gameOptions.Set("league", strconv.Itoa(league))
	}

	if replay.Blue == "" {
		return arena.TraceMatch{}, league, fmt.Errorf("%w: replay missing blue (re-fetch with `arena replay` so the username is recorded)", errReplayMismatch)
	}
	botNames := arena.ReplayPlayerNames(replay)
	blueSide := -1
	for i, name := range botNames {
		if name == replay.Blue {
			blueSide = i
			break
		}
	}
	if blueSide == -1 {
		return arena.TraceMatch{}, league, fmt.Errorf("%w: blue %q not found in players %v", errReplayMismatch, replay.Blue, botNames)
	}

	trace, finalScores := arena.RunReplay(
		factory,
		seed,
		gameOptions,
		arena.ReplayMovesFromFrames(replay),
		botNames,
		blueSide,
		0,
	)

	emitsPostEndFrame := false
	if e, ok := factory.(arena.PostEndFrameEmitter); ok {
		emitsPostEndFrame = e.EmitsPostEndFrame()
	}
	if err := verifyReplayTrace(trace, finalScores, replay, emitsPostEndFrame); err != nil {
		return arena.TraceMatch{}, league, err
	}

	return trace, league, nil
}

// verifyReplayTrace checks the engine reproduces the replay. finalScores are
// the post-OnEnd values (Player.GetScore after tie-break and deactivation
// adjustments) — the same shape the replay's gameResult.scores carries.
// trace.Scores cannot be used for this comparison: it stores the raw pre-OnEnd
// score, which diverges whenever OnEnd touches it (e.g. ties trigger a losses
// subtraction, deactivated players become -1).
//
// emitsPostEndFrame selects the trace-turn counting strategy: see
// arena.ReplayTraceTurnCount for what it means.
func verifyReplayTrace(trace arena.TraceMatch, finalScores [2]int, replay arena.CodinGameReplay[arena.CodinGameReplayFrame], emitsPostEndFrame bool) error {
	if len(replay.GameResult.Scores) < 2 {
		return fmt.Errorf("replay scores must contain two entries")
	}
	if float64(finalScores[0]) != replay.GameResult.Scores[0] || float64(finalScores[1]) != replay.GameResult.Scores[1] {
		return fmt.Errorf("%w: score mismatch: replay=[%.1f %.1f] engine=[%d %d]",
			errReplayMismatch,
			replay.GameResult.Scores[0], replay.GameResult.Scores[1], finalScores[0], finalScores[1])
	}

	expectedTurns := arena.ReplayTraceTurnCount(replay, emitsPostEndFrame)
	if len(trace.Turns) != expectedTurns {
		return fmt.Errorf("%w: turn mismatch: replay=%d engine=%d", errReplayMismatch, expectedTurns, len(trace.Turns))
	}

	return nil
}
