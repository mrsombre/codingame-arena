package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// cellGrid builds a 4-neighbour-connected grid of GRASS cells without going
// through Board, so isolated Cell tests don't pull in map generation.
func cellGrid(width, height int) [][]*Cell {
	grid := make([][]*Cell, width)
	for x := 0; x < width; x++ {
		grid[x] = make([]*Cell, height)
		for y := 0; y < height; y++ {
			grid[x][y] = NewCell(x, y)
		}
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c := grid[x][y]
			for dir := 0; dir < 4; dir++ {
				nx := x + cellDX[dir]
				ny := y + cellDY[dir]
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				c.Neighbors[dir] = grid[nx][ny]
			}
		}
	}
	return grid
}

func TestCellGetIDPacksXYBitwise(t *testing.T) {
	// Java rule: cell.getId() = x + (y << 16). Tests/Task.groupByCell rely on
	// this giving a stable deterministic order over the board.
	c := NewCell(7, 3)
	assert.Equal(t, 7+(3<<16), c.GetID())
}

func TestCellManhattanIsSymmetric(t *testing.T) {
	a := NewCell(1, 1)
	b := NewCell(4, 5)
	assert.Equal(t, 7, a.Manhattan(b))
	assert.Equal(t, 7, b.Manhattan(a))
	assert.Equal(t, 0, a.Manhattan(a))
}

func TestCellIsNearWaterAndIron(t *testing.T) {
	grid := cellGrid(3, 1)
	grid[1][0].Type = CellWATER
	assert.True(t, grid[0][0].IsNearWater())
	assert.False(t, grid[0][0].IsNearIron())

	grid[1][0].Type = CellIRON
	assert.False(t, grid[0][0].IsNearWater())
	assert.True(t, grid[0][0].IsNearIron())
}

func TestCellIsNearEdgeAtCorners(t *testing.T) {
	grid := cellGrid(3, 3)
	assert.True(t, grid[0][0].IsNearEdge(), "corner has nil neighbours")
	assert.True(t, grid[2][2].IsNearEdge(), "opposite corner")
	assert.False(t, grid[1][1].IsNearEdge(), "center is fully bounded")
}

func TestCellIsNearShackOnSelfOr4Neighbour(t *testing.T) {
	grid := cellGrid(3, 3)
	player := NewPlayer(0)
	player.Shack = grid[1][1]

	assert.True(t, grid[1][1].IsNearShack(player), "the shack itself")
	assert.True(t, grid[0][1].IsNearShack(player), "west neighbour")
	assert.True(t, grid[2][1].IsNearShack(player), "east neighbour")
	// Diagonal is *not* near — Java only checks the 4-neighbour set.
	assert.False(t, grid[0][0].IsNearShack(player), "diagonal is not near")
}
