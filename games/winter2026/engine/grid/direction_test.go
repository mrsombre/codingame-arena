package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectionCoord(t *testing.T) {
	assert.Equal(t, Coord{X: 0, Y: -1}, DirNorth.Coord())
	assert.Equal(t, Coord{X: 1, Y: 0}, DirEast.Coord())
	assert.Equal(t, Coord{X: 0, Y: 1}, DirSouth.Coord())
	assert.Equal(t, Coord{X: -1, Y: 0}, DirWest.Coord())
	assert.Equal(t, Coord{}, DirUnset.Coord())
}

func TestDirectionOpposite(t *testing.T) {
	assert.Equal(t, DirSouth, DirNorth.Opposite())
	assert.Equal(t, DirNorth, DirSouth.Opposite())
	assert.Equal(t, DirWest, DirEast.Opposite())
	assert.Equal(t, DirEast, DirWest.Opposite())
	assert.Equal(t, DirUnset, DirUnset.Opposite())
}

func TestDirectionString(t *testing.T) {
	assert.Equal(t, "N", DirNorth.String())
	assert.Equal(t, "E", DirEast.String())
	assert.Equal(t, "S", DirSouth.String())
	assert.Equal(t, "W", DirWest.String())
	assert.Equal(t, "X", DirUnset.String())
}

func TestDirectionFromCoord(t *testing.T) {
	assert.Equal(t, DirNorth, DirectionFromCoord(Coord{X: 0, Y: -1}))
	assert.Equal(t, DirEast, DirectionFromCoord(Coord{X: 1, Y: 0}))
	assert.Equal(t, DirSouth, DirectionFromCoord(Coord{X: 0, Y: 1}))
	assert.Equal(t, DirWest, DirectionFromCoord(Coord{X: -1, Y: 0}))
	assert.Equal(t, DirUnset, DirectionFromCoord(Coord{X: 3, Y: 3}))
}

func TestDirectionFromAlias(t *testing.T) {
	assert.Equal(t, DirNorth, DirectionFromAlias("N"))
	assert.Equal(t, DirEast, DirectionFromAlias("E"))
	assert.Equal(t, DirSouth, DirectionFromAlias("S"))
	assert.Equal(t, DirWest, DirectionFromAlias("W"))
	assert.Panics(t, func() { DirectionFromAlias("X") })
}
