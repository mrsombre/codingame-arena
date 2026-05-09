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
	// argsSpec is the positional signature shown in usage errors when the
	// leading game positional is missing — e.g. "<game> <seed>" for
	// `arena game serialize`. Defaults to "<game>" when empty.
	argsSpec    string
	usage       func(*pflag.FlagSet) string
	subcommands map[string]commandSpec
	subUsage    func() string
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

	spec, path, rest, err := selectCommand(command, rest)
	if err != nil {
		return err
	}
	if spec.handler == nil {
		return writeLine(stdout, spec.subUsage())
	}

	var factory arena.GameFactory
	if spec.needsFactory {
		factory, rest, err = popGame(path, spec.argsSpec, rest, games)
		if err != nil {
			return err
		}
	}

	fs := arena.NewBaseFlagSet("arena")
	spec.addFlags(fs)
	if err := fs.Parse(rest); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return fmt.Errorf(`run "arena help %s" for usage`, path)
		}
		return err
	}

	v, err := arena.NewViper(fs)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	return spec.handler(rest, stdout, factory, fs, v)
}

func commandSet() map[string]commandSpec {
	return map[string]commandSpec{
		"run": {
			addFlags:     commands.AddRunFlags,
			handler:      commands.Run,
			needsFactory: true,
			argsSpec:     "<game>",
			usage:        commands.RunUsage,
		},
		"analyze": {
			addFlags:     commands.AddAnalyzeFlags,
			handler:      commands.Analyze,
			needsFactory: true,
			argsSpec:     "<game>",
			usage:        commands.AnalyzeUsage,
		},
		"serve": {
			addFlags: commands.AddServeFlags,
			handler:  commands.Serve,
			usage:    commands.ServeUsage,
		},
		"game": {
			subcommands: map[string]commandSpec{
				"serialize": {
					addFlags:     commands.AddSerializeFlags,
					handler:      commands.Serialize,
					needsFactory: true,
					argsSpec:     "<game> <seed>",
					usage:        commands.SerializeUsage,
				},
			},
			subUsage: gameSubUsage,
		},
		"replay": {
			addFlags:     commands.AddReplayFlags,
			handler:      commands.Replay,
			needsFactory: true,
			argsSpec:     "<game> <username> [<id|url>[,<id|url>...]]",
			usage:        commands.ReplayUsage,
		},
	}
}

func normalizeCommand(args []string) (string, []string) {
	command, rest := args[0], args[1:]
	if command == "--help" || command == "-h" {
		return "help", nil
	}
	return command, rest
}

func selectCommand(command string, args []string) (commandSpec, string, []string, error) {
	spec, ok := commandSet()[command]
	if !ok {
		return commandSpec{}, "", nil, fmt.Errorf("unknown command %q; run `arena help` for usage", command)
	}
	if len(spec.subcommands) == 0 {
		return spec, command, args, nil
	}
	if len(args) == 0 {
		return spec, command, args, nil
	}

	subcommand := args[0]
	sub, ok := spec.subcommands[subcommand]
	if !ok {
		return commandSpec{}, "", nil, fmt.Errorf("unknown %s subcommand %q; run `arena help %s` for usage", command, subcommand, command)
	}
	return sub, command + " " + subcommand, args[1:], nil
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

// gameSubUsage prints the help text shown for `arena game` (with no
// subcommand) and `arena help game`. Lists every nested game-related
// subcommand registered under `commandSet["game"]`.
func gameSubUsage() string {
	return strings.TrimSpace(`arena game - Game-related helpers (engine introspection, fixtures, etc.)

Subcommands:
  serialize   <game> <seed>   Print the bot-stdin bytes (globals + turn-0 frame)
                              for a given engine seed. Use to inspect maps,
                              build deterministic test fixtures, or pipe a
                              fixed input straight into a bot binary.

Use "arena help game <subcommand>" for the full per-subcommand help.`)
}

// popGame consumes the leading game positional and returns the matching
// factory along with the remaining args. Required by every subcommand whose
// spec sets needsFactory. argsSpec is the positional signature shown in the
// usage-error message; defaults to "<game>" when empty.
func popGame(path, argsSpec string, args, games []string) (arena.GameFactory, []string, error) {
	if argsSpec == "" {
		argsSpec = "<game>"
	}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return nil, nil, fmt.Errorf("usage: arena %s %s [OPTIONS]; available games: %s", path, argsSpec, strings.Join(games, ", "))
	}
	name := args[0]
	f := arena.GetFactory(name)
	if f == nil {
		return nil, nil, fmt.Errorf("unknown game %q; available: %s", name, strings.Join(games, ", "))
	}
	return f, args[1:], nil
}
