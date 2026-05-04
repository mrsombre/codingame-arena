# Game Model

How games plug into the arena: the mandatory contract every game must satisfy
and the optional capability interfaces each game opts into.

Commands and paths below assume the repository root as the working directory.
Read interface definitions from `internal/arena/interfaces.go`.

## Shared invariants

Every game in this arena obeys the same surface contract:

- Two players, simultaneous decisions per turn.
- Integer scores. One winner or a draw; no third place.
- Disqualified player → score `-1`, recorded in `trace.Deactivated[i]`.
- First-turn timeout 1000 ms, subsequent turns 50 ms (enforced by the runner).
- A bounded turn count (≤ 200 main turns in practice).

These are not interfaces — they are the assumptions every consumer (replay
verifier, viewer, analyzer) builds on. A game that violates them does not
fit this arena.

## Mandatory contract

Three interfaces every engine must implement.

### `arena.GameFactory`

Static metadata + game construction. One factory per game, registered in
`games/game.go`.

| Method               | Purpose                                                          |
| -------------------- | ---------------------------------------------------------------- |
| `Name()`             | Internal identifier (e.g. `"spring2021"`). Used as `trace.GameID`.|
| `PuzzleID()`         | CodinGame numeric puzzle ID. Used to gate `arena replay` convert.|
| `PuzzleTitle()`      | Human-readable puzzle title.                                     |
| `LeaderboardSlug()`  | URL slug under `/multiplayer/bot-programming/<slug>`.            |
| `MaxTurns()`         | Hard cap for the engine loop (post-cap is a runner-side timeout).|
| `NewGame(seed, opts)`| Returns a `Referee` and 2-element `[]Player` for one match.      |

### `arena.Referee`

The per-match driver. Methods are called in a fixed order by the runner; see
`internal/arena/match.go` and `internal/arena/replay_runner.go`.

| Method                         | When                                                         |
| ------------------------------ | ------------------------------------------------------------ |
| `Init(players)`                | Once at match start.                                         |
| `GlobalInfoFor(player)`        | Once per player after `Init`.                                |
| `FrameInfoFor(player)`         | Per turn, for every active non-skipped player.               |
| `ParsePlayerOutputs(players)`  | Per turn, after stdout is collected.                         |
| `PerformGameUpdate(turn)`      | Per turn, after parsing.                                     |
| `ResetGameTurnData()`          | Per turn, before polling players.                            |
| `ShouldSkipPlayerTurn(player)` | Per turn, per player — gates `FrameInfoFor` + execution.     |
| `Ended()`                      | Per turn, after `PerformGameUpdate`.                         |
| `EndGame()`                    | Once when the loop exits without `Ended()` having returned true.|
| `OnEnd()`                      | Once at the very end (post-game tiebreaker, score finalize). |
| `ActivePlayers(players)`       | Helper — count of non-deactivated players.                   |

### `arena.Player`

The agent-facing handle. Engine-defined concrete type, exposed through the
interface so the runner can drive it without knowing the game.

## Turn model

Different games have different relationships between engine iterations,
player-decision turns, and CodinGame replay frames. The factory selects one
of three built-in models via `arena.TurnModeler`:

```go
type TurnModeler interface {
    TurnModel() arena.TurnModel
}

type TurnModel interface {
    ExpectedTraceTurnCount(replay) int  // matches len(trace.Turns)
    MainTurnCount(replay) int           // matches trace.MainTurns
}
```

| Model                | Shape                                                                                | Used by      |
| -------------------- | ------------------------------------------------------------------------------------ | ------------ |
| `FlatTurnModel`      | 1 engine iteration = 1 decision turn = 1 trace turn. No phase/post-end frames.       | winter2026   |
| `PostEndTurnModel`   | Same as flat plus 1 post-end "gameOverFrame" trace turn after the last decision turn.| spring2020   |
| `PhaseTurnModel`     | Each round: 1 GATHERING + N ACTIONS + 1 SUN_MOVE engine iterations / trace turns.    | spring2021   |

Factories without `TurnModeler` fall back to `FlatTurnModel`.

`MainTurnCount` is universal across all three models: it counts replay frames
where at least one player produced stdout, ignoring empty-stdout phase frames
and trailing engine markers. The trace side of the comparison is computed in
`replay_runner.go` and `match.go` as `len(turns) where Output is non-empty`.

## Optional capabilities

Each capability is an interface the factory or referee may implement to plug
into a downstream consumer. None are required — when a game omits one, the
consumer falls back to a sensible default.

### Factory-level

| Interface                | Methods                              | Effect when implemented |
| ------------------------ | ------------------------------------ | ----------------------- |
| `LeagueResolver`         | `ResolveLeague(opts) int`            | `arena run` resolves a league number from `--league` and stamps it on traces. |
| `TurnModeler`            | `TurnModel() TurnModel`              | Selects per-game replay-frame counting (see above). Default: `FlatTurnModel`. |
| `TraceMetricAnalyzer`    | `TraceMetricSpecs() []TraceMetricSpec`| `arena analyze` computes per-game per-turn metric counts and renders them in the report. |

### Referee-level

| Interface                | Methods                                   | Effect when implemented |
| ------------------------ | ----------------------------------------- | ----------------------- |
| `EndReasonProvider`      | `EndReason(turn, players, deactTurns) string` | Trace records `end_reason` (TIMEOUT / SCORE / TURNS_OUT / …). Empty otherwise. |
| `RawScoresProvider`      | `RawScores() [2]int`                      | `trace.Scores` records pre-OnEnd raw values; otherwise post-OnEnd `Player.GetScore()` is used. Required when OnEnd modifies scores (tiebreakers, bonuses). |
| `MetricsProvider`        | `Metrics() []arena.Metric`                | Per-match summary metrics (e.g. apples remaining, pellets remaining) attached to live-run results. |
| `TurnTraceProvider`      | `TurnTraces(turn, players) []TurnTrace`   | Per-turn structured event stream attached to each `TraceTurn.Traces`. Drained after `PerformGameUpdate`. |
| `GameOverFrameReporter`  | `InGameOverFrame() bool`                  | Signals the runner to skip player polling for a post-end "gameOverFrame" turn (avoids spurious deactivation). Pair with `PostEndTurnModel`. |

## End-reason vocabulary

Shared constants in `internal/arena/trace.go`:

| Constant               | Meaning                                                       |
| ---------------------- | ------------------------------------------------------------- |
| `TIMEOUT_START`        | Player timed out on turn 0 (bot startup failure).             |
| `TIMEOUT`              | Player timed out on a later turn.                             |
| `INVALID`              | Player produced bad input (not a timeout).                    |
| `ELIMINATED`           | In-game elimination condition (no fault) — e.g. all units dead.|
| `SCORE`                | Game ended on a normal scoring condition (board cleared, round cap reached).|
| `SCORE_EARLY`          | Game ended early because the score was mathematically locked. |
| `TURNS_OUT`            | Reached the engine's `MaxTurns` cap with neither side done.   |

## Per-game capability matrix

| Capability                | winter2026 | spring2020 | spring2021 |
| ------------------------- | :--------: | :--------: | :--------: |
| `LeagueResolver`          | ✅         | ✅         | ✅         |
| `TurnModeler`             | Flat       | PostEnd    | Phase      |
| `EndReasonProvider`       | ✅         | ✅         | ✅         |
| `RawScoresProvider`       | ✅         | ✅         | ✅         |
| `TurnTraceProvider`       | ✅         | ✅         | ✅         |
| `MetricsProvider`         | ✅         | ✅         | ❌         |
| `TraceMetricAnalyzer`     | ✅         | ✅         | ❌         |
| `GameOverFrameReporter`   | ❌         | ✅         | ❌         |

## Replay verification layers

`arena replay` re-simulates each downloaded CodinGame replay through the
engine and verifies the result against the recorded outcome. Three agreement
layers, all mandatory:

| Layer | Replay side                              | Trace side                |
| ----- | ---------------------------------------- | ------------------------- |
| L0    | `gameResult.scores` (-1 = DQ), `gameResult.ranks` | `trace.Scores` (post-OnEnd compared), `trace.Ranks`, `trace.Deactivated` |
| L1    | `model.MainTurnCount(replay)`            | `trace.MainTurns`         |
| L2    | `model.ExpectedTraceTurnCount(replay)`   | `len(trace.Turns)`        |

L0 has one tolerated case: the engine declares a draw (tied scores, no DQ)
but CodinGame ranks pick a winner via an unmodeled post-OnEnd tiebreaker. In
that case the trace adopts a draw `[0, 0]` (matches what the replay page
displays as 1st/1st). Other rank disagreements still fail strictly.

End-reason is engine-side only. Replays carry no end-reason field, and
parsing it from summary text would be fragile and game-specific.

## Adding a new game

1. Port the engine — see `java-port.md`.
2. Implement the mandatory contract (`GameFactory`, `Referee`, `Player`).
3. Pick a `TurnModel` (Flat unless the engine has phase frames or a post-end
   frame).
4. Layer in optional capabilities as needed:
   - `EndReasonProvider` + `RawScoresProvider` for full L0 verification.
   - `TurnTraceProvider` if you want per-turn structured events for analysis.
   - `TraceMetricAnalyzer` to expose game-specific metrics in `arena analyze`.
5. Register the factory in `games/game.go`.
6. Verify end-to-end:
   ```
   make test-arena && make test-games && make lint-arena && make build-arena
   bin/arena replay <username> --game <name> -n 5 -f
   ```
   Saved traces should report `0 skipped-mismatch`.
