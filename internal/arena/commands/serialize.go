package commands

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// SerializeUsage returns the help text shown for `arena help game serialize`.
func SerializeUsage(fs *pflag.FlagSet) string {
	extra := `Positional args:
  arena game <game> serialize <seed> [OPTIONS]
    <game>   engine slug (e.g. winter2026, spring2020).
    <seed>   RNG seed as int64 (decimal, signed). Same seed → same map and
             initial state, every time. Accepts an optional "seed=" prefix
             for paste compatibility with debug logs (e.g. seed=42).

Output:
  Two newline-terminated blocks on stdout, byte-for-byte the bytes a bot
  reads on stdin during a real match (no extra framing):
    1. Global init info — constants and map data sent once at game start.
    2. First-frame info — per-turn state for turn 0.
  Pipe straight into a bot binary to drive a single-turn invocation:
      arena game winter2026 serialize 42 | bin/bot-winter2026-cpp

Use cases:
  - Inspect an initial map without spinning up the full match runner.
  - Capture deterministic fixtures for unit tests (seeds reproduce maps).
  - Smoke-test a bot's parsing on a fresh seed before running a batch.
  - Debug fog-of-war / side-specific input differences with --player=0/1.`
	return arena.CommandUsage("game <game> serialize <seed>", "Print the bot-stdin bytes (globals + turn-0 frame) for a given engine seed.", fs, extra)
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
