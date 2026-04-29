# Command replay

Download raw replay JSON from codingame.com for offline viewing and conversion.

Two subcommands:

- `replay get` — download specific replays by ID or URL
- `replay leaderboard` — download every replay from a player's "last battles" list

`<username>` is the player you are playing for. It is recorded as the top-level `blue` field in each saved replay so the viewer and `convert` know which side is "yours".

## Quick start

```shell
# Download specific replays
bin/arena replay get mrsombre 875142454,875142455

# Download every replay listed for a player on a leaderboard
bin/arena replay leaderboard mrsombre \
  https://www.codingame.com/multiplayer/bot-programming/winter-challenge-2026/leaderboard
```

Files are saved as `<gameId>.json` under `--out` (default `./replays/`).

## `replay get`

```
arena replay get <username> <id|url>[,<id|url>...]
```

Accepts any mix of:
- numeric IDs: `875142454`
- comma-separated lists: `875142454,875142455,875142456`
- full replay URLs ending in an ID

## `replay leaderboard`

```
arena replay leaderboard <username> <puzzle-url|slug>
```

Resolves the puzzle slug, looks up the player's `agentId`, and downloads each game from their last-battles list.

Accepted slug forms:
- bare slug: `winter-challenge-2026`
- multiplayer URL: `https://www.codingame.com/multiplayer/bot-programming/winter-challenge-2026/leaderboard`
- contest URL: `https://www.codingame.com/contests/winter-challenge-2026/leaderboard/global`

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

`leaderboard` also prints resolution steps before downloading:

```
puzzle: winter-challenge-2026 -> winter-challenge-2026 (cached)
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
| `source`      | `get` or `leaderboard`                           | Which subcommand produced this file                                    |
| `fetched_at`  | RFC 3339 timestamp at download time              | Lets `analyze` filter cohorts chronologically                          |
| `leaderboard` | `replay leaderboard` only                        | `{rank, division, score}` of the player at fetch time                  |

`league` and `leaderboard.division` are deliberately separate: the former is the level a given match was played at, the latter is where the player currently sits on the ladder (Wood / Bronze / Silver / Gold / Legend, indexed from 0).

For replays saved before the seed-promotion change, `refereeInput` is still preserved on read; `arena convert` falls back to parsing it when the top-level `seed` is absent.
