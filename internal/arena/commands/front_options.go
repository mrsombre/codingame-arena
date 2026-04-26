package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// FrontOptions holds the parsed configuration for the "serve" subcommand.
type FrontOptions struct {
	Port      int
	Host      string
	TraceDir  string
	ReplayDir string
	BinDir    string
	Help      bool
}

func parseFrontOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (FrontOptions, error) {
	knownArgs, _, err := arena.SplitArgs(args, fs)
	if err != nil {
		return FrontOptions{}, err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return FrontOptions{}, err
	}

	opts := FrontOptions{
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
		return FrontOptions{}, fmt.Errorf("--port must be in 1..65535")
	}

	return opts, nil
}
