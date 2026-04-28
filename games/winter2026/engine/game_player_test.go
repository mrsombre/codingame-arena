package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayerAddMarkCapsAtFour(t *testing.T) {
	p := NewPlayer(0)
	p.Init()

	for i := 0; i < 4; i++ {
		assert.True(t, p.AddMark(Coord{X: i, Y: 0}), "mark #%d", i)
	}
	assert.False(t, p.AddMark(Coord{X: 99, Y: 0}), "5th mark rejected")
	assert.Len(t, p.Marks, 4)
}

func TestPlayerBirdByIDReturnsMatch(t *testing.T) {
	p := NewPlayer(0)
	p.Init()
	b1 := NewBird(1, p)
	b2 := NewBird(2, p)
	p.Birds = []*Bird{b1, b2}

	assert.Same(t, b1, p.BirdByID(1))
	assert.Same(t, b2, p.BirdByID(2))
	assert.Nil(t, p.BirdByID(99))
}

func TestPlayerResetClearsBirdDirectionAndMarks(t *testing.T) {
	p := NewPlayer(0)
	p.Init()
	b := NewBird(0, p)
	b.Direction = DirNorth
	b.HasMove = true
	b.Message = "hello"
	b.HasMsg = true
	p.Birds = append(p.Birds, b)
	p.AddMark(Coord{X: 1, Y: 1})

	p.Reset()

	assert.Equal(t, DirUnset, b.Direction)
	assert.False(t, b.HasMove)
	assert.Equal(t, "", b.Message)
	assert.False(t, b.HasMsg)
	assert.Empty(t, p.Marks)
}

func TestPlayerDeactivateAndScore(t *testing.T) {
	p := NewPlayer(1)
	assert.False(t, p.IsDeactivated())
	p.Deactivate("boom")
	assert.True(t, p.IsDeactivated())
	assert.Equal(t, "boom", p.DeactivationReason())

	p.SetScore(42)
	assert.Equal(t, 42, p.GetScore())
}

func TestPlayerIndexAndTimedOut(t *testing.T) {
	p := NewPlayer(1)
	assert.Equal(t, 1, p.GetIndex())
	assert.False(t, p.IsTimedOut())
	p.SetTimedOut(true)
	assert.True(t, p.IsTimedOut())
}

func TestPlayerIOBuffers(t *testing.T) {
	p := NewPlayer(0)
	p.SendInputLine("hi")
	p.SendInputLine("there")
	assert.Equal(t, []string{"hi", "there"}, p.ConsumeInputLines())
	// After consume, the buffer is emptied.
	assert.Empty(t, p.ConsumeInputLines())

	p.SetOutputs([]string{"out"})
	assert.Equal(t, []string{"out"}, p.GetOutputs())
}
