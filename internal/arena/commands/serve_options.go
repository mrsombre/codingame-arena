package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddServeFlags registers flags used by the "serve" subcommand on fs.
func AddServeFlags(fs *pflag.FlagSet) {
	fs.Int("port", 5757, "HTTP port")
	fs.String("host", "localhost", "Bind host")
	fs.String("trace-dir", "./matches", "Directory with match trace JSON files (powers /api/matches)")
	fs.String("replay-dir", "./replays", "Directory with CodinGame replay JSON files (powers /api/replays)")
	fs.String("bin-dir", "./bin", "Directory to scan for bot binaries (powers /api/bots)")
}

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
	if err := fs.Parse(args); err != nil {
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
