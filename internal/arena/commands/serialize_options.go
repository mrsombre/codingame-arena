package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// SerializeOptions holds the parsed configuration for the "serialize" subcommand.
type SerializeOptions struct {
	Seed   int64
	Player int
	Help   bool
}

func parseSerializeOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (SerializeOptions, error) {
	if err := fs.Parse(args); err != nil {
		return SerializeOptions{}, err
	}

	opts := SerializeOptions{
		Help: v.GetBool("help"),
	}
	if opts.Help {
		return opts, nil
	}

	seedRaw := v.GetString("seed")
	if seedRaw == "" {
		return SerializeOptions{}, fmt.Errorf("--seed is required")
	}
	seed, err := arena.ParseSeed(seedRaw)
	if err != nil {
		return SerializeOptions{}, fmt.Errorf("invalid integer for --seed: %s", seedRaw)
	}
	opts.Seed = seed

	opts.Player = v.GetInt("player")
	if opts.Player != 0 && opts.Player != 1 {
		return SerializeOptions{}, fmt.Errorf("--player must be 0 or 1")
	}

	return opts, nil
}
