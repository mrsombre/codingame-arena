# Spring Challenge 2021 — Trace format

This document describes the spring2021-specific parts of the arena trace
format: the match-level `setup` payload, the per-turn `state` payload, and
the per-player events emitted into `turns[].traces`. The cross-game envelope
(top-level fields, `turns[]` shape, file naming, JSON conventions) is
documented in [docs/trace.md](../../docs/trace.md).

## Identity

| Field        | Value                              |
|--------------|------------------------------------|
| `puzzleId`   | `730`                              |
| `puzzleName` | `spring2021`                       |
| Turn model   | `PhaseTurnModel`                   |

`PhaseTurnModel` means every engine frame is recorded as its own turn,
including the engine-only `gathering` and `sun` phases (no player output
on those). `mainTurns` counts only `actions`-phase turns.

## Match-level `setup` lines

`setup` is the raw `string[]` from `SerializeGlobalInfoFor`
([engine/serializer.go](engine/serializer.go)) — same shape both bots
receive on stdin. The trace records side 0's perspective; spring2021 has
no fog of war and global info is side-independent, so the lines are
identical to what either side received.

Line layout, per the upstream Java engine:

```
<numberOfCells>                                                # always 37
<cellIdx0> <richness0> <neigh0_0> <neigh1_0> … <neigh5_0>      # 37 of these
<cellIdx1> <richness1> <neigh0_1> <neigh1_1> … <neigh5_1>
...
```

| Position           | Type   | Description                                                                                  |
|--------------------|--------|----------------------------------------------------------------------------------------------|
| `numberOfCells`    | int    | First line. Always `37` for the canonical hex board.                                         |
| `cellIdx`          | int    | Cell index, `0`–`36`. Cells appear in index-ascending order.                                  |
| `richness`         | int    | `0` unusable, `1` poor, `2` ok, `3` lush.                                                     |
| `neigh0..neigh5`   | int    | Six neighbor cell indices, one per direction `0`–`5`. `-1` means no neighbor (outer ring).    |

Once `setup` is in the trace, the per-turn `state.trees` tuple's `richness`
slot is technically redundant — kept inline for now so analyzers can read
trees without re-parsing `setup` on every lookup.

## Frame model

Each `day` (0–23) is composed of three phase types, in this order:

1. **`gathering`** — sun is harvested from non-shadowed trees. One trace
   turn per day. No player output.
2. **`actions`** — players take simultaneous actions (`SEED`, `GROW`,
   `COMPLETE`, `WAIT`). One trace turn per action sub-frame; `dayActionIndex`
   counts them within the day starting at `0`. Continues until both players
   are waiting.
3. **`sun`** — the sun rotates one step (`day mod 6`). One trace turn. No
   player output. Skipped on the final day; the final round-24 frame
   keeps the previous orientation.

The current phase is in `state.phase`.

## Per-turn `state` payload

`turns[].state` is a `TraceTurnState` ([engine/traces.go](engine/traces.go)):

| Field              | Type             | Description                                                                                                         |
|--------------------|------------------|---------------------------------------------------------------------------------------------------------------------|
| `day`              | int              | Round number, `0`–`23`.                                                                                             |
| `phase`            | string           | `"gathering"`, `"actions"`, or `"sun"` — derived from the engine's `FrameType`.                                     |
| `dayActionIndex`   | int              | Action sub-frame counter within the day, starting at `0`. Present **only** on `phase == "actions"`.                 |
| `sunDirection`     | int              | Current sun orientation, `0`–`5`. Always equals `day % 6` until the final round, where it stops advancing.          |
| `nutrients`        | int              | Forest nutrient pool. Decreases on each `COMPLETE`; floored at `0`.                                                 |
| `sun`              | `[2]int`         | Per-side sun-point balance going into this frame. Index = match side (`players[i]`).                                |
| `trees`            | `[2][][4]int`    | Per-side tree list (see below). Outer index = match side; inner list ordered by cell id ascending.                  |
| `seedConflictCell` | int              | Present only on `actions` turns where both players sent a seed to the same cell — neither was planted. Cell index. |

### `trees[player][i]` tuple

Each tree is a 4-int tuple. The leading two ints are static (cell
identity + soil quality); the trailing two ints are mutable (current
growth stage + same-day dormancy).

| Position | Field        | Range / meaning                                                         |
|----------|--------------|-------------------------------------------------------------------------|
| 0        | `cellIdx`    | `0`–`36`, the hex cell the tree occupies.                               |
| 1        | `richness`   | `1` poor, `2` ok, `3` lush. (`0` is unusable and never has a tree.)     |
| 2        | `size`       | `0` seed, `1` small, `2` medium, `3` tall.                              |
| 3        | `isDormant`  | `1` if the tree has been acted on this day (cannot be acted on again until the next day), `0` otherwise. |

## Per-player events (`turns[].traces[playerIdx]`)

Events are bucketed by owner; index into `traces[]` is the match side
(same indexing as `players` / `state.sun` / `state.trees`). All structs
live in [engine/traces.go](engine/traces.go).

| `type`     | Emitted on phase | `data`                       | Notes                                                                                                          |
|------------|------------------|------------------------------|----------------------------------------------------------------------------------------------------------------|
| `GATHER`   | `gathering`      | `{cell, sun}`                | One event per non-spooky tree. `sun` equals the tree's size (`1`/`2`/`3`). Ordered by cell id ascending.       |
| `GROW`     | `actions`        | `{cell}`                     | Successful `GROW` action: the tree on `cell` advanced one size.                                                |
| `SEED`     | `actions`        | `{source, target}`           | Successful `SEED` action. If both players seed the same `target` on the same sub-frame, both `SEED` events are still emitted but the seeds are not planted — see `state.seedConflictCell`. |
| `COMPLETE` | `actions`        | `{cell, points}`             | Successful `COMPLETE` action. `points` = `nutrients + richness bonus` at the moment of the cut (richness bonus: `+0`/`+2`/`+4` for richness `1`/`2`/`3`). |
| `WAIT`     | `actions`        | none                         | Player ended the day asleep. No `data`.                                                                         |
| `DEBUG`    | `actions`        | `{value}`                    | Free-text appended to the action command (e.g. `WAIT GL HF` → `"GL HF"`). Emitted **after** the action's own event in the same frame. |

### What is **not** a per-player event

Two pieces of information that the upstream `GameSummaryManager` reports
as standalone summary lines are not emitted as `traces[]` events because
they are not per-player. Each is recoverable from the per-turn `state`:

- **`SUN_MOVE`** — derivable from `phase == "sun"` plus the next turn's
  `sunDirection` (or simply `day % 6`, except on the final round which
  skips the rotation).
- **`SEED_CONFLICT`** — promoted to `state.seedConflictCell` rather than
  duplicating the same fact into both players' event slots.

## Example

A trimmed self-play trace covering one full day (gathering → first
action → sun rotation). Top-level envelope is shown for context; turn
structure is the spring2021-specific part.

```json
{
  "createdAt": "2026-05-08T12:00:00Z",
  "type": "trace",
  "puzzleId": 730,
  "puzzleName": "spring2021",
  "traceId": 1715170800,
  "matchId": 0,
  "players": ["bot-a", "bot-b"],
  "blue": "bot-a",
  "league": 4,
  "seed": "1234567890",
  "endReason": "TURNS_OUT",
  "scores": [120.0, 118.0],
  "finalScores": [123.0, 121.0],
  "ranks": [0, 1],
  "setup": [
    "37",
    "0 3 1 2 3 4 5 6",
    "1 3 7 8 2 0 6 18",
    "..."
  ],
  "mainTurns": 85,
  "turns": [
    {
      "turn": 0,
      "isOutputTurn": [false, false],
      "score": [0, 0],
      "traces": [
        [
          { "type": "GATHER", "data": { "cell": 21, "sun": 1 } },
          { "type": "GATHER", "data": { "cell": 24, "sun": 1 } }
        ],
        [
          { "type": "GATHER", "data": { "cell": 30, "sun": 1 } },
          { "type": "GATHER", "data": { "cell": 33, "sun": 1 } }
        ]
      ],
      "state": {
        "day": 0,
        "phase": "gathering",
        "sunDirection": 0,
        "nutrients": 20,
        "sun": [0, 0],
        "trees": [
          [[21, 3, 1, 0], [24, 1, 1, 0]],
          [[30, 1, 1, 0], [33, 3, 1, 0]]
        ]
      }
    },
    {
      "turn": 1,
      "gameInput": ["0", "20", "2 0", "2 0 0", "4", "21 1 1 0", "24 1 1 0", "30 1 0 0", "33 1 0 0", "..."],
      "output": ["GROW 21", "GROW 30 GL HF"],
      "isOutputTurn": [true, true],
      "score": [0, 0],
      "traces": [
        [{ "type": "GROW", "data": { "cell": 21 } }],
        [
          { "type": "GROW", "data": { "cell": 30 } },
          { "type": "DEBUG", "data": { "value": "GL HF" } }
        ]
      ],
      "state": {
        "day": 0,
        "phase": "actions",
        "dayActionIndex": 0,
        "sunDirection": 0,
        "nutrients": 20,
        "sun": [1, 1],
        "trees": [
          [[21, 3, 2, 1], [24, 1, 1, 0]],
          [[30, 1, 2, 1], [33, 3, 1, 0]]
        ]
      }
    },
    {
      "turn": 2,
      "isOutputTurn": [false, false],
      "score": [0, 0],
      "traces": [[], []],
      "state": {
        "day": 0,
        "phase": "sun",
        "sunDirection": 0,
        "nutrients": 20,
        "sun": [1, 1],
        "trees": [
          [[21, 3, 2, 0], [24, 1, 1, 0]],
          [[30, 1, 2, 0], [33, 3, 1, 0]]
        ]
      }
    }
  ]
}
```

Things worth pointing out in the example:

- Turn 0 is the day-0 `gathering` frame. Both sides start with two
  size-1 trees on the edge; each contributes `sun: 1`.
- Turn 1 is the first `actions` sub-frame (`dayActionIndex: 0`). After
  the action, the grown trees are flagged dormant (`isDormant: 1` in
  `state.trees`) and stay that way for the rest of the day. Player 1
  appended `GL HF` to their `GROW` command, producing a `DEBUG` event
  immediately after their `GROW` event.
- Turn 2 is the day-0 `sun` frame. No player events; dormancy resets
  for all trees as the day ends (the next day's `actions` frames will
  see `isDormant: 0`). `sunDirection` stamps the orientation
  *before* this frame's rotation; the next turn's `state.sunDirection`
  will be `1`.
