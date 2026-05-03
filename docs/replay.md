# Command replay

Download raw replay JSON from codingame.com for offline viewing and conversion.

`arena replay <username> [<id|url>[,<id|url>...]]`

- With **no IDs**, downloads every replay from `<username>`'s last-battles list on the active game's leaderboard.
- With **one or more IDs/URLs**, downloads only those games.

The leaderboard slug is baked into each game engine (e.g. `winter-challenge-2026-snakebyte`, `spring-challenge-2020`), so you no longer pass the puzzle URL on the command line — the active game is selected via `--game` (or `arena.yml`).

`<username>` is the player you are playing for. It is recorded as the top-level `blue` field in each saved replay so the viewer and `convert` know which side is "yours".

## Quick start

```shell
# Download every replay from a player's leaderboard last-battles list
bin/arena replay mrsombre --game winter2026

# Download specific replays (comma- or space-separated)
bin/arena replay mrsombre --game winter2026 875142454,875142455
bin/arena replay mrsombre --game winter2026 875142454 875142455
```

Files are saved as `<gameId>.json` under `--out` (default `./replays/`).

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

## Options

| Flag           | Default     | Description                                       |
|----------------|-------------|---------------------------------------------------|
| `-o, --out`    | `./replays` | Directory to save replays as `<gameId>.json`      |
| `-n, --limit`  | `0`         | Maximum replays to download (`0` = all)           |
| `-l, --league` | `4`         | League level recorded in saved replay             |
| `--delay`      | `500ms`     | Delay between requests                            |
| `-f, --force`  | `false`     | Re-download even if file already exists           |

## Output

Per-replay status lines, then a final summary:

```
[1/3] save 875142454 (12345 bytes)
[2/3] skip 875142455 (exists)
[3/3] fail 875142456: <error>
done: 1 saved, 1 skipped, 1 failed (out=./replays)
```

Leaderboard mode also prints resolution steps before downloading:

```
puzzle: winter-challenge-2026-snakebyte -> winter-challenge-2026-snakebyte (cached)
player: mrsombre -> agentId 12345 (rank 210, division 3)
battles: 50
```

Saved replays are pretty-printed JSON ready for `arena convert` to turn into trace files.

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

For replays saved before the seed-promotion change, `refereeInput` is still preserved on read; `arena convert` falls back to parsing it when the top-level `seed` is absent.
