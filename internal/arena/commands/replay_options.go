package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// ReplayOptions holds the parsed configuration for the "replay" subcommand.
type ReplayOptions struct {
	ReplayID int64
	Out      string
	Help     bool
}

func parseReplayOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ReplayOptions, error) {
	knownArgs, _, err := arena.SplitArgs(args, fs)
	if err != nil {
		return ReplayOptions{}, err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return ReplayOptions{}, err
	}

	opts := ReplayOptions{
		Help: v.GetBool("help"),
	}
	if opts.Help {
		return opts, nil
	}

	if fs.NArg() < 1 {
		return ReplayOptions{}, fmt.Errorf("replay URL or ID is required")
	}
	id, err := parseReplayID(fs.Arg(0))
	if err != nil {
		return ReplayOptions{}, err
	}
	opts.ReplayID = id
	opts.Out = v.GetString("out")

	return opts, nil
}

var replayURLPattern = regexp.MustCompile(`(\d+)/?$`)

func parseReplayID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return id, nil
	}
	m := replayURLPattern.FindStringSubmatch(raw)
	if m == nil {
		return 0, fmt.Errorf("cannot extract replay ID from %q", raw)
	}
	return strconv.ParseInt(m[1], 10, 64)
}
