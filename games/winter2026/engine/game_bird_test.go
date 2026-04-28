package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBirdDefaults(t *testing.T) {
	owner := NewPlayer(0)
	b := NewBird(7, owner)
	assert.Equal(t, 7, b.ID)
	assert.Same(t, owner, b.Owner)
	assert.True(t, b.Alive)
	assert.Equal(t, DirUnset, b.Direction)
	assert.NotNil(t, b.Body)
	assert.Len(t, b.Body, 0)
}

func TestBirdHeadPosReturnsFirstBodyCell(t *testing.T) {
	b := NewBird(0, nil)
	b.Body = []Coord{{X: 5, Y: 5}, {X: 5, Y: 6}}
	assert.Equal(t, Coord{X: 5, Y: 5}, b.HeadPos())
}

func TestBirdFacingRequiresTwoSegments(t *testing.T) {
	b := NewBird(0, nil)
	assert.Equal(t, DirUnset, b.Facing(), "empty body")

	b.Body = []Coord{{X: 5, Y: 5}}
	assert.Equal(t, DirUnset, b.Facing(), "single segment")
}

func TestBirdFacingComputesFromHeadAndNeck(t *testing.T) {
	b := NewBird(0, nil)
	// Head is one cell north of neck → facing UP.
	b.Body = []Coord{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 7}}
	assert.Equal(t, DirNorth, b.Facing())

	// Head east of neck → facing EAST.
	b.Body = []Coord{{X: 6, Y: 5}, {X: 5, Y: 5}}
	assert.Equal(t, DirEast, b.Facing())

	// Head south of neck.
	b.Body = []Coord{{X: 5, Y: 6}, {X: 5, Y: 5}}
	assert.Equal(t, DirSouth, b.Facing())
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
