# Summer Challenge 2026 — Trace format

TBD, see [docs/trace.md](../../docs/trace.md).

Written once the engine is ported and `TurnTraceProvider` is wired
([games/docs/plan.md](../docs/plan.md) phases 4 and 12). Known so far:

| Field        | Value                                        |
| ------------ | -------------------------------------------- |
| `puzzleId`   | TBD                                           |
| `puzzleName` | `summer2026`                                  |
| Turn model   | `FlatTurnModel`                               |

`FlatTurnModel` applies because `Game.shouldSkipPlayerTurn` always returns
`false`, the referee runs exactly one `performGameUpdate` per loop turn, and
the match ends on the same turn the end condition fires — there is no phase
frame and no post-end frame. So `mainTurns` equals `len(turns)`, capped at
`MAX_TURNS = 100`.
