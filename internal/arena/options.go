package arena

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewBaseFlagSet returns a flag set pre-populated with flags shared by
// every subcommand. Subcommands register their own flags on top via
// AddRunFlags / AddSerializeFlags.
func NewBaseFlagSet(name string) *pflag.FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SortFlags = false
	fs.SetOutput(io.Discard)
	fs.BoolP("help", "h", false, "Show this help")
	fs.String("game", "", "Active game")
	return fs
}

// AddRunFlags registers flags used by the "run" subcommand on fs.
func AddRunFlags(fs *pflag.FlagSet) {
	fs.StringP("league", "l", "", "League level (default: game-specific)")
	fs.IntP("simulations", "n", 100, "Number of matches to run")
	fs.IntP("parallel", "p", runtime.NumCPU(), "Worker threads")
	fs.StringP("seed", "s", "", "Base RNG seed (default: current time)")
	fs.Int("seedx", 1, "Seed increment per match (seed_i = seed + i*N)")
	fs.Bool("output-matches", false, "Include per-match results in JSON output")
	fs.Bool("debug", false, "Force one match, fixed sides, print debug to stderr")
	fs.Bool("no-swap", false, "Disable automatic side swapping")
	fs.String("trace-dir", "", "Write per-match JSON trace files")
	fs.Int("max-turns", 200, "Maximum turns per match")
	fs.String("p0", "", "Player 0 binary (required)")
	fs.String("p1", filepath.Clean("./bin/opponent"), "Player 1 binary")
	fs.Bool("verbose", false, "Output full JSON (default: short summary line)")
}

// AddSerializeFlags registers flags used by the "serialize" subcommand on fs.
func AddSerializeFlags(fs *pflag.FlagSet) {
	fs.StringP("league", "l", "", "League level (default: game-specific)")
	fs.StringP("seed", "s", "", "RNG seed (required)")
	fs.Int("player", 0, "Player index (0 or 1)")
}

// AddReplayFlags registers flags used by the "replay" subcommand on fs.
func AddReplayFlags(fs *pflag.FlagSet) {
	fs.StringP("out", "o", "", "Output path. Empty → ./replays/<id>.json. Trailing '/' or existing directory → <id>.json inside. Otherwise → a file at that path.")
}

// AddLeaderboardFlags registers flags used by the "leaderboard" subcommand on fs.
func AddLeaderboardFlags(fs *pflag.FlagSet) {
	fs.StringP("out", "o", filepath.Clean("./replays"), "Directory to save replays as <gameId>.json")
	fs.IntP("limit", "l", 0, "Maximum number of replays to download (0 = all)")
	fs.Duration("delay", 500*time.Millisecond, "Delay between replay downloads")
}

// AddServeFlags registers flags used by the "serve" subcommand on fs.
func AddServeFlags(fs *pflag.FlagSet) {
	fs.Int("port", 5757, "HTTP port")
	fs.String("host", "localhost", "Bind host")
	fs.String("trace-dir", "./matches", "Directory with match trace JSON files (powers /api/matches)")
	fs.String("replay-dir", "./replays", "Directory with CodinGame replay JSON files (powers /api/replays)")
	fs.String("bin-dir", "./bin", "Directory to scan for bot binaries (powers /api/bots)")
}

// NewViper returns a viper instance bound to the given flag set, configured
// with the ARENA_ env prefix, and populated from an "arena" config file
// (yaml/json/toml) in the current directory if present. A missing file is
// not an error; a malformed one is.
func NewViper(fs *pflag.FlagSet) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("ARENA")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetConfigName("arena")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
	}

	if err := v.BindPFlags(fs); err != nil {
		return nil, err
	}
	return v, nil
}

// Usage returns the top-level help text, listing available commands.
func Usage(games []string) string {
	return strings.TrimSpace(fmt.Sprintf(`Available games: %s

Usage: arena <command> [OPTIONS]

Commands:
  run          Run one or more match simulations against a player binary
  serialize    Print initial game input for first turn for a given seed
  replay       Download raw replay JSON from codingame.com
  leaderboard  Download every replay from a player's last battles list
  serve        Serve the embedded web viewer

Use "arena <command> --help" for more information about a command.

Env vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).
Config: arena.yml in current directory (e.g. game: winter2026).`, strings.Join(games, ", ")))
}

// CommandUsage returns help text for a specific subcommand using fs.FlagUsages().
func CommandUsage(command, description string, fs *pflag.FlagSet, extra string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "arena %s - %s\n\nOptions:\n", command, description)
	sb.WriteString(fs.FlagUsages())
	if extra != "" {
		sb.WriteString("\n")
		sb.WriteString(extra)
		sb.WriteString("\n")
	}
	sb.WriteString("\nEnv vars: ARENA_<FLAG> (hyphens become underscores, e.g. ARENA_GAME, ARENA_SEED).\n")
	sb.WriteString("Config: arena.yml in current directory (e.g. game: winter2026).")
	return sb.String()
}

func ParseSeed(value string) (int64, error) {
	raw := strings.TrimPrefix(value, "seed=")
	return strconv.ParseInt(raw, 10, 64)
}
