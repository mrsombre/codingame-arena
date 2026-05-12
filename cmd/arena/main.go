package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	gamespkg "github.com/mrsombre/codingame-arena/games"
	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/commands"
)

type handlerFunc func(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error

type commandSpec struct {
	addFlags     func(*pflag.FlagSet)
	handler      handlerFunc
	needsFactory bool
	// argsSpec is the positional signature echoed in usage-error messages.
	// Top-level commands include <game> (e.g. "<game> <username>"); under
	// `arena game <action> <game>` the action's argsSpec is post-game
	// (e.g. "" for actions with no further positionals after <game>).
	argsSpec    string
	usage       func(*pflag.FlagSet) string
	subcommands map[string]commandSpec
	subUsage    func() string
}

func main() {
	// Cap the Go scheduler so arena's own work yields the host's remaining
	// cores to the bot subprocesses spawned per match. Two threads is enough
	// for the engine + I/O fanout while staying well under NumCPU.
	runtime.GOMAXPROCS(2)
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
	if len(arena.Games()) == 0 {
		return errors.New("no engines registered")
	}
	games := gamespkg.Order

	if len(args) == 0 {
		return writeLine(stdout, arena.Usage(games))
	}

	command, rest := normalizeCommand(args)
	if command == "help" {
		return printHelp(stdout, rest, games)
	}

	// `arena game <action> <game> [args]` mirrors the top-level CLI shape
	// (verb first, game second). Dispatch is custom because the action
	// vocabulary lives under registry["game"].subcommands rather than at
	// the top level, and `arena game list` deliberately takes no <game>.
	if command == "game" {
		return executeGame(rest, stdout, games)
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

	return runHandler(spec, path, rest, stdout, factory)
}

// executeGame handles `arena game <action> <game> [args]`. The action is
// popped first (looked up in registry["game"].subcommands), then <game> is
// resolved via popGame, and the factory is threaded into the action's
// handler. `arena game list` is the one exception — it carries no <game>
// because it introspects the live engine registry.
func executeGame(args []string, stdout io.Writer, games []string) error {
	if len(args) == 0 {
		return writeLine(stdout, gameSubUsage())
	}
	action := args[0]
	if action == "list" {
		return writeLine(stdout, strings.Join(arena.Games(), "\n"))
	}

	spec, ok := registry["game"].subcommands[action]
	if !ok {
		return fmt.Errorf("unknown game action %q; run `arena help game` for usage", action)
	}

	factory, rest, err := popGame("game "+action, "<game>", args[1:], games)
	if err != nil {
		return err
	}
	return runHandler(spec, "game "+action, rest, stdout, factory)
}

func runHandler(spec commandSpec, path string, args []string, stdout io.Writer, factory arena.GameFactory) error {
	fs := arena.NewBaseFlagSet("arena")
	spec.addFlags(fs)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return fmt.Errorf(`run "arena help %s" for usage`, path)
		}
		return err
	}

	v, err := arena.NewViper(fs)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	return spec.handler(args, stdout, factory, fs, v)
}

// registry is the dispatch table built once at package init. Lookups are
// read-only; the table is shared by execute, selectCommand, and printHelp.
var registry = map[string]commandSpec{
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
				addFlags: commands.AddSerializeFlags,
				handler:  commands.Serialize,
				argsSpec: "",
				usage:    commands.SerializeUsage,
			},
			"rules": {
				addFlags: commands.AddRulesFlags,
				handler:  commands.Rules,
				argsSpec: "",
				usage:    commands.RulesUsage,
			},
			"trace": {
				addFlags: commands.AddTraceFlags,
				handler:  commands.Trace,
				argsSpec: "",
				usage:    commands.TraceUsage,
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

func normalizeCommand(args []string) (string, []string) {
	command, rest := args[0], args[1:]
	if command == "--help" || command == "-h" {
		return "help", nil
	}
	return command, rest
}

func selectCommand(command string, args []string) (commandSpec, string, []string, error) {
	spec, ok := registry[command]
	if !ok {
		return commandSpec{}, "", nil, fmt.Errorf("unknown command %q; run `arena help` for usage", command)
	}
	// No subcommands or no candidate token → return the parent spec; the
	// caller routes the empty-handler case to subUsage.
	if len(spec.subcommands) == 0 || len(args) == 0 {
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

	spec, ok := registry[args[0]]
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

// gameSubUsage prints the help text shown for `arena game` (with no args)
// and `arena help game`. Lists every action registered under
// registry["game"].subcommands.
func gameSubUsage() string {
	return strings.TrimSpace(`arena game <action> <game> - Per-game helpers (rules, engine introspection, fixtures).

Actions:
  rules                  Print the bundled rules.md for <game> to stdout. The
                         file is embedded in the arena binary at build time.
  trace                  Print the bundled trace.md for <game> to stdout —
                         per-game trace payloads (setup, gameInput, state,
                         event labels). Embedded at build time.
  serialize              Print the bot-stdin bytes (globals + turn-0 frame)
                         for the active game. --seed/-s reproduces a specific
                         map; omitted, it picks a fresh seed each call.

Standalone:
  arena game list        Print every engine currently registered in the
                         arena binary (one per line, sorted). Diverges from
                         the chronological banner whenever an engine was
                         registered but left out of games.Order.

Use "arena help game <action>" for the full per-action help.`)
}

// popGame consumes the leading game positional and returns the matching
// factory along with the remaining args. Required by every subcommand whose
// spec sets needsFactory. argsSpec is the full positional signature shown
// in usage-error messages (e.g. "<game> <seed>") and is required.
func popGame(path, argsSpec string, args, games []string) (arena.GameFactory, []string, error) {
	available := strings.Join(games, ", ")
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return nil, nil, fmt.Errorf("usage: arena %s %s [OPTIONS]; available games: %s", path, argsSpec, available)
	}
	name := args[0]
	f := arena.GetFactory(name)
	if f == nil {
		return nil, nil, fmt.Errorf("unknown game %q; available: %s", name, available)
	}
	return f, args[1:], nil
}
