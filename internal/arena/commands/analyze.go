package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AnalyzeUsage returns the help text shown for `arena help analyze`.
func AnalyzeUsage(fs *pflag.FlagSet) string {
	extra := `Positional args:
  arena analyze <game> [OPTIONS]   <game> selects which traces to read.
  Trace files whose puzzleName != <game> are silently ignored, so the
  same --trace-dir can hold multiple games.

Inputs:
  Reads every *.json in --trace-dir (default ./traces). Two trace shapes
  are supported and treated identically here:
    - trace-<traceId>-<matchId>.json   from ` + "`arena run --trace`" + ` (self-play)
    - replay-<replayId>.json           from ` + "`arena replay`" + `       (CodinGame replay)
  Both are filtered to <game> via the trace's puzzleName field. Files that
  don't look like arena traces (missing turns and puzzleName) are skipped.
  Files missing the required ` + "`blue`" + ` field cause a hard error so a stray
  legacy trace can't silently corrupt the report.

Output (stdout, plain text):
  HEADER          <game> — N traces — <trace-dir>
  OUTCOME         decided/draw split, plus blue-side W/L/D
  MATCH           turn count, blue-vs-red avg score and timing
  END REASONS     percentage of matches by termination reason
                  (TURNS_OUT, SCORE, ELIMINATED, TIMEOUT, INVALID, …)
  METRICS         winner-vs-loser AND blue-vs-red rollups for game-defined
                  per-turn events (e.g. winter2026 DEAD, spring2021 GATHER).
                  Game-specific; only present when the game implements
                  TraceMetricAnalyzer. Per-turn rates auto-fall back to raw
                  counts when both sides average under 1% to avoid noisy
                  ratios.
  WORST           per hazard metric, blue's worst single match and trace id
                  (skipped for HigherIsBetter metrics and never-fired ones).

Notes:
  - Win/loss is always blue's perspective (the trace's ` + "`blue`" + ` field).
  - Draws are counted in OUTCOME but excluded from winner-vs-loser metrics.
  - Arena does not interpret game-specific event labels; each game owns
    its own metric meaning via TraceMetricSpec.`
	return arena.CommandUsage("analyze <game>", "Aggregate trace files into an outcome + game-metric report.", fs, extra)
}

// Analyze is the entry point for the "analyze" subcommand.
func Analyze(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseAnalyzeOptions(args, fs, v)
	if err != nil {
		return err
	}

	input, err := buildTraceAnalysisInput(opts.TraceDir, factory)
	if err != nil {
		return err
	}

	metricAnalyzer, _ := factory.(arena.TraceMetricAnalyzer)
	report, err := arena.AnalyzeTraceFiles(input, metricAnalyzer)
	if err != nil {
		return err
	}
	return report.Write(stdout)
}

func buildTraceAnalysisInput(traceDir string, factory arena.GameFactory) (arena.TraceAnalysisInput, error) {
	files, err := loadAnalyzeTraceFiles(traceDir)
	if err != nil {
		return arena.TraceAnalysisInput{}, err
	}
	if len(files) == 0 {
		return arena.TraceAnalysisInput{}, fmt.Errorf("no arena trace JSON files found in %s", traceDir)
	}

	puzzleName := factory.Name()
	gameFiles := filterTraceFilesByGame(files, puzzleName)
	if len(gameFiles) == 0 {
		return arena.TraceAnalysisInput{}, fmt.Errorf("no %s trace JSON files found in %s", puzzleName, traceDir)
	}

	return arena.TraceAnalysisInput{
		TraceDir:   traceDir,
		Files:      gameFiles,
		PuzzleName: puzzleName,
	}, nil
}

func loadAnalyzeTraceFiles(traceDir string) ([]arena.TraceFile, error) {
	entries, err := os.ReadDir(traceDir)
	if err != nil {
		return nil, fmt.Errorf("read trace directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	files := make([]arena.TraceFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		trace, err := readAnalyzeTraceFile(filepath.Join(traceDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if !looksLikeArenaTrace(trace) {
			continue
		}
		if err := validateAnalyzeTrace(entry.Name(), trace); err != nil {
			return nil, err
		}

		files = append(files, arena.TraceFile{
			Name:  entry.Name(),
			Trace: trace,
		})
	}

	return files, nil
}

func readAnalyzeTraceFile(path string) (arena.TraceMatch, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return arena.TraceMatch{}, fmt.Errorf("read %s: %w", path, err)
	}

	var trace arena.TraceMatch
	if err := json.Unmarshal(data, &trace); err != nil {
		return arena.TraceMatch{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return trace, nil
}

func looksLikeArenaTrace(trace arena.TraceMatch) bool {
	return trace.PuzzleName != "" || len(trace.Turns) > 0
}

func validateAnalyzeTrace(name string, trace arena.TraceMatch) error {
	if trace.Blue == "" {
		return fmt.Errorf("%s: trace missing blue (re-run or re-convert; analyze requires every trace to identify the user side)", name)
	}
	if trace.BlueSide() == -1 {
		return fmt.Errorf("%s: blue %q not found in players %v", name, trace.Blue, trace.Players)
	}
	return nil
}

func filterTraceFilesByGame(files []arena.TraceFile, puzzleName string) []arena.TraceFile {
	out := make([]arena.TraceFile, 0, len(files))
	for _, file := range files {
		if file.Trace.PuzzleName == puzzleName {
			out = append(out, file)
		}
	}
	return out
}
