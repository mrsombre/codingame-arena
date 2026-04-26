package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/codingame"
)

// Replay is the entry point for the "replay" subcommand. It downloads the
// raw replay JSON for a CodinGame game and writes it to disk untouched.
func Replay(args []string, stdout io.Writer, _ arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	knownArgs, _, err := arena.SplitArgs(args, fs)
	if err != nil {
		return err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return err
	}

	if v.GetBool("help") {
		_, err := fmt.Fprintln(stdout, arena.CommandUsage(
			"replay <url|id>",
			"Download raw replay JSON from codingame.com.",
			fs,
			"",
		))
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("replay URL or ID is required")
	}

	id, err := parseReplayID(fs.Arg(0))
	if err != nil {
		return err
	}

	body, err := codingame.New().FetchReplay(id)
	if err != nil {
		return err
	}

	outPath, err := resolveOutPath(v.GetString("out"), id)
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
