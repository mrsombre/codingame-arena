# Command game rules

Print a game's bundled `rules.md` to stdout. The markdown is embedded into the `arena` binary at build time via `//go:embed`, so the docs travel with the CLI — no separate checkout, no filesystem path, no network call. The same arena binary on a remote machine answers the same way.

## Quick start

```shell
bin/arena game winter2026 rules | head
```

```
# Winter Challenge 2026 — SnakeByte

## Goal
...
```

## Options

`arena game <game> rules` — no flags, no further positionals.

## How it works

Each game package (`games/<game>/`) contains a `rules.go` source file that pins `rules.md` into the binary:

```go
package winter2026

import _ "embed"

//go:embed rules.md
var Rules string
```

The engine factory exposes the embedded string via the `arena.RulesProvider` interface (`Rules() string`). The `rules` action reads it from the resolved factory and writes it verbatim to stdout — no rendering, no transformation.

## Use cases

- Refresh on a game's rules without leaving the terminal or finding the repo on disk.
- Pipe to a pager or markdown renderer:
  ```shell
  bin/arena game winter2026 rules | less
  bin/arena game winter2026 rules | glow -
  ```
- Feed to an LLM agent so it can read the rules straight from the binary that runs the engine — agents using arena as a tool no longer need the source tree mounted alongside the binary.

## Adding a new game

Two steps:

1. Drop a `rules.md` next to your game's `engine/` directory (`games/<game>/rules.md`).
2. Make sure the package at `games/<game>/` (the directory, not `engine/`) carries a `rules.go` with the `//go:embed` directive shown above, and that the factory implements `Rules() string`.

`arena game <game> rules` then works for the new game with no further wiring.
