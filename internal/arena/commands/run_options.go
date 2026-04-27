package commands

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// RunOptions holds the parsed configuration for the "run" subcommand.
type RunOptions struct {
	arena.BatchOptions
	P0Bin    string
	P1Bin    string
	MaxTurns int
	TraceDir string
	Debug    bool
	NoSwap   bool
	Verbose  bool
	Help     bool
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
		Debug:    v.GetBool("debug"),
		NoSwap:   v.GetBool("no-swap"),
		Verbose:  v.GetBool("verbose"),
		Help:     v.GetBool("help"),
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
	if !opts.Help && opts.P0Bin == "" {
		return RunOptions{}, fmt.Errorf("--p0 is required")
	}
	if opts.Debug {
		opts.Simulations = 1
		opts.Parallel = 1
	}

	return opts, nil
}
