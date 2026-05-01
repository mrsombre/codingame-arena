package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn, attached to trace output for replay viewers
// and analyzers. Each type has a typed meta payload defined below; the wire
// shape is `{"type": <const>, "meta": {...}}`.
//
// Naming follows the analyze-report family prefix: HIT_* for non-fatal head
// intersections (lose 1 segment), DEAD_FALL for fall removals, DEAD for
// beheading removals (the cause subtype lives in BirdDeathMeta.Cause so the
// wire shape stays one event type).
const (
	TraceEat      = "EAT"
	TraceHitSelf  = "HIT_SELF"
	TraceHitWall  = "HIT_WALL"
	TraceHitEnemy = "HIT_ENEMY"
	TraceDead     = "DEAD"
	TraceDeadFall = "DEAD_FALL"
)

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

func (g *Game) trace(t arena.TurnTrace) {
	g.traces = append(g.traces, t)
}
