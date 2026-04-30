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

	files, err := loadAnalyzeTraceFiles(opts.TraceDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no arena trace JSON files found in %s", opts.TraceDir)
	}

	gameID, err := resolveAnalyzeGame(v.GetString("game"), files)
	if err != nil {
		return err
	}

	gameFiles := filterTraceFilesByGame(files, gameID)
	if len(gameFiles) == 0 {
		return fmt.Errorf("no %s trace JSON files found in %s", gameID, opts.TraceDir)
	}

	factory := arena.GetFactory(gameID)
	if factory == nil {
		return fmt.Errorf("unknown game %q", gameID)
	}
	metricAnalyzer, _ := factory.(arena.TraceMetricAnalyzer)

	report, err := arena.AnalyzeTraceFiles(arena.TraceAnalysisInput{
		TraceDir: opts.TraceDir,
		Files:    gameFiles,
		GameID:   gameID,
	}, metricAnalyzer)
	if err != nil {
		return err
	}
	return report.Write(stdout)
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

		path := filepath.Join(traceDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var trace arena.TraceMatch
		if err := json.Unmarshal(data, &trace); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if !looksLikeArenaTrace(trace) {
			continue
		}

		files = append(files, arena.TraceFile{
			Name:  entry.Name(),
			Trace: trace,
		})
	}

	return files, nil
}

func looksLikeArenaTrace(trace arena.TraceMatch) bool {
	return trace.GameID != "" || len(trace.Turns) > 0
}

func resolveAnalyzeGame(configured string, files []arena.TraceFile) (string, error) {
	if configured != "" {
		return configured, nil
	}

	seen := make(map[string]struct{})
	for _, file := range files {
		if file.Trace.GameID == "" {
			continue
		}
		seen[file.Trace.GameID] = struct{}{}
	}
	if len(seen) == 1 {
		for gameID := range seen {
			return gameID, nil
		}
	}
	if len(seen) == 0 {
		return "", fmt.Errorf("cannot infer game from trace files; pass --game")
	}

	games := make([]string, 0, len(seen))
	for gameID := range seen {
		games = append(games, gameID)
	}
	sort.Strings(games)
	return "", fmt.Errorf("multiple games in trace files (%v); pass --game", games)
}

func filterTraceFilesByGame(files []arena.TraceFile, gameID string) []arena.TraceFile {
	out := make([]arena.TraceFile, 0, len(files))
	for _, file := range files {
		if file.Trace.GameID == gameID {
			out = append(out, file)
		}
	}
	return out
}
