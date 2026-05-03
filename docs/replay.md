# Command replay

Download raw replay JSON from codingame.com and immediately convert each freshly-downloaded file into a verified arena trace.

`arena replay <username> [<id|url>[,<id|url>...]]`

- With **no IDs**, downloads every replay from `<username>`'s last-battles list on the active game's leaderboard.
- With **one or more IDs/URLs**, downloads only those games.

The leaderboard slug is baked into each game engine (e.g. `winter-challenge-2026-snakebyte`, `spring-challenge-2020`), so you no longer pass the puzzle URL on the command line — the active game is selected via `--game` (or `arena.yml`).

`<username>` is the player you are playing for. It is recorded as the top-level `blue` field in each saved replay so the viewer and the trace know which side is "yours".

## Quick start

```shell
# Download + convert every replay from a player's leaderboard last-battles list
bin/arena replay mrsombre --game winter2026

# Download + convert specific replays (comma- or space-separated)
bin/arena replay mrsombre --game winter2026 875142454,875142455
bin/arena replay mrsombre --game winter2026 875142454 875142455
```

Replay JSON is saved under `--out` (default `./replays/`) as `<gameId>.json`. The matching trace is written to `--trace-dir` (default `./traces/`) as `replay-<gameId>-0.json`.

## Argument forms

IDs may be passed as multiple args, or as a comma-separated list within a single arg, or both. Each token is one of:

- numeric ID: `875142454`
- full replay URL ending in an ID

## Leaderboard mode

When invoked with no IDs, the command:

1. Reads the active game's `LeaderboardSlug()` (e.g. `winter-challenge-2026-snakebyte`).
2. Resolves the slug → puzzle leaderboard ID via the CodinGame API.
3. Looks up the player's `agentId` on that leaderboard.
4. Downloads every game from their last-battles list.

Puzzle slug and agent ID lookups are cached in `db.sqlite3` to avoid repeated API calls.

## Auto-convert step

For every replay that is **freshly downloaded** (the file did not yet exist on disk, or `-f` was used), the command immediately runs the same conversion the old `arena convert` performed:

1. Parse the replay JSON and extract the seed.
2. Re-run the engine using the recorded player moves and the league parsed from the replay title.
3. Verify final scores and turn count match the replay; if they diverge, skip writing the trace (logged as a `skip` with `replay mismatch:` detail).
4. Write the verified trace to `--trace-dir` as a replay-typed trace file keyed by replay ID.

Replays that are skipped during download (already on disk, no `-f`) are **not** re-converted — their existing trace is left alone. Mismatches are logged but do not abort the batch.

## Options

| Flag           | Default     | Description                                       |
|----------------|-------------|---------------------------------------------------|
| `-o, --out`    | `./replays` | Directory to save replays as `<gameId>.json`      |
| `--trace-dir`  | `./traces`  | Directory to write converted trace files          |
| `-n, --limit`  | `0`         | Maximum replays to download (`0` = all)           |
| `-l, --league` | `4`         | League level recorded in saved replay             |
| `--delay`      | `500ms`     | Delay between requests                            |
| `-f, --force`  | `false`     | Re-download (and re-convert) even if file exists  |

## Output

Per-replay status lines: a `save` (or `skip`/`fail`) for the download, followed by a `trace` (or `skip`/`fail`) line whenever the replay was freshly downloaded. Two summary lines at the end:

```
[1/3] save 875142454 (12345 bytes)
[1/3] trace 875142454 (league=4 turns=187 scores=24.0:18.0)
[2/3] skip 875142455 (exists)
[3/3] save 875142456 (98765 bytes)
[3/3] skip 875142456 (replay mismatch: score mismatch: replay=[...] engine=[...])
done: 2 saved, 1 skipped, 0 failed (out=./replays)
traces: 1 saved, 0 skipped-existing, 0 skipped-puzzle, 1 skipped-mismatch, 0 failed (out=./traces)
```

Leaderboard mode also prints resolution steps before downloading:

```
puzzle: winter-challenge-2026-snakebyte -> winter-challenge-2026-snakebyte (cached)
player: mrsombre -> agentId 12345 (rank 210, division 3)
battles: 50
```

A `skipped-mismatch` trace usually means the engine doesn't yet match the recorded league behavior — the replay JSON is preserved so the engine fix can be re-tested with `-f`.

## Saved file shape

Each saved replay is the upstream CodinGame `gameResult` body with viewer-only payloads stripped (top-level `viewer`, `shareable`; `gameResult.metadata`, `tooltips`; per-frame `view`, `gameInformation`, `keyframe`), the seed lifted to the top level, and arena-only annotations layered on:

| Field         | Source                                           | Description                                                            |
|---------------|--------------------------------------------------|------------------------------------------------------------------------|
| `seed`        | extracted from `gameResult.refereeInput`         | Match RNG seed; JSON-string-encoded int64. Replaces `refereeInput`.    |
| `blue`        | `<username>` argument                            | Player we are playing for (the analyze "us" side)                      |
| `league`      | parsed from `questionTitle` (e.g. `level3` → 3)  | League level the match was played at                                   |
| `source`      | `get` or `leaderboard`                           | Which mode produced this file (no IDs → `leaderboard`, IDs → `get`)    |
| `fetched_at`  | RFC 3339 timestamp at download time              | Lets `analyze` filter cohorts chronologically                          |
| `leaderboard` | leaderboard mode only                            | `{rank, division, score}` of the player at fetch time                  |

`league` and `leaderboard.division` are deliberately separate: the former is the level a given match was played at, the latter is where the player currently sits on the ladder (Wood / Bronze / Silver / Gold / Legend, indexed from 0).

For replays saved before the seed-promotion change, `refereeInput` is still preserved on read and parsed as a fallback when the top-level `seed` is absent.
