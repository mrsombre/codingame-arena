package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

func TestMovementResolutionTracksMovedAndBlocked(t *testing.T) {
	mr := NewMovementResolution()

	a := &Pacman{ID: 1}
	b := &Pacman{ID: 2}
	c := &Pacman{ID: 3}

	mr.AddMoved(a)
	mr.AddBlocked(b)
	mr.BlockedBy[b] = c

	assert.Equal(t, []*Pacman{a}, mr.MovedPacmen)
	assert.Equal(t, []*Pacman{b}, mr.BlockedPacmen)
	assert.Equal(t, c, mr.BlockerOf(b))
	assert.Nil(t, mr.BlockerOf(a))
}

func TestBumpCoupleSymmetricUniqueness(t *testing.T) {
	from := grid.Coord{X: 0, Y: 0}
	to := grid.Coord{X: 1, Y: 0}

	first := BumpCouple{From: from, FromBlocker: to, To: to, Distance: 1}
	// Symmetric variant — same pair swapped.
	second := BumpCouple{From: to, FromBlocker: from, To: to, Distance: 1}
	// Different pair.
	other := BumpCouple{From: grid.Coord{X: 2, Y: 0}, FromBlocker: grid.Coord{X: 3, Y: 0}}

	list := []BumpCouple{}
	list = AddUniqueBumpCouple(list, first)
	assert.Len(t, list, 1)

	list = AddUniqueBumpCouple(list, second)
	assert.Len(t, list, 1, "symmetric pair must not double-insert")

	list = AddUniqueBumpCouple(list, other)
	assert.Len(t, list, 2)
}
