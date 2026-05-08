# Spring Challenge 2020 — Trace format

This document describes the spring2020-specific parts of the arena trace
format: the match-level `setup` payload, the per-turn `state` and
`gameInput` payloads, and the per-player traces emitted into
`turns[].traces`. The cross-game envelope (top-level fields, `turns[]`
shape, file naming, JSON conventions) is documented in
[docs/trace.md](../../docs/trace.md).

## Identity

| Field        | Value                              |
|--------------|------------------------------------|
| `puzzleId`   | `592`                              |
| `puzzleName` | `spring2020`                       |
| Turn model   | `PostEndTurnModel`                 |

`PostEndTurnModel` reflects that Spring 2020's referee emits a separate
post-game-over frame (Java's `gameOverFrame` branch). The runner skips
player polling on that final frame, so it appears as a no-output trailing
turn. `mainTurns` counts only the player-decision turns and excludes the
post-end frame.

There is no separate phase model — every emitted turn is implicitly the
"actions" frame, with zero or one engine-folded SPEED sub-step before
the next bot read (see [Frame model](#frame-model)).

## Match-level `setup` lines

`setup` is the raw `string[]` from `SerializeGlobalInfoFor`
([engine/serializer.go](engine/serializer.go)) — same shape both bots
receive on stdin. The grid contains only walls and floors and is
side-symmetric, so the lines are identical regardless of perspective.

Line layout, per the upstream Java engine:

```
<width> <height>                               # e.g. "31 15"
<row 0>                                        # height rows of '#' (wall) and ' ' (floor)
<row 1>
...
```

| Position    | Type   | Description                                                                                            |
|-------------|--------|--------------------------------------------------------------------------------------------------------|
| `width`     | int    | Grid width in cells. Typical maps are `31`–`35`; tetris-based generation may pick others.              |
| `height`    | int    | Grid height in cells. Typical maps are `13`–`17`.                                                      |
| row chars   | char   | `'#'` = wall (impassable), `' '` = floor (walkable, eligible for pellets/pacs/cherries).               |

Map width wraps when `MAP_WRAPS` is enabled (left/right edges connect —
true for all leagues). Height never wraps. The `setup` block captures
only static geometry; pellet/cherry/pac placement is part of `state`
and `gameInput` on every turn.

## Frame model

Each main turn maps 1:1 to a Java main turn:

1. The runner reads each side's command line (one pacman command per
   `|`-separated token).
2. The engine runs `executePacmenAbilities` (SPEED/SWITCH activations) →
   `updateAbilityModifiers` → `processPacmenIntent` → `resolveMovement`.
3. If any SPEED'd pac still has path remaining, the engine immediately
   loops `resolveMovement` once more without re-reading bot input
   (Java's `performGameSpeedUpdate`). Both steps emit into the same
   trace turn — `EAT`/`COLLIDE_*`/`KILLED` accumulate; command traces
   (`MOVE`/`SPEED`/`SWITCH`/`WAIT`) only fire on the first step
   (commands are read once per main turn before movement).

After the final main turn, a single post-game-over frame is recorded
(`isOutputTurn = [false, false]`, `output` empty, `traces` empty).
This frame holds the post-`PerformGameOver` pellet absorption when one
side was eliminated; `state` reflects the absorbed result.

## Per-turn `state` payload

`turns[].state` is a `TraceTurnState` ([engine/traces.go](engine/traces.go)),
sampled by the arena runner **after** command parsing and **before**
`PerformGameUpdate`. So values reflect the state players saw when
choosing actions: cooldowns/durations are already ticked, but no
SPEED/SWITCH activation or movement from this turn has applied yet.

| Field          | Type             | Description                                                                                                                |
|----------------|------------------|----------------------------------------------------------------------------------------------------------------------------|
| `pellets`      | int              | Count of regular (1-point) pellets currently on the board.                                                                 |
| `superPellets` | int              | Count of super-pellets / cherries (10 points each) currently on the board. Capped at `NUMBER_OF_CHERRIES` per league (4).  |
| `pacs`         | `[2][]TracePac`  | Per-side pac roster. Outer index = match side (`players[i]`); inner list ordered by global `Pacman.ID` ascending.          |

### `pacs[side][i]` entry

| Field      | Type     | Description                                                                                                       |
|------------|----------|-------------------------------------------------------------------------------------------------------------------|
| `id`       | int      | Global `Pacman.ID` (matches the `pac` field on event payloads). Per-player number is `id / 2`.                    |
| `coord`    | `[2]int` | Current `[x, y]` cell.                                                                                            |
| `type`     | string   | `"ROCK"` / `"PAPER"` / `"SCISSORS"` / `"NEUTRAL"` (no-switch leagues), or `"DEAD"` once eliminated.               |
| `isSpeed`  | int      | `1` if SPEED is currently active (`abilityDuration > 0`), `0` otherwise. Remaining duration lives in `gameInput`. |
| `cooldown` | int      | Turns remaining until the next SPEED/SWITCH can fire (`0` = ready). Decrements at turn start.                     |

Dead pacs are included in the roster regardless of the league's
`PROVIDE_DEAD_PACS` flag (which only gates whether bots see them on
stdin); their `coord` / `isSpeed` / `cooldown` are frozen at the
moment of death.

## Per-turn `gameInput` payload

`turns[].gameInput` is the raw `string[]` from `SerializeTraceFrameInfo`
([engine/serializer.go](engine/serializer.go)) — a god-mode view that
**bypasses fog of war**, so analyzers see every pacman and every
pellet/cherry each turn rather than the filtered view either bot's
stdin received during the live match. Bots themselves still get the
fog-filtered `SerializeFrameInfoFor` lines; the trace path is the only
caller of `SerializeTraceFrameInfo`.

Lines are emitted from match side 0's perspective (`mine` flag uses
`Players[0]` as the reference player); spring2020 input is otherwise
side-symmetric.

```
<myScore> <opponentScore>                                                    # pellet counts
<visiblePacCount>                                                            # always == total pacs (no fog in trace)
<pacIdx> <mine> <x> <y> <type> <abilityDuration> <abilityCooldown>           # one per pac, sorted by global ID
... (visiblePacCount lines)
<visiblePelletCount>                                                         # pellets + cherries
<x> <y> <value>                                                              # value 1 = pellet, 10 = cherry/super-pellet
... (visiblePelletCount lines)
```

### Score line

| Position | Field            | Description                                                                                |
|----------|------------------|--------------------------------------------------------------------------------------------|
| 0        | `myScore`        | Side 0's pellet count going into this turn.                                                |
| 1        | `opponentScore`  | Side 1's pellet count going into this turn.                                                |

### Pacman line

`<pacIdx> <mine> <x> <y> <type> <abilityDuration> <abilityCooldown>`

| Position | Field              | Range / meaning                                                                                                     |
|----------|--------------------|---------------------------------------------------------------------------------------------------------------------|
| 0        | `pacIdx`           | **Per-player** pac number (0..`PacsPerPlayer-1`). Both sides reuse the same indices — disambiguate with `mine`.     |
| 1        | `mine`             | `1` if owned by the reference player (side 0 in the trace), `0` otherwise.                                          |
| 2..3     | `x`, `y`           | Cell coordinates.                                                                                                   |
| 4        | `type`             | `"ROCK"` / `"PAPER"` / `"SCISSORS"` / `"NEUTRAL"` (no-switch leagues), or `"DEAD"` once eliminated.                 |
| 5        | `abilityDuration`  | Remaining SPEED-active turns (`0` if not speeding). Decrements at turn start; `5` when freshly activated.           |
| 6        | `abilityCooldown`  | Remaining cooldown until next SPEED/SWITCH (`0` = ready). Decrements at turn start; `10` when freshly used.         |

Line ordering inside the visible-pacs block is by **global pac ID**
ascending (`Pacman.ID`, the interleaved 0..2N-1 numbering used in
trace events — see below). Dead pacs are appended only when
`PROVIDE_DEAD_PACS` is on (league ≥ 4); when off, the count drops
the dead pacs entirely.

### Pellet/cherry line

`<x> <y> <value>`

`value` is `1` for a regular pellet and `10` (`CHERRY_SCORE`) for a
super-pellet/cherry. The block lists pellets first, then cherries; both
in `Grid.Cells` traversal order. `<visiblePelletCount>` is the **sum**
of pellets + cherries on this single line — bots can't tell from the
count alone which is which.

The typed `state.pellets` and `state.superPellets` fields are the
counted shortcut for analyzers — see [Per-turn `state` payload](#per-turn-state-payload).

## Per-player traces (`turns[].traces[playerIdx]`)

Traces are bucketed by owner; index into `traces[]` is the match side
(same indexing as `players` / score arrays). All structs live in
[engine/traces.go](engine/traces.go).

Trace payloads identify pacs by **global** `Pacman.ID`
(0..`TotalPacs-1`, alternating side: `0`→side 0, `1`→side 1, `2`→side 0, …),
**not** the per-player `pacIdx` shown on stdin. To recover the
per-player number, halve the ID: side 0's pacs have IDs `0, 2, 4, …`
and per-player numbers `0, 1, 2, …`; side 1's pacs have IDs
`1, 3, 5, …` with the same per-player numbering.

Each turn's trace list is split into two buckets, in this order:

1. **Commands** — what each alive pac was asked to do (one per pac per
   turn, leading the slot). Carry the optional `debug` field.
2. **Events** — engine-derived consequences of movement resolution
   (variable count per turn). Do **not** carry `debug`.

### Commands

One command trace per alive pac per main turn, emitted at frame start
before any ability or movement resolution. The SPEED sub-step does not
add additional command traces — bots are not re-prompted between steps.
Commands are emitted **even when the engine ignores the requested
effect** (e.g. SPEED issued while cooldown > 0 still emits the SPEED
command trace; only the activation side-effect is suppressed).

| `type`   | `data`                  | Notes                                                                                                                                                                                                                                              |
|----------|-------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `MOVE`   | `{pac, target, debug?}` | The bot sent `MOVE pacIdx x y`. `target` is `[x, y]` — the bot-supplied destination, not where the pac ended up. The engine pathfinds and steps the pac one cell along that path (or two with SPEED active). Where the pac actually lands is recoverable from the next turn's `state.pacs[*].coord`. |
| `SPEED`  | `{pac, debug?}`         | The bot sent `SPEED pacIdx`. Activates SPEED for the next 5 turns (10-turn cooldown) **iff** cooldown is 0 and the league enables SPEED. Trace fires regardless.                                                                                  |
| `SWITCH` | `{pac, type, debug?}`   | The bot sent `SWITCH pacIdx <type>`. `type` is the requested shape: `"ROCK"` / `"PAPER"` / `"SCISSORS"`. Same cooldown channel as SPEED (10 turns). Trace fires regardless of whether the type change actually applied (cooldown / league flag may suppress the effect). |
| `WAIT`   | `{pac, debug?}`         | Either the bot sent `WAIT pacIdx` explicitly, **or** sent no command for this pac at all. Both leave the pac idle for the turn and look identical from the trace's perspective; distinguish via `debug` — an explicit `WAIT 0 chilling` carries the message, an unsent command does not. |

### Events

Variable count per turn; emitted by the engine in response to ability
activation, movement, and collision resolution. Multiple events can
fire for the same pac in one turn (e.g. a SPEED'd pac stepping onto
two pellets emits two `EAT` events). Cross-owner events are mirrored
into both players' slots so each side sees its involvement.

| `type`          | `data`                  | Notes                                                                                                                                                                                                                              |
|-----------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `EAT`           | `{pac, coord, cost}`    | One event per pellet/cherry consumed. `coord` is `[x, y]`. `cost` = `1` for a regular pellet, `10` for a super-pellet/cherry. When two opposing pacs land on the same cell, **both** sides get an `EAT` (each crediting the same `cost`); same-team duplicates are folded into a single credit. May fire on either step of a SPEED main turn — events from both steps share the trace turn. |
| `KILLED`        | `{pac, coord, killer}`  | Emitted **on the victim's side slot**. `pac` is the dead pac (subject); `coord` is where it died; `killer` is the global pac ID of the winner. The killer's side does not get a corresponding event — analyzers derive `KILLS` from `KILLED` by attributing to the opposite side (see [analyzer.go](engine/analyzer.go)). |
| `COLLIDE_SELF`  | `{pac}`                 | The owning pac's intended move was blocked by **a same-team pac** (FRIENDLY_BODY_BLOCK). Emitted into the owner's slot only. May fire on either of the two SPEED-turn movement steps.                                              |
| `COLLIDE_ENEMY` | `{pac}`                 | The owning pac's intended move was blocked by **an enemy pac** (BODY_BLOCK without RPS-eat conditions, e.g. same type). **Mirrored into both players' slots** so each side sees its involvement: `traces[ownerSide]` carries the entry from the blocked pac's perspective; `traces[otherSide]` is an identical copy with the same `pac` field. |

### Inline `debug` field (commands only)

Free-text appended to any per-pac command (`MOVE 0 5 1 hello`,
`SPEED 0 go go go`, `SWITCH 1 PAPER counter`, `WAIT 2 thinking`, …)
rides on the **command** trace as `data.debug`. The text is captured
into `Pacman.Message` by `CommandManager.HandlePacmanCommand`
(truncated to 48 chars per Java's `Pacman.setMessage`) and cleared on
the next turn's `TurnReset`, so a non-empty `debug` always reflects
the message sent on the same turn the command was emitted.

The field is omitted from JSON when empty, so a present `data.debug`
is unambiguously a bot-authored chat string. Event traces (`EAT`,
`KILLED`, `COLLIDE_SELF`, `COLLIDE_ENEMY`) deliberately **do not
carry** the debug field — the chat string belongs to the command, not
to its downstream fallout. Errored commands (parse failure or
player deactivation) drop the message entirely: there's no command
trace to attach it to.

### What is **not** a per-player trace

Several things the upstream `GameSummaryManager` reports as standalone
summary lines are not emitted as `traces[]` entries because they are
not per-player. Each is recoverable from the per-turn `state` (or
match-level fields) instead:

- **Pellet inventory** — `state.pellets` and `state.superPellets`
  carry the global counts; per-pellet cells live in `gameInput` for
  the rare analyzer that needs cell-level detail.
- **Per-pac type / position / cooldowns** — promoted to
  `state.pacs[side][i]` rather than emitted as a turn-by-turn diff;
  analyzers diff consecutive turns to detect changes.
- **Sub-step boundaries within a SPEED turn** — events from main
  step and sub-step share the same `traces[]` array. Step
  attribution is not currently recoverable.
- **`KILLS` (the killer's perspective)** — derivable from `KILLED`
  by attributing to the side opposite the slot the event landed in
  (handled in [analyzer.go](engine/analyzer.go)).

## League rules and trace shape

Per-league toggles (see [spring2020_league_rules.go](engine/spring2020_league_rules.go))
change which traces can appear:

- **League 1**: 1 pac per side, no fog, no abilities → no `SPEED`,
  `SWITCH`, `COLLIDE_*` (only one pac), `KILLED` (no RPS combat
  without abilities ever switching types). `MOVE` and `WAIT` only.
- **League 2**: 2..5 pacs, no fog, no abilities → no `SPEED`/`SWITCH`
  command traces; `COLLIDE_*` and `KILLED` events possible.
- **League 3**: full game minus dead-pac reporting → `gameInput` drops
  `DEAD` lines once a pac is eliminated; `state.pacs` and trace
  events are unchanged.
- **League ≥ 4**: full rules, dead pacs reported on stdin.

The `state` and `gameInput` payloads always reflect the **active
league's rules** (e.g. no SPEED activations possible in league 1) but
always use the god-mode no-fog view regardless of `FOG_OF_WAR`.

## Example

A trimmed self-play trace covering one main turn plus the post-end
frame. The trace below uses a 5-pac-per-side league-4 match; large
pellet lists are elided.

```json
{
  "createdAt": "2026-05-08T12:00:00Z",
  "type": "trace",
  "puzzleId": 592,
  "puzzleName": "spring2020",
  "traceId": 1715170800,
  "matchId": 0,
  "players": ["bot-a", "bot-b"],
  "blue": "bot-a",
  "league": 4,
  "seed": "1234567890",
  "endReason": "SCORE_EARLY",
  "scores": [129.0, 123.0],
  "finalScores": [129.0, 123.0],
  "ranks": [0, 1],
  "setup": [
    "31 15",
    "###############################",
    "    #   #   #     #   #   #    ",
    "..."
  ],
  "mainTurns": 123,
  "turns": [
    {
      "turn": 0,
      "gameInput": [
        "0 0",
        "4",
        "0 1 2 1 ROCK 0 0",
        "0 0 28 1 ROCK 0 0",
        "1 1 23 10 PAPER 0 0",
        "1 0 7 10 PAPER 0 0",
        "221",
        "..."
      ],
      "output": ["SPEED 0 | SPEED 1", "SPEED 0|SPEED 1"],
      "isOutputTurn": [true, true],
      "score": [0, 0],
      "traces": [
        [
          { "type": "SPEED", "data": { "pac": 0 } },
          { "type": "SPEED", "data": { "pac": 2 } }
        ],
        [
          { "type": "SPEED", "data": { "pac": 1 } },
          { "type": "SPEED", "data": { "pac": 3 } }
        ]
      ],
      "state": {
        "pellets": 217,
        "superPellets": 4,
        "pacs": [
          [
            { "id": 0, "coord": [2, 1],   "type": "ROCK",  "isSpeed": 0, "cooldown": 0 },
            { "id": 2, "coord": [23, 10], "type": "PAPER", "isSpeed": 0, "cooldown": 0 }
          ],
          [
            { "id": 1, "coord": [28, 1],  "type": "ROCK",  "isSpeed": 0, "cooldown": 0 },
            { "id": 3, "coord": [7, 10],  "type": "PAPER", "isSpeed": 0, "cooldown": 0 }
          ]
        ]
      }
    },
    {
      "turn": 1,
      "gameInput": [
        "0 0",
        "4",
        "0 1 2 1 ROCK 5 9",
        "0 0 28 1 ROCK 5 9",
        "1 1 23 10 PAPER 5 9",
        "1 0 7 10 PAPER 5 9",
        "221",
        "..."
      ],
      "output": ["MOVE 0 5 1 hunt | MOVE 1 23 4", "MOVE 0 25 1|MOVE 1 7 4"],
      "isOutputTurn": [true, true],
      "score": [0, 0],
      "traces": [
        [
          { "type": "MOVE", "data": { "pac": 0, "target": [5, 1], "debug": "hunt" } },
          { "type": "MOVE", "data": { "pac": 2, "target": [23, 4] } },
          { "type": "EAT",  "data": { "pac": 0, "coord": [3, 1], "cost": 1 } },
          { "type": "EAT",  "data": { "pac": 0, "coord": [4, 1], "cost": 1 } },
          { "type": "EAT",  "data": { "pac": 2, "coord": [22, 4], "cost": 1 } }
        ],
        [
          { "type": "MOVE", "data": { "pac": 1, "target": [25, 1] } },
          { "type": "MOVE", "data": { "pac": 3, "target": [7, 4] } },
          { "type": "EAT",  "data": { "pac": 1, "coord": [21, 5], "cost": 1 } },
          { "type": "EAT",  "data": { "pac": 3, "coord": [8, 4], "cost": 1 } }
        ]
      ],
      "state": {
        "pellets": 212,
        "superPellets": 4,
        "pacs": [
          [
            { "id": 0, "coord": [2, 1],   "type": "ROCK",  "isSpeed": 1, "cooldown": 9 },
            { "id": 2, "coord": [23, 10], "type": "PAPER", "isSpeed": 1, "cooldown": 9 }
          ],
          [
            { "id": 1, "coord": [28, 1],  "type": "ROCK",  "isSpeed": 1, "cooldown": 9 },
            { "id": 3, "coord": [7, 10],  "type": "PAPER", "isSpeed": 1, "cooldown": 9 }
          ]
        ]
      }
    }
  ]
}
```

Things worth pointing out in the example:

- Turn 0 is the first player-decision frame. Both sides issue SPEED
  commands for their two ROCK/PAPER pacs. The `SPEED` command trace
  fires once per pac per main turn at frame start, before any
  ability or movement resolution. `state.pacs` shows pre-activation
  values (`isSpeed: 0`, `cooldown: 0`); the next turn's `state` will
  reflect the post-tick values.
- Turn 1 shows command-then-event ordering inside each slot: every
  alive pac leads with its `MOVE` command (carrying the bot-supplied
  `target` and any `debug` string), then the engine-derived `EAT`
  events follow. The `debug: "hunt"` rides on the side-0 pac-0 MOVE
  command; the EATs that movement caused do **not** carry it.
- Turn 1 also illustrates SPEED sub-step accumulation: each pac took
  two movement steps inside the same trace turn, so a pac that
  crossed two pellets emits two `EAT` events. Both steps share the
  single `MOVE` command — bots are not re-prompted between steps.
- Pac IDs in `traces[].data.pac` and `state.pacs[*].id` are global
  (`0`, `1`, `2`, `3`): side 0 owns even IDs; side 1 owns odd IDs.
  The corresponding `pacIdx` shown on stdin is `id / 2` for both
  sides.
- The post-end frame (not shown — `mainTurns: 123` means turn 123 is
  the last decision; turn 124 would be the post-end frame with
  `isOutputTurn: [false, false]`, empty `output`/`traces`, and a
  `state` reflecting any post-`PerformGameOver` pellet absorption).
