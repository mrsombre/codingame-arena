package commands

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Serialize is the entry point for the "serialize" subcommand. It creates a
// game for the given seed and prints the initial global info followed by the
// first frame info for the selected player.
func Serialize(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	knownArgs, gameOptions, err := arena.SplitArgs(args, fs)
	if err != nil {
		return err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return err
	}

	arena.MergeConfigGameOptions(v, fs, gameOptions)
	arena.InjectLeague(v, gameOptions)

	if v.GetBool("help") {
		_, err := fmt.Fprintln(stdout, arena.CommandUsage("serialize", "Print initial game input for first turn for a given seed.", fs, ""))
		return err
	}

	seedRaw := v.GetString("seed")
	if seedRaw == "" {
		return fmt.Errorf("--seed is required")
	}
	seed, err := arena.ParseSeed(seedRaw)
	if err != nil {
		return fmt.Errorf("invalid integer for --seed: %s", seedRaw)
	}

	playerIdx := v.GetInt("player")
	if playerIdx != 0 && playerIdx != 1 {
		return fmt.Errorf("--player must be 0 or 1")
	}

	referee, players := factory.NewGame(seed, gameOptions)
	referee.Init(players)

	player := players[playerIdx]
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
