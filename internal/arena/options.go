package arena

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	fs.String("game", "", "Active game")
	return fs
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
  replay       Download replay JSON (sub: get <url|id>..., leaderboard <url> <nick>)
  convert      Convert replay JSON files into arena trace files
  analyze      Analyze trace outcomes and game-owned metrics
  serialize    Print initial game input for first turn for a given seed
  serve        Serve the embedded web viewer

Use "arena help <command>" for more information about a command.

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
