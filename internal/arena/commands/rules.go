package commands

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddRulesFlags registers flags for `arena game <game> rules`. The action
// has no flags today; the function exists so the dispatch table can call
// it uniformly.
func AddRulesFlags(_ *pflag.FlagSet) {}

// RulesUsage returns the help text shown for `arena help game rules`.
func RulesUsage(fs *pflag.FlagSet) string {
	extra := `Positional args:
  arena game <game> rules    No further positionals or flags.

Output:
  Writes the bundled rules.md for <game> to stdout, verbatim. The markdown
  is embedded in the arena binary at build time, so the docs travel with
  the CLI — no separate checkout, no filesystem path, and the same arena
  binary on a remote machine answers the same way.

Use cases:
  - Refresh on a game's rules without leaving the terminal or finding the
    repo on disk.
  - Pipe to a pager (` + "`arena game winter2026 rules | less`" + `) or to a markdown
    renderer (` + "`arena game winter2026 rules | glow -`" + `).
  - Feed to an LLM agent so it can read the rules straight from the binary
    that runs the engine, no sidecar files needed.`
	return arena.CommandUsage("game <game> rules", "Print the bundled rules.md for a game.", fs, extra)
}

// Rules is the entry point for `arena game <game> rules`.
func Rules(_ []string, stdout io.Writer, factory arena.GameFactory, _ *pflag.FlagSet, _ *viper.Viper) error {
	provider, ok := factory.(arena.RulesProvider)
	if !ok {
		return fmt.Errorf("game %q does not bundle rules", factory.Name())
	}
	_, err := io.WriteString(stdout, provider.Rules())
	return err
}
