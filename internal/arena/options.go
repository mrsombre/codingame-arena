package arena

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// NewBaseFlagSet returns a flag set pre-populated with flags shared by
// every subcommand. Subcommands register their own flags on top via
// AddRunFlags / AddSerializeFlags.
func NewBaseFlagSet(name string) *pflag.FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SortFlags = false
	fs.SetOutput(io.Discard)
	return fs
}

// Usage returns the top-level help text, listing available commands.
func Usage(games []string) string {
	return strings.TrimSpace(fmt.Sprintf(`Available games: %s

Usage: arena <command> [<game>] [OPTIONS]

Commands:
  run         <game>            Run one or more match simulations against a player binary
  replay      <game>            Download replay JSON for a player and convert it to traces
  analyze     <game>            Analyze trace outcomes and game-owned metrics
  game        <action> [<game>] Per-game helpers (rules, trace, serialize, list)

Use "arena help <command>" for more information about a command.`, strings.Join(games, ", ")))
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
	return strings.TrimRight(sb.String(), "\n")
}

func ParseSeed(value string) (int64, error) {
	raw := strings.TrimPrefix(value, "seed=")
	return strconv.ParseInt(raw, 10, 64)
}
