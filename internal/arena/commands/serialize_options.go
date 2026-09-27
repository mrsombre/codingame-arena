package commands

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
	fs.StringP("seed", "s", "", "RNG seed as int64 (default: current Unix nanoseconds). Same seed → same map and initial state, every time. Accepts an optional \"seed=\" prefix for paste compatibility with debug logs.")
	fs.IntP("league", "l", 0, "League level for the active game (0 = game default; game-specific, check the game's docs)")
	fs.Int("player", 0, "Whose perspective to render: 0 = engine left slot, 1 = right slot")
}

// SerializeOptions holds the parsed configuration for the "serialize" subcommand.
type SerializeOptions struct {
	Seed   int64
	League int
	Player int
}

func parseSerializeOptions(args []string, fs *pflag.FlagSet) (SerializeOptions, error) {
	if err := fs.Parse(args); err != nil {
		return SerializeOptions{}, err
	}

	if fs.NArg() > 0 {
		return SerializeOptions{}, fmt.Errorf("unexpected positional argument %q; pass the seed via --seed/-s", fs.Arg(0))
	}

	var opts SerializeOptions

	if raw, _ := fs.GetString("seed"); raw != "" {
		seed, err := arena.ParseSeed(raw)
		if err != nil {
			return SerializeOptions{}, fmt.Errorf("invalid integer for --seed: %s", raw)
		}
		opts.Seed = seed
	} else {
		opts.Seed = time.Now().UnixNano()
	}

	opts.League, _ = fs.GetInt("league")
	if opts.League < 0 {
		return SerializeOptions{}, fmt.Errorf("--league must be >= 0")
	}

	opts.Player, _ = fs.GetInt("player")
	if opts.Player != 0 && opts.Player != 1 {
		return SerializeOptions{}, fmt.Errorf("--player must be 0 or 1")
	}

	return opts, nil
}
