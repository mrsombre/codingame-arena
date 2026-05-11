package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
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
	// argsSpec is the positional signature echoed in usage-error messages.
	// Top-level commands include <game> (e.g. "<game> <username>"); under
	// `arena game <game> <action>` the action's argsSpec is action-only
	// (e.g. "<seed>", or "" for actions with no positionals).
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

	// `arena game <game> <action> [args]` puts <game> before the action,
	// matching the rest of the CLI. Dispatch is custom because the generic
	// path expects subcommand-then-game.
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

// executeGame handles `arena game <game> <action> [args]`. <game> is popped
// first, the next token names an action under registry["game"].subcommands,
// and the resolved factory is threaded into the action's handler.
func executeGame(args []string, stdout io.Writer, games []string) error {
	if len(args) == 0 {
		return writeLine(stdout, gameSubUsage())
	}
	factory, rest, err := popGame("game", "<game> <action>", args, games)
	if err != nil {
		return err
	}
	if len(rest) == 0 || strings.HasPrefix(rest[0], "-") {
		return fmt.Errorf("usage: arena game <game> <action> [OPTIONS]; actions: %s", gameActionsList())
	}
	action := rest[0]
	rest = rest[1:]

	spec, ok := registry["game"].subcommands[action]
	if !ok {
		return fmt.Errorf("unknown game action %q; run `arena help game` for usage", action)
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
				argsSpec: "<seed>",
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
	return strings.TrimSpace(`arena game <game> <action> - Per-game helpers (rules, engine introspection, fixtures).

Actions:
  rules                  Print the bundled rules.md for <game> to stdout. The
                         file is embedded in the arena binary at build time.
  trace                  Print the bundled trace.md for <game> to stdout —
                         per-game trace payloads (setup, gameInput, state,
                         event labels). Embedded at build time.
  serialize <seed>       Print the bot-stdin bytes (globals + turn-0 frame)
                         for a given seed. Use to inspect maps, build
                         deterministic test fixtures, or pipe a fixed input
                         straight into a bot binary.

Use "arena help game <action>" for the full per-action help.`)
}

// gameActionsList returns a sorted, comma-separated list of registered game
// actions for usage-error messages.
func gameActionsList() string {
	names := make([]string, 0, len(registry["game"].subcommands))
	for name := range registry["game"].subcommands {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
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
