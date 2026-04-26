package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// ServeOptions holds the parsed configuration for the "serve" subcommand.
type ServeOptions struct {
	Port      int
	Host      string
	TraceDir  string
	ReplayDir string
	BinDir    string
	Help      bool
}

func parseServeOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ServeOptions, error) {
	knownArgs, _, err := arena.SplitArgs(args, fs)
	if err != nil {
		return ServeOptions{}, err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return ServeOptions{}, err
	}

	opts := ServeOptions{
		Help: v.GetBool("help"),
	}
	if opts.Help {
		return opts, nil
	}

	opts.Port = v.GetInt("port")
	opts.Host = v.GetString("host")
	opts.TraceDir = v.GetString("trace-dir")
	opts.ReplayDir = v.GetString("replay-dir")
	opts.BinDir = v.GetString("bin-dir")

	if opts.Port < 1 || opts.Port > 65535 {
		return ServeOptions{}, fmt.Errorf("--port must be in 1..65535")
	}

	return opts, nil
}
