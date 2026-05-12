package commands

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
	fs.StringP("seed", "s", "", "RNG seed as int64 (default: current Unix nanoseconds). Same seed → same map and initial state, every time. Accepts an optional \"seed=\" prefix for paste compatibility with debug logs.")
	fs.StringP("league", "l", "", "League level for the active game (game-specific; check the game's docs)")
	fs.Int("player", 0, "Whose perspective to render: 0 = engine left slot, 1 = right slot")
}

// SerializeOptions holds the parsed configuration for the "serialize" subcommand.
type SerializeOptions struct {
	Seed   int64
	Player int
}

func parseSerializeOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (SerializeOptions, error) {
	if err := fs.Parse(args); err != nil {
		return SerializeOptions{}, err
	}

	if fs.NArg() > 0 {
		return SerializeOptions{}, fmt.Errorf("unexpected positional argument %q; pass the seed via --seed/-s", fs.Arg(0))
	}

	var opts SerializeOptions

	if raw := v.GetString("seed"); raw != "" {
		seed, err := arena.ParseSeed(raw)
		if err != nil {
			return SerializeOptions{}, fmt.Errorf("invalid integer for --seed: %s", raw)
		}
		opts.Seed = seed
	} else {
		opts.Seed = time.Now().UnixNano()
	}

	opts.Player = v.GetInt("player")
	if opts.Player != 0 && opts.Player != 1 {
		return SerializeOptions{}, fmt.Errorf("--player must be 0 or 1")
	}

	return opts, nil
}
