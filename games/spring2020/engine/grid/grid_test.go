package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCellTypeFromChar(t *testing.T) {
	tests := []struct {
		c      byte
		isWall bool
		pellet bool
	}{
		{'#', true, false},
		{'x', true, false},
		{' ', false, false},
		{'.', false, true},
		{'o', false, true},
	}
	for _, tc := range tests {
		name := string(tc.c)
		if tc.isWall {
			assert.Equal(t, CellWall, cellTypeFromChar(tc.c), name)
		} else {
			assert.Equal(t, CellFloor, cellTypeFromChar(tc.c), name)
		}
		assert.Equal(t, tc.pellet, cellHasPelletFromChar(tc.c), name)
	}
}

func TestCellTypeFromCharPanicsOnUnknown(t *testing.T) {
	assert.Panics(t, func() { cellTypeFromChar('Z') })
	assert.Panics(t, func() { cellHasPelletFromChar('Z') })
}

func TestNewGridFromRows(t *testing.T) {
	rows := []string{
		"###",
		"#.#",
		"###",
	}
	g := NewGridFromRows(rows, false)

	assert.Equal(t, 3, g.Width)
	assert.Equal(t, 3, g.Height)

	assert.True(t, g.GetXY(0, 0).IsWall())
	assert.True(t, g.GetXY(1, 1).IsFloor())
	assert.True(t, g.GetXY(1, 1).HasPellet)
	assert.False(t, g.GetXY(0, 0).HasPellet)
}

func TestGridGetOutOfBoundsReturnsNoCell(t *testing.T) {
	g := NewGridFromRows([]string{"   ", "   "}, false)
	got := g.GetXY(-1, 0)
	assert.False(t, got.IsValid())
	got = g.GetXY(0, 5)
	assert.False(t, got.IsValid())
	got = g.Get(Coord{X: 5, Y: 1})
	assert.False(t, got.IsValid())
}

func TestGridNeighboursNoWrap(t *testing.T) {
	g := NewGridFromRows([]string{"   ", "   ", "   "}, false)

	corner := g.Neighbours(Coord{X: 0, Y: 0})
	assert.ElementsMatch(t, []Coord{{X: 1, Y: 0}, {X: 0, Y: 1}}, corner)

	center := g.Neighbours(Coord{X: 1, Y: 1})
	assert.ElementsMatch(t,
		[]Coord{{X: 0, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 2}},
		center)
}

func TestGridNeighboursWrapsHorizontally(t *testing.T) {
	g := NewGridFromRows([]string{"   "}, true)

	// Left edge wraps to x=2.
	n := g.Neighbours(Coord{X: 0, Y: 0})
	assert.Contains(t, n, Coord{X: 2, Y: 0})
	assert.Contains(t, n, Coord{X: 1, Y: 0})

	// Right edge wraps to x=0.
	n = g.Neighbours(Coord{X: 2, Y: 0})
	assert.Contains(t, n, Coord{X: 0, Y: 0})
	assert.Contains(t, n, Coord{X: 1, Y: 0})
}

func TestGridGetCoordNeighbourWrap(t *testing.T) {
	g := NewGridFromRows([]string{"   "}, true)

	n, ok := g.GetCoordNeighbour(Coord{X: 0, Y: 0}, Coord{X: -1, Y: 0})
	assert.True(t, ok)
	assert.Equal(t, Coord{X: 2, Y: 0}, n)

	n, ok = g.GetCoordNeighbour(Coord{X: 2, Y: 0}, Coord{X: 1, Y: 0})
	assert.True(t, ok)
	assert.Equal(t, Coord{X: 0, Y: 0}, n)

	// Vertical is never wrapped.
	_, ok = g.GetCoordNeighbour(Coord{X: 0, Y: 0}, Coord{X: 0, Y: -1})
	assert.False(t, ok)
}

func TestGridCalculateDistanceWrap(t *testing.T) {
	g := NewGridFromRows([]string{"          "}, true) // width=10

	// Direct = 8, wrap = 2. Take wrap.
	assert.Equal(t, 2, g.CalculateDistance(Coord{X: 1, Y: 0}, Coord{X: 9, Y: 0}))
	// Direct = 3, no wrap gain.
	assert.Equal(t, 3, g.CalculateDistance(Coord{X: 1, Y: 0}, Coord{X: 4, Y: 0}))
}

func TestGridAllPelletsCherriesOrder(t *testing.T) {
	g := NewGridFromRows([]string{"o.", ".#"}, false)
	g.GetXY(0, 0).HasCherry = true
	g.GetXY(0, 0).HasPellet = false

	pellets := g.AllPellets()
	assert.Equal(t,
		[]Coord{{X: 1, Y: 0}, {X: 0, Y: 1}},
		pellets,
	)

	cherries := g.AllCherries()
	assert.Equal(t, []Coord{{X: 0, Y: 0}}, cherries)
}

func TestGridCoordsInsertionOrder(t *testing.T) {
	g := NewGrid(2, 2, false)
	assert.Equal(t,
		[]Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
		g.Coords(),
	)
}

func TestCellCopyFromSource(t *testing.T) {
	src := NewCell(Coord{X: 0, Y: 0})
	src.Type = CellFloor
	src.HasPellet = true

	dst := NewCell(Coord{X: 1, Y: 1})
	dst.HasCherry = true
	dst.Copy(src)

	assert.Equal(t, CellFloor, dst.Type)
	assert.True(t, dst.HasPellet)
	// Copy does not touch HasCherry, matching Java Cell.copy.
	assert.True(t, dst.HasCherry)
}

func TestNoCellIsInvalid(t *testing.T) {
	assert.False(t, NoCell.IsValid())
}
