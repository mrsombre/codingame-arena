# Arena Docs

Per-command reference for the `arena` CLI. Each page covers flags, expected output, and command-specific behaviour.

For a project overview, see the [top-level README](../README.md).

## Commands

| Command                          | Purpose                                                  |
|----------------------------------|----------------------------------------------------------|
| [`run`](run.md)                  | Run one or more match simulations against a player       |
| [`replay`](replay.md)            | Download replay JSON and convert it into trace files     |
| [`trace`](trace.md)              | On-disk trace file format produced by `run` and `replay` |
| [`analyze`](analyze.md)          | Aggregate trace outcomes and game-owned metrics          |
| [`game rules`](rules.md)         | Print the bundled rules.md for a game                    |
| [`game serialize`](serialize.md) | Print initial game input for the first turn of a seed    |
| [`serve`](serve.md)              | Serve the embedded web viewer                            |

## Typical flow

```
arena run <game> --trace                 ─▶ traces/trace-<id>-<n>.json   (self-play)
arena replay <game> <user> [ids]         ─▶ replays/<id>.json + traces/replay-<id>.json
arena analyze <game>                     ─▶ outcome and game-metric report
arena game <game> rules                  ─▶ bundled rules.md to stdout
arena game <game> serialize <seed>       ─▶ first-turn stdin for a given seed
arena serve                              ─▶ web viewer over both dirs
```

`run`, `replay`, and `analyze` take the game slug as their first positional argument. Game-specific helpers live under `arena game <game> <action>` (currently `rules`, `serialize`). `serve` lists every registered game.

## Configuration

Flags can be supplied via CLI, environment variables (`ARENA_<FLAG>`, hyphens become underscores), or an `arena.yml` config file in the current directory.
