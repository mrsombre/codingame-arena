# Trace files

A **trace** is the canonical per-match JSON record arena produces. It bundles the metadata downstream tools need (analyze, the web viewer, training pipelines) without forcing them to re-read CodinGame's bulky replay payload.

Two commands write traces:

- [`arena run --trace`](run.md) — self-play matches → `traces/trace-<traceId>-<matchId>.json` (`type: "trace"`)
- [`arena replay`](replay.md) — converted CodinGame replays → `traces/replay-<replayId>.json` (`type: "replay"`)

[`arena analyze`](analyze.md) and [`arena serve`](serve.md) read whatever `*.json` files live under `--trace-dir`; the `type` field and filename let consumers tell self-play apart from replay-derived traces.

## Quick start

```shell
# Self-play: 100 matches, one trace file per match
bin/arena run --game=winter2026 \
  --blue=bin/bot-winter2026-cpp \
  --red=bin/bot-winter2026-py \
  --trace --trace-dir=./traces

# Replay-derived: download + auto-convert (writes both replays/<id>.json and traces/replay-<id>.json)
bin/arena replay mrsombre --game=winter2026 875142454

# Print one self-play match's trace to stdout (handy for piping into jq)
bin/arena run --game=winter2026 --blue=... --debug
```

A self-play batch shares a `traceId` (the batch start timestamp) across every file it writes, so one batch's traces can be filtered together. Replay traces use the CodinGame replay id as their `traceId`, so a single replay always has `matchId: 0`.

## File naming

| File pattern              | Producer                | `type`    | `traceId` source               |
|---------------------------|-------------------------|-----------|--------------------------------|
| `trace-<traceId>-<matchId>.json` | `arena run --trace`     | `trace`   | batch start timestamp          |
| `replay-<replayId>.json`  | `arena replay`          | `replay`  | CodinGame replay id            |

`TraceTypeFromFileName` infers the type from the filename when consumers need to disambiguate without parsing JSON.

## Top-level fields

Fields appear in the JSON in this order. Side-indexed arrays (`players`, `scores`, `finalScores`, `ranks`, `disqualified`, `timing.*`) are always two-element with index 0 = left side of the map, index 1 = right side. `blue` names which side is "ours" — analyzers locate it by scanning `players[i] == blue`.

| #  | Field          | Type        | Description                                                                                                   |
|----|----------------|-------------|---------------------------------------------------------------------------------------------------------------|
| 1  | `createdAt`    | string      | RFC 3339 timestamp the trace was produced. Self-play: stamped at match completion. Replay: copied from the source replay's `fetchedAt`. |
| 2  | `type`         | string      | `trace` (self-play) or `replay` (converted CodinGame replay).                                                 |
| 3  | `puzzleId`     | int         | Canonical CodinGame puzzle id (e.g. `730` for Spring Challenge 2021). Used to gate cross-game replay loads.   |
| 4  | `puzzleName`   | string      | Active arena game id (e.g. `spring2021`, `winter2026`). Filters traces in `arena analyze`.                    |
| 5  | `traceId`      | int64       | Batch identifier. Self-play: shared across a batch. Replay: equals the CodinGame replay id.                   |
| 6  | `matchId`      | int         | Per-batch match index. Always `0` for replay traces.                                                          |
| 7  | `players`      | `[2]string` | Bot/agent display name on each side. Self-play: bot binary basenames. Replay: CodinGame pseudos / boss nicknames. |
| 8  | `blue`         | string      | Side we are playing for (the analyze "us" side). Always equals one of `players`. Required.                   |
| 9  | `league`       | int         | League level the match was played at.                                                                         |
| 10 | `seed`         | string      | Match RNG seed; JSON-string-encoded int64 to survive 53-bit JS number precision.                              |
| 11 | `endReason`    | string      | How the match ended. Game-specific; shared values: `TIMEOUT_START`, `TIMEOUT`, `INVALID`, `ELIMINATED`, `SCORE`, `SCORE_EARLY`, `TURNS_OUT`. Empty when the referee doesn't expose `EndReasonProvider`. |
| 12 | `disqualified` | `[2]bool`   | Side `i` was deactivated (timeout / bad command). Omitted when neither side was DQ'd. Used to attribute fault end reasons. |
| 13 | `scores`       | `[2]float`  | Raw pre-`OnEnd` score (intrinsic in-game count, e.g. spring2021 tree segments before the sun bonus). Always engine truth. |
| 14 | `finalScores`  | `[2]float`  | Post-`OnEnd` score (with bonuses and tiebreakers). Self-play: engine truth. Replay (DQ): inherited from CG's `gameResult.scores` so the trace matches the official replay. |
| 15 | `ranks`        | `[2]int`    | CodinGame-style ranks: `0` = first place, `[0,0]` = draw. For replays, normalized from CG's `gameResult.ranks`. |
| 16 | `timing`       | object      | Aggregate response timings in milliseconds (see below). Omitted on replay traces (no live bot to time).       |
| 17 | `mainTurns`    | int         | Count of player-decision turns. Excludes non-decision phase frames (spring2021 GATHERING/SUN_MOVE) and post-end frames (spring2020 gameOverFrame). `0` on legacy traces written before this field existed. |
| 18 | `turns`        | array       | Per-turn entries (see below).                                                                                 |

Score values are always emitted with at least one fractional digit (`127` → `127.0`) so the file shape matches CodinGame's replay encoding.

### `timing`

| Field              | Type         | Description                                                                |
|--------------------|--------------|----------------------------------------------------------------------------|
| `firstResponse`    | `[2]float`   | Turn-0 latency per side (typically dominated by bot startup; not steady-state). |
| `responseAverage`  | `[2]float`   | Mean response time per side, **excluding** turn 0.                         |
| `responseMedian`   | `[2]float`   | Median response time per side, **excluding** turn 0.                       |

## Per-turn fields (`turns[]`)

| Field           | Type         | Description                                                                                                   |
|-----------------|--------------|---------------------------------------------------------------------------------------------------------------|
| `turn`          | int          | Turn number, starting at `0`.                                                                                 |
| `gameInput`     | `[]string`   | Lines the engine fed **blue's** stdin this turn. Symmetric-input games: identical for both sides. Fog-of-war games: blue's perspective only. Absent on turns where blue did not execute. |
| `output`        | `[2]string`  | Raw stdout each side emitted this turn. Empty when the side was deactivated or skipped. Omitted when both sides were silent. |
| `isOutputTurn`  | `[2]bool`    | Side `i` was prompted for output this turn. `false` on engine-only frames (spring2021 GATHERING/SUN_MOVE, post-end frames) and on skipped/deactivated sides. Lets analyzers find the first turn a bot was actually asked to act. |
| `timing`        | object       | `{ "response": [2]float }` — per-side response time in milliseconds. `0` when the side did not execute.       |
| `score`         | `[2]int`     | Raw score going into this turn, sampled from `RawScoresProvider` before `PerformGameUpdate`. Zero when the referee doesn't expose raw scores. |
| `traces`        | `[2][]event` | Structured events partitioned by owner. `traces[0]` = player 0's events; `traces[1]` = player 1's. Cross-owner events (spring2020 `COLLIDE_ENEMY`, winter2026 `HIT_ENEMY`) are mirrored into both slots. Each event is `{ "type": string, "data": object }`. |
| `state`         | object       | Opaque per-turn payload owned by the game. Arena never inspects it; games marshal a typed struct describing whatever board / scoring / phase info downstream consumers need. |

`turns[].traces` event labels (`GATHER`, `EAT`, `DEAD`, `HIT_ENEMY`, …) are **game-defined**. Arena passes them through; `arena analyze` asks each game's `TraceMetricSpec` how to aggregate them.

## Sample trace file

Trimmed spring2021 replay trace (truncated at the second turn for brevity):

```json
{
  "createdAt": "2026-05-07T11:08:14Z",
  "type": "replay",
  "puzzleId": 730,
  "puzzleName": "spring2021",
  "traceId": 886355641,
  "matchId": 0,
  "players": ["mrsombre", "vemynona"],
  "blue": "mrsombre",
  "league": 4,
  "seed": "7564536647070201000",
  "endReason": "SCORE",
  "scores": [126.0, 122.0],
  "finalScores": [129.0, 131.0],
  "ranks": [1, 0],
  "timing": {
    "firstResponse":   [0, 0],
    "responseAverage": [0, 0],
    "responseMedian":  [0, 0]
  },
  "mainTurns": 85,
  "turns": [
    {
      "turn": 0,
      "timing": { "response": [0, 0] },
      "score": [0, 0],
      "traces": [
        [
          { "type": "GATHER", "data": { "cell": 30, "sun": 1 } },
          { "type": "GATHER", "data": { "cell": 33, "sun": 1 } }
        ],
        [
          { "type": "GATHER", "data": { "cell": 21, "sun": 1 } },
          { "type": "GATHER", "data": { "cell": 24, "sun": 1 } }
        ]
      ],
      "state": {
        "day": 0,
        "phase": "gathering",
        "sun_direction": 0,
        "nutrients": 20,
        "sun":   [0, 0],
        "trees": [[[30, 1, 1], [33, 1, 1]], [[21, 1, 1], [24, 1, 1]]]
      }
    }
  ]
}
```

A self-play trace looks the same except `type` is `"trace"`, `timing` carries real numbers, and per-turn entries include `gameInput` and `output`.

## Reading traces in Go

Trace files are produced and consumed by `internal/arena.TraceMatch`. Round-tripping is `json.Unmarshal` → struct → `json.MarshalIndent`:

```go
import (
    "encoding/json"
    "os"

    "github.com/mrsombre/codingame-arena/internal/arena"
)

func loadTrace(path string) (arena.TraceMatch, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return arena.TraceMatch{}, err
    }
    var t arena.TraceMatch
    if err := json.Unmarshal(data, &t); err != nil {
        return arena.TraceMatch{}, err
    }
    return t, nil
}
```

`TraceMatch.BlueSide()` returns the side index (0 or 1) for `blue`, so callers can do side-relative analysis without re-deriving the mapping. `TraceWinnerFromScores(scores, disqualified)` decides the winner with deactivation precedence (a DQ'd side can never win, even with a higher raw score).
