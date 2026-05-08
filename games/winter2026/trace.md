# Winter Challenge 2026 — Trace format

This document describes the winter2026-specific parts of the arena trace
format: the match-level `setup` payload, the per-turn `gameInput`
payload, and the per-player events emitted into `turns[].traces`. The
cross-game envelope (top-level fields, `turns[]` shape, file naming,
JSON conventions) is documented in [docs/trace.md](../../docs/trace.md).

## Identity

| Field        | Value                                |
|--------------|--------------------------------------|
| `puzzleId`   | `13771`                              |
| `puzzleName` | `winter2026`                         |
| Turn model   | `FlatTurnModel`                      |

`FlatTurnModel` reflects that winter2026's referee runs one
`PerformGameUpdate` per loop turn and ends the match on the same turn
game-over is detected — there is no engine-only phase frame and no
post-end frame. Every recorded turn is a player-decision turn, so
`mainTurns` always equals `len(turns)` (capped at 200 by the upstream
`gameManager.setMaxTurns(200)`).

## Match-level `setup` lines

`setup` is the raw `string[]` from `SerializeGlobalInfoFor`
([engine/game_serializer.go](engine/game_serializer.go)) — same shape
both bots receive on stdin. The trace records side 0's perspective; the
grid contains only walls and free cells and is side-symmetric, so the
geometry rows are identical regardless of perspective. The
`myId` header and the snakebot-id ordering, however, **are**
side-specific: the trace's `setup` has `myId == 0` and lists side 0's
snakebot ids first, then side 1's.

Line layout, per the upstream Java engine:

```
<myId>                                           # 0 — recorded from side 0's perspective
<width>                                          # e.g. 44
<height>                                         # e.g. 24
<row 0>                                          # height rows of '#' (wall) and '.' (empty)
<row 1>
...
<snakebotsPerPlayer>                             # e.g. 4
<myBird0Id>                                      # snakebotsPerPlayer entries, side 0's ids ascending
...
<oppBird0Id>                                     # snakebotsPerPlayer entries, side 1's ids ascending
...
```

| Position             | Type | Description                                                                                          |
|----------------------|------|------------------------------------------------------------------------------------------------------|
| `myId`               | int  | Reference player index. Always `0` in the trace (recorded from side 0's perspective).                |
| `width`              | int  | Grid width in cells. League-dependent through the grid generator (`15..45`).                          |
| `height`             | int  | Grid height in cells. League-dependent (`10..30`).                                                    |
| row chars            | char | `'#'` = wall (impassable / solid for gravity), `'.'` = empty (walkable, eligible for apples).         |
| `snakebotsPerPlayer` | int  | Snakebots each player controls; both sides always have the same count.                               |
| my bird ids          | int  | Side 0's snakebot ids, ascending. Side 0 always owns ids `0..snakebotsPerPlayer-1`.                   |
| opp bird ids         | int  | Side 1's snakebot ids, ascending. Side 1 always owns ids `snakebotsPerPlayer..2·snakebotsPerPlayer-1`.|

The grid does **not** wrap on either axis; cells outside the grid are
treated as walls for movement (heads stepping off the side are
beheaded) and as open air for gravity (a bird supported only by
out-of-bounds cells falls). The `setup` block captures only static
geometry; apple and snakebot placement is part of `gameInput` on every
turn.

### Bird id convention

Unlike spring2020's interleaved global pac IDs, winter2026 assigns
snakebot ids **block-by-player**: side 0's birds are `[0,
snakebotsPerPlayer)`, side 1's birds are `[snakebotsPerPlayer,
2·snakebotsPerPlayer)`. The same numbering appears in
`traces[].data.bird` — analyzers can recover the owning side directly
from the id (`id < snakebotsPerPlayer ? 0 : 1`), so `traces[]` slot
attribution and bird-id derivation never disagree.

## Frame model

Each loop turn maps 1:1 to a Java game turn:

1. The runner reads each side's command line (one snakebot or `MARK`
   command per `;`-separated token, up to 30 tokens per side).
2. The engine runs `PerformGameUpdate`:
   `DoMoves` → `DoEats` → `DoBeheadings` → `DoFalls`
   ([engine/game_game.go](engine/game_game.go)).
3. If `IsGameOver` returns true (no apples remaining or any player has
   no live snakebots / has been deactivated), the match ends on the
   same turn. There is no separate post-end frame.

A bird with no command for the turn keeps moving in its current facing
(`Bird.Facing()`, derived from `body[0] - body[1]`). The starting
direction is **up** for every freshly spawned bird (spawn body lays
out top-down so head − second segment = `(0,-1)`).

The trace turn accumulates two kinds of entries:

- **Command traces** (`MOVE`, `WAIT`, `MARK`) emitted by the
  `CommandManager` during step 1, one per accepted command token in
  the bot's output line. They lead the turn's `traces[]` array in
  bot-input order.
- **Event traces** (`EAT`, `HIT_*`, `DEAD`, `DEAD_FALL`) emitted by
  the `DoEats` / `DoBeheadings` / `DoFalls` phases of step 2. They
  follow the command traces, in engine-phase order.

The trace buffer is cleared at the top of `ResetGameTurnData` (turn
start) so command and event traces from the same turn share the
slice.

## Per-turn `gameInput` payload

`turns[].gameInput` is the raw `string[]` from `SerializeFrameInfoFor`
([engine/game_serializer.go](engine/game_serializer.go)) — the same
lines either bot's stdin received during the live match (winter2026
has no fog of war, so side 0's and side 1's serializations are
identical for the apple list and bird-body block). The trace records
side 0's perspective.

`gameInput` is sampled **after** command parsing and **before**
`PerformGameUpdate`. Apple positions and bird bodies reflect the state
players saw when choosing actions for this turn; movement and
collision resolution from this turn have not yet applied.

```
<applesCount>                                                      # remaining apples on the grid
<x0> <y0>                                                          # one per apple
... (applesCount lines)
<liveBirdsCount>                                                   # alive snakebots only
<id0> <x,y:x,y:x,y...>                                             # one per live bird, head first
... (liveBirdsCount lines)
```

### Apple lines

| Position | Field | Description                                                                                       |
|----------|-------|---------------------------------------------------------------------------------------------------|
| 0..1     | `x y` | Cell coordinates of one remaining apple. Apples cleared on the previous turn are absent here.     |

Apples appear in `Grid.Apples` insertion order. There is no separate
apple-count metric beyond `applesCount`; the `apples_remaining` match
metric ([engine/game_referee.go](engine/game_referee.go)) is the same
count surfaced through `Referee.Metrics()`.

### Bird lines

`<id> <body>` where `<body>` is colon-separated `<x>,<y>` cells, head
first.

| Position | Field    | Description                                                                                                |
|----------|----------|------------------------------------------------------------------------------------------------------------|
| 0        | `id`     | Snakebot id (matches the `bird` field on event payloads, same scheme as `setup`).                          |
| 1        | `body`   | `head_x,head_y:next_x,next_y:...` — body length ≥ 1; the engine prepends the new head each turn.           |

Dead snakebots are **dropped** from the bird block — `liveBirdsCount`
counts only alive birds. Once a bird is removed (DEAD beheading or
DEAD_FALL), it never reappears in subsequent `gameInput` payloads.

To recover a bird's facing from `gameInput`, compute
`(body[0] - body[1])` and map back to `UP/DOWN/LEFT/RIGHT`. Length-1
bodies (cannot occur during normal play — a bird with body ≤ 3 dies
on a beheading) have no defined facing.

## Per-turn `state` payload

`turns[].state` is a `TraceTurnState` ([engine/traces.go](engine/traces.go)),
sampled by the arena runner **after** command parsing and **before**
`PerformGameUpdate`. So values reflect the state players saw when
choosing actions: any snake removed on the previous turn is already
absent, but heads/sizes from this turn's moves have not been applied
yet. The shape is intentionally a typed shortcut over `gameInput` —
analyzers that only need head and size don't have to re-parse the
colon-separated body string on every turn.

| Field    | Type             | Description                                                                                                    |
|----------|------------------|----------------------------------------------------------------------------------------------------------------|
| `apples` | int              | Count of remaining power sources on the grid going into this turn. Equal to `len(gameInput[0])` apples.        |
| `snakes` | `[2][]TraceSnake`| Per-side roster of **alive** snakebots. Outer index = match side (`players[i]`); inner list ordered by `id` ascending. |

### `snakes[side][i]` entry

| Field  | Type     | Description                                                                                                  |
|--------|----------|--------------------------------------------------------------------------------------------------------------|
| `id`   | int      | Snakebot id (same scheme as `setup` / `gameInput` / event payloads). `id < snakebotsPerPlayer` ⇒ side 0.      |
| `size` | int      | Body length in cells, including the head. Spawn size is 3; grows by 1 per `EAT`; drops by 1 per non-fatal `HIT_*`. |
| `head` | `[2]int` | Head cell `[x, y]`. Equal to `body[0]` in the matching `gameInput` bird line.                                |

Dead snakebots are dropped from the roster, mirroring the gameInput
live-birds block. Once a snake is removed by `DEAD` or `DEAD_FALL` it
never reappears in subsequent `state.snakes`. Reconstructing the full
body or facing direction still requires the matching `gameInput` line —
the typed state intentionally exposes only the fields most analyzers
actually need (apple budget, segment counts, head positions for
distance/proximity queries).

## Per-player events (`turns[].traces[playerIdx]`)

Events are bucketed by owner; index into `traces[]` is the match side
(same indexing as `players` / score arrays). All structs live in
[engine/traces.go](engine/traces.go).

Event payloads identify snakebots by their `bird` id (the same id
shown in `setup`, `gameInput`, and the bot's stdin). Side 0 owns ids
`[0, snakebotsPerPlayer)`; side 1 owns ids `[snakebotsPerPlayer,
2·snakebotsPerPlayer)`.

Each turn's trace list is split into two buckets, in this order:

1. **Commands** — what the bot was asked to do this turn (one per
   accepted token in the bot's output line, leading the slot in
   command-issue order). Carry the optional `debug` field on `MOVE`.
2. **Events** — engine-derived consequences of movement and gravity
   resolution (variable count per turn). Do **not** carry `debug`.

### Commands

One trace per **accepted** command token in the bot's `output[i]` line,
in input order. A token rejected during parsing emits no trace and
appears only as a summary error:

- `MOVE` for the bird whose id matched, alive, not already moved this
  turn, and not requesting a reverse direction. Soft-rejected moves
  (unknown bird id, dead bird, double-move on the same bird,
  backwards) drop their trace and any `debug` along with it.
- `MARK` for the first four `MARK x y` per side per turn. The 5th+
  is rejected by `Player.AddMark` and emits no trace.
- `WAIT` for every `WAIT` token (no rejection conditions).

Commands that fail the top-level regex (`DANCE`, malformed tokens)
deactivate the player on the spot, so they never produce a trace
either.

| `type`  | `data`                                 | Notes                                                                                                                                                                                                                                                                                                                                  |
|---------|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `MOVE`  | `{bird, direction, debug?}`            | The bot sent `<bird> UP/DOWN/LEFT/RIGHT [debug]`. `direction` is the bot's wire token (`"UP"` / `"DOWN"` / `"LEFT"` / `"RIGHT"`), not the engine's NESW alias, so the trace mirrors what the bot actually wrote. `debug` is the optional trailing free-text (truncated to 48 chars by `Bird.SetMessage`). Field is omitted when the bot sent a bare directional command. |
| `WAIT`  | _(no `data`)_                          | The bot sent a `WAIT` token. Winter 2026's `WAIT` is global ("do nothing") — it has no bird id and no message group in the regex, so the trace carries the type alone. Multiple `WAIT` tokens in one output line each emit their own trace.                                                                                              |
| `MARK`  | `{coord}`                              | The bot sent `MARK x y`. `coord` is `[x, y]`. Not associated with any specific bird; markers are pure viewer-side debugging affordances. Cap of 4 per side per turn.                                                                                                                                                                    |

A bird that received no command this turn emits no command trace. It
keeps moving in its current facing inertially, and the next turn's
`gameInput` shows where it ended up. Distinguishing "explicit
`WAIT`" from "no-command-for-this-bird" is therefore a side-level
question (`WAIT` trace present in the slot) rather than a per-bird
question — winter2026 has no per-bird WAIT.

### Events

| `type`      | `data`                              | Notes                                                                                                                                                                                                                                                                          |
|-------------|-------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `EAT`       | `{bird, coord}`                     | One event per bird that ate an apple this turn. `coord` is `[x, y]` of the eaten cell (the bird's new head). When multiple snakebot heads collide on the same apple, **each** participating bird emits its own `EAT` event (the apple is consumed once, but every snake grows). |
| `HIT_WALL`  | `{bird, coord}`                     | The owning bird's head landed on a wall cell. Emitted into the owner's slot only. `coord` is the head cell after the move (the wall cell). Fires regardless of body length; whether the bird also dies is recorded by a same-turn `DEAD` (cause `WALL`).                       |
| `HIT_ENEMY` | `{bird, coord}`                     | The owning bird's head landed on a cell occupied by **an enemy snakebot's** body. **Attributed to the moving bird's owner only — not mirrored into the enemy's slot.** This matches winter2026's "active bird pays the segment cost" semantics: per-match `HIT_ENEMY` counts segments lost, not encounters had. |
| `HIT_SELF`  | `{bird, coord}`                     | The owning bird's head landed on **its own or a teammate's** body cell. Emitted into the owner's slot only.                                                                                                                                                                    |
| `DEAD`      | `{bird, cause}`                     | Bird removed by beheading (body length ≤ 3 at intersection time). `cause` records which intersection killed it: `"WALL"` / `"ENEMY"` / `"SELF"`, picked by **priority `WALL > ENEMY > SELF`** when multiple hazards coincide (the engine's intersection-check order). The dying bird's full segment count is added to `Losses[ownerSide]`.        |
| `DEAD_FALL` | `{bird, segments}`                  | Bird fell off the grid (every body cell ended at `y ≥ height + 1` after gravity resolution). `segments` = body length at fall time, equal to the segments lost to the fall. Variable cost — unlike `DEAD` beheadings (always remove a length-≤3 snake), a fall removes a snake of any length. |

Engine events are emitted in this order within the turn:

1. **`EAT`** (`DoEats`).
2. **`HIT_WALL` / `HIT_ENEMY` / `HIT_SELF`** (`DoBeheadings`, first
   pass). A bird can emit multiple `HIT_*` in the same turn if its
   head cell overlaps more than one hazard class (e.g.
   `HIT_ENEMY` + `HIT_SELF`).
3. **`DEAD`** (`DoBeheadings`, second pass). Only emitted for birds
   beheaded with body length ≤ 3; long birds beheaded but still
   ≥ 3 segments after losing the head emit `HIT_*` only.
4. **`DEAD_FALL`** (`DoFalls`).

Cross-owner events do not exist: winter2026 attributes every event to
the moving / owning bird's side. Per-match metric counts therefore
correspond to "segments my snakes lost to this hazard," not
"encounters my snakes had with this hazard."

### Multi-event semantics within a turn

- **Multiple `EAT` from the same bird** is impossible (a snakebot
  occupies a single head cell per turn).
- **Multiple `HIT_*` from the same bird** is possible: a head landing
  on a cell that is simultaneously an enemy body and a self-body cell
  emits both `HIT_ENEMY` and `HIT_SELF`. `HIT_WALL` precludes any
  body intersection in the same cell (a wall cell cannot also hold a
  body), so `HIT_WALL` is always solo.
- **`HIT_*` without `DEAD`** means the bird was long enough to
  survive (body length > 3 at intersection time): the head segment
  drops, the next segment becomes the new head, and the bird keeps
  playing. `Losses[ownerSide]` increments by 1 per non-fatal
  beheading.
- **`DEAD` without `HIT_*`** cannot occur: every `DEAD` event is
  preceded in the same turn by at least one `HIT_*` recording the
  hazard. (`DEAD_FALL` is the exception — falls bypass the
  `HIT_*` channel entirely.)
- **`DEAD_FALL` without `HIT_*`** is the normal shape — falls happen
  in `DoFalls`, after `DoBeheadings`, when a bird is left airborne
  (no body cell rests above a wall, apple, or grounded bird's body).

### Inline `debug` field (MOVE only)

Free-text appended to a movement command (`0 LEFT chasing apple`,
`3 UP go go go`) rides on the **`MOVE`** trace as `data.debug`. The
text is captured into `Bird.Message` by `CommandManager.ParseCommands`
(truncated to 48 chars per `Bird.SetMessage`) and cleared on the next
turn's `Player.Reset`, so a non-empty `debug` always reflects the
message sent on the same turn the trace was emitted.

The field is omitted from JSON when the bot sent a bare directional
command, so a present `data.debug` is unambiguously a bot-authored
chat string. `WAIT` traces never carry debug — winter2026's WAIT
regex has no message group. `MARK` likewise drops trailing text:
`MARK 0 0 hello` would fail the regex (`MARK x y` is the only
accepted form). Event traces (`EAT`, `HIT_*`, `DEAD`, `DEAD_FALL`)
deliberately **do not carry** the debug field — the chat string
belongs to the command, not to its downstream fallout.

Errored commands drop the message entirely: a soft-rejected `MOVE`
(unknown bird, dead bird, double-move, backwards) emits no trace, so
its message has no carrier. A hard parse failure deactivates the
player and breaks out of the parse loop before reaching subsequent
tokens, dropping their messages too.

### What is **not** a per-player trace

A few things visible in the live game are not emitted as `traces[]`
entries because they're either available through other trace fields
or not per-player:

- **The raw bot output line** is in `turns[].output[i]` verbatim —
  command traces decode it into typed entries, but the original
  string remains for debugging mismatches between intent and the
  parsed action.
- **Per-bird message strings beyond `MOVE`'s `debug`** — there are
  none. `Bird.Message` only ever populates from a successful
  movement command, so reconstructing it is just reading the latest
  `MOVE.data.debug` for that bird.
- **Marker positions / counts** — markers exist only for the viewer.
  `MARK` traces capture the placement but the `Player.Marks` slice
  itself is not serialized.
- **Per-side score / `Losses`** — `turns[].score` is the per-turn
  raw score sampled before `PerformGameUpdate`; the `Losses` counters
  are recoverable as the running sum of `HIT_*` plus `DEAD.segments`
  (3 per `DEAD`) plus `DEAD_FALL.segments`, and surface in
  `Referee.Metrics()` as `losses_p0` / `losses_p1`.

## League rules and trace shape

The `league` option only affects grid generation (skew on the
height-distribution `pow` and the resulting wall density:
bronze=2 → tight maps, silver=1, gold=0.8, legend=0.3 → larger,
sparser maps; see `GridMaker.Make`
[engine/game_grid_grid_maker.go](engine/game_grid_grid_maker.go)). The
game protocol, command set, event vocabulary, and trace shape are
identical across leagues — analyzers can ignore `league` for any
non-grid-shape analysis.

## Example

A trimmed self-play trace covering the first three turns and one
mid-game beheading. Bird ids `0..3` are side 0; ids `4..7` are side 1
(`snakebotsPerPlayer == 4`); apple list is elided after the first
turn. Side 0's bot tags every move with a debug index (`(0)..(3)`)
and drops a `MARK` after each move — typical instrumentation pattern.

```json
{
  "createdAt": "2026-05-08T12:00:00Z",
  "type": "trace",
  "puzzleId": 13771,
  "puzzleName": "winter2026",
  "traceId": 1715170800,
  "matchId": 0,
  "players": ["bot-a", "bot-b"],
  "blue": "bot-a",
  "league": 4,
  "seed": "1234567890",
  "endReason": "TURNS_OUT",
  "scores": [19.0, 31.0],
  "finalScores": [19.0, 31.0],
  "ranks": [1, 0],
  "setup": [
    "0",
    "44",
    "24",
    "............................................",
    ".#........................................#.",
    "...",
    "############################################",
    "4",
    "0",
    "1",
    "2",
    "3",
    "4",
    "5",
    "6",
    "7"
  ],
  "mainTurns": 200,
  "turns": [
    {
      "turn": 0,
      "gameInput": ["..."],
      "output": [
        "0 LEFT (0);MARK 0 10;1 LEFT (1);MARK 12 4;2 UP (2);MARK 19 15;3 RIGHT (3);MARK 32 6;",
        "6 UP;4 LEFT;5 LEFT;7 RIGHT"
      ],
      "isOutputTurn": [true, true],
      "score": [12, 12],
      "traces": [
        [
          { "type": "MOVE", "data": { "bird": 0, "direction": "LEFT",  "debug": "(0)" } },
          { "type": "MARK", "data": { "coord": [0, 10] } },
          { "type": "MOVE", "data": { "bird": 1, "direction": "LEFT",  "debug": "(1)" } },
          { "type": "MARK", "data": { "coord": [12, 4] } },
          { "type": "MOVE", "data": { "bird": 2, "direction": "UP",    "debug": "(2)" } },
          { "type": "MARK", "data": { "coord": [19, 15] } },
          { "type": "MOVE", "data": { "bird": 3, "direction": "RIGHT", "debug": "(3)" } },
          { "type": "MARK", "data": { "coord": [32, 6] } },
          { "type": "EAT",  "data": { "bird": 2, "coord": [19, 15] } }
        ],
        [
          { "type": "MOVE", "data": { "bird": 6, "direction": "UP" } },
          { "type": "MOVE", "data": { "bird": 4, "direction": "LEFT" } },
          { "type": "MOVE", "data": { "bird": 5, "direction": "LEFT" } },
          { "type": "MOVE", "data": { "bird": 7, "direction": "RIGHT" } },
          { "type": "EAT",  "data": { "bird": 6, "coord": [16, 15] } }
        ]
      ],
      "state": {
        "apples": 60,
        "snakes": [
          [
            { "id": 0, "size": 3, "head": [4, 8] },
            { "id": 1, "size": 3, "head": [15, 3] },
            { "id": 2, "size": 3, "head": [19, 16] },
            { "id": 3, "size": 3, "head": [29, 4] }
          ],
          [
            { "id": 4, "size": 3, "head": [31, 8] },
            { "id": 5, "size": 3, "head": [20, 3] },
            { "id": 6, "size": 3, "head": [16, 16] },
            { "id": 7, "size": 3, "head": [6, 4] }
          ]
        ]
      }
    },
    {
      "turn": 10,
      "gameInput": ["..."],
      "output": [
        "0 RIGHT;1 RIGHT;2 RIGHT;3 LEFT",
        "WAIT;4 UP;5 RIGHT;6 LEFT;7 LEFT"
      ],
      "isOutputTurn": [true, true],
      "score": [16, 14],
      "traces": [
        [
          { "type": "MOVE",      "data": { "bird": 0, "direction": "RIGHT" } },
          { "type": "MOVE",      "data": { "bird": 1, "direction": "RIGHT" } },
          { "type": "MOVE",      "data": { "bird": 2, "direction": "RIGHT" } },
          { "type": "MOVE",      "data": { "bird": 3, "direction": "LEFT"  } },
          { "type": "EAT",       "data": { "bird": 1, "coord": [41, 19] } },
          { "type": "HIT_ENEMY", "data": { "bird": 3, "coord": [32, 12] } },
          { "type": "DEAD",      "data": { "bird": 3, "cause": "ENEMY" } }
        ],
        [
          { "type": "WAIT" },
          { "type": "MOVE", "data": { "bird": 4, "direction": "UP" } },
          { "type": "MOVE", "data": { "bird": 5, "direction": "RIGHT" } },
          { "type": "MOVE", "data": { "bird": 6, "direction": "LEFT" } },
          { "type": "MOVE", "data": { "bird": 7, "direction": "LEFT" } },
          { "type": "EAT",  "data": { "bird": 5, "coord": [2, 19] } }
        ]
      ],
      "state": {
        "apples": 40,
        "snakes": [
          [
            { "id": 0, "size": 6, "head": [1, 15] },
            { "id": 1, "size": 7, "head": [11, 6] },
            { "id": 2, "size": 4, "head": [20, 17] },
            { "id": 3, "size": 6, "head": [35, 13] }
          ],
          [
            { "id": 4, "size": 7, "head": [26, 16] },
            { "id": 5, "size": 4, "head": [18, 15] },
            { "id": 6, "size": 4, "head": [15, 15] },
            { "id": 7, "size": 6, "head": [8, 14] }
          ]
        ]
      }
    }
  ]
}
```

Things worth pointing out in the example:

- The `setup` block ends with two blocks of bird ids (`0,1,2,3` then
  `4,5,6,7`) — winter2026 assigns ids by player block, not
  interleaved, so id `< 4` ⇒ side 0 and id `≥ 4` ⇒ side 1.
- Turn 0 starts with `score: [12, 12]` (4 birds × 3 segments per
  side). Side 0's bot tags every move with a debug index and drops
  a `MARK` between each move, so the trace lists alternating
  `MOVE`/`MARK` pairs in **bot-input order** (not bird-id order).
  Side 1 issues no MARKs and no debug — its `MOVE` traces lack the
  `debug` field. The single `EAT` event lands at the end of each
  slot, after all command traces, because engine events are emitted
  by `DoEats` after `CommandManager.ParseCommands` has finished.
- Turn 10 shows the full command-then-event ordering inside one slot:
  side 0's four `MOVE` commands lead, then `EAT` (`DoEats`), then
  bird `3` hit an enemy body (`HIT_ENEMY` from `DoBeheadings`) and
  was beheaded with body length ≤ 3 (`DEAD` with `cause: "ENEMY"`
  from the same phase). Side 1's bot opens its line with a `WAIT`
  token (rendered as `{"type": "WAIT"}` with no data field), then
  four moves, then the `EAT` from bird `5`; the `HIT_ENEMY` from
  bird `3` is attributed to side 0 only and is **not** mirrored into
  side 1's slot.
- The `state` block on each turn is sampled **before**
  `PerformGameUpdate`, so the snake heads / sizes match what the bots
  saw on stdin this turn — not where they're about to land. Turn 0's
  state shows the spawn (size 3 everywhere, 60 apples); turn 10's
  state shows growth across the early game (sizes 4–7) and 20
  apples consumed (60 → 40). Note that bird `3` is still in the
  `state` block at turn 10 — it's only **removed** from the next
  turn's state, since the `DEAD` event in the same turn fires
  during `PerformGameUpdate`, after the state is captured.
- The match in the example ran the full 200 turns
  (`endReason: "TURNS_OUT"`, `mainTurns: 200`); raw `scores` and
  `finalScores` match because winter2026's `OnEnd` only adjusts
  scores on a tie (subtracting `Losses[i]`), and the scores here
  differ.
