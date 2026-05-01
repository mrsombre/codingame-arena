package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn, attached to trace output for replay viewers
// and analyzers. Each type has a typed meta payload defined below; the wire
// shape is `{"type": <const>, "meta": {...}}`.
const (
	TraceEat          = "EAT"
	TraceKilled       = "KILLED"
	TraceSpeed        = "SPEED"
	TraceSwitch       = "SWITCH"
	TraceCollideSelf  = "COLLIDE_SELF"
	TraceCollideEnemy = "COLLIDE_ENEMY"
)

// PacMeta is the meta for trace events whose only subject is one pac (SPEED,
// COLLIDE_SELF, COLLIDE_ENEMY).
type PacMeta struct {
	Pac int `json:"pac"`
}

// EatMeta is the meta for the EAT trace. Cost > 1 marks a super pellet.
type EatMeta struct {
	Pac   int    `json:"pac"`
	Coord [2]int `json:"coord"`
	Cost  int    `json:"cost"`
}

// KilledMeta is the meta for the KILLED trace. Pac is the dead pac (subject);
// Killer is the global pac ID that caused the kill.
type KilledMeta struct {
	Pac    int    `json:"pac"`
	Coord  [2]int `json:"coord"`
	Killer int    `json:"killer"`
}

// SwitchMeta is the meta for the SWITCH trace.
type SwitchMeta struct {
	Pac  int    `json:"pac"`
	Type string `json:"type"`
}

func coordPair(c Coord) [2]int {
	return [2]int{c.X, c.Y}
}

func (g *Game) trace(t arena.TurnTrace) {
	g.traces = append(g.traces, t)
}
