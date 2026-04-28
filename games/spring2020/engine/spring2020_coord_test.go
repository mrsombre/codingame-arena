package engine

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordAdd(t *testing.T) {
	c := Coord{X: 3, Y: 5}.Add(Coord{X: -1, Y: 2})
	assert.Equal(t, Coord{X: 2, Y: 7}, c)
}

func TestCoordAddXY(t *testing.T) {
	c := Coord{X: 3, Y: 5}.AddXY(-1, 2)
	assert.Equal(t, Coord{X: 2, Y: 7}, c)
}

func TestCoordSubtract(t *testing.T) {
	c := Coord{X: 3, Y: 5}.Subtract(Coord{X: 1, Y: 2})
	assert.Equal(t, Coord{X: 2, Y: 3}, c)
}

func TestCoordManhattan(t *testing.T) {
	assert.Equal(t, 0, Coord{X: 4, Y: 4}.ManhattanTo(Coord{X: 4, Y: 4}))
	assert.Equal(t, 5, Coord{X: 1, Y: 1}.ManhattanTo(Coord{X: 4, Y: 3}))
	assert.Equal(t, 5, Coord{X: 4, Y: 3}.ManhattanTo(Coord{X: 1, Y: 1}))
	assert.Equal(t, 7, Coord{X: 0, Y: 0}.ManhattanToXY(-3, 4))
}

func TestCoordChebyshev(t *testing.T) {
	assert.Equal(t, 4, Coord{X: 1, Y: 1}.ChebyshevToXY(5, 3))
	assert.Equal(t, 3, Coord{X: 1, Y: 1}.ChebyshevToXY(3, 4))
}

func TestCoordEuclidean(t *testing.T) {
	d := Coord{X: 0, Y: 0}.EuclideanTo(Coord{X: 3, Y: 4})
	assert.InDelta(t, 5.0, d, 1e-9)
	d = Coord{X: 2, Y: -1}.EuclideanTo(Coord{X: 2, Y: -1})
	assert.Equal(t, 0.0, d)
	assert.False(t, math.IsNaN(Coord{}.EuclideanTo(Coord{X: 1, Y: 1})))
}

func TestCoordLess(t *testing.T) {
	assert.True(t, Coord{X: 1, Y: 5}.Less(Coord{X: 2, Y: 0}))
	assert.True(t, Coord{X: 2, Y: 0}.Less(Coord{X: 2, Y: 1}))
	assert.False(t, Coord{X: 2, Y: 1}.Less(Coord{X: 2, Y: 1}))
	assert.False(t, Coord{X: 3, Y: 0}.Less(Coord{X: 1, Y: 9}))
}

func TestCoordString(t *testing.T) {
	assert.Equal(t, "(2, 7)", Coord{X: 2, Y: 7}.String())
}

func TestCoordUnitVector(t *testing.T) {
	tests := []struct {
		in   Coord
		want Coord
	}{
		{Coord{X: 0, Y: 0}, Coord{X: 0, Y: 0}},
		{Coord{X: 5, Y: 0}, Coord{X: 1, Y: 0}},
		{Coord{X: -3, Y: 0}, Coord{X: -1, Y: 0}},
		{Coord{X: 0, Y: 7}, Coord{X: 0, Y: 1}},
		{Coord{X: -2, Y: -8}, Coord{X: -1, Y: -1}},
		{Coord{X: 4, Y: -2}, Coord{X: 1, Y: -1}},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.in.UnitVector(), tc.in.String())
	}
}
