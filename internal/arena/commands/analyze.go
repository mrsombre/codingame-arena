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
	return arena.CommandUsage("analyze", "Analyze trace outcomes and game-owned metrics.", fs, "")
}

// Analyze is the entry point for the "analyze" subcommand.
func Analyze(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseAnalyzeOptions(args, fs, v)
	if err != nil {
		return err
	}

	input, metricAnalyzer, err := buildTraceAnalysisInput(opts.TraceDir, v.GetString("game"))
	if err != nil {
		return err
	}

	report, err := arena.AnalyzeTraceFiles(input, metricAnalyzer)
	if err != nil {
		return err
	}
	return report.Write(stdout)
}

func buildTraceAnalysisInput(traceDir, configuredGame string) (arena.TraceAnalysisInput, arena.TraceMetricAnalyzer, error) {
	files, err := loadAnalyzeTraceFiles(traceDir)
	if err != nil {
		return arena.TraceAnalysisInput{}, nil, err
	}
	if len(files) == 0 {
		return arena.TraceAnalysisInput{}, nil, fmt.Errorf("no arena trace JSON files found in %s", traceDir)
	}

	puzzleName, err := resolveAnalyzeGame(configuredGame, files)
	if err != nil {
		return arena.TraceAnalysisInput{}, nil, err
	}

	gameFiles := filterTraceFilesByGame(files, puzzleName)
	if len(gameFiles) == 0 {
		return arena.TraceAnalysisInput{}, nil, fmt.Errorf("no %s trace JSON files found in %s", puzzleName, traceDir)
	}

	metricAnalyzer, err := resolveTraceMetricAnalyzer(puzzleName)
	if err != nil {
		return arena.TraceAnalysisInput{}, nil, err
	}

	return arena.TraceAnalysisInput{
		TraceDir:   traceDir,
		Files:      gameFiles,
		PuzzleName: puzzleName,
	}, metricAnalyzer, nil
}

func resolveTraceMetricAnalyzer(puzzleName string) (arena.TraceMetricAnalyzer, error) {
	factory := arena.GetFactory(puzzleName)
	if factory == nil {
		return nil, fmt.Errorf("unknown game %q", puzzleName)
	}
	metricAnalyzer, _ := factory.(arena.TraceMetricAnalyzer)
	return metricAnalyzer, nil
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

func resolveAnalyzeGame(configured string, files []arena.TraceFile) (string, error) {
	if configured != "" {
		return configured, nil
	}

	seen := make(map[string]struct{})
	for _, file := range files {
		if file.Trace.PuzzleName == "" {
			continue
		}
		seen[file.Trace.PuzzleName] = struct{}{}
	}
	if len(seen) == 1 {
		for puzzleName := range seen {
			return puzzleName, nil
		}
	}
	if len(seen) == 0 {
		return "", fmt.Errorf("cannot infer game from trace files; pass --game")
	}

	games := make([]string, 0, len(seen))
	for puzzleName := range seen {
		games = append(games, puzzleName)
	}
	sort.Strings(games)
	return "", fmt.Errorf("multiple games in trace files (%v); pass --game", games)
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
