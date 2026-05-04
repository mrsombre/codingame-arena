package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rule: "The forest is made up of 37 hexagonal cells."
func TestBoardHas37Cells(t *testing.T) {
	g := newScenario(4)
	assert.Len(t, g.Board.Cells, 37)
	assert.Len(t, g.Board.Coords, 37)
	assert.Len(t, g.Board.Map, 37)
}

// Richness rings: center+inner=lush, middle=ok, outer=poor.
func TestBoardRichnessRings(t *testing.T) {
	g := newScenario(4)
	for i := 0; i <= 6; i++ {
		assert.Equalf(t, RICHNESS_LUSH, g.Board.Cells[i].GetRichness(), "cell %d", i)
	}
	for i := 7; i <= 18; i++ {
		assert.Equalf(t, RICHNESS_OK, g.Board.Cells[i].GetRichness(), "cell %d", i)
	}
	for i := 19; i <= 36; i++ {
		assert.Equalf(t, RICHNESS_POOR, g.Board.Cells[i].GetRichness(), "cell %d", i)
	}
}

func TestBoardCellByIndexOutOfRangeReturnsNilSentinel(t *testing.T) {
	g := newScenario(4)
	assert.Nil(t, g.Board.CellByIndex(-1))
	assert.Nil(t, g.Board.CellByIndex(37))
	assert.False(t, g.Board.CellByIndex(-1).IsValid())
}

func TestBoardCoordByIndexBounds(t *testing.T) {
	g := newScenario(4)
	_, ok := g.Board.CoordByIndex(-1)
	assert.False(t, ok)
	_, ok = g.Board.CoordByIndex(37)
	assert.False(t, ok)
	_, ok = g.Board.CoordByIndex(0)
	assert.True(t, ok)
}

func TestBoardCellAtUnknownCoord(t *testing.T) {
	g := newScenario(4)
	assert.Nil(t, g.Board.CellAt(NewCubeCoord(99, -50, -49)))
}

func TestBoardCellsIndexedByPosition(t *testing.T) {
	g := newScenario(4)
	for i, cell := range g.Board.Cells {
		assert.Equalf(t, i, cell.GetIndex(), "Cells slice indexed by Cell.Index")
	}
}
