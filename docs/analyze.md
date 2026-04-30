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
winter2026 analysis: 100 trace files analyzed [./traces]
Decided matches: 66.0% / Draws: 34.0%
Side wins: p0 31.0% / p1 35.0%
Blue: Wins: 32.0% Losses: 34.0% Draws: 34.0%
Turns: avg 155.4 min 71 max 200
Scores: avg p0 16.4 p1 17.0 margin 4.8
Timing: first_response 29msx198ms avg_turn_response 0msx0ms

End reasons:
  TURNS_OUT       52.0%
  SCORE           44.0%
  ELIMINATED       4.0% (blue: 0.0%)
  TIMEOUT_START    0.0%
  TIMEOUT          0.0%
  INVALID          0.0%

Winner vs loser metrics:
  DEAD           winner  1.20/match  loser  2.10/match  (winner only 57% as often as loser; loser 1.8x winner)
  NO_EAT         winner  22.0%  loser  31.0%  (winner only 71% as often as loser; loser 1.4x winner)

Blue vs enemy metrics:
  DEAD           blue  1.80/match  enemy  1.60/match  (blue 1.1x enemy; enemy only 89% as often as blue)
  NO_EAT         blue  25.0%  enemy  20.0%  (blue 1.2x enemy; enemy only 80% as often as blue)
```

Sections:

- **Header** — file count, trace directory, decided/draw split, side wins, and blue-side win/loss/draw rates when `blue` can be resolved from trace players.
- **Generic stats** — turn count, score, and timing summaries from arena trace fields.
- **End reasons** — match termination reasons as a percentage of trace files. Side-specific rows can include the percentage attributable to blue.
- **Metric comparisons** — game-selected metrics rendered as counts per match or per-turn percentages.

Draws are counted in the header but excluded from winner-vs-loser metric aggregation.
