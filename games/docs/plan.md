# Adding a New Game

End-to-end plan for porting a CodinGame multiplayer game into the arena.
Each phase produces a reviewable, committable artefact. Skipping a phase
makes the next one harder to verify.

This document is the orchestration layer. Detailed how-tos live in:

- `java-port.md` — Java → Go porting rules, file layout, parity tests.
- `game-model.md` — mandatory contract, optional capabilities, replay
  verification layers.

Commands and paths assume the repository root as the working directory.
Use `winter2026` as the running example for a new game id.

## Phases

### 1. Import upstream source

Add the upstream Java engine as a git subtree under `source/<RepoName>`.

- Locate the official CodinGame repo (e.g. `CodinGame/WinterChallenge2026-Exotec`).
- Add it as a subtree, not a submodule — local diffs need to be greppable.
- Never modify files under `source/` after import.
- Commit on its own so the import is auditable.

### 2. Stub the game directory

Create `games/<game>/` with the minimum scaffolding:

```
games/<game>/
├─ README.md
├─ rules.md          # filled in phase 3
├─ engine/           # filled in phase 4
└─ agents/           # filled in phase 10
```

`README.md` template:

```markdown
# <Codename> - <Season Year> Challenge

https://www.codingame.com/multiplayer/bot-programming/<slug>

---

Source: https://github.com/CodinGame/<RepoName>
```

### 3. Compile `rules.md`

Distil a human- and agent-readable rules document from the upstream
statement and source comments. The statement is usually at
`source/<Repo>/config/statement_en.html.tpl`, supplemented by rule
comments inside the Java engine.

Coverage requirements are enumerated in `java-port.md` → "Rules
Documentation". When the statement and the Java source disagree, the Java
source wins and the divergence is noted in `rules.md`.

### 4. Port the engine

Follow `java-port.md` end-to-end — file layout, naming, snippets, parity
rules. Output: `games/<game>/engine/` package implementing the mandatory
contract from `game-model.md`:

- `arena.GameFactory`
- `arena.Referee`
- `arena.Player`

Pick a `TurnModel` (Flat / PostEnd / Phase) per `game-model.md` →
"Turn model".

### 5. Engine acceptance tests

Add `engine/acceptance_test.go` covering the rules-visible behavior of a
full match: setup, input serialization, simultaneous actions, movement,
collisions, scoring, end conditions, league differences. These are the
tests that exercise the `Referee` lifecycle as the runner sees it.

### 6. Per-file unit tests

Add focused tests next to each Go file that owns isolated logic:
coordinates, parsers, validators, grids, pathfinding, RNG helpers, etc.
Naming follows the source file:

```
spring2020_coord.go        -> spring2020_coord_test.go
spring2020_pathfinder.go   -> spring2020_pathfinder_test.go
```

### 7. Grid-generator parity tests

For any random map / grid generator, add an acceptance test that compares
generated output against real CodinGame output for two known seeds:

```
468706172918629800
-468706172918629800
```

See `games/spring2020/engine/spring2020_maps_tetris_based_map_generator_test.go`
as the reference pattern.

### 8. Bind optional capabilities

Implement the optional interfaces the game needs (see `game-model.md` →
"Optional capabilities"):

- `LeagueResolver` if the game has leagues.
- `EndReasonProvider` + `RawScoresProvider` for full L0 verification.
- `TurnTraceProvider` for per-turn structured events.
- `MetricsProvider` + `TraceMetricAnalyzer` to expose game-specific
  metrics in `arena analyze`.
- `GameOverFrameReporter` paired with `PostEndTurnModel` when the engine
  emits a post-end frame.

Update the per-game capability matrix in `game-model.md`.

### 9. Replay verification

Download the latest CodinGame replays for this game and re-simulate them
through the engine:

```
bin/arena replay mrsombre --game <game>
```

Expected outcome: zero `skipped-mismatch` entries. Discrepancies surface
as L0 / L1 / L2 disagreements (see `game-model.md` → "Replay verification
layers") and must be fixed in the engine, not papered over.

### 10. Reference bots

Write a minimal C++ and Python bot under `games/<game>/agents/`:

```
games/<game>/agents/
├─ bot.cpp
└─ bot.py
```

Wire `build-<game>-agents` and `match-<game>` targets into the `Makefile`
following the `winter2026` example.

### 11. Local simulated matches

Run a batch of local matches to exercise the engine + bots end-to-end:

```
bin/arena --game=<game> \
    --blue=./bin/bot-<game>-cpp \
    --red=./bin/bot-<game>-py \
    --seed=100030005000700089 \
    --simulations 50 \
    --trace
```

Target ~100 simulations once the engine is stable. The `--trace` flag
writes per-turn traces to `traces/` for inspection.

### 12. Meaningful per-turn traces

Extend `TurnTraceProvider` with structured events that go beyond the raw
CodinGame replay payload — anything that helps diagnose bot behavior,
analyzer queries, or rule edge cases. The replay payload is the floor,
not the ceiling.

### 13. Analyzer

Implement `TraceMetricAnalyzer` so `arena analyze --game <game>` reports
per-turn metric counts derived from the traces produced in phase 12.
This closes the loop: replays → traces → metrics → rendered report.

## Validation gate

Before declaring a phase complete, run the project gates:

```
make test-arena && make test-games && make lint-arena && make build-arena
```

Phases 9 and 11 also require a clean run of `bin/arena replay …` and the
relevant `match-<game>` target.
