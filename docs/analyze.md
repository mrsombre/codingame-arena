# Command analyze

Aggregate trace files into a winner-vs-loser report — what does the winning side tend to do that the losing side doesn't?

Reads every `*.json` trace in `--trace-dir` (both self-play `trace-*.json` files written by `arena run --trace` and `replay-*.json` files written by `arena convert`), groups them by outcome, and emits a per-game report.

## Quick start

```shell
# Analyze every trace in ./traces/
bin/arena analyze --game=spring2020

# Read traces from a different directory
bin/arena analyze --game=spring2020 --trace-dir=./traces/experiment-A
```

## Options

| Flag          | Default    | Description                                                                    |
|---------------|------------|--------------------------------------------------------------------------------|
| `--trace-dir` | `./traces` | Directory to scan for trace JSON files                                         |
| `--game`      | inferred   | Active game (or set in `arena.yml`); inferred when all traces share a `gameId` |

If `--game` is omitted and the trace dir contains traces from multiple games, the command exits with an error listing the games it found — pass `--game` to disambiguate.

## How it works

1. Read every `*.json` file in `--trace-dir`; ignore non-trace JSON.
2. Filter to traces whose `gameId` matches `--game` (or the inferred game).
3. Hand the filtered list to the active game's analyzer (`TraceAnalyzer`).
4. Per game, classify each trace as winner / loser / draw based on `ranks` (with `scores` as tie-breaker), then aggregate command and event statistics.
5. Render the report to stdout.

## Game support

| Game                  | Analyzer                |
|-----------------------|-------------------------|
| Spring Challenge 2020 | implemented             |
| Winter Challenge 2026 | not yet — see `TODO.md` |

A game without an analyzer returns:

```
game "winter2026" does not implement trace analysis
```

## Output (Spring 2020)

```
Spring 2020 trace analysis: 42 trace files from ./traces
Decided matches: 40  draws: 2
Side wins: p0=18 p1=22
Winner score: avg 32.4 vs loser 18.7 (margin +13.7), avg turns 178.3

Winner command rates (% of decisions per side):
  MOVE         winner 78.10%  loser 71.20%  diff +6.90pp
  SPEED        winner 14.50%  loser 12.30%  diff +2.20pp
  SWITCH       winner  7.40%  loser 16.50%  diff -9.10pp

Winner pac events (avg per decided match):
  EAT          winner 28.40  loser 16.10  diff +12.30 (winner 1.8x loser; loser only 57% as often as winner)
  KILLED       winner  0.30  loser  1.70  diff -1.40 (winner only 18% as often as loser; loser 5.7x winner)

Command rates normalize for surviving pacs; events are absolute counts. Diff = winner - loser.
```

Sections:

- **Header** — file count, decided/draw split, per-side win counts, winner-vs-loser score margin.
- **Command rates** — frequency of each command issued by the winning side vs the losing side, expressed as a percentage of that side's total commands. Normalizes for asymmetric unit counts (e.g. a side with fewer alive pacs naturally emits fewer commands).
- **Events** — absolute counts of trace events (e.g. `EAT`, `KILLED` in Spring 2020) averaged per decided match, with a plain-language explanation of the gap.

`Diff` is always **winner − loser**: positive means winners do it more, negative means losers do it more.

## Caveats

- The current report compares **winner vs loser** anonymously. It does not yet split by "us-when-winning vs us-when-losing" — that requires propagating the `blue` field from replays into traces (tracked in `TODO.md`).
- All traces in `--trace-dir` are treated equally regardless of source (self-play vs converted replay), league, opponent, or bot version. For meaningful comparisons, keep cohorts in separate directories or trace dirs.
- Draws (equal `ranks` and equal `scores`) are counted but excluded from winner/loser aggregation.
