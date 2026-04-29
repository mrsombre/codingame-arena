package main

import (
	"errors"
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

	if command == "help" {
		if err := printHelp(rest, games); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	fs := arena.NewBaseFlagSet("arena")

	var handler handlerFunc
	var needsFactory bool
	switch command {
	case "run":
		commands.AddRunFlags(fs)
		handler = commands.Run
		needsFactory = true
	case "analyze":
		commands.AddAnalyzeFlags(fs)
		handler = commands.Analyze
	case "convert":
		commands.AddConvertFlags(fs)
		handler = commands.Convert
		needsFactory = true
	case "serve":
		commands.AddServeFlags(fs)
		handler = commands.Serve
		needsFactory = true
	case "serialize":
		commands.AddSerializeFlags(fs)
		handler = commands.Serialize
		needsFactory = true
	case "replay":
		if len(rest) == 0 {
			fmt.Println(commands.ReplayUsage())
			return
		}
		sub := rest[0]
		rest = rest[1:]
		switch sub {
		case "get":
			commands.AddReplayGetFlags(fs)
			handler = commands.ReplayGet
		case "leaderboard":
			commands.AddReplayLeaderboardFlags(fs)
			handler = commands.ReplayLeaderboard
		default:
			fmt.Fprintf(os.Stderr, "unknown replay subcommand %q; run `arena help replay` for usage\n", sub)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q; run `arena help` for usage\n", command)
		os.Exit(1)
	}

	if err := fs.Parse(rest); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			fmt.Fprintln(os.Stderr, `run "arena help" for usage`)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
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

// printHelp dispatches `arena help [command [subcommand]]` and prints the
// matching usage text to stdout.
func printHelp(args []string, games []string) error {
	if len(args) == 0 {
		fmt.Println(arena.Usage(games))
		return nil
	}

	fs := arena.NewBaseFlagSet("arena")
	switch args[0] {
	case "run":
		commands.AddRunFlags(fs)
		fmt.Println(commands.RunUsage(fs))
	case "analyze":
		commands.AddAnalyzeFlags(fs)
		fmt.Println(commands.AnalyzeUsage(fs))
	case "convert":
		commands.AddConvertFlags(fs)
		fmt.Println(commands.ConvertUsage(fs))
	case "serve":
		commands.AddServeFlags(fs)
		fmt.Println(commands.ServeUsage(fs))
	case "serialize":
		commands.AddSerializeFlags(fs)
		fmt.Println(commands.SerializeUsage(fs))
	case "replay":
		if len(args) < 2 {
			fmt.Println(commands.ReplayUsage())
			return nil
		}
		switch args[1] {
		case "get":
			commands.AddReplayGetFlags(fs)
			fmt.Println(commands.ReplayGetUsage(fs))
		case "leaderboard":
			commands.AddReplayLeaderboardFlags(fs)
			fmt.Println(commands.ReplayLeaderboardUsage(fs))
		default:
			return fmt.Errorf("unknown replay subcommand %q; run `arena help replay` for usage", args[1])
		}
	default:
		return fmt.Errorf("unknown command %q; run `arena help` for usage", args[0])
	}
	return nil
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
