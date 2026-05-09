package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddRunFlags registers flags used by the "run" subcommand on fs.
func AddRunFlags(fs *pflag.FlagSet) {
	fs.StringP("blue", "b", "", "Our bot executable (required); speaks the CodinGame stdin/stdout protocol")
	fs.StringP("red", "r", filepath.Clean("./bin/opponent"), "Opponent bot executable")
	fs.IntP("simulations", "n", 100, "Number of matches to play in the batch")
	fs.IntP("parallel", "p", runtime.NumCPU(), "Concurrent match workers (default: number of CPU cores)")
	// Override the resolved integer (e.g. 10) with a zero so pflag suppresses
	// the auto-rendered "(default N)" suffix; the description explains it.
	fs.Lookup("parallel").DefValue = "0"
	fs.StringP("seed", "s", "", "Base RNG seed as int64 (default: current Unix nanoseconds)")
	fs.Int("seedx", 1, "Per-match seed stride: seed_i = seed + i*seedx (i = 0..simulations-1)")
	fs.Int("max-turns", 200, "Hard cap on turns per match before the engine ends the game")
	fs.StringP("league", "l", "", "League level (game-specific; check the game's docs for valid values)")
	fs.Bool("no-swap", false, "Disable automatic side swapping (blue is locked to the engine's left slot)")
	fs.Bool("trace", false, "Write one JSON trace file per match to --trace-dir")
	fs.String("trace-dir", "./traces", "Output directory for trace files (used with --trace; created if missing)")
	fs.Bool("verbose", false, "Print the full JSON summary on stdout instead of the one-line summary")
	fs.Bool("output-matches", false, "Include each match's result inline in the verbose JSON summary")
	fs.Bool("debug", false, "Single-match debug: forces -n=1 -p=1, locks sides, prints the match's full trace JSON to stdout (no file written, even with --trace), and prints each turn's bot stderr under a per-turn header")
}

// RunOptions holds the parsed configuration for the "run" subcommand.
type RunOptions struct {
	arena.BatchOptions
	BlueBotBin string
	RedBotBin  string
	MaxTurns   int
	TraceDir   string
	Trace      bool
	Debug      bool
	NoSwap     bool
	Verbose    bool
}

func parseRunOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (RunOptions, error) {
	if err := fs.Parse(args); err != nil {
		return RunOptions{}, err
	}

	opts := runOptionsFromConfig(v)

	if raw := v.GetString("seed"); raw != "" {
		n, err := arena.ParseSeed(raw)
		if err != nil {
			return RunOptions{}, fmt.Errorf("invalid integer for --seed: %s", raw)
		}
		opts.Seed = n
	} else {
		opts.Seed = time.Now().UnixNano()
	}

	if err := validateRunOptions(opts); err != nil {
		return RunOptions{}, err
	}
	if opts.Debug {
		opts.Simulations = 1
		opts.Parallel = 1
	}

	return opts, nil
}

func runOptionsFromConfig(v *viper.Viper) RunOptions {
	return RunOptions{
		BatchOptions: arena.BatchOptions{
			Simulations:   v.GetInt("simulations"),
			Parallel:      v.GetInt("parallel"),
			SeedIncrement: int64(v.GetInt("seedx")),
			OutputMatches: v.GetBool("output-matches"),
		},
		BlueBotBin: v.GetString("blue"),
		RedBotBin:  v.GetString("red"),
		MaxTurns:   v.GetInt("max-turns"),
		TraceDir:   v.GetString("trace-dir"),
		Trace:      v.GetBool("trace"),
		Debug:      v.GetBool("debug"),
		NoSwap:     v.GetBool("no-swap"),
		Verbose:    v.GetBool("verbose"),
	}
}

func validateRunOptions(opts RunOptions) error {
	if opts.Simulations < 1 {
		return fmt.Errorf("--simulations must be >= 1")
	}
	if opts.Parallel < 1 {
		return fmt.Errorf("--parallel must be >= 1")
	}
	if opts.MaxTurns < 1 {
		return fmt.Errorf("--max-turns must be >= 1")
	}
	if opts.SeedIncrement < 1 {
		return fmt.Errorf("--seedx must be >= 1")
	}
	if opts.BlueBotBin == "" {
		return fmt.Errorf("--blue is required")
	}
	if err := checkBotBinary("--blue", opts.BlueBotBin); err != nil {
		return err
	}
	if err := checkBotBinary("--red", opts.RedBotBin); err != nil {
		return err
	}
	return nil
}

func checkBotBinary(flag, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: file does not exist: %s", flag, path)
		}
		return fmt.Errorf("%s: %w", flag, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s: not a file: %s", flag, path)
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("%s: not executable: %s", flag, path)
	}
	return nil
}
