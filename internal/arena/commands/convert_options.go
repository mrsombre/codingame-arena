package commands

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddConvertFlags registers flags used by the "convert" subcommand on fs.
func AddConvertFlags(fs *pflag.FlagSet) {
	fs.StringP("league", "l", "", "League level override (default: parse from replay title, else game-specific)")
	fs.String("trace-dir", filepath.Clean("./traces"), "Directory to save converted trace files")
	fs.String("replay-dir", filepath.Clean("./replays"), "Directory to scan for replay JSON files")
	fs.BoolP("force", "f", false, "Overwrite trace files even if they already exist")
}

// ConvertOptions holds the parsed configuration for the "convert" subcommand.
type ConvertOptions struct {
	TraceDir  string
	ReplayDir string
	Force     bool
	League    string
	IDs       []int64
}

func parseConvertOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ConvertOptions, error) {
	if err := fs.Parse(args); err != nil {
		return ConvertOptions{}, err
	}

	opts := ConvertOptions{
		TraceDir:  v.GetString("trace-dir"),
		ReplayDir: v.GetString("replay-dir"),
		Force:     v.GetBool("force"),
		League:    v.GetString("league"),
	}
	if opts.TraceDir == "" {
		opts.TraceDir = "traces"
	}
	if opts.ReplayDir == "" {
		opts.ReplayDir = "replays"
	}
	if opts.League != "" {
		n, err := strconv.Atoi(opts.League)
		if err != nil || n < 1 {
			return ConvertOptions{}, fmt.Errorf("--league must be a positive integer")
		}
	}
	if fs.NArg() > 0 {
		opts.IDs = make([]int64, 0, fs.NArg())
		for i := 0; i < fs.NArg(); i++ {
			id, err := strconv.ParseInt(fs.Arg(i), 10, 64)
			if err != nil || id < 1 {
				return ConvertOptions{}, fmt.Errorf("invalid replay id %q", fs.Arg(i))
			}
			opts.IDs = append(opts.IDs, id)
		}
	}

	return opts, nil
}
