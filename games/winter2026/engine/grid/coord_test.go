package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordAddAndAddXY(t *testing.T) {
	c := Coord{X: 3, Y: 5}.Add(Coord{X: -1, Y: 2})
	assert.Equal(t, Coord{X: 2, Y: 7}, c)

	c = Coord{X: 3, Y: 5}.AddXY(-1, 2)
	assert.Equal(t, Coord{X: 2, Y: 7}, c)
}

func TestCoordManhattan(t *testing.T) {
	assert.Equal(t, 0, Coord{X: 4, Y: 4}.ManhattanTo(Coord{X: 4, Y: 4}))
	assert.Equal(t, 5, Coord{X: 1, Y: 1}.ManhattanTo(Coord{X: 4, Y: 3}))
	assert.Equal(t, 7, Coord{}.ManhattanToXY(-3, 4))
}

func TestCoordChebyshev(t *testing.T) {
	assert.Equal(t, 4, Coord{X: 1, Y: 1}.ChebyshevTo(Coord{X: 5, Y: 3}))
	assert.Equal(t, 3, Coord{X: 1, Y: 1}.ChebyshevToXY(3, 4))
}

func TestCoordEuclidean(t *testing.T) {
	assert.InDelta(t, 5.0, Coord{}.EuclideanTo(Coord{X: 3, Y: 4}), 1e-9)
	assert.InDelta(t, 5.0, Coord{}.EuclideanToXY(3, 4), 1e-9)
	assert.InDelta(t, 25.0, Coord{}.SqrEuclideanTo(Coord{X: 3, Y: 4}), 1e-9)
	assert.InDelta(t, 25.0, Coord{}.SqrEuclideanToXY(3, 4), 1e-9)
}

func TestCoordLess(t *testing.T) {
	assert.True(t, Coord{X: 1, Y: 5}.Less(Coord{X: 2, Y: 0}))
	assert.True(t, Coord{X: 2, Y: 0}.Less(Coord{X: 2, Y: 1}))
	assert.False(t, Coord{X: 2, Y: 1}.Less(Coord{X: 2, Y: 1}))
	assert.False(t, Coord{X: 3, Y: 0}.Less(Coord{X: 1, Y: 9}))
}

func TestCoordStringAndIntString(t *testing.T) {
	assert.Equal(t, "(2, 7)", Coord{X: 2, Y: 7}.String())
	assert.Equal(t, "2 7", Coord{X: 2, Y: 7}.ToIntString())
}
