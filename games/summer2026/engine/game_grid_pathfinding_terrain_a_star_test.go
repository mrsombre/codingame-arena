package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// terrainGrid builds a grid whose regions come from a rune layer, numbered by
// first appearance in row-major order. Terrain type is irrelevant to
// TerrainAStar, so every cell is plains.
func terrainGrid(t *testing.T, regions []string) *Grid {
	t.Helper()
	grid := NewGrid(len(regions[0]), len(regions))
	assignRegions(t, grid, regions)
	return grid
}

func TestTerrainAStarWalksTerrainRegardlessOfTrack(t *testing.T) {
	grid := terrainGrid(t, []string{"aaaa"})
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{3, 0})

	path, ok := NewTerrainAStar(grid, from, to).Search()

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, path)
}

// Mountains and rivers cost extra paint but are traversable, so terrain alone
// never blocks a route.
func TestTerrainAStarCrossesMountainsAndRivers(t *testing.T) {
	grid := terrainGrid(t, []string{"aaa"})
	grid.GetXY(1, 0).Type = TYPE_MOUNTAIN
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{2, 0})

	_, ok := NewTerrainAStar(grid, from, to).Search()

	assert.True(t, ok)
}

func TestTerrainAStarRoutesAroundAnInkedRegion(t *testing.T) {
	grid := terrainGrid(t, []string{
		"aba",
		"aaa",
	})
	grid.Zones[1].Inked = true
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{2, 0})

	path, ok := NewTerrainAStar(grid, from, to).Search()

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 0}, {0, 1}, {1, 1}, {2, 1}, {2, 0}}, path)
}

func TestTerrainAStarFailsWhenInkingWallsOffTheDestination(t *testing.T) {
	grid := terrainGrid(t, []string{
		"aba",
		"aba",
	})
	grid.Zones[1].Inked = true
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{2, 0})

	_, ok := NewTerrainAStar(grid, from, to).Search()

	assert.False(t, ok)
}

// The start cell is never tested against the inked filter, so a town whose own
// region has been inked can still be routed out of.
func TestTerrainAStarLeavesAnInkedStartRegion(t *testing.T) {
	grid := terrainGrid(t, []string{"abb"})
	grid.Zones[0].Inked = true
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{2, 0})

	_, ok := NewTerrainAStar(grid, from, to).Search()

	assert.True(t, ok)
}

// The destination is a successor like any other, so inking its region blocks
// the route even though the search would end there.
func TestTerrainAStarFailsWhenTheDestinationRegionIsInked(t *testing.T) {
	grid := terrainGrid(t, []string{"aab"})
	grid.Zones[1].Inked = true
	from, to := NewTown(0, Coord{0, 0}), NewTown(1, Coord{2, 0})

	_, ok := NewTerrainAStar(grid, from, to).Search()

	assert.False(t, ok)
}

func TestTerrainAStarReturnsTheSingleCellPathForATownToItself(t *testing.T) {
	grid := terrainGrid(t, []string{"aa"})
	town := NewTown(0, Coord{1, 0})

	path, ok := NewTerrainAStar(grid, town, town).Search()

	require.True(t, ok)
	assert.Equal(t, []Coord{{1, 0}}, path)
}
