package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
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

	var opts SerializeOptions

	if fs.NArg() < 1 {
		return SerializeOptions{}, fmt.Errorf("<seed> is required; usage: arena game serialize <game> <seed> [OPTIONS]")
	}
	seedRaw := strings.TrimSpace(fs.Arg(0))
	seed, err := arena.ParseSeed(seedRaw)
	if err != nil {
		return SerializeOptions{}, fmt.Errorf("invalid integer for seed: %s", seedRaw)
	}
	opts.Seed = seed

	opts.Player = v.GetInt("player")
	if opts.Player != 0 && opts.Player != 1 {
		return SerializeOptions{}, fmt.Errorf("--player must be 0 or 1")
	}

	return opts, nil
}
