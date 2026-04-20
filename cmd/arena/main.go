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
	if arena.Factory == nil {
		fmt.Fprintln(os.Stderr, "no engine selected — build with: go build -tags winter2026 ./cmd/arena")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(arena.Usage(arena.Factory))
		return
	}

	command, rest := args[0], args[1:]
	// If the first token isn't a subcommand name, assume implicit "run"
	// and feed every arg as a flag (preserves `arena --p0-bin …` ergonomics).
	if strings.HasPrefix(command, "-") {
		command, rest = "run", args
	}

	fs := arena.NewBaseFlagSet("arena")

	var handler handlerFunc
	switch command {
	case "serialize":
		arena.AddSerializeFlags(fs)
		handler = commands.Serialize
	case "replay":
		arena.AddReplayFlags(fs)
		handler = commands.Replay
	case "run":
		arena.AddRunFlags(fs)
		handler = commands.Run
	case "serve":
		arena.AddFrontFlags(fs)
		handler = commands.Front
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q; run `arena --help` for usage\n", command)
		os.Exit(1)
	}

	v, err := arena.NewViper(fs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	if err := handler(rest, os.Stdout, arena.Factory, fs, v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
