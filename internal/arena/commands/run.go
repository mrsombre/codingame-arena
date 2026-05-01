package commands

import (
	"encoding/json"
	"io"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

const worstLossLimit = 5

type runnerOutput struct {
	BlueBotBin  string                 `json:"blue_bin"`
	RedBotBin   string                 `json:"red_bin"`
	Runner      runnerMetadata         `json:"runner"`
	Summary     arena.MatchSummary     `json:"summary"`
	BadCommands []arena.BadCommandInfo `json:"bad_commands,omitempty"`
	WorstLosses []json.RawMessage      `json:"worst_losses,omitempty"`
	Matches     []json.RawMessage      `json:"matches,omitempty"`
}

type runnerMetadata struct {
	Simulations   int    `json:"simulations"`
	Parallel      int    `json:"parallel"`
	Seed          int64  `json:"seed"`
	SeedIncrement int64  `json:"seed_increment"`
	OutputMatches bool   `json:"output_matches"`
	TraceDir      string `json:"trace_dir,omitempty"`
	MaxTurns      int    `json:"max_turns"`
	NoSwap        bool   `json:"no_swap"`
	BlueLeft      int    `json:"blue_left"`
	BlueRight     int    `json:"blue_right"`
}

// RunUsage returns the help text shown for `arena help run`.
func RunUsage(fs *pflag.FlagSet) string {
	return arena.CommandUsage("run", "Run one or more match simulations against a player binary.", fs, "")
}

// Run is the entry point for the "run" subcommand.
func Run(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseRunOptions(args, fs, v)
	if err != nil {
		return err
	}

	startedAt := time.Now()
	results := runMatches(factory, opts, v, startedAt)
	elapsed := time.Since(startedAt)

	return writeRunOutput(stdout, opts, results, elapsed)
}

func runMatches(factory arena.GameFactory, opts RunOptions, v *viper.Viper, startedAt time.Time) []arena.MatchResult {
	traceDir := ""
	if opts.Trace {
		traceDir = opts.TraceDir
	}
	traceWriter := arena.NewTraceWriter(traceDir, startedAt.Unix())

	runner := arena.NewRunner(factory, arena.MatchOptions{
		MaxTurns:    opts.MaxTurns,
		BlueBotBin:  opts.BlueBotBin,
		RedBotBin:   opts.RedBotBin,
		Debug:       opts.Debug,
		NoSwap:      opts.NoSwap,
		TraceWriter: traceWriter,
		GameOptions: v,
	})

	return arena.RunMatches(opts.BatchOptions, runner.RunMatch)
}

func writeRunOutput(stdout io.Writer, opts RunOptions, results []arena.MatchResult, elapsed time.Duration) error {
	if opts.Debug {
		_, err := io.WriteString(stdout, results[0].RenderMatch())
		return err
	}

	out := buildRunnerOutput(opts, results)
	if !opts.Verbose {
		return arena.WriteShortSummary(stdout, out.Summary, elapsed)
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func buildRunnerOutput(opts RunOptions, results []arena.MatchResult) runnerOutput {
	blueLeft := countBlueLeft(results)
	out := runnerOutput{
		BlueBotBin: opts.BlueBotBin,
		RedBotBin:  opts.RedBotBin,
		Runner: runnerMetadata{
			Simulations:   opts.Simulations,
			Parallel:      opts.Parallel,
			Seed:          opts.Seed,
			SeedIncrement: opts.SeedIncrement,
			OutputMatches: opts.OutputMatches,
			TraceDir:      traceDir(opts),
			MaxTurns:      opts.MaxTurns,
			NoSwap:        opts.NoSwap,
			BlueLeft:      blueLeft,
			BlueRight:     len(results) - blueLeft,
		},
		Summary:     arena.SummarizeMatches(results),
		BadCommands: collectBadCommands(results),
	}

	worstIndices := arena.FindWorstLosses(results, worstLossLimit)
	if len(worstIndices) > 0 {
		out.WorstLosses = make([]json.RawMessage, 0, len(worstIndices))
		for _, idx := range worstIndices {
			out.WorstLosses = append(out.WorstLosses, json.RawMessage(results[idx].RenderMatch()))
		}
	}

	if opts.OutputMatches {
		out.Matches = make([]json.RawMessage, 0, len(results))
		for _, result := range results {
			out.Matches = append(out.Matches, json.RawMessage(result.RenderMatch()))
		}
	}

	return out
}

func traceDir(opts RunOptions) string {
	if !opts.Trace {
		return ""
	}
	return opts.TraceDir
}

func countBlueLeft(results []arena.MatchResult) int {
	count := 0
	for _, result := range results {
		if !result.Swapped {
			count++
		}
	}
	return count
}

func collectBadCommands(results []arena.MatchResult) []arena.BadCommandInfo {
	var all []arena.BadCommandInfo
	for _, result := range results {
		all = append(all, result.BadCommands...)
	}
	return all
}
