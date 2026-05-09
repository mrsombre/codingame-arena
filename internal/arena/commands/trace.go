package commands

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AddTraceFlags registers flags for `arena game <game> trace`. The action
// has no flags today; the function exists so the dispatch table can call
// it uniformly.
func AddTraceFlags(_ *pflag.FlagSet) {}

// TraceUsage returns the help text shown for `arena help game trace`.
func TraceUsage(fs *pflag.FlagSet) string {
	extra := `Positional args:
  arena game <game> trace    No further positionals or flags.

Output:
  Writes the bundled trace.md for <game> to stdout, verbatim. The markdown
  describes the per-game trace payloads (setup, gameInput, state, traces[].type)
  and is embedded in the arena binary at build time, so the docs travel with
  the CLI — no separate checkout, no filesystem path, and the same arena
  binary on a remote machine answers the same way.

  For the cross-game trace envelope (top-level fields, file naming, timing),
  see ` + "`docs/trace.md`" + ` in the arena repo.

Use cases:
  - Look up a game's per-turn ` + "`state`" + ` shape or trace event labels without
    leaving the terminal.
  - Pipe to a pager (` + "`arena game winter2026 trace | less`" + `) or to a markdown
    renderer (` + "`arena game winter2026 trace | glow -`" + `).
  - Feed to an LLM agent so it can read the trace format straight from the
    binary that produces the traces, no sidecar files needed.`
	return arena.CommandUsage("game <game> trace", "Print the bundled trace.md for a game.", fs, extra)
}

// Trace is the entry point for `arena game <game> trace`.
func Trace(_ []string, stdout io.Writer, factory arena.GameFactory, _ *pflag.FlagSet, _ *viper.Viper) error {
	provider, ok := factory.(arena.TraceProvider)
	if !ok {
		return fmt.Errorf("game %q does not bundle trace docs", factory.Name())
	}
	_, err := io.WriteString(stdout, provider.Trace())
	return err
}
