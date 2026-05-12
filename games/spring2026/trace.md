# Spring Challenge 2026 — Trace format

This document describes the spring2026-specific parts of the arena trace
format: the match-level `setup` payload, the per-turn `gameInput` /
`state` payload, and the per-player events emitted into
`turns[].traces`. The cross-game envelope (top-level fields, `turns[]`
shape, file naming, JSON conventions) is documented in
[docs/trace.md](../../docs/trace.md).

## Identity

| Field        | Value                                |
|--------------|--------------------------------------|
| `puzzleId`   | `0` (community contest, no puzzleId) |
| `puzzleName` | `spring2026`                         |
| Turn model   | `FlatTurnModel`                      |

`FlatTurnModel` reflects that the spring2026 referee runs one
`PerformGameUpdate` per loop turn and ends the match on the same turn
the stall / turn-cap condition fires — there is no engine-only phase
frame and no post-end frame. Every recorded turn is a player-decision
turn, so `mainTurns` always equals `len(turns)` (capped at `300` for
leagues 3+, `100` for leagues 1–2 via `MainTurnsForLeague`).

## Match-level `setup` lines

`setup` is the raw `string[]` from `Board.GetInitialInputs(0)`
([engine/engine_board.go](engine/engine_board.go)) — the same shape
either bot receives on stdin at match start. The trace records side
0's perspective (so the shack-owner digit on the side-0 shack is `0`
and the side-1 shack is `1`).

```
<width> <height>                                # e.g. "16 8"
<row 0>                                         # height rows of width chars
<row 1>
...
```

| Position    | Type   | Description                                                                                                                |
|-------------|--------|----------------------------------------------------------------------------------------------------------------------------|
| `width`     | int    | Grid width in cells. `2 * height`. League-dependent (`16..22`).                                                            |
| `height`    | int    | Grid height in cells. League-dependent (`8..11`; pinned to `MAP_MIN_HEIGHT == 8` for leagues 1–2).                         |
| row chars   | char   | `'.'` = grass (walkable, plantable), `'~'` = water, `'#'` = rock, `'+'` = iron (league 3+), `'0'`/`'1'` = shack owner.     |

The grid is side-symmetric: cell `(x, y)` mirrors to
`(width-1-x, height-1-y)`. Plant placement, terrain (rocks, iron),
rivers (league 3+), and starting inventory all mirror across this
axis, so both bots see geometrically equivalent maps regardless of
which side they play.

## Frame model

Each loop turn maps 1:1 to a Java game turn:

1. The runner reads each side's command line (one or more
   semicolon-separated tokens — `MOVE`, `HARVEST`, `PLANT`, `PICK`,
   `DROP`, `TRAIN`, `CHOP`, `MINE`, `WAIT`, `MSG <text>`).
2. The engine parses commands into tasks
   ([engine/engine_task_task_manager.go](engine/engine_task_task_manager.go))
   and `Board.Tick` drains them by priority bucket
   (`MOVE(1) → HARVEST(2) → PLANT(3) → CHOP(4) → PICK(5) → TRAIN(6) → DROP(7) → MINE(8)`),
   then ticks all surviving plants and recomputes scores.
3. If the match has stalled (`Board.HasStalled`) or the turn cap is
   reached, the engine flags `ended = true` and the runner stops on
   the same turn. There is no separate post-end frame.

The trace buffer is cleared at the top of `Referee.ResetGameTurnData`
so `MSG` and `WAIT` traces emitted during command parsing live in the
same turn's bucket as the task-applied events.

## Per-turn `gameInput` payload

`turns[].gameInput` is the raw `string[]` from
`Board.GetTurnInputs(0)` ([engine/engine_board.go](engine/engine_board.go))
— the same lines side 0's bot received on stdin during the live match.
Spring 2026 has no fog of war, but the input is recipient-relative:
side 0's inventory is listed first, then side 1's, and the per-troll
`outputId` field reorders the owner-side digit so each bot sees its
own trolls as `0`. The trace's `gameInput` therefore reflects side 0's
perspective: `outputId == 0` rows are side 0's trolls.

```
<my_PLUM> <my_LEMON> <my_APPLE> <my_BANANA> <my_IRON> <my_WOOD>    # side 0's shack inventory
<opp_PLUM> <opp_LEMON> <opp_APPLE> <opp_BANANA> <opp_IRON> <opp_WOOD>
<plantsCount>
<TYPE> <x> <y> <size> <health> <resources> <cooldown>              # plantsCount lines
...
<unitsCount>
<id> <ownerId> <x> <y> <moveSpeed> <carryCapacity> <harvestPower> <chopPower> <carry_PLUM> <carry_LEMON> <carry_APPLE> <carry_BANANA> <carry_IRON> <carry_WOOD>
...
```

`gameInput` is sampled **after** command parsing and **before**
`Board.Tick` — inventory, plant cooldowns, and troll positions reflect
the state players saw when choosing actions for this turn. Plants
killed earlier this match are absent. Trained-this-turn trolls are
absent (the `TRAIN` task spawns the unit during `Tick`).

## Per-turn `state` payload

`turns[].state` is a `TraceTurnState`
([engine/traces.go](engine/traces.go)), sampled at the same point as
`gameInput`. The shape is a typed shortcut over the inputs — analyzers
that need shack inventories, troll attributes, or live plants don't
have to re-parse the space-separated strings.

| Field         | Type                  | Description                                                                                                              |
|---------------|-----------------------|--------------------------------------------------------------------------------------------------------------------------|
| `turn`        | int                   | Loop turn number, starts at `1` (per the arena runner's match loop).                                                     |
| `inventories` | `[2][6]int`           | Per-side shack inventory, indexed by Item ordinal `[PLUM, LEMON, APPLE, BANANA, IRON, WOOD]`. Outer index = match side.   |
| `units`       | `[2][]TraceUnit`      | Per-side troll roster (alive trolls only — trolls never die in spring 2026, so this matches the all-time roster).         |
| `plants`      | `[]TracePlant`        | All live plants on the board, in `Board.Plants` insertion order. Omitted when no plants remain.                          |

### `units[side][i]` entry

| Field           | Type    | Description                                                                                                  |
|-----------------|---------|--------------------------------------------------------------------------------------------------------------|
| `id`            | int     | Troll id (matches `unit` on event payloads and the `id` field on the corresponding `gameInput` troll line).   |
| `pos`           | `[2]int`| Troll cell `[x, y]`. Matches the `gameInput` troll line's `x y` fields.                                       |
| `moveSpeed`     | int     | Cells the troll can step per `MOVE` (set at training time).                                                   |
| `carryCapacity` | int     | Maximum total items the troll can carry.                                                                      |
| `harvestPower`  | int     | Maximum plant resource tier the troll can harvest (`0` ⇒ cannot harvest).                                     |
| `chopPower`     | int     | Damage per `CHOP`; also the iron rate per `MINE` (`0` ⇒ cannot chop / mine). League 3+ feature.               |
| `carry`         | `[6]int`| Items the troll is currently carrying, indexed by Item ordinal `[PLUM, LEMON, APPLE, BANANA, IRON, WOOD]`.    |

### `plants[i]` entry

| Field       | Type    | Description                                                                                                  |
|-------------|---------|--------------------------------------------------------------------------------------------------------------|
| `type`      | string  | Plant kind: `"PLUM"`, `"LEMON"`, `"APPLE"`, or `"BANANA"`.                                                   |
| `pos`       | `[2]int`| Plant cell `[x, y]`.                                                                                          |
| `size`      | int     | Growth stage, `0..PLANT_MAX_SIZE`.                                                                            |
| `health`    | int     | Current health. Drops on `CHOP`; reaches `0` ⇒ plant is dead and removed from the board.                      |
| `resources` | int     | Available fruits to harvest, `0..PLANT_MAX_RESOURCES`. Each `HARVEST` consumes one.                          |
| `cooldown`  | int     | Ticks remaining before the plant grows / yields again. `0` ⇒ ready next `Tick`.                              |

## Per-player events (`turns[].traces[playerIdx]`)

Events are bucketed by owner; index into `traces[]` is the match side
(same indexing as `players`, `inventories`, and `units`). Cross-owner
events don't exist — every action is attributed to the issuing side.

Each turn's trace list interleaves two emission points:

1. **Parse-time** — `MSG` and `WAIT` traces, emitted during
   `TaskManager.ParseTasks` in the bot's command-token order.
2. **Apply-time** — task event traces, emitted during `Board.Tick` in
   priority order (`MOVE → HARVEST → PLANT → CHOP → PICK → TRAIN → DROP → MINE`).
   The order within a priority bucket follows the engine's
   per-task application loop (`groupByCell` / per-task linear pass).

All event structs live in [engine/traces.go](engine/traces.go).

| `type`     | Emitted on            | `data`                                              | Notes                                                                                                                                                                                                                              |
|------------|-----------------------|-----------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `MSG`      | `MSG <text>` parsing  | `{text}`                                            | Bot's verbatim message after `Player.SetMessage` trims it to 50 chars. Emitted at parse time before any task event in the same turn. Multiple `MSG` tokens overwrite `Player.Message` but each still emits its own trace.          |
| `WAIT`     | `WAIT` parsing        | _(no `data`)_                                       | The bot sent a bare `WAIT` token. Spring 2026's `WAIT` is "do nothing this turn" — no troll id, no message group. Multiple `WAIT` tokens each emit their own trace.                                                                |
| `MOVE`     | `MOVE` task applied   | `{unit, to}`                                        | The troll stepped one `moveSpeed`-bounded step toward the bot's `MOVE id x y` target. `to` is the cell the troll ended on this turn. Soft-blocked moves (target cell occupied, cycle unresolvable) drop the trace and surface as a `MOVE_BLOCKED` error in the summary tape. |
| `HARVEST`  | `HARVEST` task applied| `{unit, cell, type, amount}`                        | One trace per troll that actually picked fruit. Multiple trolls on the same cell can share a plant in one `HARVEST` task bucket; each emits its own trace with the per-troll `amount`. Zero-amount harvests (plant ran dry mid-bucket) drop the trace. |
| `PLANT`    | `PLANT` task applied  | `{unit, cell, type}`                                | The troll consumed one seed from its carry inventory and planted on its current grass cell. Conflicting same-cell plants between the two sides drop the trace and surface an `OPPONENT_BLOCKING` error.                          |
| `CHOP`     | `CHOP` task applied   | `{unit, cell, damage, wood, killed?}`               | League 3+. `damage` is the troll's `chopPower` (the damage dealt this turn); `wood` is the wood added to the troll's carry as the plant fell (`0` unless this chop felled the tree). `killed` is `true` when the plant ended dead. Same-cell chops emit one trace per troll. |
| `PICK`     | `PICK` task applied   | `{unit, type}`                                      | The troll pulled one item from the player's shack. Always one item per `PICK`; emit `PICK` four times for four items.                                                                                                              |
| `TRAIN`    | `TRAIN` task applied  | `{unit, talents}`                                   | A new troll spawned on the player's shack. `unit` is the freshly assigned troll id; `talents` is `[moveSpeed, carryCapacity, harvestPower, chopPower]` — the bot's requested attribute vector.                                     |
| `DROP`     | `DROP` task applied   | `{unit, items}`                                     | The troll emptied its carry into the player's shack. `items[i]` is the count of each item handed over, indexed by Item ordinal `[PLUM, LEMON, APPLE, BANANA, IRON, WOOD]`.                                                       |
| `MINE`     | `MINE` task applied   | `{unit, cell, iron}`                                | League 3+. The troll mined `iron` units of iron from a neighboring iron cell. `cell` is the troll's current grass cell (adjacent to iron); `iron` is bounded by `chopPower` and the troll's free carry capacity.                |

### Skipped commands and rejected tasks

Commands that fail parsing (unknown command, out-of-board target, no
plant at the troll's cell, no capacity, opponent contradicts, …) emit
**no trace** — they surface only as error lines in the summary tape
(`P<idx>: [failed] <message>`). The trace records what actually
resolved into game state, not what the bot tried to do. The raw
output line is still preserved in `turns[].output[i]` so analyzers
can correlate intent with outcome.

A bot can issue at most one task per troll per turn; the second task
for the same troll is rejected with an `ALREADY_USED` error and
drops its trace. `WAIT` and `MSG` are global tokens (not per-troll)
and have no `ALREADY_USED` rule.

### Inline `debug` field

Spring 2026's protocol does not have a per-action debug ride-along.
Free-text belongs on the `MSG <text>` token, which emits its own
`MSG` trace and persists in `Player.Message` for one turn (the
referee mirrors it to the summary tape via the message prefix on
errors). Analyzers reconstructing per-troll chatter should read the
`MSG` trace immediately preceding the action of interest.

### What is **not** a per-player trace

A few engine-internal happenings are intentionally not emitted as
`traces[]` entries because they're already available through other
trace fields or are not per-action:

- **Plant ticks (growth and yield)** — recoverable from
  `state.plants` deltas across consecutive turns.
- **Plant deaths** — derivable from a missing entry in the next
  turn's `state.plants` plus a `CHOP` event with `killed: true`.
- **Score deltas** — `turns[].score` already carries the per-turn raw
  score (`PLUM + LEMON + APPLE + BANANA + 4·WOOD`) sampled before
  `Board.Tick`. The change between consecutive turns is the net
  award for that turn.
- **Stall / game-over** — surfaced through `endReason`
  (`SCORE_EARLY` for `Board.HasStalled` ending the match early,
  `SCORE` for hitting the turn cap, `TIMEOUT` / `TIMEOUT_START` /
  `INVALID` for player faults).

## League rules and trace shape

The `league` option affects map generation and which commands are
available:

- League 1: 100 turns, fixed map height `8`, no starting inventory,
  no rivers / iron / rocks, only `MOVE` / `HARVEST` / `DROP` / `WAIT`.
- League 2: 100 turns, fixed map height `8`, starting inventory,
  adds `PLANT` / `PICK` / `TRAIN`.
- League 3+: 300 turns, variable map height `8..11`, rivers, iron,
  rocks, and `CHOP` / `MINE` enabled; trolls can be trained with
  `chopPower`.

Trace shape is identical across leagues — only the event vocabulary
that can appear differs (no `CHOP` / `MINE` / `TRAIN` events in
league 1, etc.). Analyzers can ignore `league` for any non-vocab
analysis.

## Example

A trimmed self-play trace covering one full turn (parse-time `MSG` +
`WAIT`, then one `MOVE`, one `HARVEST`, one `TRAIN`). Top-level
envelope is shown for context; turn structure is the spring2026-
specific part.

```json
{
  "createdAt": "2026-05-12T12:00:00Z",
  "type": "trace",
  "puzzleName": "spring2026",
  "traceId": 1715170800,
  "matchId": 0,
  "players": ["bot-a", "bot-b"],
  "blue": "bot-a",
  "league": 4,
  "seed": "1234567890",
  "endReason": "SCORE",
  "scores": [42.0, 38.0],
  "finalScores": [42.0, 38.0],
  "ranks": [0, 1],
  "setup": [
    "16 8",
    "................",
    "...0............",
    "................",
    "...~~...........",
    "...........~~...",
    "............1...",
    "................",
    "................"
  ],
  "mainTurns": 300,
  "turns": [
    {
      "turn": 5,
      "gameInput": [
        "3 2 4 1 0 0",
        "2 5 1 3 0 1",
        "3",
        "PLUM 7 2 2 8 1 4",
        "APPLE 11 5 1 12 0 3",
        "BANANA 4 3 3 4 0 2",
        "2",
        "0 0 3 1 1 1 1 0 0 0 0 1 0 0",
        "1 1 12 5 1 1 1 0 0 1 0 0 0 0"
      ],
      "output": [
        "MSG harvest plum;MOVE 0 7 2",
        "HARVEST 1;TRAIN 1 1 1 0"
      ],
      "isOutputTurn": [true, true],
      "score": [10, 8],
      "traces": [
        [
          { "type": "MSG", "data": { "text": "harvest plum" } },
          { "type": "MOVE", "data": { "unit": 0, "to": [4, 1] } }
        ],
        [
          { "type": "HARVEST", "data": { "unit": 1, "cell": [11, 5], "type": "APPLE", "amount": 1 } },
          { "type": "TRAIN",   "data": { "unit": 2, "talents": [1, 1, 1, 0] } }
        ]
      ],
      "state": {
        "turn": 5,
        "inventories": [
          [3, 2, 4, 1, 0, 0],
          [2, 5, 1, 3, 0, 1]
        ],
        "units": [
          [
            { "id": 0, "pos": [3, 1], "moveSpeed": 1, "carryCapacity": 1, "harvestPower": 1, "chopPower": 1,
              "carry": [0, 0, 0, 0, 1, 0] }
          ],
          [
            { "id": 1, "pos": [12, 5], "moveSpeed": 1, "carryCapacity": 1, "harvestPower": 1, "chopPower": 1,
              "carry": [0, 0, 1, 0, 0, 0] }
          ]
        ],
        "plants": [
          { "type": "PLUM",   "pos": [7, 2],  "size": 2, "health": 8,  "resources": 1, "cooldown": 4 },
          { "type": "APPLE",  "pos": [11, 5], "size": 1, "health": 12, "resources": 0, "cooldown": 3 },
          { "type": "BANANA", "pos": [4, 3],  "size": 3, "health": 4,  "resources": 0, "cooldown": 2 }
        ]
      }
    }
  ]
}
```

Things worth pointing out in the example:

- Side 0's bot sent two semicolon-separated tokens: `MSG harvest
  plum` and `MOVE 0 7 2`. The `MSG` trace lands first (parse-time,
  before any task is applied); then the `MOVE` trace lands during
  `Board.Tick` after the move task resolved. `to: [4, 1]` is the
  cell troll `0` stepped to — one cell closer to the requested
  `(7, 2)` target. The bot's requested `(7, 2)` is collapsed into
  the resolved next-step at parse time, so the trace records what
  actually happened, not the multi-turn intent.
- Side 1's bot sent two task tokens (no `MSG`, no `WAIT`). Trace
  order follows priority: `HARVEST` (priority 2) before `TRAIN`
  (priority 6). The `TRAIN` trace's `unit: 2` is the freshly
  assigned id for the new troll, which spawned on side 1's shack
  but is not present in the same turn's `state.units` (state is
  sampled before `Tick`).
- `state.plants` shows three live plants in `Board.Plants` insertion
  order; the apple's `resources: 0` reflects the post-harvest count
  but is sampled before `Tick` and so still appears with its prior
  cooldown. The next turn's state will show `resources: 0,
  cooldown: 3 → 2` after the plant ticks.
- `score: [10, 8]` is the per-turn raw score sampled before `Tick`,
  computed as `PLUM + LEMON + APPLE + BANANA + 4·WOOD` from the
  player's shack inventory. The final `scores: [42, 38]` are the
  end-of-match totals after 300 turns.
