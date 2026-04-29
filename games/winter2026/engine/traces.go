package engine

import (
	"strconv"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace labels emitted per turn, attached to trace output for replay viewers.
const (
	TraceEat      = "EAT"
	TraceHitWall  = "HIT_WALL"
	TraceHitSelf  = "HIT_ITSELF"
	TraceHitEnemy = "HIT_ENEMY"
	TraceDead     = "DEAD"
	TraceFall     = "FALL"
)

func traceBirdCoordPayload(birdID int, c Coord) string {
	return strconv.Itoa(birdID) + " " + strconv.Itoa(c.X) + "," + strconv.Itoa(c.Y)
}

func traceBirdPayload(birdID int) string {
	return strconv.Itoa(birdID)
}

func (g *Game) trace(label, payload string) {
	g.traces = append(g.traces, arena.TurnTrace{Label: label, Payload: payload})
}
