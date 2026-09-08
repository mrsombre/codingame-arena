package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordArithmetic(t *testing.T) {
	c := Coord{3, 5}

	assert.Equal(t, Coord{4, 7}, c.Add(Coord{1, 2}))
	assert.Equal(t, Coord{4, 7}, c.AddXY(1, 2))
	assert.Equal(t, Coord{2, 3}, c.Sub(Coord{1, 2}))
	assert.Equal(t, Coord{2, 3}, c.SubXY(1, 2))
}

func TestCoordDistances(t *testing.T) {
	c := Coord{1, 1}

	assert.Equal(t, 7, c.ManhattanTo(Coord{4, 5}))
	assert.Equal(t, 4, c.ChebyshevTo(Coord{4, 5}))
	assert.Equal(t, 5.0, c.EuclideanTo(Coord{4, 5}))
	assert.Equal(t, 25.0, c.SqrEuclideanTo(4, 5))
}

func TestCoordDistancesAreSymmetricAcrossNegativeDeltas(t *testing.T) {
	a, b := Coord{4, 5}, Coord{1, 1}

	assert.Equal(t, b.ManhattanTo(a), a.ManhattanTo(b))
	assert.Equal(t, b.ChebyshevTo(a), a.ChebyshevTo(b))
}

func TestCoordIntNormalizeTruncatesToACardinalStep(t *testing.T) {
	assert.Equal(t, Coord{1, -1}, Coord{7, -3}.IntNormalize())
	assert.Equal(t, Coord{0, 1}, Coord{0, 9}.IntNormalize())
	assert.Equal(t, Coord{0, 0}, Coord{0, 0}.IntNormalize())
}

func TestCoordCompareToOrdersByXThenY(t *testing.T) {
	assert.Equal(t, -1, Coord{1, 9}.CompareTo(Coord{2, 0}))
	assert.Equal(t, 1, Coord{2, 0}.CompareTo(Coord{1, 9}))
	assert.Equal(t, -1, Coord{2, 0}.CompareTo(Coord{2, 1}))
	assert.Equal(t, 0, Coord{2, 1}.CompareTo(Coord{2, 1}))
}

func TestCoordEqualityIsValueBased(t *testing.T) {
	assert.True(t, Coord{2, 3} == NewCoord(2, 3))
	assert.False(t, Coord{2, 3} == NewCoord(3, 2))
}

// JavaHash reproduces Coord.hashCode(); GridMaker indexes a HashSet<Coord>
// after random.nextInt(size), so the value feeds map-generation parity.
func TestCoordJavaHashMatchesJavaHashCode(t *testing.T) {
	assert.Equal(t, int32(961), Coord{0, 0}.JavaHash())
	assert.Equal(t, int32(993), Coord{1, 1}.JavaHash())
	assert.Equal(t, int32(1677), Coord{23, 3}.JavaHash())
	assert.Equal(t, int32(930), Coord{-1, 0}.JavaHash())
}

func TestCoordFormatting(t *testing.T) {
	assert.Equal(t, "(2, 3)", Coord{2, 3}.String())
	assert.Equal(t, "2 3", Coord{2, 3}.ToIntString())
}

func TestCoordIsItsOwnPositionable(t *testing.T) {
	c := Coord{2, 3}

	assert.Equal(t, c, c.Position())
	assert.Equal(t, 1.0, c.DistanceMultiplier())
}
