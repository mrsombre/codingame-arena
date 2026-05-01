# Command run

Run one or more match simulations between two player binaries.

`run` is the implicit default — `arena --blue=... --red=...` is equivalent to `arena run --blue=... --red=...`.

## Quick start

```shell
bin/arena --game=winter2026 \
  --blue=bin/bot-winter2026-cpp \
  --red=bin/bot-winter2026-py

# Summary: 100 matches played (3.15s)
# Stats: wins=29% losses=32% draws=39% avg_score=16.4x17.0 avg_turns=155
# Timing: avg_first_response=29msx198ms avg_turn_response=0msx0ms
```

## Options

| Flag                  | Default          | Description                                                  |
|-----------------------|------------------|--------------------------------------------------------------|
| `-b, --blue` *(required)* | —            | Our bot binary                                               |
| `-r, --red`           | `./bin/opponent` | Opponent bot binary                                          |
| `-n, --simulations`   | `100`            | Number of matches to run                                     |
| `-p, --parallel`      | `NumCPU`         | Worker threads                                               |
| `-s, --seed`          | current time     | Base RNG seed (deterministic when set)                       |
| `--seedx`             | `1`              | Seed increment per match (`seed_i = seed + i*seedx`)         |
| `--max-turns`         | `200`            | Maximum turns per match                                      |
| `-l, --league`        | game-specific    | League level                                                 |
| `--no-swap`           | `false`          | Disable automatic side swapping (see below)                  |
| `--trace`             | `false`          | Write per-match JSON trace files                             |
| `--trace-dir`         | `./traces`       | Directory for trace files                                    |
| `--output-matches`    | `false`          | Include per-match results in JSON output                     |
| `--verbose`           | `false`          | Print full JSON instead of short summary                     |
| `--debug`             | `false`          | Single match, fixed sides, bot stderr passthrough            |

## Output

**Short summary** (default):

```
Summary: <n> matches played (<elapsed>)
Stats: wins=<%> losses=<%> draws=<%> avg_score=<blue>x<red> avg_turns=<n>
Timing: avg_first_response=<blue>x<red> avg_turn_response=<blue>x<red>
```

Win/loss/draw counts are from blue's (our) perspective.

**Verbose JSON** (`--verbose`): full summary with per-metric averages, runner metadata, bad-command list, and the five worst losses for blue.

**Debug** (`--debug`): forces `--simulations=1` and `--parallel=1`, prints bot stderr to your terminal, and emits the match trace JSON to stdout.

## Side swapping

By default blue and red alternate left/right engine slots across matches to neutralize positional bias. The runner reports `blue_left` / `blue_right` counts in verbose mode. Use `--no-swap` to lock blue to the left slot.

## Tracing

`--trace` writes one JSON file per match to `--trace-dir` (default `./traces/`). Trace files are inputs to `arena analyze` and the web viewer (`arena serve`).
