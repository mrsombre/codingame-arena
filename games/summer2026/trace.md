# Summer Challenge 2026 — Trace format

This document describes the summer2026-specific parts of the arena trace
format: the match-level `setup` payload, the per-turn `gameInput` lines, and
the per-player events emitted into `turns[].traces`. The cross-game envelope
(top-level fields, `turns[]` shape, file naming, JSON conventions) is
documented in [docs/trace.md](../../docs/trace.md).

## Identity

| Field        | Value                                |
| ------------ | ------------------------------------ |
| `puzzleId`   | `0` (community contest, no puzzleId) |
| `puzzleName` | `summer2026`                         |
| Turn model   | `FlatTurnModel`                      |

`FlatTurnModel` applies because `Game.ShouldSkipPlayerTurn` always returns
`false`, the referee runs exactly one `PerformGameUpdate` per loop turn, and
the match ends on the same turn the end condition fires — there is no phase
frame and no post-end frame. So `mainTurns` equals `len(turns)`, capped at
`MAX_TURNS = 100`.

## `endReason`

| Value           | Meaning                                                                             |
| --------------- | ----------------------------------------------------------------------------------- |
| `SCORE`         | Both normal endings: the turn cap, or no desired connection has any route left.     |
| `TIMEOUT`       | A player failed to answer in time on a turn after its first.                        |
| `TIMEOUT_START` | A player failed to answer on the very first turn it was prompted.                   |
| `INVALID`       | A player emitted an unparseable command and was disqualified by `CommandManager`.   |

The engine does not distinguish the turn cap from the no-route-left ending, so
`SCORE` covers both; `mainTurns < 100` is what tells them apart.

Rejections that do **not** end the match — an off-grid cell, an inked region,
an exhausted paint budget — are not end reasons. They appear as `FAILED`
events on the turn they happened.

## Match-level `setup` lines

`setup` is the raw `string[]` from `SerializeGlobalInfoFor(players[0], game)`
([engine/game_serializer.go](engine/game_serializer.go)) — the initialization
block either bot receives on stdin at match start. Only the first line differs
between the sides, so the trace records side 0's copy.

```
<playerId>                 # 0 in the trace — side 0's copy
<width>
<height>
<regionId> <type>          # width*height lines, row-major from (0, 0)
...
<townCount>
<id> <x> <y> <desired>     # townCount lines
...
```

| Field      | Type   | Description                                                                                                     |
| ---------- | ------ | --------------------------------------------------------------------------------------------------------------- |
| `playerId` | int    | Recipient's side index. Always `0` in a trace.                                                                   |
| `width`    | int    | Grid width in cells (`height * 1.5`, rounded).                                                                   |
| `height`   | int    | Grid height in cells, `14..20`.                                                                                  |
| `regionId` | int    | Index into the region list. Regions are the unit disruption and inking operate on.                               |
| `type`     | int    | Terrain: `0` plains, `1` river, `2` mountain, `3` point of interest (never generated — the side quest is inert). |
| `id`       | int    | Town id, also the `from` / `to` of a `SCORE` event.                                                              |
| `x` `y`    | int    | Town cell.                                                                                                       |
| `desired`  | string | Comma-separated ids this town wants connected to, or `x` for none. Unilateral — the counterpart need not list back. |

## Frame model

Each loop turn maps 1:1 to one Java game turn. `turns[].turn` is the arena's
loop counter and starts at `0`; the engine's own counter starts at `1`, so the
`turn` field inside a `TURN` event is always one higher than the enclosing
`turns[].turn`.

Within a turn:

1. The runner sends each side its frame info and reads one line back (one or
   more semicolon-separated commands — `AUTOPLACE`, `PLACE_TRACK` /
   `PLACE_TRACKS`, `DISRUPT`, `MESSAGE`, `WAIT`).
2. `CommandManager.ParseCommands` turns the line into intents. An unparseable
   command disqualifies the player on the spot; a parseable one is queued.
3. `Game.PerformGameUpdate` runs, in this order:
   `DoIncome` → `ComputeAutobuilds` → `DoActions` → `doDisruptions` →
   `DoInstabilityCheck` → `MoveTrains` → `ComputeTileStates` → game-over check.
4. The turn's events are handed to the runner.

The trace buffer is cleared in `Game.ResetGameTurnData`, which runs before
command parsing, so the `MESSAGE` events parsing emits share a bucket with the
events the update resolves.

## Per-turn `gameInput` payload

`turns[].gameInput` is the raw `string[]` from
`SerializeFrameInfoFor(players[0], game)`
([engine/game_serializer.go](engine/game_serializer.go)) — the lines side 0's
bot received on stdin this turn. There is no fog of war; the only
recipient-relative part is the score pair, which is always "mine then theirs",
so the trace's first score line is side 0's.

```
<myScore>
<oppScore>
<track> <instability> <inked> <connections>   # width*height lines, row-major
...
```

| Field         | Type   | Description                                                                                                    |
| ------------- | ------ | -------------------------------------------------------------------------------------------------------------- |
| `myScore`     | int    | Side 0's running score.                                                                                         |
| `oppScore`    | int    | Side 1's running score.                                                                                         |
| `track`       | int    | Track owner on the cell: `-1` none, `0` / `1` a side, `2` contested (neutral — scores for nobody).              |
| `instability` | int    | The cell's region's instability count. Inks out at `INSTABILITY_THRESHOLD_BASE = 4`.                            |
| `inked`       | int    | `1` when the cell's region is inked out — permanently unbuildable and no longer disruptable.                     |
| `connections` | string | Comma-separated `fromTownId-toTownId` pairs whose current path crosses this cell, or `x`. Sorted as strings, so `10-2` precedes `2-3`. |

`gameInput` is sampled before the bots are prompted, so it is the board they
chose their commands against.

There is no `state` payload: `gameInput` already carries the full board (no fog
of war), and the per-player ledger lives on the `TURN` event.

## Per-turn `score`

`turns[].score` is the raw pre-`OnEnd` score pair going into the turn, sampled
before `PerformGameUpdate`. The `TURN` event's `score` is the same counter
sampled *after* the update, so `turns[n].traces[i].TURN.score` equals
`turns[n+1].score[i]`.

## Per-player events (`turns[].traces[playerIdx]`)

The index into `traces[]` is the match side, the same indexing as `players`,
`score` and `gameInput`'s score lines. Two event types belong to no single
side — `CONTESTED` and `INK` — and are mirrored identically into both buckets.

All event structs live in [engine/traces.go](engine/traces.go).

Events land in engine order, which is fixed:

`MESSAGE` (parse time) → `AUTOPLACE` → `TRACK` → `CONTESTED` → `DISRUPT` →
`INK` → `SCORE` → `TURN`, with `FAILED` interleaved at the point the rejected
command would have resolved.

| `type`      | Emitted on                              | `data`                                     | Notes                                                                                                                                                                                                              |
| ----------- | --------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `MESSAGE`   | `MESSAGE <text>` parsing                | `{text}`                                    | The bot's verbatim message. Emitted at parse time, before any update event of the same turn. Several `MESSAGE` tokens each emit their own event even though only the last survives in `Player.Message`.             |
| `AUTOPLACE` | An honoured `AUTOPLACE` is expanded     | `{from, to, placements}`                    | One per turn per player at most; a second `AUTOPLACE` is rejected as `FAILED` instead. `placements` is the rail count of the winning plan — `0` when no route exists, the only way an `AUTOPLACE` silently does nothing. |
| `TRACK`     | A rail is paid for                      | `{cell, cost, terrain, autoplaced?}`        | One per rail. `cost` is what that cell charged; `terrain` is the kind that set it. `autoplaced` marks a rail the planner chose. Ownership is written only after both sides are charged, so a cell that also has a `CONTESTED` event this turn ended up neutral despite this event naming its buyer. |
| `CONTESTED` | Both sides bought the same cell         | `{cell}`                                    | Mirrored into both buckets. Neither side is refunded and the cell goes neutral (`track == 2`), scoring for nobody.                                                                                                   |
| `DISRUPT`   | An honoured disruption                  | `{zone, instability}`                       | `instability` is the region's count *after* this blot. A player gets `BLOT_POINTS_PER_TURN = 1` a turn, so naming two regions in one line lands one and rejects the other.                                          |
| `INK`       | A region reaches the threshold          | `{zone, credited?, tracksLost}`             | Mirrored into both buckets. `credited` lists the sides whose last honoured blot named this region, and is absent when the region tipped over on nobody's blot — such an inking counts for neither side's metrics. `tracksLost` is indexed by track owner: `[0]` and `[1]` the sides, `[2]` the contested rails credited to nobody. |
| `SCORE`     | A connection pays out                   | `{from, to, points, pathLength, detour}`    | Once per connection per turn, every turn it stands — not once when it is made. Connections are directional, so a mutual pair yields two events. `points` is one per rail the side owns along the path; `pathLength` counts the whole path including both towns; `detour` is `pathLength` minus the Manhattan distance between them. |
| `FAILED`    | A parseable-but-illegal command          | `{reason}`                                  | The engine's verbatim rejection message, minus the player prefix and the viewer's colour markers. Unparseable commands are **not** reported here — they disqualify the player and surface as `endReason: INVALID`. |
| `TURN`      | End of every turn, for every side        | see below                                   | Always the last event in a bucket, and always present — including for a side that was disqualified earlier in the match.                                                                                             |

### `TURN` data

The per-turn ledger the other events add up to, so a turn can be read without
replaying its event stream.

| Field               | Type | Description                                                                                                                                      |
| ------------------- | ---- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `turn`              | int  | The engine's own turn counter, starting at `1` — one higher than the enclosing `turns[].turn`.                                                    |
| `paintAvailable`    | int  | Paint granted this turn, `PASSIVE_INCOME = 3`. Granted fresh each turn, never accumulated.                                                        |
| `paintSpent`        | int  | Paint the rails in this turn's `TRACK` events cost.                                                                                              |
| `paintLeft`         | int  | Paint unspent at the end of the turn. Lost, not saved — waste, not savings. Read off the live budget rather than derived, so the three paint fields failing to add up means a spend the trace did not see. |
| `tracksPlaced`      | int  | Rails laid this turn; equals the count of `TRACK` events.                                                                                        |
| `tracksAutoplaced`  | int  | How many of those the `AUTOPLACE` planner chose rather than the bot naming them. `tracksPlaced - tracksAutoplaced` is the hand-placed count.      |
| `disruptAvailable`  | int  | Disruption points granted this turn, `BLOT_POINTS_PER_TURN = 1`. Also granted fresh each turn.                                                    |
| `disruptSpent`      | int  | Disruptions honoured; equals the count of `DISRUPT` events.                                                                                      |
| `connectionsActive` | int  | Connections that **paid this side** this turn, i.e. the count of `SCORE` events — not every connection standing on the board.                     |
| `pointsEarned`      | int  | Points scored this turn; the sum over this turn's `SCORE` events.                                                                                |
| `score`             | int  | Running total after this turn, before the end-of-game verdict `OnEnd` may replace it with (`-1` for a disqualified side, or the tutorial verdict). |

### What is **not** a per-player event

- **Income** — constant and reported on every `TURN` event as
  `paintAvailable` / `disruptAvailable`.
- **Connection paths** — recoverable from `gameInput`'s per-cell
  `connections` column.
- **Region instability short of the threshold** — carried on `gameInput`'s
  per-cell `instability` column, and on `DISRUPT` events for the changes a
  side caused itself.
- **Game over** — surfaced through the envelope's `endReason` and
  `mainTurns`.

## League rules and trace shape

`league` is stamped on the envelope. Leagues 3–5 run identical rules and
differ only in what the CodinGame statement reveals, so the trace shape is the
same. Leagues 1–2 are tutorials: the win condition comes from
`TutorialManager` and both final scores are replaced by its verdict, so
`finalScores` bears no relation to `scores` (the raw in-play points) there.
Event vocabulary is unchanged in every league.

## Example

One turn of a real self-play trace, trimmed. The envelope is shown for context;
the turn structure is the summer2026-specific part.

```json
{
  "createdAt": "2026-09-08T12:29:50Z",
  "type": "trace",
  "puzzleName": "summer2026",
  "traceId": 1788870589,
  "matchId": 1,
  "players": ["bot-summer2026-py", "bot-summer2026-cpp"],
  "blue": "bot-summer2026-cpp",
  "league": 5,
  "seed": "-468706172918629801",
  "endReason": "SCORE",
  "scores": [6440, 9170],
  "finalScores": [6440, 9170],
  "ranks": [1, 0],
  "mainTurns": 100,
  "turns": [
    {
      "turn": 3,
      "output": [
        "PLACE_TRACKS 6 9;PLACE_TRACKS 7 9;PLACE_TRACKS 8 9",
        "PLACE_TRACKS 6 4;PLACE_TRACKS 6 3;PLACE_TRACKS 7 3"
      ],
      "isOutputTurn": [true, true],
      "score": [8, 0],
      "traces": [
        [
          { "type": "TRACK", "data": { "cell": [6, 9], "cost": 1, "terrain": "PLAINS" } },
          { "type": "TRACK", "data": { "cell": [7, 9], "cost": 1, "terrain": "PLAINS" } },
          { "type": "TRACK", "data": { "cell": [8, 9], "cost": 1, "terrain": "PLAINS" } },
          { "type": "SCORE", "data": { "from": 1, "to": 3, "points": 4, "pathLength": 10, "detour": 1 } },
          { "type": "SCORE", "data": { "from": 4, "to": 1, "points": 4, "pathLength": 6,  "detour": 1 } },
          { "type": "TURN",  "data": {
              "turn": 4, "paintAvailable": 3, "paintSpent": 3, "paintLeft": 0,
              "tracksPlaced": 3, "tracksAutoplaced": 0,
              "disruptAvailable": 1, "disruptSpent": 0,
              "connectionsActive": 2, "pointsEarned": 8, "score": 16 } }
        ],
        [
          { "type": "TRACK", "data": { "cell": [6, 4], "cost": 1, "terrain": "PLAINS" } },
          { "type": "TRACK", "data": { "cell": [6, 3], "cost": 1, "terrain": "PLAINS" } },
          { "type": "TRACK", "data": { "cell": [7, 3], "cost": 1, "terrain": "PLAINS" } },
          { "type": "TURN",  "data": {
              "turn": 4, "paintAvailable": 3, "paintSpent": 3, "paintLeft": 0,
              "tracksPlaced": 3, "tracksAutoplaced": 0,
              "disruptAvailable": 1, "disruptSpent": 0,
              "connectionsActive": 0, "pointsEarned": 0, "score": 0 } }
        ]
      ]
    }
  ]
}
```

Things worth pointing out:

- Both sides spent their whole budget on three plains rails, one paint each.
  Neither used `AUTOPLACE`, so no rail is marked `autoplaced`.
- Side 0's rails completed two connections. `1-3` and `4-1` are separate
  directional connections, so they pay separately: 4 points each, 8 for the
  turn. `turns[].score` is `[8, 0]` — the total *going into* the turn — and the
  `TURN` event's `score: 16` is the total after it.
- The `4-1` connection runs a 6-cell path with a `detour` of 1, so the rails
  take one step more than the straight line between the two towns; the `1-3`
  connection is 10 cells with the same one-step detour.
- Side 1 owns no rail on any active path yet, so it has no `SCORE` event and
  its `TURN` event reports `connectionsActive: 0`.
- `endReason: SCORE` with `mainTurns: 100` means the match ran out the turn
  cap rather than ending early on an unreachable board.
