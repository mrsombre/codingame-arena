package commands

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// SerializeUsage returns the help text shown for `arena help serialize`.
func SerializeUsage(fs *pflag.FlagSet) string {
	return arena.CommandUsage("serialize", "Print initial game input for first turn for a given seed.", fs, "")
}

// Serialize is the entry point for the "serialize" subcommand. It creates a
// game for the given seed and prints the initial global info followed by the
// first frame info for the selected player.
func Serialize(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseSerializeOptions(args, fs, v)
	if err != nil {
		return err
	}

	referee, players := factory.NewGame(opts.Seed, v)
	referee.Init(players)

	player := players[opts.Player]
	for _, line := range referee.GlobalInfoFor(player) {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return err
		}
	}
	referee.ResetGameTurnData()
	for _, line := range referee.FrameInfoFor(player) {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return err
		}
	}
	return nil
}
