package commands

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LeaderboardOptions holds the parsed configuration for the "leaderboard" subcommand.
type LeaderboardOptions struct {
	Slug     string
	Nickname string
	OutDir   string
	Limit    int
	Delay    time.Duration
	Help     bool
}

func parseLeaderboardOptions(args []string, fs *pflag.FlagSet, v *viper.Viper) (LeaderboardOptions, error) {
	if err := fs.Parse(args); err != nil {
		return LeaderboardOptions{}, err
	}

	opts := LeaderboardOptions{
		Help: v.GetBool("help"),
	}
	if opts.Help {
		return opts, nil
	}

	if fs.NArg() < 2 {
		return LeaderboardOptions{}, fmt.Errorf("leaderboard URL and nickname are required")
	}

	slug, err := parseLeaderboardSlug(fs.Arg(0))
	if err != nil {
		return LeaderboardOptions{}, err
	}
	nickname := strings.TrimSpace(fs.Arg(1))
	if nickname == "" {
		return LeaderboardOptions{}, fmt.Errorf("nickname is required")
	}

	opts.Slug = slug
	opts.Nickname = nickname
	opts.OutDir = v.GetString("out")
	if opts.OutDir == "" {
		opts.OutDir = "replays"
	}
	opts.Limit = v.GetInt("limit")
	opts.Delay = v.GetDuration("delay")

	return opts, nil
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
