package commands

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddServeFlags registers flags used by the "serve" subcommand on fs.
func AddServeFlags(fs *pflag.FlagSet) {
	fs.String("host", "localhost", "Bind host (use 0.0.0.0 to expose on the LAN)")
	fs.Int("port", 5757, "HTTP port (1..65535)")
	fs.String("trace-dir", "./traces", "Directory of arena trace JSON files served via /api/matches and /api/matches/{id}")
	fs.String("replay-dir", "./replays", "Directory of CodinGame replay JSON files served via /api/replays and /api/replays/{id}")
	fs.String("bin-dir", "./bin", "Directory scanned for bot executables (any executable file whose name contains 'bot') exposed via /api/bots")
}

// ServeOptions holds the parsed configuration for the "serve" subcommand.
type ServeOptions struct {
	Port      int
	Host      string
	TraceDir  string
	ReplayDir string
	BinDir    string
}

func parseServeOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ServeOptions, error) {
	if err := fs.Parse(args); err != nil {
		return ServeOptions{}, err
	}

	var opts ServeOptions
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
