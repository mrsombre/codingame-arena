package commands

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddReplayFlags registers flags used by the "replay" command on fs.
func AddReplayFlags(fs *pflag.FlagSet) {
	fs.StringP("out", "o", filepath.Clean("./replays"), "Output directory for raw replay JSON, saved as <gameId>.json (created if missing)")
	fs.String("trace-dir", filepath.Clean("./traces"), "Output directory for auto-converted traces, saved as replay-<gameId>.json (created if missing)")
	fs.IntP("limit", "n", 0, "Cap on the number of replays to download in one invocation (0 = no cap, fetch every id)")
	fs.IntP("league", "l", 4, "League level stamped into each saved replay's 'league' field (does not filter what gets downloaded)")
	fs.Duration("delay", 500*time.Millisecond, "Sleep between successive HTTP fetches (rate-limit politeness; skipped-existing replays do not reset it)")
	fs.BoolP("force", "f", false, "Re-fetch replays already on disk and re-run the auto-convert step, overwriting both files in place")
}

// ReplayOptions holds the parsed configuration for the "replay" command.
//
// IDs is empty when the user invoked `arena replay <username>` with no
// trailing arguments — that selects leaderboard mode (download every
// replay from the player's last battles list). When non-empty, only the
// listed games are fetched.
type ReplayOptions struct {
	Username string
	IDs      []int64
	OutDir   string
	TraceDir string
	League   int
	Limit    int
	Delay    time.Duration
	Force    bool
}

func parseReplayOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ReplayOptions, error) {
	if err := fs.Parse(args); err != nil {
		return ReplayOptions{}, err
	}

	var opts ReplayOptions

	if fs.NArg() < 1 {
		return ReplayOptions{}, fmt.Errorf("<username> is required; usage: arena replay <game> <username> [<id|url>[,<id|url>...]]")
	}

	username := strings.TrimSpace(fs.Arg(0))
	if username == "" {
		return ReplayOptions{}, fmt.Errorf("<username> must not be blank")
	}
	opts.Username = username

	ids := make([]int64, 0, fs.NArg()-1)
	for i := 1; i < fs.NArg(); i++ {
		for _, raw := range strings.Split(fs.Arg(i), ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := parseReplayID(raw)
			if err != nil {
				return ReplayOptions{}, err
			}
			ids = append(ids, id)
		}
	}
	opts.IDs = ids

	opts.OutDir = v.GetString("out")
	if opts.OutDir == "" {
		opts.OutDir = "replays"
	}
	opts.TraceDir = v.GetString("trace-dir")
	if opts.TraceDir == "" {
		opts.TraceDir = "traces"
	}
	opts.League = v.GetInt("league")
	if opts.League < 0 {
		return ReplayOptions{}, fmt.Errorf("--league must be >= 0")
	}
	opts.Limit = v.GetInt("limit")
	opts.Delay = v.GetDuration("delay")
	opts.Force = v.GetBool("force")

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
