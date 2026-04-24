package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
)

func TestNewBirdDefaults(t *testing.T) {
	owner := NewPlayer(0)
	b := NewBird(7, owner)
	assert.Equal(t, 7, b.ID)
	assert.Same(t, owner, b.Owner)
	assert.True(t, b.Alive)
	assert.Equal(t, grid.DirUnset, b.Direction)
	assert.NotNil(t, b.Body)
	assert.Len(t, b.Body, 0)
}

func TestBirdHeadPosReturnsFirstBodyCell(t *testing.T) {
	b := NewBird(0, nil)
	b.Body = []grid.Coord{{X: 5, Y: 5}, {X: 5, Y: 6}}
	assert.Equal(t, grid.Coord{X: 5, Y: 5}, b.HeadPos())
}

func TestBirdFacingRequiresTwoSegments(t *testing.T) {
	b := NewBird(0, nil)
	assert.Equal(t, grid.DirUnset, b.Facing(), "empty body")

	b.Body = []grid.Coord{{X: 5, Y: 5}}
	assert.Equal(t, grid.DirUnset, b.Facing(), "single segment")
}

func TestBirdFacingComputesFromHeadAndNeck(t *testing.T) {
	b := NewBird(0, nil)
	// Head is one cell north of neck → facing UP.
	b.Body = []grid.Coord{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 7}}
	assert.Equal(t, grid.DirNorth, b.Facing())

	// Head east of neck → facing EAST.
	b.Body = []grid.Coord{{X: 6, Y: 5}, {X: 5, Y: 5}}
	assert.Equal(t, grid.DirEast, b.Facing())

	// Head south of neck.
	b.Body = []grid.Coord{{X: 5, Y: 6}, {X: 5, Y: 5}}
	assert.Equal(t, grid.DirSouth, b.Facing())
}

func TestBirdSetMessageTruncates(t *testing.T) {
	b := NewBird(0, nil)
	b.SetMessage("hi")
	assert.True(t, b.HasMessage())
	assert.Equal(t, "hi", b.Message)

	long := strings.Repeat("x", 60)
	b.SetMessage(long)
	assert.True(t, strings.HasSuffix(b.Message, "..."))
	assert.Equal(t, 49, len(b.Message))
}

func TestBirdIsAlive(t *testing.T) {
	b := NewBird(0, nil)
	assert.True(t, b.IsAlive())
	b.Alive = false
	assert.False(t, b.IsAlive())
}
