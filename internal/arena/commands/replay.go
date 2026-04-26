package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/codingame"
)

// Replay is the entry point for the "replay" subcommand. It downloads the
// raw replay JSON for a CodinGame game and writes it to disk untouched.
func Replay(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	opts, err := parseReplayOptions(args, fs, v)
	if err != nil {
		return err
	}

	if opts.Help {
		_, err := fmt.Fprintln(stdout, arena.CommandUsage(
			"replay <url|id>",
			"Download raw replay JSON from codingame.com.",
			fs,
			"",
		))
		return err
	}

	body, err := codingame.New().FetchReplay(opts.ReplayID)
	if err != nil {
		return err
	}

	outPath, err := resolveOutPath(opts.Out, opts.ReplayID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, body, 0644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	_, _ = fmt.Fprintf(stdout, "saved %d bytes to %s\n", len(body), outPath)
	return nil
}

// resolveOutPath decides the final output file path based on user input.
//
//   - empty               → ./replays/<id>.json (creates ./replays if missing)
//   - trailing "/"        → <out>/<id>.json (creates <out> if missing)
//   - existing directory  → <out>/<id>.json
//   - anything else       → treated as a file path, written as-is
func resolveOutPath(out string, id int64) (string, error) {
	filename := fmt.Sprintf("%d.json", id)

	if out == "" {
		dir := "replays"
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("create %s: %w", dir, err)
		}
		return filepath.Join(dir, filename), nil
	}

	if strings.HasSuffix(out, string(os.PathSeparator)) {
		dir := strings.TrimRight(out, string(os.PathSeparator))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("create %s: %w", dir, err)
		}
		return filepath.Join(dir, filename), nil
	}

	if info, err := os.Stat(out); err == nil && info.IsDir() {
		return filepath.Join(out, filename), nil
	}

	return out, nil
}
