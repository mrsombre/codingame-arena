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

---

## P0 — Blocking gaps

### 1. Identify "us" in every trace

The replay JSON already carries `blue: <username>` (set by `replay get`/`leaderboard`), but `convert` discards it — `RunReplay` writes `Players: [pseudo0, pseudo1]` with no flag pointing to which side is "us". Without this, `analyze` can only compute anonymous **winner-vs-loser**, not **us-when-winning vs us-when-losing**, which is the only comparison that tells us how to improve.

- Add `Blue string` (and ideally `BlueSide int` 0/1) to `TraceMatch`.
- `convert.go` propagates `replay.Blue` plus matches it against `replay.GameResult.Agents[].codingamer.pseudo` to set `BlueSide`.
- For self-play traces, expose a `--trace-blue p0|p1|<basename>` flag on `arena run` so the same field is set when comparing two of our own bots.

`internal/arena/trace.go:37`, `internal/arena/replay_runner.go:160`, `internal/arena/commands/convert.go:81`.

### 2. Winter 2026 has no analyzer

`Factory.AnalyzeTraces` is implemented for Spring 2020 only (`games/spring2020/engine/analyzer.go:14`). For winter2026, `commands.Analyze` errors with `"game does not implement trace analysis"`.

Implement `games/winter2026/engine/analyzer.go` with the same shape but Snake-relevant breakdowns:

- Command rates: `UP/DOWN/LEFT/RIGHT/MARK/WAIT` per side (rate, not absolute, per spring2020's pattern at `analyzer.go:270`).
- Event counts (per match avg, winner-vs-loser diff): `EAT`, `HIT_WALL`, `HIT_SELF`, `HIT_ENEMY`, `DEAD`, `FALL` — labels already emitted at `games/winter2026/engine/traces.go:11`.
- Apples eaten / total apples ratio per side.
- First-death-turn distribution per side.
- Average bird length over time per side (proxy for territorial dominance).
- Losses count from `Referee.Metrics()` (already exposes `losses_p0`, `losses_p1`).

### 3. Switch analyze from "winner vs loser" to "us vs them"

Once `Blue` lands, change the default reporting axis from anonymous winner/loser to **us-when-winning** vs **us-when-losing**. That's the comparison that maps directly to "what should I change in the bot".

- Three buckets per analyze run: `our_wins`, `our_losses`, `draws`.
- For each event/command, show the **diff between our wins and our losses** (not winner vs loser of the field).
- Keep the old winner/loser mode reachable behind `--axis winner-loser` for completeness.

### 4. Surface worst losses from disk

`run.go:106` reports `worst_losses` for the current batch only — useful in CI, useless for analyzing 200 accumulated traces. `analyze` should additionally:

- List the **N largest-margin losses** with the trace filename, replay URL (if `Type == replay`), seed, opponent name, final scores, and turn the first death happened.
- List the **N closest losses** (margin ≤ small threshold) — these are the matches a slightly better bot would have won, and are the highest-leverage fixes.

Output as a section in the existing report; add `--worst N` and `--close N` flags.

### 5. Filter trace cohorts in analyze

Today `analyze` reads every `*.json` in `traces/` and lumps simulator runs and real-CG replays together. They have different distributions and should be analyzed separately.

Add filters to `commands/analyze.go`:

- `--type trace|replay|all` (use `arena.TraceTypeFromFileName`, `internal/arena/trace.go:164`).
- `--player <basename>` to keep only matches where one side equals that player.
- `--league N` once league is recorded on the trace (see #6).
- `--seeds-from <file>` so we can re-run a fixed cohort of seeds.

---

## P1 — High lever

### 6. Record provenance on every trace

Without provenance, traces decay quickly: we can't tell which bot version produced trace from yesterday, which league it was played at, or whether it's still relevant.

Add to `TraceMatch`:

- `League int` — copied from converted replays' parsed league, set on self-play from the configured league.
- `BotVersions [2]string` — git short SHA at run time (computed once per `arena run` invocation), or content hash of the bot binary.
- `TraceLabel string` — value of a new `--trace-label` flag on `arena run` (cohort tagging like `experiment-mark-strategy`).
- `Source string` — `"self-play"` or `"codingame"`.

Then `analyze --label experiment-A` becomes possible, and we can keep traces from incompatible bot versions in the same dir without polluting reports.

### 7. Per-turn diagnostic for our bot

The single highest-signal report for snake-style games: **when do we die, and why?** Build a per-trace summary that the analyzer aggregates:

- Histogram of `DEAD` events binned by turn (early-death = strategic blunder; late-death = endgame mistake).
- Death cause distribution: `HIT_WALL` / `HIT_SELF` / `HIT_ENEMY` / `FALL` for our side vs theirs.
- Apples eaten before vs after first death.
- For losses, the last 5 turns of game_input + our output (so we can see the actual "last decision before disaster" without opening the viewer).

Surface a `last-decisions` section in `analyze --worst N`.

### 8. Pairwise matchup table

`replay leaderboard` pulls real CG matches against many opponents. Tabulate winrate vs each opponent (group by `Players[non-us]` after #1). Lets us spot a single opponent we keep losing to and study just those traces.

Output in JSON or text; add `--by opponent` flag.

### 9. Convert mismatch quarantine

`convert` currently logs `skipped-mismatch` lines to stdout (`commands/convert.go:76`). Each mismatch is engineering signal — the engine doesn't yet match CG behavior, and our trace dataset is silently smaller than expected.

- Write skipped replays + reason to `<trace-dir>/_mismatch.json` with the replay id, league, expected vs actual scores, expected vs actual turns.
- `arena convert --report-mismatches` prints aggregate counts per mismatch category for a quick "engine fidelity" health check.
- A future `make engine-fidelity` target can fail if mismatch rate increases.

### 10. JSON output for analyze

Today `analyze` prints text only (`spring2020/engine/analyzer.go:161`). Make the report `Write(w, format)` aware (`text` default, `json` opt-in). Enables:

- Diffing two analyze runs across bot versions to spot regressions.
- Feeding analyze results into a separate viewer page.
- CI checks (e.g. fail PR if `EAT` rate dropped > 5%).

### 11. Self-play regression seedset

A frozen list of seeds we re-run with `arena run --seeds-from seedset.txt` (combined with #5). New seeds added when we discover a loss pattern. Becomes our regression suite — Make target `make regression-winter2026` runs N seeds, compares aggregate stats to a recorded baseline, fails on regression.

---

## P2 — Polish / future

### 12. Mortality heatmap per tile (winter2026)

The board is small (grid is ~30 cells). For each `DEAD`/`HIT_*` event we already have `Coord` in the payload (`games/winter2026/engine/traces.go:19`). Aggregate per-cell counts in losses → emit as ASCII heatmap, or a JSON the viewer can render.

Reveals sticky death-zones (apples in dangerous corners, recurring trap cells).

### 13. Apple-contention metric

For each apple eaten, record turn + side. Compute (per match) the "apple race" balance: did we get the apples within reach, or did the opponent reach them first? Diff this between our wins/losses.

### 14. Auto-rerun replay seeds in simulator

After `convert` saves `replay-<id>.json`, optionally also run `arena run --seed <replay.seed> --p0 <our-current-bot> --p1 <something>` and emit a paired `trace-replay-<id>.json`. We get to see how our **current** bot would have played the same seed, side-by-side with the historical CG match.

CLI: `arena convert --rerun --p0 bin/bot-winter2026-cpp` writes both files; analyze reports the delta.

### 15. Track leaderboard rank at replay-fetch time

`replay leaderboard` doesn't record the player's elo/rank when downloading. Save it alongside the replay (e.g., `replays/<id>.meta.json`) so analyze can group by elo bands ("losses against top-100" vs "losses against ladder peers").

### 16. Move legality / wasted-action analysis

Some moves are no-ops (WAIT, MARK in conditions where it doesn't help) or self-destructive (MOVE into wall). Count rate of:

- WAIT turns by us — every WAIT is a possibly missed opportunity.
- MOVE → immediate `HIT_WALL`/`HIT_SELF` on the same turn — i.e., suicidal moves.
- MARK followed by no apparent strategy effect.

These are hands-on, fix-this-class-of-bug signals.

### 17. Replay viewer integration

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
