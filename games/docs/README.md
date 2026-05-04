# Games Documentation

Reference docs for porting and integrating CodinGame multiplayer games
into the arena.

## Contents

- [plan.md](plan.md) — end-to-end plan for adding a new game (13 phases,
  from upstream subtree import to analyzer wiring). Start here.
- [java-port.md](java-port.md) — Java → Go porting rules: file layout,
  naming, snippet conventions, parity tests, and the per-engine
  checklist.
- [game-model.md](game-model.md) — arena contract: shared invariants,
  mandatory interfaces (`GameFactory`, `Referee`, `Player`), turn models,
  optional capabilities, end-reason vocabulary, and replay verification
  layers.

## Reading order

1. `plan.md` to scope the work and pick the phase you are on.
2. `java-port.md` while implementing the engine itself.
3. `game-model.md` when wiring the engine into the arena and deciding
   which optional capabilities to implement.
