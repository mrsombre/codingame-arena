package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectionCoordsAndAliases(t *testing.T) {
	assert.Equal(t, Coord{0, -1}, NORTH.Coord())
	assert.Equal(t, Coord{1, 0}, EAST.Coord())
	assert.Equal(t, Coord{0, 1}, SOUTH.Coord())
	assert.Equal(t, Coord{-1, 0}, WEST.Coord())
	assert.Equal(t, Coord{0, 0}, UNSET.Coord())

	assert.Equal(t, "N", NORTH.String())
	assert.Equal(t, "E", EAST.String())
	assert.Equal(t, "S", SOUTH.String())
	assert.Equal(t, "W", WEST.String())
	assert.Equal(t, "X", UNSET.String())
}

func TestDirectionOpposite(t *testing.T) {
	assert.Equal(t, SOUTH, NORTH.Opposite())
	assert.Equal(t, WEST, EAST.Opposite())
	assert.Equal(t, NORTH, SOUTH.Opposite())
	assert.Equal(t, EAST, WEST.Opposite())
	assert.Equal(t, UNSET, UNSET.Opposite())
}

func TestDirectionFromCoord(t *testing.T) {
	assert.Equal(t, NORTH, DirectionFromCoord(Coord{0, -1}))
	assert.Equal(t, WEST, DirectionFromCoord(Coord{-1, 0}))
	assert.Equal(t, UNSET, DirectionFromCoord(Coord{1, 1}))
	assert.Equal(t, UNSET, DirectionFromCoord(Coord{0, 0}))
}

func TestDirectionFromAlias(t *testing.T) {
	assert.Equal(t, NORTH, DirectionFromAlias("N"))
	assert.Equal(t, EAST, DirectionFromAlias("E"))
	assert.Equal(t, SOUTH, DirectionFromAlias("S"))
	assert.Equal(t, WEST, DirectionFromAlias("W"))
	assert.Panics(t, func() { DirectionFromAlias("X") })
}
