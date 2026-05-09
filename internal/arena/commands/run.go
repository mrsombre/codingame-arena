package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	extra := `Positional args:
  arena run <game> [OPTIONS]   <game> selects the engine (e.g. winter2026, spring2020)

Concurrency:
  --simulations is the total number of matches the batch will play.
  --parallel is the number of worker threads that dispatch those matches in
  parallel — purely a wall-clock speedup; results don't depend on it.
  Do NOT set --parallel above the number of CPU cores (oversubscription
  degrades throughput) and do NOT start a second ` + "`arena run`" + ` (or any
  other CPU-heavy job) on the same machine while a batch is in flight:
  workers will compete for CPU and inflate the engine's response-time
  measurements, skewing trace timings and bot stats.

Sides:
  blue = our bot, red = opponent. By default blue alternates between the engine's
  left/right slots match-by-match to neutralize positional bias (--no-swap to
  lock blue left). Win/loss/draw counts in the summary are from blue's perspective.

Seeding:
  Each match in the batch gets a deterministic per-match seed:
      seed_i = --seed + i * --seedx     (i = 0..simulations-1)
  Pin --seed for a reproducible batch; default --seed is the current Unix
  nanosecond timestamp.

Output channels:
  default      one-line summary on stdout
  --verbose    full JSON summary on stdout (per-metric averages, runner metadata,
               bad-command list, five worst losses from blue's perspective)
  --debug      forces -n=1 -p=1, locks sides, prints the match's full trace
               JSON to stdout (same shape ` + "`arena run --trace`" + ` writes to disk;
               nothing is written to --trace-dir, even if --trace is also
               set), and prints each turn's bot stderr to your terminal under
               a "--- turn N <side> stderr ---" header (silent turns omitted)

Tracing:
  --trace writes one JSON file per match to --trace-dir (default ./traces).
  Trace files feed ` + "`arena analyze`" + ` and the web viewer (` + "`arena serve`" + `).`
	return arena.CommandUsage("run <game>", "Play a batch of head-to-head matches between two bot binaries.", fs, extra)
}

// Run is the entry point for the "run" subcommand.
func Run(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseRunOptions(args, fs, v)
	if err != nil {
		return err
	}

	startedAt := time.Now()
	var debugSink *debugTraceCapture
	if opts.Debug {
		debugSink = &debugTraceCapture{traceID: startedAt.Unix()}
	}
	results := runMatches(factory, opts, v, startedAt, debugSink)
	elapsed := time.Since(startedAt)

	return writeRunOutput(stdout, opts, results, elapsed, debugSink)
}

// debugTraceCapture is the in-memory TraceSink used by --debug. Stamps the
// trace with the same TraceID/Type defaults that TraceWriter applies on
// disk, so the JSON printed to stdout is byte-for-byte the file `arena run
// --trace` would have produced — only without writing anything.
type debugTraceCapture struct {
	traceID int64
	match   arena.TraceMatch
}

func (c *debugTraceCapture) WriteMatch(m arena.TraceMatch) error {
	m.TraceID = c.traceID
	if m.Type == "" {
		m.Type = arena.TraceTypeTrace
	}
	c.match = m
	return nil
}

func runMatches(factory arena.GameFactory, opts RunOptions, v *viper.Viper, startedAt time.Time, debugSink *debugTraceCapture) []arena.MatchResult {
	matchOpts := arena.MatchOptions{
		MaxTurns:    opts.MaxTurns,
		BlueBotBin:  opts.BlueBotBin,
		RedBotBin:   opts.RedBotBin,
		Debug:       opts.Debug,
		NoSwap:      opts.NoSwap,
		GameOptions: v,
	}
	switch {
	case debugSink != nil:
		// Debug always captures in-memory and never writes to disk, so
		// --trace / --trace-dir are intentionally bypassed here.
		matchOpts.TraceSink = debugSink
	case opts.Trace:
		matchOpts.TraceSink = arena.NewTraceWriter(opts.TraceDir, startedAt.Unix())
	}

	runner := arena.NewRunner(factory, matchOpts)

	batchOpts := opts.BatchOptions
	batchOpts.Progress = newProgressLogger(opts.Simulations, os.Stderr)

	return arena.RunMatches(batchOpts, runner.RunMatch)
}

// progressStep is the report cadence: every 100 matches up to and
// including 1000 total, every 1000 thereafter. Returns 0 when no
// progress lines should fire at all (single-match runs, e.g. --debug).
func progressStep(total int) int {
	switch {
	case total <= 1:
		return 0
	case total <= 1000:
		return 100
	default:
		return 1000
	}
}

// newProgressLogger returns a BatchOptions.Progress callback that prints
// "Completed N matches" to w at every progressStep(total) milestone.
// Returns nil when no milestones would fire (saves the per-match callback
// invocation).
func newProgressLogger(total int, w io.Writer) func(completed, total int) {
	step := progressStep(total)
	if step == 0 {
		return nil
	}
	return func(completed, _ int) {
		if completed%step == 0 {
			_, _ = fmt.Fprintf(w, "Completed %d matches\n", completed)
		}
	}
}

func writeRunOutput(stdout io.Writer, opts RunOptions, results []arena.MatchResult, elapsed time.Duration, debugSink *debugTraceCapture) error {
	if opts.Debug {
		data, err := json.MarshalIndent(debugSink.match, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		_, err = stdout.Write(data)
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
