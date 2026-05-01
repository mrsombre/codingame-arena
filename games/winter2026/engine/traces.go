package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn, attached to trace output for replay viewers
// and analyzers. Each type has a typed meta payload defined below; the wire
// shape is `{"type": <const>, "meta": {...}}`.
const (
	TraceEat      = "EAT"
	TraceHitWall  = "HIT_WALL"
	TraceHitSelf  = "HIT_ITSELF"
	TraceHitEnemy = "HIT_ENEMY"
	TraceDead     = "DEAD"
	TraceFall     = "FALL"
)

// BirdMeta is the meta for trace events whose only subject is one bird (DEAD).
type BirdMeta struct {
	Bird int `json:"bird"`
}

// BirdCoordMeta is the meta for trace events that include the bird's head
// position at the moment of the event (EAT, HIT_*).
type BirdCoordMeta struct {
	Bird  int    `json:"bird"`
	Coord [2]int `json:"coord"`
}

// BirdSegmentsMeta is the meta for FALL events. Segments records the bird's
// body length the moment it fell off the grid, which equals the segments lost
// to that fall — needed because a fall kills a snake regardless of length, so
// the cost is variable (unlike DEAD beheadings, which always cost 3).
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
