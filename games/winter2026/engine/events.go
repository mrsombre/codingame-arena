package engine

import (
	"strconv"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Event labels emitted per turn, attached to trace output for replay viewers.
const (
	EventEat      = "EAT"
	EventHitWall  = "HIT_WALL"
	EventHitSelf  = "HIT_ITSELF"
	EventHitEnemy = "HIT_ENEMY"
	EventDead     = "DEAD"
	EventFall     = "FALL"
)

func eventBirdCoordPayload(birdID int, c grid.Coord) string {
	return strconv.Itoa(birdID) + " " + strconv.Itoa(c.X) + "," + strconv.Itoa(c.Y)
}

func eventBirdPayload(birdID int) string {
	return strconv.Itoa(birdID)
}

func (g *Game) emit(label, payload string) {
	g.events = append(g.events, arena.TurnEvent{Label: label, Payload: payload})
}
