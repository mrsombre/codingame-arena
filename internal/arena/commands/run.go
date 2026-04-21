package commands

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

type runnerOutput struct {
	P0Bin       string                 `json:"p0_bin"`
	P1Bin       string                 `json:"p1_bin"`
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
	SeedIncrement *int64 `json:"seed_increment,omitempty"`
	OutputMatches bool   `json:"output_matches"`
	TraceDir      string `json:"trace_dir,omitempty"`
	MaxTurns      int    `json:"max_turns"`
	NoSwap        bool   `json:"no_swap"`
	P0Left        int    `json:"p0_left"`
	P0Right       int    `json:"p0_right"`
}

// Run is the entry point for the "run" subcommand.
func Run(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	parsed, err := arena.ParseRunArgs(args, fs, v)
	if err != nil {
		return err
	}

	if parsed.Help {
		_, err = fmt.Fprintln(stdout, arena.Usage(arena.Games()))
		return err
	}

	traceWriter := arena.NewTraceWriter(parsed.TraceDir)

	runner := arena.NewRunner(factory, arena.MatchOptions{
		MaxTurns:    parsed.MaxTurns,
		P0Bin:       parsed.P0Bin,
		P1Bin:       parsed.P1Bin,
		Debug:       parsed.Debug,
		Timing:      parsed.Timing,
		NoSwap:      parsed.NoSwap,
		TraceWriter: traceWriter,
		GameOptions: parsed.GameOptions,
	})

	results := arena.RunMatches(parsed.BatchOptions, runner.RunMatch)

	p0Left := 0
	for _, r := range results {
		if !r.Swapped {
			p0Left++
		}
	}

	var allBadCommands []arena.BadCommandInfo
	for _, r := range results {
		allBadCommands = append(allBadCommands, r.BadCommands...)
	}

	out := runnerOutput{
		P0Bin: parsed.P0Bin,
		P1Bin: parsed.P1Bin,
		Runner: runnerMetadata{
			Simulations:   parsed.Simulations,
			Parallel:      parsed.Parallel,
			Seed:          parsed.Seed,
			SeedIncrement: parsed.SeedIncrement,
			OutputMatches: parsed.OutputMatches,
			TraceDir:      parsed.TraceDir,
			MaxTurns:      parsed.MaxTurns,
			NoSwap:        parsed.NoSwap,
			P0Left:        p0Left,
			P0Right:       len(results) - p0Left,
		},
		Summary:     arena.SummarizeMatches(results),
		BadCommands: allBadCommands,
	}

	worstIndices := arena.FindWorstLosses(results, 5)
	if len(worstIndices) > 0 {
		out.WorstLosses = make([]json.RawMessage, 0, len(worstIndices))
		for _, idx := range worstIndices {
			out.WorstLosses = append(out.WorstLosses, json.RawMessage(results[idx].RenderMatch()))
		}
	}

	if parsed.OutputMatches {
		out.Matches = make([]json.RawMessage, 0, len(results))
		for _, result := range results {
			out.Matches = append(out.Matches, json.RawMessage(result.RenderMatch()))
		}
	}

	if !parsed.Verbose {
		return arena.WriteShortSummary(stdout, out.Summary)
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
