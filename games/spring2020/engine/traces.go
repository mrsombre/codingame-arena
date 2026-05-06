package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn, attached to trace output for replay viewers
// and analyzers. Each type has a typed payload defined below; the wire
// shape is `{"type": <const>, "data": {...}}`.
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

// tracePlayer appends a player-owned event into the per-player slot.
// Slots out of [0, 1] are silently dropped (defensive).
func (g *Game) tracePlayer(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(g.traces) {
		return
	}
	g.traces[playerIdx] = append(g.traces[playerIdx], t)
}

// traceBoth mirrors a cross-owner event into both player slots so each side
// sees its involvement (e.g. COLLIDE_ENEMY).
func (g *Game) traceBoth(t arena.TurnTrace) {
	for i := range g.traces {
		g.traces[i] = append(g.traces[i], t)
	}
}
