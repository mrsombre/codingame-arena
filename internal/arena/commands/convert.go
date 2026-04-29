package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

var convertReplayFilePattern = regexp.MustCompile(`^\d+\.json$`)

// errReplayMismatch flags a verification failure (engine output disagrees with
// the replay) so the convert loop can skip writing the trace and move on
// instead of aborting the whole batch.
var errReplayMismatch = errors.New("replay mismatch")

// ConvertUsage returns the help text shown for `arena help convert`.
func ConvertUsage(fs *pflag.FlagSet) string {
	return arena.CommandUsage("convert", "Convert replay JSON files into arena trace files.", fs, "")
}

// Convert scans replay JSON files, re-simulates matching games, verifies the
// results, and writes arena trace files keyed by replay id.
func Convert(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseConvertOptions(args, fs, v)
	if err != nil {
		return err
	}

	targets, err := convertReplayTargets(opts.ReplayDir, opts.IDs)
	if err != nil {
		return err
	}

	var saved, skippedExisting, skippedPuzzle, skippedMismatch int
	for i, target := range targets {
		tracePath := filepath.Join(opts.TraceDir, arena.TraceFileName(arena.TraceTypeReplay, target.ID, 0))
		if !opts.Force {
			if _, err := os.Stat(tracePath); err == nil {
				skippedExisting++
				_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (trace exists)\n", i+1, len(targets), target.ID)
				continue
			} else if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("stat %s: %w", tracePath, err)
			}
		}

		data, err := os.ReadFile(target.Path)
		if err != nil {
			return fmt.Errorf("read %s: %w", target.Path, err)
		}

		var replay arena.CodinGameReplay[arena.CodinGameReplayFrame]
		if err := json.Unmarshal(data, &replay); err != nil {
			return fmt.Errorf("parse %s: %w", target.Path, err)
		}
		if replay.PuzzleID != factory.PuzzleID() {
			skippedPuzzle++
			_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (puzzleId %d != %d)\n", i+1, len(targets), target.ID, replay.PuzzleID, factory.PuzzleID())
			continue
		}

		trace, league, err := convertReplayTrace(factory, replay, opts.League)
		if err != nil {
			if errors.Is(err, errReplayMismatch) {
				skippedMismatch++
				_, _ = fmt.Fprintf(stdout, "[%d/%d] skip %d (%v)\n", i+1, len(targets), target.ID, err)
				continue
			}
			return fmt.Errorf("convert replay %d: %w", target.ID, err)
		}
		trace.MatchID = 0
		trace.Type = arena.TraceTypeReplay
		trace.Blue = replay.Blue
		trace.League = replay.League
		trace.CreatedAt = replay.FetchedAt

		if err := arena.NewTraceWriter(opts.TraceDir, target.ID).WriteMatch(trace); err != nil {
			return fmt.Errorf("write trace for replay %d: %w", target.ID, err)
		}

		_, _ = fmt.Fprintf(stdout, "[%d/%d] save %d (league=%d turns=%d scores=%.1f:%.1f)\n",
			i+1, len(targets), target.ID, league, len(trace.Turns), trace.Scores[0], trace.Scores[1])
		saved++
	}

	_, _ = fmt.Fprintf(stdout, "done: %d saved, %d skipped-existing, %d skipped-puzzle, %d skipped-mismatch (replays=%d out=%s)\n",
		saved, skippedExisting, skippedPuzzle, skippedMismatch, len(targets), opts.TraceDir)
	return nil
}

type convertReplayTarget struct {
	ID   int64
	Path string
}


func convertReplayTargets(replayDir string, ids []int64) ([]convertReplayTarget, error) {
	if len(ids) > 0 {
		targets := make([]convertReplayTarget, 0, len(ids))
		for _, id := range ids {
			path := filepath.Join(replayDir, fmt.Sprintf("%d.json", id))
			info, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("replay not found: %s", path)
				}
				return nil, fmt.Errorf("stat %s: %w", path, err)
			}
			if info.IsDir() {
				return nil, fmt.Errorf("replay is a directory: %s", path)
			}
			targets = append(targets, convertReplayTarget{ID: id, Path: path})
		}
		return targets, nil
	}

	entries, err := os.ReadDir(replayDir)
	if err != nil {
		return nil, fmt.Errorf("read replay directory: %w", err)
	}

	targets := make([]convertReplayTarget, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !convertReplayFilePattern.MatchString(entry.Name()) {
			continue
		}
		replayID, err := strconv.ParseInt(entry.Name()[:len(entry.Name())-len(".json")], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse replay id from %s: %w", entry.Name(), err)
		}
		targets = append(targets, convertReplayTarget{
			ID:   replayID,
			Path: filepath.Join(replayDir, entry.Name()),
		})
	}
	return targets, nil
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

	botNames := arena.ReplayPlayerNames(replay)
	blueSide := 0
	if replay.Blue != "" && botNames[1] == replay.Blue {
		blueSide = 1
	}

	trace := arena.RunReplay(
		factory,
		seed,
		gameOptions,
		arena.ReplayMovesFromFrames(replay),
		botNames,
		blueSide,
		0,
	)

	if err := verifyReplayTrace(trace, replay); err != nil {
		return arena.TraceMatch{}, league, err
	}

	return trace, league, nil
}

func verifyReplayTrace(trace arena.TraceMatch, replay arena.CodinGameReplay[arena.CodinGameReplayFrame]) error {
	if len(replay.GameResult.Scores) < 2 {
		return fmt.Errorf("replay scores must contain two entries")
	}
	if float64(trace.Scores[0]) != replay.GameResult.Scores[0] || float64(trace.Scores[1]) != replay.GameResult.Scores[1] {
		return fmt.Errorf("%w: score mismatch: replay=[%.1f %.1f] engine=[%.1f %.1f]",
			errReplayMismatch,
			replay.GameResult.Scores[0], replay.GameResult.Scores[1], trace.Scores[0], trace.Scores[1])
	}

	expectedTurns := arena.ReplayTraceTurnCount(replay)
	if len(trace.Turns) != expectedTurns {
		return fmt.Errorf("%w: turn mismatch: replay=%d engine=%d", errReplayMismatch, expectedTurns, len(trace.Turns))
	}

	return nil
}
