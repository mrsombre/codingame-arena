package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "github.com/mrsombre/codingame-arena/games"
	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/commands"
)

type handlerFunc func(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error

func main() {
	games := arena.Games()
	if len(games) == 0 {
		fmt.Fprintln(os.Stderr, "no engines registered")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(arena.Usage(games))
		return
	}

	command, rest := args[0], args[1:]
	// If the first token isn't a subcommand name, assume implicit "run"
	// and feed every arg as a flag (preserves `arena --p0 …` ergonomics).
	if strings.HasPrefix(command, "-") {
		if command == "--help" || command == "-h" {
			fmt.Println(arena.Usage(games))
			return
		}
		command, rest = "run", args
	}

	fs := arena.NewBaseFlagSet("arena")

	var handler handlerFunc
	var needsFactory bool
	switch command {
	case "serialize":
		arena.AddSerializeFlags(fs)
		handler = commands.Serialize
		needsFactory = true
	case "replay":
		arena.AddReplayFlags(fs)
		handler = commands.Replay
	case "leaderboard":
		arena.AddLeaderboardFlags(fs)
		handler = commands.Leaderboard
	case "run":
		arena.AddRunFlags(fs)
		handler = commands.Run
		needsFactory = true
	case "serve":
		arena.AddServeFlags(fs)
		handler = commands.Serve
		needsFactory = true
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q; run `arena --help` for usage\n", command)
		os.Exit(1)
	}

	v, err := arena.NewViper(fs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	var factory arena.GameFactory
	if needsFactory {
		factory, err = resolveFactory(v, games)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := handler(rest, os.Stdout, factory, fs, v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveFactory picks the active game factory from config or auto-selects
// when only one game is registered.
func resolveFactory(v *viper.Viper, games []string) (arena.GameFactory, error) {
	name := v.GetString("game")
	if name == "" {
		if len(games) == 1 {
			name = games[0]
		} else {
			return nil, fmt.Errorf("multiple games available (%s); set 'game' in arena.yml or pass ARENA_GAME env", strings.Join(games, ", "))
		}
	}
	f := arena.GetFactory(name)
	if f == nil {
		return nil, fmt.Errorf("unknown game %q; available: %s", name, strings.Join(games, ", "))
	}
	return f, nil
}
