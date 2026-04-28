package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

