# Porting a CodinGame Java Engine to Go

Guide for porting any CodinGame multiplayer game engine from Java to Go
inside `games/<engine>/engine/`.

Commands and import paths below assume they are run from the repository root.
Read Java sources from `source/`, but never modify files under `source/`.

## Scope: Simulation Only

Port only the **core game simulation** — the logic needed for the referee to
read input from two agents, run turn updates, and send the next turn's input
back. Skip everything that only exists for the CodinGame viewer:

- animation / event / frame-time systems
- sprites, tooltips, tile graphics, module setup
- replay snapshots, viewer-only serializers
- score tooltips and human-readable score text

Concretely, do **not** port the Java `View`, `Animation`, `Module`,
`TooltipModule`, `ViewerEvents`, `Serializer.serializeFrameData`, or any
sibling types whose only consumer is the CodinGame frontend. If a method on
`Game`/`Referee` exists purely to feed the viewer (`SnapshotJSON`,
`ViewerEvents`, `TurnEvents`, `FrameData`, `GlobalData`), omit it — the arena
`Referee` interface does not require it and the optional
`TraceProvider`/`TurnEventProvider` interfaces can be left unimplemented.

Parity applies to simulation behavior (turn order, RNG, deactivation reasons,
command validation, scoring), not to viewer output.

The game engine must be treated as a correctness-critical Java port. Local
simulations depend on exact behavior; small differences in ordering, rounding,
visibility, validation, or endgame scoring can invalidate bot results.

## Directory Layout

The whole engine lives in **one flat Go package** at `games/<engine>/engine/`.
Do **not** create Go sub-packages that mirror Java sub-packages — Go forbids
the cyclic package dependencies that Java tolerates, and a flat layout is the
simplest way to avoid the problem entirely. The cost is long filenames; that
is acceptable.

Encode each Java file's full package path in its Go filename. Take the Java
package below `com.codingame`, lowercase it, and join the parts with the
filename using `_` as a separator:

```
com.codingame.game.Player                   -> game_player.go
com.codingame.spring2020.Pacman             -> spring2020_pacman.go
com.codingame.spring2020.Cell               -> spring2020_cell.go
com.codingame.spring2020.action.MoveAction  -> spring2020_action_move_action.go
com.codingame.spring2020.maps.TetrisBasedMapGenerator
                                            -> spring2020_maps_tetris_based_map_generator.go
com.codingame.spring2020.pathfinder.AStar   -> spring2020_pathfinder_a_star.go
```

```
games/<engine>/
  rules.md           # ported from the source game statement/documentation
  engine/
    engine.go                                # GameFactory implementation (arena-only)
    register.go                              # init() -> arena.Register(NewFactory())
    serializer.go                            # replay format (arena-only)
    game_<file>.go                           # ports of com.codingame.game/*.java
    <gamepkg>_<file>.go                      # ports of com.codingame.<gamepkg>/*.java
    <gamepkg>_<subpkg>_<file>.go             # ports of com.codingame.<gamepkg>.<subpkg>/*.java
```

Every file declares `package engine`. There are no other Go packages inside
the engine directory.

Every Java source file in the simulation scope must have a Go counterpart with
the same base name converted to snake_case, prefixed with the full Java
package path. Do not merge a Java file into another Go file only because it
is small; one-file-per-source keeps roots searchable and audit-friendly.

Each Go file must include source provenance at the top: the source repository
name plus the Java file it was ported from. The header is provenance only — do
**not** paste the full Java source as one big block at the top of the file.

```go
// Package <pkg>
// Source: <RepoName>/src/main/java/com/codingame/game/<pkg>/File.java
package <pkg>
```

Instead, put **focused Java snippets next to the Go code that ports them**, one
snippet per type, method, enum, or constant group. The snippet is the audit
anchor for the Go code that immediately follows it, so it should be small
enough to read at a glance and contain only the lines that matter for the
adjacent Go. Drop fields, helpers, comments, and unrelated overloads that the
Go port does not reproduce.

The first line inside each Java block must identify the source file and line
range in `Java: <source-path>:<start>-<end>` format. Use line ranges from the
checked-in `source/` tree so grep, editor search, and code review can jump back
to the exact Java source quickly. Keep the line range tight to the snippet you
quoted:

```go
// Package grid
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java
package grid

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:81-88

public List<Coord> getNeighbours(Coord pos) {
    return Arrays
        .stream(Config.ADJACENCY)
        .map(delta -> getCoordNeighbour(pos, delta))
        .filter(Optional::isPresent)
        .map(Optional::get)
        .collect(Collectors.toList());
}
*/

// Neighbours returns all 4-directional valid neighbours.
func (g *Grid) Neighbours(pos Coord) []Coord {
    out := make([]Coord, 0, 4)
    for _, delta := range Adjacency4 {
        if n, ok := g.GetCoordNeighbour(pos, delta); ok {
            out = append(out, n)
        }
    }
    return out
}
```

A trivial single-symbol file (a small enum or a constants group) may keep one
snippet at the top, since the file itself is the one ported symbol. For any
file with multiple methods or types, prefer one snippet per Go symbol over a
single file-top blob — the goal is that whoever is reading the Go can see the
exact Java lines it came from without scrolling.

If a Java file has no simulation behavior to port but still belongs to the
simulation source set, keep an empty Go counterpart with only package,
provenance, and a short comment explaining why it is intentionally empty:

```go
// Package <pkg>
// Source: <RepoName>/src/main/java/com/codingame/game/<pkg>/ViewerOnly.java
package <pkg>

// Intentionally empty: ViewerOnly.java only wires CodinGame viewer animation
// state, which is outside the local simulation port.
```

## Registration

Two files wire the engine into the arena binary:

**1. `register.go`** — self-registers the factory via `init()`:

```go
package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

func init() {
    arena.Register(NewFactory())
}
```

**2. `games/game.go`** — add a blank import for the engine package:

```go
package games

import _ "github.com/mrsombre/codingame-arena/games/<engine>/engine"
```

Build with: `make build-arena`.

## Rules Documentation

Each game must have `games/<engine>/rules.md`. Port it from the upstream source
statement and source documentation, usually `source/<Repo>/config/statement_*.html.tpl`
plus any rule comments in the Java engine. Do not invent or simplify rules from
memory.

The rules document is the human checklist for the port. It should describe all
normative gameplay behavior that bots rely on:

- map generation constraints and symmetry
- entity setup, IDs, ownership, and starting state
- visibility and player input protocol
- valid commands, invalid-command handling, and messages
- action order inside a turn
- movement, collisions, gravity, abilities, cooldowns, and timers
- scoring, tie-breakers, game-end conditions, and turn limits
- league-level differences and configurable constants

When the Java source and the statement disagree, make the Go engine match the
Java source. Keep `rules.md` aligned with the implemented Java behavior and note
any known source-vs-statement divergence explicitly.

## Interfaces to Implement

The engine must provide `arena.GameFactory`, `arena.Referee`, and `arena.Player`.
See `internal/arena/interfaces.go` for the full signatures.

### GameFactory

```go
type GameFactory interface {
    Name() string
    NewGame(seed int64, options *viper.Viper) (Referee, []Player)
    MaxTurns() int
}
```

- `Name` returns the CLI game id, matching the directory name.
- `NewGame` creates a fresh game state from a seed and CLI options.
- `MaxTurns` returns the hard turn limit for this game.
- `options` is the viper instance carrying flags, config, env, and any
  `--key value` pairs not consumed by the arena core. Read game-specific
  keys via the standard viper API (`options.GetString`, `GetInt`, …).

### Referee

```go
type Referee interface {
    Init(players []Player)
    GlobalInfoFor(player Player) []string
    FrameInfoFor(player Player) []string
    ParsePlayerOutputs(players []Player)
    PerformGameUpdate(turn int)
    ResetGameTurnData()
    Ended() bool
    EndGame()
    OnEnd()
    ShouldSkipPlayerTurn(player Player) bool
    ActivePlayers(players []Player) int
}
```

Maps directly to the CodinGame SDK referee lifecycle:

1. `Init` — called once before the first turn; set up initial game state.
2. Each turn:
   - `ResetGameTurnData` — clear per-turn state.
   - `GlobalInfoFor` / `FrameInfoFor` — produce input lines sent to each player.
   - Player execution (handled by arena framework).
   - `ParsePlayerOutputs` — read and validate player commands.
   - `PerformGameUpdate` — apply game logic for the turn.
   - `Ended` — check if the game is over.
3. `EndGame` — compute final scores.
4. `OnEnd` — cleanup.

### Player

```go
type Player interface {
    GetIndex() int
    GetScore() int
    SetScore(int)
    IsDeactivated() bool
    Deactivate(reason string)
    DeactivationReason() string
    IsTimedOut() bool
    SetTimedOut(bool)
    GetExpectedOutputLines() int
    SendInputLine(string)
    ConsumeInputLines() []string
    GetOutputs() []string
    SetOutputs([]string)
    GetOutputError() error
    SetExecuteFunc(func() error)
    Execute() error
}
```

Implement as a concrete struct in the engine package. The arena framework
calls `SetExecuteFunc` and `Execute` to wire subprocess I/O — the engine
just calls `SendInputLine` to queue input and reads `GetOutputs` after execution.

### Optional Interfaces

The `Referee` may also implement these for richer output:

- `MetricsProvider` — `Metrics() []Metric` — game-specific per-match stats.
- `TraceProvider` — `SnapshotTurn(turn, players) json.RawMessage` — per-turn JSON state for replay/debug.
- `TurnEventProvider` — `TurnEvents(turn, players) []TurnEvent` — structured per-turn events.
- `RawScoresProvider` — `RawScores() [2]int` — scores before arena/endgame adjustments.

## Conversion Principles

The safest port is a natural Go implementation with obvious Java roots.
Do not mechanically translate Java syntax, but keep enough names and file
structure that every behavior can be traced back to the source quickly.

- Preserve simulation semantics before improving structure.
- Derive `rules.md` from the upstream source statement and keep the engine,
  rules document, and tests in sync.
- Keep Java class names as Go type names (`GridMaker`, `CommandManager`, `Bird`).
- Keep Java constants in the same Go files with the same names whenever they
  are part of simulation behavior.
- Export all ported types, fields, functions, and methods by default so tests
  and future engines can compose them directly.
- Convert Java accessor names into clear Go names while preserving the root
  (`getClosestTargets` -> `ClosestTargets`, `isValid` -> `IsValid`).
- Convert Java exceptions and invalid-input flow into Go errors or arena
  deactivation, not panics or copied exception hierarchies.
- Prefer small, explicit Go data structures over Java framework patterns,
  dependency injection, annotations, and viewer-only abstractions.

## Java-to-Go Porting Rules

### Parity First

Prefer a direct semantic port before optimizing or reshaping the design.
The first working version should preserve:

- turn order and lifecycle side effects
- random number generation behavior
- collection iteration order where Java code depends on it
- integer division, modulo, rounding, clamping, and overflow behavior
- command validation order and exact deactivation reasons
- output line ordering and formatting

After parity tests pass, optimize only the hotspots that matter.

### Names, Exports, and Searchability

Use Go naming, but keep the Java root name visible. Every source Java class
should have an obvious Go type or file to search for:

| Java | Go |
|---|---|
| `class GridMaker` | `type GridMaker struct` |
| `class CommandManager` | `type CommandManager struct` |
| `getClosestTargets` | `ClosestTargets` |
| `setType` | `SetType` |
| `isValid` | `IsValid` |
| `hasApple` | `HasApple` |
| `detectAirPockets` | `DetectAirPockets` |
| `fromIndex` | `LeagueRulesFromIndex` or `FromIndex` when unambiguous |

Export all ported fields, package functions, and methods. A Java private field
such as `birdByIDCache` should become `BirdByIDCache`, not hidden behind extra
getters. This keeps tests, serializers, and future engines able to compose the
ported engine without reflection or package hacks. Keep only tiny Go-only local
helpers private when they have no Java source root and no composition value.

Do not keep Java's `get` prefix just because the source used it. Keep `Get`
only when it is the natural Go operation name, such as `Grid.Get(c)` or
`Player.GetScore()` for an arena interface. For normal domain methods, drop
the prefix and export the root name.

### Constants and Static State

Port every Java simulation constant into the same Go file as the Java source
that declared it. Preserve the Java constant identifier exactly when practical,
including `SCREAMING_SNAKE_CASE`, because constants are common parity anchors:

```go
const (
    MIN_GRID_HEIGHT = 10
    MAX_GRID_HEIGHT = 24
    ASPECT_RATIO    = float32(1.8)
)
```

If a Go-style public name is useful for package ergonomics, add it as an alias
instead of replacing the Java-name constant during the first port:

```go
const MinGridHeight = MIN_GRID_HEIGHT
```

Do not leave mutable Java `static` fields as package globals unless the Java
value is truly process-wide. Prefer per-match state on `Game`, `Config`,
`GridMaker`, or another owner type so repeated simulations cannot leak state.
Keep those fields exported and named from the Java root (`MAP_MIN_WIDTH` may
become `MapMinWidth` inside `Config`, or stay `MAP_MIN_WIDTH` if exact source
matching is more useful).

### Type Mapping

| Java | Go |
|---|---|
| `class` with fields | exported `type ClassName struct` |
| `enum` | `type X int` + `iota` constants |
| `enum` with methods/fields | `iota` + lookup arrays indexed by enum value |
| `interface` | `interface` |
| `null` | zero value, sentinel, or `Has*` boolean flag |
| `Optional<T>` | `(T, bool)` return or `Has*` field |
| `List<T>` | `[]T` |
| `Map<K,V>` | `map[K]V` |
| `Set<T>` | `map[T]struct{}` |
| `LinkedHashMap` | `[]K` key order + `map[K]V`, or a slice of entries |
| `throws Exception` | `(T, error)` return |
| `Pattern.compile` | `regexp.MustCompile` in `init()` |
| constructor | `NewClassName(...) *ClassName` or value return |

### Iteration Order

Never rely on Go `map` iteration order. Java `HashMap` order is not guaranteed
either, but many CodinGame engines accidentally rely on the observed order of
`LinkedHashMap`, sorted lists, arrays, or insertion-ordered collections.

When order affects gameplay, keep it explicit:

```go
type OrderedMap[K comparable, V any] struct {
    Keys []K
    Data map[K]V
}
```

For deterministic output from a plain map, collect keys and sort them before
iteration.

### Value Types vs Pointers

Use value types for small immutable data (coordinates, directions, actions).
Benefits:
- `==` works natively (no `equals`/`hashCode` needed)
- Usable as map keys
- No nil checks, no allocation overhead

Use pointer receivers for mutable state owners (game state, grid, player).

### Enums

Java enums with ordinals become `iota`. If the enum has methods or associated
data, use fixed-size arrays indexed by the enum value:

```go
type MyEnum int

const (
    EnumA MyEnum = iota
    EnumB
    EnumC
)

// Lookup array — O(1), no switch needed.
var enumLabels = [3]string{"a", "b", "c"}

func (e MyEnum) Label() string { return enumLabels[e] }
```

### Nullable / Optional Fields

Use `Has*` boolean flags instead of pointers for optional value fields:

```go
type Command struct {
    TargetID    int
    HasTargetID bool  // instead of *int
    Position    Coord
    HasPosition bool  // instead of *Coord
}
```

### Sentinel Values

For "no result" returns, prefer `(T, bool)` for value types:

```go
func (g *Grid) Get(c Coord) (Tile, bool) {
    if outOfBounds { return Tile{}, false }
    return g.cells[idx], true
}
```

Use a sentinel only when it matches the Java shape better. Keep sentinels
immutable and never return a writable pointer to a shared package variable:

```go
var NoTile = Tile{Valid: false}

func (g *Grid) Get(c Coord) Tile {
    if outOfBounds { return NoTile }
    return g.cells[idx]
}
```

Callers check `tile.IsValid()` instead of `tile != null`. If callers must
mutate the returned tile, return an index or a pointer to real grid storage
only, and use `nil` for out-of-bounds.

### Collections as Flat Arrays

Java's `Map<Coord, Tile>` for 2D grids becomes a flat slice with index
arithmetic for O(1) cache-friendly access:

```go
type Grid struct {
    Width, Height int
    Cells         []Cell  // row-major: Cells[y*Width+x]
}
```

BFS visited sets use `map[Coord]struct{}` (works because value-type coords
are comparable).

### Init Patterns

Use `init()` for:
- Engine registration (`register.go`)
- Compiling regexes (`regexp.MustCompile`)
- Building lookup tables

Avoid `init()` for game state. Match setup should happen inside `NewGame` and
`Referee.Init` so repeated simulations do not leak state between games.

### Java Random

Go's `math/rand` does not match `java.util.Random`. If map generation or engine
logic must match the Java referee exactly, use or extend `internal/util/javarand`
for `java.util.Random` behavior and `internal/util/sha1prng` when the source
uses Java's SHA1PRNG provider. Expose the same source methods used by the Java
engine, such as `NextInt(bound)`, `NextDouble`, `NextBoolean`, or
`NextIntRange(origin, bound)`.

Pay attention to inclusive/exclusive bounds. Java `nextInt(n)` returns values in
`[0, n)`.

### Numeric Semantics

Java `int` is always 32-bit. Go `int` is platform-sized. Use `int` for normal
arena logic when ranges are small, but use `int32`, `int64`, or explicit casts
when Java overflow or exact serialized values matter.

Important differences:

- Java integer division truncates toward zero, same as Go.
- Java `%` keeps the sign of the dividend, same as Go.
- Java `Math.floorMod` is not the same as `%` for negative values.
- Java `Math.round`, `floor`, `ceil`, and casts should be ported explicitly
  with `math` helpers and tests around boundary values.

### Sorting and Priority Queues

Port Java comparators directly and test tie-breakers. If Java uses a stable sort
or preserves insertion order for equal elements, use `sort.SliceStable` or add
an explicit original index tie-breaker.

Java `PriorityQueue` does not guarantee ordering among equal-priority elements.
If equal priorities can affect gameplay, preserve whatever the original engine
observes with tests instead of assuming Go heap behavior will match.

### Error Handling

Do not port Java exception classes mechanically. Java exceptions become Go
errors at the boundary where failure is expected: parsers, validators, map
builders, and command interpreters. Custom error types implement the `error`
interface:

```go
type ParseError struct{ Message string }
func (e *ParseError) Error() string { return e.Message }
```

Parse functions return `(T, error)` instead of throwing.

For invalid player output, prefer the arena flow over propagating exceptions:
return the parse error from the parser, then let the command manager deactivate
the player with the same message and order as Java. Use `panic` only for
impossible programmer errors or malformed hardcoded fixtures, never for normal
gameplay validation.

## Rule Coverage and Parity Tests

The port is not complete until tests cover `rules.md`. Treat every normative
sentence in the rules document as a test requirement. Use unit tests for isolated
mechanics and acceptance tests for full turn flows.

Minimum test expectations:

- Value and domain tests cover coordinates, directions, enum mappings, constants,
  parsers, validators, grids, pathfinding, and RNG helpers.
- Unit tests cover every edge case that can be isolated without a full match:
  wrapping, neighbour order, rounding, sorting tie-breakers, cooldown decrement,
  timer expiry, sentinel values, command parsing, and exact error text.
- Acceptance tests cover complete rules-visible behavior: initial setup, input
  serialization, simultaneous actions, movement, collisions, deaths/removals,
  item collection, visibility, scoring, game-end conditions, and league changes.
- Invalid-output tests cover missing output, malformed commands, illegal targets,
  duplicate commands, unknown entity IDs, commands for dead entities, and any
  source-specific validation order.
- Seed/parity tests cover map generation and any random setup against known Java
  outputs or recorded fixtures for representative seeds.
- Regression tests are required for any discovered source-vs-statement mismatch.

Maintain a traceable relationship between `rules.md` and tests. Test names should
quote or closely mirror the rule they cover, for example
`TestSpeedSubTurnDelivers2Steps` for "SPEED allows a pac to move by 2 steps".
When adding or changing a rule, add or update the corresponding unit or
acceptance test in the same change.

Do not accept "close enough" behavior. If a test has to choose between idiomatic
Go behavior and Java behavior, Java behavior wins.

## File Porting Order

Port the rules document first, then port leaf packages and work up the
dependency graph:

1. **Rules document** — port `rules.md` from the source statement/docs
2. **Value types** — coordinates, directions, enums (no deps)
3. **Domain primitives** — tiles, cells, board/grid (uses value types)
4. **Action parsing** — player command parsing (uses domain types)
5. **Game entities** — units, items, player state (uses domain + grid)
6. **Game logic** — referee, command manager, serializer (uses everything)
7. **Map generation** — board/map builder (uses grid + RNG)
8. **Wire `engine.go`** — implement `NewGame` with actual game logic
9. **Rule coverage** — add unit and acceptance tests for every rule

## Checklist for a New Engine

- [ ] Create `games/<engine>/engine/` directory
- [ ] Create `games/<engine>/rules.md` from source statement/docs
- [ ] Port leaf packages first (value types, domain primitives)
- [ ] Add `// Source:` comments to every ported file
- [ ] Preserve Java simulation constants in the same Go files with the same names
- [ ] Keep Java class roots as exported Go type names
- [ ] Export ported fields, functions, and methods for composition
- [ ] Create `engine.go` with `GameFactory` implementation
- [ ] Create `register.go` with `init()` registration
- [ ] Add the engine blank import to `games/game.go`
- [ ] Implement `Referee` interface (match the CodinGame SDK lifecycle)
- [ ] Implement `Player` interface (concrete struct with I/O buffer fields)
- [ ] Optionally implement `MetricsProvider`, `TraceProvider`, `TurnEventProvider`, `RawScoresProvider`
- [ ] Add unit tests for all isolated rule mechanics and Java edge cases
- [ ] Add acceptance tests for all cross-turn rules in `rules.md`
- [ ] Add seed/parity fixtures for random generation or setup behavior
- [ ] Verify: `make test-games`
- [ ] Verify: `make test-arena`
- [ ] Verify: `make lint-arena`
- [ ] Verify: `make build-arena`
