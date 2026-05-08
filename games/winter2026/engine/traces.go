package engine

import (
	"encoding/json"
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn, attached to trace output for replay viewers
// and analyzers. Each type has a typed payload defined below; the wire
// shape is `{"type": <const>, "data": {...}}`.
//
// Two categories live in this enum:
//
//   - Commands (MOVE/WAIT/MARK) — emitted during command parsing, one per
//     accepted token in the bot's output line. Carry whatever the bot asked
//     for (direction, debug message, mark coord); fire only when the engine
//     accepted the command.
//   - Events (EAT/HIT_*/DEAD/DEAD_FALL) — emitted by the four DoMoves/DoEats/
//     DoBeheadings/DoFalls phases of PerformGameUpdate, recording engine
//     consequences of those commands plus inertial movement.
//
// Naming follows the analyze-report family prefix: HIT_* for non-fatal head
// intersections (lose 1 segment), DEAD_FALL for fall removals, DEAD for
// beheading removals (the cause subtype lives in BirdDeathMeta.Cause so the
// wire shape stays one event type).
const (
	// Commands.
	TraceMove = "MOVE"
	TraceWait = "WAIT"
	TraceMark = "MARK"

	// Events.
	TraceEat      = "EAT"
	TraceHitSelf  = "HIT_SELF"
	TraceHitWall  = "HIT_WALL"
	TraceHitEnemy = "HIT_ENEMY"
	TraceDead     = "DEAD"
	TraceDeadFall = "DEAD_FALL"
)

// MoveMeta is the meta for MOVE command traces — one per accepted UP/DOWN/
// LEFT/RIGHT command. Direction is the bot-issued cardinal name (`"UP"` /
// `"DOWN"` / `"LEFT"` / `"RIGHT"`) so the trace mirrors the wire token, not
// the internal Direction enum's NESW alias. Debug carries the optional
// trailing free-text the bot appended (truncated to 48 chars by
// Bird.SetMessage); empty when the bot sent a bare directional command.
//
// Trace fires only when the engine accepted the command — bird exists,
// alive, has not already moved this turn, and is not trying to reverse.
// Soft-rejected moves (unknown bird id, dead bird, double-move, backwards)
// land in the summary as errors and emit no trace, mirroring spring2020's
// "errored commands drop the message entirely" convention.
type MoveMeta struct {
	Bird      int    `json:"bird"`
	Direction string `json:"direction"`
	Debug     string `json:"debug,omitempty"`
}

// MarkMeta is the meta for MARK command traces — one per accepted MARK x y,
// up to four per side per turn (engine cap from Player.AddMark). Coord is
// `[x, y]`; not associated with any specific bird since markers are a pure
// viewer-side debugging affordance. The 5th+ MARK in a turn is rejected
// with a summary error and emits no trace.
type MarkMeta struct {
	Coord [2]int `json:"coord"`
}

// BirdMeta is the meta for trace events whose only subject is one bird. Used
// as a fallback decode shape: any meta that carries a `bird` field decodes
// into BirdMeta successfully (extra fields are ignored), so analyzers can
// extract bird ownership without knowing the full meta type.
type BirdMeta struct {
	Bird int `json:"bird"`
}

// DEAD-event cause values. WALL > ENEMY > SELF priority — order matches the
// engine's intersection check sequence and the HIT_* trace emission order, so
// a single cause is always recorded even when intersections could overlap
// (only ENEMY+SELF can co-occur; WALL precludes any body in the same cell).
const (
	DeathCauseWall  = "WALL"
	DeathCauseEnemy = "ENEMY"
	DeathCauseSelf  = "SELF"
)

// BirdDeathMeta is the meta for DEAD events. Cause records which intersection
// killed the snake, so analyzers can split deaths by hazard (DEAD_WALL,
// DEAD_ENEMY, DEAD_SELF) — different causes point at different bot bugs
// (navigation vs tactics vs planning).
type BirdDeathMeta struct {
	Bird  int    `json:"bird"`
	Cause string `json:"cause"`
}

// BirdCoordMeta is the meta for trace events that include the bird's head
// position at the moment of the event (EAT, HIT_*).
type BirdCoordMeta struct {
	Bird  int    `json:"bird"`
	Coord [2]int `json:"coord"`
}

// BirdSegmentsMeta is the meta for DEAD_FALL events. Segments records the
// bird's body length the moment it fell off the grid, which equals the
// segments lost to that fall — needed because a fall kills a snake regardless
// of length, so the cost is variable (unlike DEAD beheadings, which always
// cost 3).
type BirdSegmentsMeta struct {
	Bird     int `json:"bird"`
	Segments int `json:"segments"`
}

func coordPair(c Coord) [2]int {
	return [2]int{c.X, c.Y}
}

// tracePlayer appends a player-owned event into the per-player slot.
// Slots out of [0, 1] are silently dropped (defensive).
func (g *Game) tracePlayer(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(g.traces) {
		return
	}
	g.traces[playerIdx] = append(g.traces[playerIdx], t)
}

// TraceTurnState is the winter2026-owned per-turn payload written into
// TraceTurn.State. Captures the apple count and a per-side roster of
// alive snakes, so analyzers and viewers don't have to re-parse
// `gameInput`'s body strings just to recover head/length.
//
// Sampled by the arena runner *after* command parsing and *before*
// PerformGameUpdate (the standard TraceTurnDecorator hook point), so
// values reflect the state players saw when choosing actions: the new
// turn's heads/sizes have not been applied yet, and any snake removed
// on the previous turn (DEAD or DEAD_FALL) is already absent.
type TraceTurnState struct {
	Apples int             `json:"apples"`
	Snakes [2][]TraceSnake `json:"snakes"`
}

// TraceSnake is the per-snake entry in TraceTurnState.Snakes. Bucketing
// is positional: the outer index is the match side (same convention as
// `players` / `traces[]`), so the snake's owning side is recoverable
// without an explicit owner field. Only **alive** snakes appear here —
// once a snake is removed it never re-enters this list, mirroring the
// gameInput bird block.
type TraceSnake struct {
	ID   int    `json:"id"`
	Size int    `json:"size"`
	Head [2]int `json:"head"`
}

// DecorateTraceTurn marshals the per-turn TraceTurnState payload onto
// TraceTurn.State. Matches arena.TraceTurnDecorator.
func (g *Game) DecorateTraceTurn(_ int, _ []arena.Player) json.RawMessage {
	state := TraceTurnState{
		Apples: len(g.Grid.Apples),
		Snakes: g.traceSnakes(),
	}
	raw, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return raw
}

// traceSnakes builds the per-side snake roster for TraceTurnState, in
// snake-id ascending order within each slot. Dead birds are skipped —
// the gameInput's live-birds block uses the same filter, so the state
// stays consistent with what bots saw.
func (g *Game) traceSnakes() [2][]TraceSnake {
	var snakes [2][]TraceSnake
	for _, p := range g.Players {
		side := p.GetIndex()
		if side < 0 || side >= len(snakes) {
			continue
		}
		for _, b := range p.Birds {
			if !b.Alive {
				continue
			}
			snakes[side] = append(snakes[side], TraceSnake{
				ID:   b.ID,
				Size: len(b.Body),
				Head: coordPair(b.HeadPos()),
			})
		}
		sort.SliceStable(snakes[side], func(a, c int) bool {
			return snakes[side][a].ID < snakes[side][c].ID
		})
	}
	return snakes
}
