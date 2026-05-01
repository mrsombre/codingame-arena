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

type commandSpec struct {
	addFlags     func(*pflag.FlagSet)
	handler      handlerFunc
	needsFactory bool
	usage        func(*pflag.FlagSet) string
	subcommands  map[string]commandSpec
	subUsage     func() string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if err := execute(args, stdout); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func execute(args []string, stdout io.Writer) error {
	games := arena.Games()
	if len(games) == 0 {
		return errors.New("no engines registered")
	}

	if len(args) == 0 {
		return writeLine(stdout, arena.Usage(games))
	}

	command, rest := normalizeCommand(args)
	if command == "help" {
		return printHelp(stdout, rest, games)
	}

	spec, rest, err := selectCommand(command, rest)
	if err != nil {
		return err
	}
	if spec.handler == nil {
		return writeLine(stdout, spec.subUsage())
	}

	fs := arena.NewBaseFlagSet("arena")
	spec.addFlags(fs)
	if err := fs.Parse(rest); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return fmt.Errorf(`run "arena help %s" for usage`, command)
		}
		return err
	}

	v, err := arena.NewViper(fs)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	var factory arena.GameFactory
	if spec.needsFactory {
		factory, err = resolveFactory(v, games)
		if err != nil {
			return err
		}
	}

	return spec.handler(rest, stdout, factory, fs, v)
}

func commandSet() map[string]commandSpec {
	return map[string]commandSpec{
		"run": {
			addFlags:     commands.AddRunFlags,
			handler:      commands.Run,
			needsFactory: true,
			usage:        commands.RunUsage,
		},
		"analyze": {
			addFlags: commands.AddAnalyzeFlags,
			handler:  commands.Analyze,
			usage:    commands.AnalyzeUsage,
		},
		"convert": {
			addFlags:     commands.AddConvertFlags,
			handler:      commands.Convert,
			needsFactory: true,
			usage:        commands.ConvertUsage,
		},
		"serve": {
			addFlags:     commands.AddServeFlags,
			handler:      commands.Serve,
			needsFactory: true,
			usage:        commands.ServeUsage,
		},
		"serialize": {
			addFlags:     commands.AddSerializeFlags,
			handler:      commands.Serialize,
			needsFactory: true,
			usage:        commands.SerializeUsage,
		},
		"replay": {
			subUsage: commands.ReplayUsage,
			subcommands: map[string]commandSpec{
				"get": {
					addFlags:     commands.AddReplayGetFlags,
					handler:      commands.ReplayGet,
					needsFactory: true,
					usage:        commands.ReplayGetUsage,
				},
				"leaderboard": {
					addFlags:     commands.AddReplayLeaderboardFlags,
					handler:      commands.ReplayLeaderboard,
					needsFactory: true,
					usage:        commands.ReplayLeaderboardUsage,
				},
			},
		},
	}
}

func normalizeCommand(args []string) (string, []string) {
	command, rest := args[0], args[1:]
	if command == "--help" || command == "-h" {
		return "help", nil
	}
	// If the first token isn't a subcommand name, assume implicit "run"
	// and feed every arg as a flag (preserves `arena --p0 ...` ergonomics).
	if strings.HasPrefix(command, "-") {
		return "run", args
	}
	return command, rest
}

func selectCommand(command string, args []string) (commandSpec, []string, error) {
	spec, ok := commandSet()[command]
	if !ok {
		return commandSpec{}, nil, fmt.Errorf("unknown command %q; run `arena help` for usage", command)
	}
	if len(spec.subcommands) == 0 {
		return spec, args, nil
	}
	if len(args) == 0 {
		return spec, args, nil
	}

	subcommand := args[0]
	sub, ok := spec.subcommands[subcommand]
	if !ok {
		return commandSpec{}, nil, fmt.Errorf("unknown %s subcommand %q; run `arena help %s` for usage", command, subcommand, command)
	}
	return sub, args[1:], nil
}

// printHelp dispatches `arena help [command [subcommand]]` and prints the
// matching usage text to stdout.
func printHelp(stdout io.Writer, args []string, games []string) error {
	if len(args) == 0 {
		return writeLine(stdout, arena.Usage(games))
	}

	spec, ok := commandSet()[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q; run `arena help` for usage", args[0])
	}
	if len(spec.subcommands) > 0 {
		if len(args) < 2 {
			return writeLine(stdout, spec.subUsage())
		}
		sub, ok := spec.subcommands[args[1]]
		if !ok {
			return fmt.Errorf("unknown %s subcommand %q; run `arena help %s` for usage", args[0], args[1], args[0])
		}
		spec = sub
	}

	fs := arena.NewBaseFlagSet("arena")
	spec.addFlags(fs)
	return writeLine(stdout, spec.usage(fs))
}

func writeLine(w io.Writer, line string) error {
	_, err := fmt.Fprintln(w, line)
	return err
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
