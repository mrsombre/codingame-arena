# Command game

Per-game helpers: print bundled docs (rules, trace format) and inspect engine fixtures (initial game input). Everything under `arena game <game> <action>` operates on a single resolved game, with `<game>` as the first positional argument.

```shell
bin/arena game <game> <action> [args] [OPTIONS]
```

## Table of contents

- [`rules`](#rules) — print the bundled `rules.md` for a game
- [`trace`](#trace) — print the bundled per-game `trace.md` for a game
- [`serialize`](#serialize) — print the initial game input for a seed

## rules

Print a game's bundled `rules.md` to stdout. The markdown is embedded into the `arena` binary at build time via `//go:embed`, so the docs travel with the CLI — no separate checkout, no filesystem path, no network call. The same arena binary on a remote machine answers the same way.

### Quick start

```shell
bin/arena game winter2026 rules | head
```

```
# Winter Challenge 2026 — SnakeByte

## Goal
...
```

### Options

`arena game <game> rules` — no flags, no further positionals.

### How it works

Each game package (`games/<game>/`) contains an embed source file that pins `rules.md` into the binary:

```go
package winter2026

import _ "embed"

//go:embed rules.md
var Rules string
```

The engine factory exposes the embedded string via the `arena.RulesProvider` interface (`Rules() string`). The `rules` action reads it from the resolved factory and writes it verbatim to stdout — no rendering, no transformation.

### Use cases

- Refresh on a game's rules without leaving the terminal or finding the repo on disk.
- Pipe to a pager or markdown renderer:
  ```shell
  bin/arena game winter2026 rules | less
  bin/arena game winter2026 rules | glow -
  ```
- Feed to an LLM agent so it can read the rules straight from the binary that runs the engine — agents using arena as a tool no longer need the source tree mounted alongside the binary.

### Adding a new game

Two steps:

1. Drop a `rules.md` next to your game's `engine/` directory (`games/<game>/rules.md`).
2. Make sure the package at `games/<game>/` (the directory, not `engine/`) carries an embed declaration with the `//go:embed` directive shown above, and that the factory implements `Rules() string`.

`arena game <game> rules` then works for the new game with no further wiring.

## trace

Print a game's bundled `trace.md` to stdout — the per-game trace-format reference (`setup` lines, `gameInput` lines, `state` shape, `traces[].type` event labels). Same embedding mechanism as `rules`: the markdown ships inside the arena binary.

For the cross-game trace envelope (file naming, top-level fields, per-turn shape), see [`docs/trace.md`](trace.md).

### Quick start

```shell
bin/arena game winter2026 trace | head
```

```
# Winter Challenge 2026 — Trace format

This document describes the winter2026-specific parts of the arena trace
...
```

### Options

`arena game <game> trace` — no flags, no further positionals.

### How it works

Mirrors `rules`. The game package embeds `trace.md` and the factory implements `arena.TraceProvider` (`Trace() string`). The `trace` action reads it from the resolved factory and writes it verbatim to stdout.

### Use cases

- Look up a game's per-turn `state` shape or trace event labels (`GATHER`, `EAT`, `HIT_ENEMY`, …) without leaving the terminal.
- Pipe to a pager or markdown renderer:
  ```shell
  bin/arena game winter2026 trace | less
  bin/arena game winter2026 trace | glow -
  ```
- Feed to an LLM agent so it can read the trace format straight from the binary that produces the traces, no sidecar files needed.

### Adding a new game

Two steps:

1. Drop a `trace.md` next to your game's `engine/` directory (`games/<game>/trace.md`).
2. Make sure the package at `games/<game>/` carries an `//go:embed trace.md` declaration and the factory implements `Trace() string`.

## serialize

Print the game input that a bot would receive on its first turn for a given seed. Useful for inspecting initial map state, capturing fixtures for unit tests, or feeding a bot a fixed input via stdin without running the engine.

### Quick start

```shell
bin/arena game winter2026 serialize 100030005000
```

Output is the raw lines a bot reads from stdin:

```
0
34
19
..........##..........##..........
...
```

### Options

`arena game <game> serialize <seed> [OPTIONS]` — `<game>` first, then the `serialize` action, then `<seed>`.

| Flag           | Default       | Description                          |
|----------------|---------------|--------------------------------------|
| `--player`     | `0`           | Player perspective (`0` or `1`)      |
| `-l, --league` | game-specific | League level                         |

### Output

Two blocks separated by newline-terminated lines:

1. **Global init info** — constants and map data sent once at game start.
2. **First-frame info** — per-turn state for turn 0.

Format matches exactly what bots receive on stdin during a real match. Pipe it into a bot binary to drive a single-turn invocation:

```shell
bin/arena game winter2026 serialize 42 | bin/bot-winter2026-cpp
```
