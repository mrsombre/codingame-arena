package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrainStateHasNoPreviousState(t *testing.T) {
	state := NewTrainState(Coord{2, 3})

	assert.Equal(t, Coord{2, 3}, state.Coord)
	assert.Nil(t, state.Prev)
}

func TestNewTrainStateFromKeepsTheBackPointer(t *testing.T) {
	first := NewTrainState(Coord{0, 0})
	second := NewTrainStateFrom(Coord{1, 0}, first)

	assert.Equal(t, Coord{1, 0}, second.Coord)
	assert.Same(t, first, second.Prev)
}

// The stray comma is Java's — the toString concatenation opens with one.
func TestTrainStateStringMatchesJava(t *testing.T) {
	assert.Equal(t, "State{, coord=(2, 3)}", NewTrainState(Coord{2, 3}).String())
}
