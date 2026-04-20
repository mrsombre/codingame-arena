// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java
package engine

import "github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"

type Bird struct {
	ID        int
	Body      []grid.Coord
	Owner     *Player
	Alive     bool
	Direction grid.Direction
	HasMove   bool
	Message   string
	HasMsg    bool
}

func NewBird(id int, owner *Player) *Bird {
	return &Bird{
		ID:        id,
		Owner:     owner,
		Alive:     true,
		Direction: grid.DirUnset,
		Body:      make([]grid.Coord, 0),
	}
}

func (b *Bird) HeadPos() grid.Coord {
	return b.Body[0]
}

func (b *Bird) Facing() grid.Direction {
	if len(b.Body) < 2 {
		return grid.DirUnset
	}
	return grid.DirectionFromCoord(grid.Coord{
		X: b.Body[0].X - b.Body[1].X,
		Y: b.Body[0].Y - b.Body[1].Y,
	})
}

func (b *Bird) IsAlive() bool {
	return b.Alive
}

func (b *Bird) SetMessage(message string) {
	b.Message = message
	b.HasMsg = true
	if len(message) > 48 {
		b.Message = message[:46] + "..."
	}
}

func (b *Bird) HasMessage() bool {
	return b.HasMsg
}
