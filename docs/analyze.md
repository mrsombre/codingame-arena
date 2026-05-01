# Command analyze

Aggregate trace files into a standard outcome report plus game-owned metrics.

The command reads every `*.json` trace in `--trace-dir` (self-play `trace-*.json` files from `arena run --trace` and converted `replay-*.json` files from `arena convert`), filters them to one selected game, and emits an arena-owned report.

## Quick start

```shell
# Analyze every Spring 2020 trace in ./traces/
bin/arena analyze --game=spring2020

# Read traces from a different directory
bin/arena analyze --game=winter2026 --trace-dir=./traces/experiment-A
```

## Options

| Flag          | Default    | Description                                                                    |
|---------------|------------|--------------------------------------------------------------------------------|
| `--trace-dir` | `./traces` | Directory to scan for trace JSON files                                         |
| `--game`      | inferred   | Active game (or set in `arena.yml`); inferred when all traces share a `gameId` |

If `--game` is omitted and the trace dir contains traces from multiple games, the command exits with an error listing the games it found. Pass `--game` to disambiguate.

## How it works

1. Read every `*.json` file in `--trace-dir`; ignore non-trace JSON.
2. Filter to traces whose `gameId` matches `--game` (or the inferred game).
3. Render generic multiplayer facts: win/draw split, side wins, blue-side results, turns, scores, timing, and end reasons.
4. If the game implements trace metrics, ask it to interpret opaque `turns[].traces` and return per-side metric counts.
5. Arena aggregates those metrics as either average counts per match or average per-match turn rates.

Games decide metric meaning and side attribution. Arena never interprets labels such as `EAT`, `DEAD`, or `COLLIDE_SELF`.

## Output

```text
winter2026 — 100 traces — ./traces

OUTCOME
  Decided   66.0%   Draws   34.0%
  Blue     W  32.0%   L  34.0%   D  34.0%

MATCH
  Turns    avg 155.4   min 71   max 200
  Scores   blue 16.4   red 17.0   margin 4.8
  Timing   first  blue 29ms / red 198ms
           turn   blue 0ms / red 0ms

END REASONS
  TURNS_OUT       52.0%
  SCORE           44.0%
  ELIMINATED       4.0%  (blue 0.0%)
  TIMEOUT_START    0.0%
  TIMEOUT          0.0%
  INVALID          0.0%

METRICS — winner vs loser
  DEAD       winner  1.20/match   loser  2.10/match   (loser 1.75x winner)
  NO_EAT     winner  22.0%   loser  31.0%   (loser 1.41x winner)

METRICS — blue vs red
  DEAD       blue  1.80/match   red  1.60/match   (blue 1.13x red)
  NO_EAT     blue  25.0%   red  20.0%   (blue 1.25x red)
```

Sections:

- **Header** — game id, trace count, and trace directory.
- **OUTCOME** — decided/draw split across all matches, plus blue-side W/L/D when `blue` can be resolved from trace players.
- **MATCH** — turn count plus blue-vs-red score and timing summaries (only when `blue` is resolved; otherwise omitted, since per-side averages are noise under random side swap).
- **END REASONS** — match termination reasons as a percentage of trace files. Side-specific rows include the share attributable to blue.
- **METRICS** — game-selected metrics rendered as counts per match or per-turn percentages, with a single ratio summarizing the gap (`bigger Nx smaller`). When both sides of a per-turn-rate row average under 1%, the row switches to raw cumulative event counts so noisy 0.0%-vs-0.1% comparisons don't read as a 5x gap.

Draws are counted in the header but excluded from winner-vs-loser metric aggregation.
