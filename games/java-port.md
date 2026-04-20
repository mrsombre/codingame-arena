# Porting a CodinGame Java Engine to Go

Guide for porting any CodinGame multiplayer game engine from Java to Go
inside `arena/internal/game/<engine>/`.

Commands and import paths below assume they are run from the `arena/` module
directory unless stated otherwise.

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

## Directory Layout

Mirror the Java package structure as Go sub-packages:

```
internal/game/<engine>/
  engine.go          # GameFactory implementation
  register.go        # init() → arena.Register(NewFactory())
  <domain>/          # game-specific domain packages
```

Each Go file must include source provenance at the top: the source repository
name plus every Java file it was ported from. Most files map one-to-one, but
small Java files may be merged when that keeps the Go package clearer.

```go
// Package <pkg>
// Source: <RepoName>/src/main/java/com/codingame/game/<pkg>/File.java
package <pkg>
```

For merged ports, repeat `// Source:` once per Java file:

```go
// Package action
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionException.java
package action
```

## Build Tag Registration

Three files wire the engine into the arena binary:

**1. `register.go`** — self-registers the factory via `init()`:

```go
package <engine>

import "github.com/mrsombre/codingame-arena/internal/arena"

func init() {
    arena.Register(NewFactory())
}
```

**2. `internal/game/<engine>.go`** — build-tagged blank import:

```go
//go:build <engine>

package game

import _ "github.com/mrsombre/codingame-arena/internal/game/<engine>"
```

**3. `internal/game/game.go`** — already exists, no changes needed.

Build with: `go build -tags <engine> ./cmd/arena`

## Interfaces to Implement

The engine must provide `arena.GameFactory`, `arena.Referee`, and `arena.Player`.
See `internal/arena/interfaces.go` for the full signatures.

### GameFactory

```go
type GameFactory interface {
    NewGame(seed int64, options map[string]string) (Referee, []Player)
    MaxTurns() int
}
```

- `NewGame` creates a fresh game state from a seed and CLI options.
- `MaxTurns` returns the hard turn limit for this game.
- `options` contains any `--key value` flags not consumed by the arena core.

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

### Type Mapping

| Java | Go |
|---|---|
| `class` with fields | `struct` (value type if immutable, pointer receiver if mutable) |
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

### Iteration Order

Never rely on Go `map` iteration order. Java `HashMap` order is not guaranteed
either, but many CodinGame engines accidentally rely on the observed order of
`LinkedHashMap`, sorted lists, arrays, or insertion-ordered collections.

When order affects gameplay, keep it explicit:

```go
type OrderedMap[K comparable, V any] struct {
    keys []K
    data map[K]V
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
var NoTile = Tile{valid: false}

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
    cells         []Cell  // row-major: cells[y*Width+x]
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
logic must match the Java referee exactly, port Java's 48-bit LCG and expose the
same methods used by the source engine, such as `nextInt(bound)`, `nextDouble`,
and `nextBoolean`.

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

Java exceptions become Go errors. Custom error types implement the `error`
interface:

```go
type ParseError struct{ Message string }
func (e *ParseError) Error() string { return e.Message }
```

Parse functions return `(T, error)` instead of throwing.

## File Porting Order

Port leaf packages first, then work up the dependency graph:

1. **Value types** — coordinates, directions, enums (no deps)
2. **Domain primitives** — tiles, cells, board/grid (uses value types)
3. **Action parsing** — player command parsing (uses domain types)
4. **Game entities** — units, items, player state (uses domain + grid)
5. **Game logic** — referee, command manager, serializer (uses everything)
6. **Map generation** — board/map builder (uses grid + RNG)
7. **Wire `engine.go`** — implement `NewGame` with actual game logic

## Checklist for a New Engine

- [ ] Create `internal/game/<engine>/` directory
- [ ] Port leaf packages first (value types, domain primitives)
- [ ] Add `// Source:` comments to every ported file
- [ ] Create `engine.go` with `GameFactory` implementation
- [ ] Create `register.go` with `init()` registration
- [ ] Create `internal/game/<engine>.go` with build tag + blank import
- [ ] Implement `Referee` interface (match the CodinGame SDK lifecycle)
- [ ] Implement `Player` interface (concrete struct with I/O buffer fields)
- [ ] Optionally implement `MetricsProvider`, `TraceProvider`, `TurnEventProvider`
- [ ] Verify: `go build -tags <engine> ./cmd/arena`
- [ ] Verify: `go test -tags <engine> ./internal/game/<engine>/...`
