package commands

import (
	"path/filepath"

	"github.com/spf13/pflag"
)

// AddAnalyzeFlags registers flags used by the "analyze" subcommand on fs.
func AddAnalyzeFlags(fs *pflag.FlagSet) {
	fs.String("trace-dir", filepath.Clean("./traces"), "Directory to scan for trace JSON files (top-level *.json only; subdirectories are ignored)")
}

// AnalyzeOptions holds the parsed configuration for the "analyze" subcommand.
type AnalyzeOptions struct {
	TraceDir string
}

func parseAnalyzeOptions(args []string, fs *pflag.FlagSet) (AnalyzeOptions, error) {
	if err := fs.Parse(args); err != nil {
		return AnalyzeOptions{}, err
	}

	var opts AnalyzeOptions
	opts.TraceDir, _ = fs.GetString("trace-dir")
	if opts.TraceDir == "" {
		opts.TraceDir = "traces"
	}
	return opts, nil
}
