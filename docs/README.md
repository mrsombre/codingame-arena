# Arena Docs

Per-command reference for the `arena` CLI. Each page covers flags, expected output, and command-specific behaviour.

For a project overview, see the [top-level README](../README.md).

## Commands

| Command                     | Purpose                                                  |
|-----------------------------|----------------------------------------------------------|
| [`run`](run.md)             | Run one or more match simulations against a player       |
| [`analyze`](analyze.md)     | Aggregate trace files into a winner-vs-loser report      |
| [`serialize`](serialize.md) | Print initial game input for the first turn of a seed    |
| [`convert`](convert.md)     | Re-simulate downloaded replays into arena trace files    |
| [`replay`](replay.md)       | Download replay JSON (`get`, `leaderboard` subcommands)  |
| [`serve`](serve.md)         | Serve the embedded web viewer                            |

## Typical flow

```
arena run --trace             ─▶ traces/trace-<id>-<n>.json    (self-play)
arena replay get | leaderboard ─▶ replays/<id>.json            (real CG matches)
arena convert                 ─▶ traces/replay-<id>.json       (verified)
arena analyze                 ─▶ winner-vs-loser report
arena serve                   ─▶ web viewer over both dirs
```

## Configuration

Flags can be supplied via CLI, environment variables (`ARENA_<FLAG>`, hyphens become underscores), or an `arena.yml` config file in the current directory. See [`arena.example.yml`](../arena.example.yml).
