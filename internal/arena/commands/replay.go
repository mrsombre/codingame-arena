package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

const codingameReplayAPI = "https://www.codingame.com/services/gameResult/findInformationById"

// Replay is the entry point for the "replay" subcommand. It downloads the
// raw replay JSON for a Codingame game and writes it to disk.
func Replay(args []string, stdout io.Writer, factory arena.GameFactory, fs *pflag.FlagSet, v *viper.Viper) error {
	knownArgs, _, err := arena.SplitArgs(args, fs)
	if err != nil {
		return err
	}
	if err := fs.Parse(knownArgs); err != nil {
		return err
	}

	if v.GetBool("help") {
		_, err := fmt.Fprintln(stdout, arena.Usage(factory))
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("replay URL or ID is required")
	}

	id, err := parseReplayID(fs.Arg(0))
	if err != nil {
		return err
	}

	body, err := fetchReplay(id)
	if err != nil {
		return err
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("parse replay json: %w", err)
	}
	delete(decoded, "viewer")
	stripFrameViews(decoded)

	pretty, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		return fmt.Errorf("format replay json: %w", err)
	}

	outPath, err := resolveOutPath(v.GetString("out"), id)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, pretty, 0644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	fmt.Fprintf(stdout, "saved %d bytes to %s\n", len(pretty), outPath)
	return nil
}

// stripFrameViews removes the verbose "view" field from every frame in
// gameResult.frames. The view holds per-turn graphics payloads we do not use.
func stripFrameViews(root map[string]any) {
	gr, ok := root["gameResult"].(map[string]any)
	if !ok {
		return
	}
	frames, ok := gr["frames"].([]any)
	if !ok {
		return
	}
	for _, f := range frames {
		if frame, ok := f.(map[string]any); ok {
			delete(frame, "view")
		}
	}
}

// resolveOutPath decides the final output file path based on user input.
//
//   - empty               → ./replays/replay-<id>.json (creates ./replays if missing)
//   - trailing "/"        → <out>/replay-<id>.json (creates <out> if missing)
//   - existing directory  → <out>/replay-<id>.json
//   - anything else       → treated as a file path, written as-is
func resolveOutPath(out string, id int64) (string, error) {
	filename := fmt.Sprintf("replay-%d.json", id)

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

func fetchReplay(id int64) ([]byte, error) {
	payload := fmt.Sprintf("[%d,null]", id)
	req, err := http.NewRequest(http.MethodPost, codingameReplayAPI, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("codingame API %d: %s", resp.StatusCode, body)
	}
	return body, nil
}
