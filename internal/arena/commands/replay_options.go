package commands

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// addReplayBatchFlags registers the flags shared by every replay subcommand
// that downloads one or more files into a directory.
func addReplayBatchFlags(fs *pflag.FlagSet) {
	fs.StringP("out", "o", filepath.Clean("./replays"), "Directory to save replays as <gameId>.json")
	fs.IntP("limit", "n", 0, "Maximum number of replays to download (0 = all)")
	fs.IntP("league", "l", 4, "League level recorded in saved replay")
	fs.Duration("delay", 500*time.Millisecond, "Delay between replay downloads")
	fs.BoolP("force", "f", false, "Re-download replays even if they already exist on disk")
}

// AddReplayGetFlags registers flags used by the "replay get" subcommand on fs.
func AddReplayGetFlags(fs *pflag.FlagSet) {
	addReplayBatchFlags(fs)
}

// AddReplayLeaderboardFlags registers flags used by the "replay leaderboard"
// subcommand on fs.
func AddReplayLeaderboardFlags(fs *pflag.FlagSet) {
	addReplayBatchFlags(fs)
}

// ReplayGetOptions holds the parsed configuration for the "replay get"
// subcommand.
type ReplayGetOptions struct {
	Username string
	IDs      []int64
	OutDir   string
	League   int
	Limit    int
	Delay    time.Duration
	Force    bool
}

func parseReplayGetOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ReplayGetOptions, error) {
	if err := fs.Parse(args); err != nil {
		return ReplayGetOptions{}, err
	}

	var opts ReplayGetOptions

	if fs.NArg() < 2 {
		return ReplayGetOptions{}, fmt.Errorf("usage: arena replay get <username> <id|url>[,<id|url>...]")
	}

	username := strings.TrimSpace(fs.Arg(0))
	if username == "" {
		return ReplayGetOptions{}, fmt.Errorf("username is required")
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
				return ReplayGetOptions{}, err
			}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ReplayGetOptions{}, fmt.Errorf("at least one replay URL or ID is required")
	}
	opts.IDs = ids
	opts.OutDir = v.GetString("out")
	if opts.OutDir == "" {
		opts.OutDir = "replays"
	}
	opts.League = v.GetInt("league")
	if opts.League < 0 {
		return ReplayGetOptions{}, fmt.Errorf("--league must be >= 0")
	}
	opts.Limit = v.GetInt("limit")
	opts.Delay = v.GetDuration("delay")
	opts.Force = v.GetBool("force")

	return opts, nil
}

// ReplayLeaderboardOptions holds the parsed configuration for the
// "replay leaderboard" subcommand.
type ReplayLeaderboardOptions struct {
	Username string
	Slug     string
	OutDir   string
	League   int
	Limit    int
	Delay    time.Duration
	Force    bool
}

func parseReplayLeaderboardOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (ReplayLeaderboardOptions, error) {
	if err := fs.Parse(args); err != nil {
		return ReplayLeaderboardOptions{}, err
	}

	var opts ReplayLeaderboardOptions

	if fs.NArg() < 2 {
		return ReplayLeaderboardOptions{}, fmt.Errorf("usage: arena replay leaderboard <username> <puzzle-url|slug>")
	}
	username := strings.TrimSpace(fs.Arg(0))
	if username == "" {
		return ReplayLeaderboardOptions{}, fmt.Errorf("username is required")
	}
	slug, err := parseLeaderboardSlug(fs.Arg(1))
	if err != nil {
		return ReplayLeaderboardOptions{}, err
	}

	opts.Username = username
	opts.Slug = slug
	opts.OutDir = v.GetString("out")
	if opts.OutDir == "" {
		opts.OutDir = "replays"
	}
	opts.League = v.GetInt("league")
	if opts.League < 0 {
		return ReplayLeaderboardOptions{}, fmt.Errorf("--league must be >= 0")
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

// parseLeaderboardSlug pulls the puzzle pretty-id out of a CodinGame
// leaderboard URL. Accepts:
//   - https://www.codingame.com/multiplayer/<kind>/<slug>/leaderboard
//   - https://www.codingame.com/contests/<slug>/leaderboard/global
//   - bare slug "<slug>"
func parseLeaderboardSlug(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("leaderboard URL is required")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		if !validSlug(raw) {
			return "", fmt.Errorf("cannot extract puzzle slug from %q", raw)
		}
		return raw, nil
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	switch {
	case len(parts) >= 4 && parts[0] == "multiplayer":
		return parts[2], nil
	case len(parts) >= 2 && parts[0] == "contests":
		return parts[1], nil
	}
	return "", fmt.Errorf("cannot extract puzzle slug from %q", raw)
}

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func validSlug(s string) bool {
	return slugPattern.MatchString(s)
}
