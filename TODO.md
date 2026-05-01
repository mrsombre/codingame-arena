# TODO — Make the Bot Stronger

Goal: turn the **simulator + real replay → trace → analyze** pipeline into a feedback loop that surfaces our bot's actual weaknesses, ranked by frequency and impact, so every iteration of the bot is informed by data.

The current flow:

```
arena run --trace             ─▶ trace-<id>-<n>.json   (self-play)
arena replay get/leaderboard  ─▶ replays/<id>.json
arena convert                 ─▶ replay-<id>.json     (real CG matches)
arena analyze                 ─▶ winner-vs-loser report
```

Below are improvements grouped by impact. P0 = blocking the goal, P1 = high lever, P2 = polish.

## Status snapshot (2026-05-01)

| #  | Item                                  | Status      |
|----|---------------------------------------|-------------|
| 1  | Identify "us" in every trace          | DONE        |
| 2  | Winter 2026 analyzer                  | DONE        |
| 3  | Switch axis to us-vs-them             | DONE        |
| 4  | Worst losses from disk                | NOT DONE    |
| 5  | Filter trace cohorts                  | NOT DONE    |
| 6  | Provenance on every trace             | PARTIAL     |
| 7  | Per-turn death diagnostic             | NOT DONE    |
| 8  | Pairwise matchup table                | NOT DONE    |
| 9  | Convert mismatch quarantine           | NOT DONE    |
| 10 | JSON output for analyze               | NOT DONE    |
| 11 | Regression seedset                    | NOT DONE    |
| 12 | Mortality heatmap                     | NOT DONE    |
| 13 | Apple-contention metric               | NOT DONE    |
| 14 | Auto-rerun replay seeds               | NOT DONE    |
| 15 | Leaderboard rank at fetch             | NOT DONE    |
| 16 | Move legality / wasted-action         | NOT DONE    |
| 17 | Replay viewer integration             | NOT DONE    |

P0 milestone (#1+#2+#3) is done. Suggested next move: #4 + #5 to make the report navigable across hundreds of accumulated traces.

---

## P0 — Blocking gaps

### 1. Identify "us" in every trace — DONE

`TraceMatch.Blue` is required on every loaded trace. Self-play sets it to `filepath.Base(--p0)` (P0 is always "us" — `--p0` is required by the CLI). Replays carry blue from `replay get`/`leaderboard` (username is required there too); `convert` errors out on a replay without blue or where blue doesn't match either player. `analyze` rejects any on-disk trace that lacks blue, so `BlueSide()` is treated as 0/1 throughout the report and the "Blue not identified" branch is gone. A `--trace-blue` flag on `arena run` was deemed unnecessary while P0 == us.

### 2. Winter 2026 has no analyzer — DONE

`games/winter2026/engine/analyzer.go` implements `TraceMetricAnalyzer` (specs at `analyzer.go:13`, aggregation at `analyzer.go:27`); `commands/analyze.go:51` already type-asserts and feeds it. Original brief preserved below — confirm the listed breakdowns (command rates, event counts, apples-ratio, first-death turn, length-over-time, losses) are all wired up; if some aren't, treat that as the next refinement on this item.

`Factory.AnalyzeTraces` is implemented for Spring 2020 only (`games/spring2020/engine/analyzer.go:14`). For winter2026, `commands.Analyze` errors with `"game does not implement trace analysis"`.

Implement `games/winter2026/engine/analyzer.go` with the same shape but Snake-relevant breakdowns:

- Command rates: `UP/DOWN/LEFT/RIGHT/MARK/WAIT` per side (rate, not absolute, per spring2020's pattern at `analyzer.go:270`).
- Event counts (per match avg, winner-vs-loser diff): `EAT`, `HIT_WALL`, `HIT_SELF`, `HIT_ENEMY`, `DEAD`, `FALL` — labels already emitted at `games/winter2026/engine/traces.go:11`.
- Apples eaten / total apples ratio per side.
- First-death-turn distribution per side.
- Average bird length over time per side (proxy for territorial dominance).
- Losses count from `Referee.Metrics()` (already exposes `losses_p0`, `losses_p1`).

### 3. Switch analyze from "winner vs loser" to "us vs them" — DONE

The report prints three metric axes side by side: `METRICS — winner vs loser` (anonymous field-wide), `METRICS — blue vs red` (us vs opponent across all matches), and `METRICS — blue wins vs blue losses` (the diagnostic axis: our metrics in matches we won vs our metrics in matches we lost). The OUTCOME block carries `Blue W/L/D` rates as the summary. The originally-proposed `--axis` flag was dropped — the report is small enough that printing all three axes at once is fine.

Once `Blue` lands, change the default reporting axis from anonymous winner/loser to **us-when-winning** vs **us-when-losing**. That's the comparison that maps directly to "what should I change in the bot".

- Three buckets per analyze run: `our_wins`, `our_losses`, `draws`.
- For each event/command, show the **diff between our wins and our losses** (not winner vs loser of the field).
- Keep the old winner/loser mode reachable behind `--axis winner-loser` for completeness.

### 4. Surface worst losses from disk — NOT DONE

`run.go:106` reports `worst_losses` for the current batch only — useful in CI, useless for analyzing 200 accumulated traces. `analyze` should additionally:

- List the **N largest-margin losses** with the trace filename, replay URL (if `Type == replay`), seed, opponent name, final scores, and turn the first death happened.
- List the **N closest losses** (margin ≤ small threshold) — these are the matches a slightly better bot would have won, and are the highest-leverage fixes.

Output as a section in the existing report; add `--worst N` and `--close N` flags.

### 5. Filter trace cohorts in analyze — NOT DONE

Today `analyze` reads every `*.json` in `traces/` and lumps simulator runs and real-CG replays together. They have different distributions and should be analyzed separately.

Add filters to `commands/analyze.go`:

- `--type trace|replay|all` (use `arena.TraceTypeFromFileName`, `internal/arena/trace.go:164`).
- `--player <basename>` to keep only matches where one side equals that player.
- `--league N` once league is recorded on the trace (see #6).
- `--seeds-from <file>` so we can re-run a fixed cohort of seeds.

---

## P1 — High lever

### 6. Record provenance on every trace — PARTIAL

`League int` and `CreatedAt string` are on `TraceMatch` and populated for both self-play (`match.go:285`) and convert (`convert.go:84`). Still missing: `BotVersions [2]string`, `TraceLabel string`, and `Source string` (the "self-play" / "codingame" label exists on `CodinGameReplay` but is not propagated onto the trace). `--trace-label` flag on `arena run` is not implemented. The TraceType (`trace` vs `replay`) on `TraceMatch.Type` partially fills the `Source` role today.

Without provenance, traces decay quickly: we can't tell which bot version produced trace from yesterday, which league it was played at, or whether it's still relevant.

Add to `TraceMatch`:

- `League int` — copied from converted replays' parsed league, set on self-play from the configured league.
- `BotVersions [2]string` — git short SHA at run time (computed once per `arena run` invocation), or content hash of the bot binary.
- `TraceLabel string` — value of a new `--trace-label` flag on `arena run` (cohort tagging like `experiment-mark-strategy`).
- `Source string` — `"self-play"` or `"codingame"`.

Then `analyze --label experiment-A` becomes possible, and we can keep traces from incompatible bot versions in the same dir without polluting reports.

### 7. Per-turn diagnostic for our bot — NOT DONE

The single highest-signal report for snake-style games: **when do we die, and why?** Build a per-trace summary that the analyzer aggregates:

- Histogram of `DEAD` events binned by turn (early-death = strategic blunder; late-death = endgame mistake).
- Death cause distribution: `HIT_WALL` / `HIT_SELF` / `HIT_ENEMY` / `FALL` for our side vs theirs.
- Apples eaten before vs after first death.
- For losses, the last 5 turns of game_input + our output (so we can see the actual "last decision before disaster" without opening the viewer).

Surface a `last-decisions` section in `analyze --worst N`.

### 8. Pairwise matchup table — NOT DONE

`replay leaderboard` pulls real CG matches against many opponents. Tabulate winrate vs each opponent (group by `Players[non-us]` after #1). Lets us spot a single opponent we keep losing to and study just those traces.

Output in JSON or text; add `--by opponent` flag.

### 9. Convert mismatch quarantine — NOT DONE

`convert` currently logs `skipped-mismatch` lines to stdout (`commands/convert.go:76`). Each mismatch is engineering signal — the engine doesn't yet match CG behavior, and our trace dataset is silently smaller than expected.

- Write skipped replays + reason to `<trace-dir>/_mismatch.json` with the replay id, league, expected vs actual scores, expected vs actual turns.
- `arena convert --report-mismatches` prints aggregate counts per mismatch category for a quick "engine fidelity" health check.
- A future `make engine-fidelity` target can fail if mismatch rate increases.

### 10. JSON output for analyze — NOT DONE

Today `analyze` prints text only (`spring2020/engine/analyzer.go:161`). Make the report `Write(w, format)` aware (`text` default, `json` opt-in). Enables:

- Diffing two analyze runs across bot versions to spot regressions.
- Feeding analyze results into a separate viewer page.
- CI checks (e.g. fail PR if `EAT` rate dropped > 5%).

### 11. Self-play regression seedset — NOT DONE

A frozen list of seeds we re-run with `arena run --seeds-from seedset.txt` (combined with #5). New seeds added when we discover a loss pattern. Becomes our regression suite — Make target `make regression-winter2026` runs N seeds, compares aggregate stats to a recorded baseline, fails on regression.

---

## P2 — Polish / future

### 12. Mortality heatmap per tile (winter2026) — NOT DONE

The board is small (grid is ~30 cells). For each `DEAD`/`HIT_*` event we already have `Coord` in the payload (`games/winter2026/engine/traces.go:19`). Aggregate per-cell counts in losses → emit as ASCII heatmap, or a JSON the viewer can render.

Reveals sticky death-zones (apples in dangerous corners, recurring trap cells).

### 13. Apple-contention metric — NOT DONE

For each apple eaten, record turn + side. Compute (per match) the "apple race" balance: did we get the apples within reach, or did the opponent reach them first? Diff this between our wins/losses.

### 14. Auto-rerun replay seeds in simulator — NOT DONE

After `convert` saves `replay-<id>.json`, optionally also run `arena run --seed <replay.seed> --p0 <our-current-bot> --p1 <something>` and emit a paired `trace-replay-<id>.json`. We get to see how our **current** bot would have played the same seed, side-by-side with the historical CG match.

CLI: `arena convert --rerun --p0 bin/bot-winter2026-cpp` writes both files; analyze reports the delta.

### 15. Track leaderboard rank at replay-fetch time — NOT DONE

`replay leaderboard` doesn't record the player's elo/rank when downloading. Save it alongside the replay (e.g., `replays/<id>.meta.json`) so analyze can group by elo bands ("losses against top-100" vs "losses against ladder peers").

### 16. Move legality / wasted-action analysis — NOT DONE

Some moves are no-ops (WAIT, MARK in conditions where it doesn't help) or self-destructive (MOVE into wall). Count rate of:

- WAIT turns by us — every WAIT is a possibly missed opportunity.
- MOVE → immediate `HIT_WALL`/`HIT_SELF` on the same turn — i.e., suicidal moves.
- MARK followed by no apparent strategy effect.

These are hands-on, fix-this-class-of-bug signals.

### 17. Replay viewer integration — NOT DONE

`arena serve` already renders traces. Two add-ons:

- Filter the match list by `Source` / `Blue` / outcome so the user can binge-watch their losses.
- Click a `worst-loss` entry in the analyze report and have it open in the viewer with the right turn pre-selected.

---

## Suggested implementation order

1. #1 + #2 + #3 land together — once "us" is identified and winter2026 has any analyzer, the report becomes useful immediately. (~2 days)
2. #4 + #5 — surface the matches worth watching, filter cohorts. (~1 day)
3. #7 — death analysis. Highest-leverage signal for snake games. (~1 day)
4. #6 + #11 — provenance + regression seedset. Now the loop is self-correcting. (~1 day)
5. Everything else opportunistically.

After #1–#7 we have a real "look at traces, see exactly where our bot loses, fix it, re-run, watch the metric move" loop — which is the actual product.
