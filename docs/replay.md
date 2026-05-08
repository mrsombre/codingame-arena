# Command replay

Download raw replay JSON from codingame.com. Each freshly-saved replay is **auto-converted** to a verified arena trace under the same id (`replays/<id>.json` → `traces/replay-<id>.json`).

`arena replay <username> [<id|url>[,<id|url>...]]`

- With **no IDs**, downloads every replay from `<username>`'s last-battles list on the active game's leaderboard.
- With **one or more IDs/URLs**, downloads only those games.

The leaderboard slug is baked into each game engine (e.g. `winter-challenge-2026-snakebyte`, `spring-challenge-2020`), so you no longer pass the puzzle URL on the command line — the active game is selected via `--game` (or `arena.yml`).

`<username>` is the player you are playing for. It is recorded as the top-level `blue` field in each saved replay so the viewer and the trace know which side is "yours".

## Quick start

```shell
# Download + auto-convert every replay from a player's leaderboard last-battles list
bin/arena replay mrsombre --game winter2026

# Download + auto-convert specific replays (comma- or space-separated)
bin/arena replay mrsombre --game winter2026 875142454,875142455
bin/arena replay mrsombre --game winter2026 875142454 875142455

# Re-download AND re-convert in place (overwrites both files under the same id)
bin/arena replay mrsombre --game winter2026 875142454 -f
```

Replay JSON is saved under `--out` (default `./replays/`) as `<gameId>.json`. The matching trace is written to `--trace-dir` (default `./traces/`) as `replay-<gameId>.json`.

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
3. Verify final scores, ranks, and turn count match the replay.
4. Write the trace to `--trace-dir` as `replay-<gameId>.json`. Verifier disagreements are written as a `MISMATCH` trace (so engine output stays comparable to the raw replay) instead of being skipped; pre-run failures (e.g. missing seed, unknown blue) are skipped with no trace.

When a replay is skipped during download (already on disk, no `-f`), its existing trace is left alone. Re-converting just the trace without re-downloading is supported: delete the trace file and re-run — the replay is left in place and only the missing trace is regenerated.

## Options

| Flag           | Default     | Description                                                       |
|----------------|-------------|-------------------------------------------------------------------|
| `-o, --out`    | `./replays` | Directory to save replays as `<gameId>.json`                      |
| `--trace-dir`  | `./traces`  | Directory to write converted trace files as `replay-<gameId>.json`|
| `-n, --limit`  | `0`         | Maximum replays to download (`0` = all)                           |
| `-l, --league` | `4`         | League level recorded in the saved replay's `league` field        |
| `--delay`      | `500ms`     | Delay between successive downloads (skipped-existing replays don't reset it) |
| `-f, --force`  | `false`     | Re-download replay AND re-convert trace, overwriting both         |

## Output

Per-replay status lines: a `save` (or `skip`/`fail`) for the download, followed by a `trace` (or `skip`/`fail`) line whenever the replay was processed by the auto-convert step. Two summary lines at the end:

```
[1/3] save 875142454 (12345 bytes)
[1/3] trace 875142454 (league=4 turns=187 scores=24.0:18.0)
[2/3] skip 875142455 (exists)
[3/3] save 875142456 (98765 bytes)
[3/3] trace 875142456 MISMATCH (replay mismatch: final score mismatch: replay=[...] engine=[...])
done: 2 saved, 0 skipped-existing, 0 skipped-puzzle, 0 failed (out=./replays)
traces: 1 saved, 1 saved-mismatch, 0 skipped-existing, 0 skipped-mismatch, 0 failed (out=./traces)
```

The progress line's `scores` is the post-OnEnd value matching CG's `gameResult.scores` (the verifier's only score check). The trace file additionally stores the raw pre-OnEnd value as `scores` and the post-OnEnd value as `finalScores`; the two diverge whenever `OnEnd()` adjusts the raw in-game count — e.g. spring2021 adds `floor(sun/3)` to each player on game end.

Leaderboard mode also prints resolution steps before downloading:

```
puzzle: winter-challenge-2026-snakebyte -> winter-challenge-2026-snakebyte (cached)
player: mrsombre -> agentId 12345 (rank 210, division 3)
battles: 50
```

A `MISMATCH` trace usually means the engine doesn't yet match the recorded league behavior — the trace is still written so engine output and replay can be compared, and `-f` re-runs the conversion after an engine fix.

## Saved file shape

Each saved replay is the upstream CodinGame `gameResult` body with viewer-only payloads stripped (top-level `viewer`, `shareable`; `gameResult.metadata`, `tooltips`; per-frame `view`, `gameInformation`, `keyframe`), the seed lifted to the top level, the bulky `frames` array hoisted out of `gameResult` to sit last in the file, and arena-only annotations layered on.

Top-level keys are emitted in this fixed order — annotation metadata first, then the bulky payloads at the end:

| # | Field           | Source                                          | Description                                                                |
|---|-----------------|-------------------------------------------------|----------------------------------------------------------------------------|
| 1 | `fetchedAt`     | RFC 3339 timestamp at download time             | Lets `analyze` filter cohorts chronologically                              |
| 2 | `source`        | `get` or `leaderboard`                          | Which mode produced this file (no IDs → `leaderboard`, IDs → `get`)        |
| 3 | `puzzleId`      | CG API; backfilled from `factory.PuzzleID()` when CG returned 0 | Canonical CodinGame puzzle id; gates cross-game replay rejection |
| 4 | `puzzleTitle`   | CG API; backfilled from `factory.PuzzleTitle()` when CG returned 0 | Two-element array as CG's API emits it                          |
| 5 | `questionTitle` | CG API                                          | Full match title (e.g. `Spring Challenge 2021 - Level 4`); `league` is parsed from it |
| 6 | `replayId`      | mirror of `gameResult.gameId`                   | Canonical replay/match id, hoisted to the top level for convenience        |
| 7 | `players`       | `gameResult.agents[].codingamer.pseudo` (or `arenaboss.nickname` for boss matches) | `[left, right]` display names, indexed by agent `index`         |
| 8 | `blue`          | `<username>` argument                           | Player we are playing for (the analyze "us" side)                          |
| 9 | `leaderboard`   | leaderboard mode only                           | `{rank, division, score}` of the player at fetch time                      |
| 10| `league`        | parsed from `questionTitle` (e.g. `Level 3` → 3) | League level the match was played at                                      |
| 11| `seed`          | extracted from `gameResult.refereeInput`        | Match RNG seed; JSON-string-encoded int64. Replaces `refereeInput`.        |
| 12| `gameResult`    | upstream CG payload (viewer-only fields stripped, `frames` hoisted out) | Agents, scores, ranks, gameId                              |
| 13| `frames`        | hoisted from `gameResult.frames`                | Per-turn engine output (the bulky payload, kept last so the metadata stays grouped at the top of the file) |

`league` and `leaderboard.division` are deliberately separate: the former is the level a given match was played at, the latter is where the player currently sits on the ladder (Wood / Bronze / Silver / Gold / Legend, indexed from 0).

For replays saved before the seed-promotion change, `refereeInput` is still preserved on read and parsed as a fallback when the top-level `seed` is absent. For replays saved before the `frames` hoist, the nested `gameResult.frames` shape is still accepted on read.

## Sample replay file

```json
{
  "fetchedAt": "2026-05-07T09:06:10Z",
  "source": "get",
  "puzzleId": 730,
  "puzzleTitle": [
    "Spring Challenge 2021",
    "Spring Challenge 2021"
  ],
  "questionTitle": "Spring Challenge 2021 - Level 4",
  "replayId": 886403710,
  "players": ["mrsombre", "MiyazaBoss"],
  "blue": "mrsombre",
  "league": 4,
  "seed": "-4436915910920504000",
  "gameResult": {
    "agents": [
      {
        "index": 0,
        "codingamer": { "userId": 469626, "pseudo": "mrsombre", "avatar": 161451360826472 },
        "agentId": -1
      },
      {
        "index": 1,
        "agentId": 3664808,
        "score": 45.27,
        "valid": true,
        "arenaboss": { "nickname": "MiyazaBoss", "league": { "divisionIndex": 4 } }
      }
    ],
    "gameId": 886403710,
    "ranks": [0, 1],
    "scores": [133.0, 131.0]
  },
  "frames": [
    { "agentId": -1, "summary": "" },
    { "agentId": 0, "stdout": "WAIT GL HF\n" },
    { "agentId": 1, "stdout": "WAIT\n", "summary": "Round 0/23\n$0 is waiting\n$1 is waiting\n" }
  ]
}
```

The boss-vs-player case shown above puts the opponent under `arenaboss.nickname` instead of `codingamer.pseudo`; both shapes resolve to a player display name in the trace.
