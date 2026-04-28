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
	fs.StringP("league", "l", "", "League level (default: game-specific)")
	fs.IntP("simulations", "n", 100, "Number of matches to run")
	fs.IntP("parallel", "p", runtime.NumCPU(), "Worker threads")
	fs.StringP("seed", "s", "", "Base RNG seed (default: current time)")
	fs.Int("seedx", 1, "Seed increment per match (seed_i = seed + i*N)")
	fs.Bool("output-matches", false, "Include per-match results in JSON output")
	fs.Bool("debug", false, "Force one match, fixed sides, bot debug to stderr, match trace JSON to stdout")
	fs.Bool("no-swap", false, "Disable automatic side swapping")
	fs.Bool("trace", false, "Write per-match JSON trace files for every match")
	fs.String("trace-dir", "./traces", "Directory for trace files (used with --trace)")
	fs.Int("max-turns", 200, "Maximum turns per match")
	fs.String("p0", "", "Player 0 binary (required)")
	fs.String("p1", filepath.Clean("./bin/opponent"), "Player 1 binary")
	fs.Bool("verbose", false, "Output full JSON (default: short summary line)")
}

// RunOptions holds the parsed configuration for the "run" subcommand.
type RunOptions struct {
	arena.BatchOptions
	P0Bin    string
	P1Bin    string
	MaxTurns int
	TraceDir string
	Trace    bool
	Debug    bool
	NoSwap   bool
	Verbose  bool
}

func parseRunOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (RunOptions, error) {
	if err := fs.Parse(args); err != nil {
		return RunOptions{}, err
	}

	opts := RunOptions{
		BatchOptions: arena.BatchOptions{
			Simulations:   v.GetInt("simulations"),
			Parallel:      v.GetInt("parallel"),
			SeedIncrement: int64(v.GetInt("seedx")),
			OutputMatches: v.GetBool("output-matches"),
		},
		P0Bin:    v.GetString("p0"),
		P1Bin:    v.GetString("p1"),
		MaxTurns: v.GetInt("max-turns"),
		TraceDir: v.GetString("trace-dir"),
		Trace:    v.GetBool("trace"),
		Debug:    v.GetBool("debug"),
		NoSwap:   v.GetBool("no-swap"),
		Verbose:  v.GetBool("verbose"),
	}

	if raw := v.GetString("seed"); raw != "" {
		n, err := arena.ParseSeed(raw)
		if err != nil {
			return RunOptions{}, fmt.Errorf("invalid integer for --seed: %s", raw)
		}
		opts.Seed = n
	} else {
		opts.Seed = time.Now().UnixNano()
	}

	if opts.Simulations < 1 {
		return RunOptions{}, fmt.Errorf("--simulations must be >= 1")
	}
	if opts.Parallel < 1 {
		return RunOptions{}, fmt.Errorf("--parallel must be >= 1")
	}
	if opts.MaxTurns < 1 {
		return RunOptions{}, fmt.Errorf("--max-turns must be >= 1")
	}
	if opts.SeedIncrement < 1 {
		return RunOptions{}, fmt.Errorf("--seedx must be >= 1")
	}
	if opts.P0Bin == "" {
		return RunOptions{}, fmt.Errorf("--p0 is required")
	}
	if err := checkBotBinary("--p0", opts.P0Bin); err != nil {
		return RunOptions{}, err
	}
	if err := checkBotBinary("--p1", opts.P1Bin); err != nil {
		return RunOptions{}, err
	}
	if opts.Debug {
		opts.Simulations = 1
		opts.Parallel = 1
	}

	return opts, nil
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
