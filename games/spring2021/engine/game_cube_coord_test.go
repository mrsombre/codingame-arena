package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCubeCoordDistanceTo(t *testing.T) {
	c := NewCubeCoord(0, 0, 0)
	assert.Equal(t, 0, c.DistanceTo(c))
	assert.Equal(t, 1, c.DistanceTo(NewCubeCoord(1, -1, 0)))
	assert.Equal(t, 2, c.DistanceTo(NewCubeCoord(2, -2, 0)))
	assert.Equal(t, 3, c.DistanceTo(NewCubeCoord(0, 3, -3)))
}

func TestCubeCoordNeighborAt(t *testing.T) {
	c := NewCubeCoord(0, 0, 0)
	for orient := 0; orient < 6; orient++ {
		n1 := c.Neighbor(orient)
		n2 := c.NeighborAt(orient, 2)
		assert.Equal(t, 1, c.DistanceTo(n1), "Neighbor(%d) should be 1 cell away", orient)
		assert.Equal(t, 2, c.DistanceTo(n2), "NeighborAt(%d,2) should be 2 cells away", orient)
	}
}

func TestCubeCoordOpposite(t *testing.T) {
	c := NewCubeCoord(2, -1, -1)
	assert.Equal(t, NewCubeCoord(-2, 1, 1), c.Opposite())
	assert.Equal(t, c, c.Opposite().Opposite())
}

func TestCubeCoordString(t *testing.T) {
	assert.Equal(t, "1 -2 1", NewCubeCoord(1, -2, 1).String())
}
