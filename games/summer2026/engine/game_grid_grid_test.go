package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGridIsRowMajorAndAddressableByCoord(t *testing.T) {
	grid := NewGrid(3, 2)

	require.Len(t, grid.Cells, 6)
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			assert.Equal(t, Coord{x, y}, grid.GetXY(x, y).Coord)
		}
	}
	assert.Equal(t, []Coord{
		{0, 0}, {1, 0}, {2, 0},
		{0, 1}, {1, 1}, {2, 1},
	}, grid.Coords())
}

// Neighbour order is load-bearing: shortest-path ties break NORTH, EAST,
// SOUTH, WEST.
func TestGridNeighboursWalkNorthEastSouthWest(t *testing.T) {
	grid := NewGrid(3, 3)

	assert.Equal(t, []Coord{
		{1, 0}, {2, 1}, {1, 2}, {0, 1},
	}, grid.Neighbours(Coord{1, 1}))
}

func TestGridNeighboursSkipCellsOffTheGrid(t *testing.T) {
	grid := NewGrid(3, 3)

	assert.Equal(t, []Coord{{1, 0}, {0, 1}}, grid.Neighbours(Coord{0, 0}))
	assert.Equal(t, []Coord{{2, 1}, {1, 2}}, grid.Neighbours(Coord{2, 2}))
}

func TestGridNeighboursWithEightAdjacency(t *testing.T) {
	grid := NewGrid(3, 3)

	assert.Equal(t, []Coord{
		{1, 0}, {2, 1}, {1, 2}, {0, 1},
		{0, 0}, {2, 2}, {2, 0}, {0, 2},
	}, grid.NeighboursWith(Coord{1, 1}, ADJACENCY_8))
}

func TestGridGetReturnsNilOutOfBounds(t *testing.T) {
	grid := NewGrid(2, 2)

	assert.Nil(t, grid.GetXY(-1, 0))
	assert.Nil(t, grid.GetXY(0, -1))
	assert.Nil(t, grid.GetXY(2, 0))
	assert.Nil(t, grid.GetXY(0, 2))
	assert.NotNil(t, grid.Get(Coord{1, 1}))
}

// Java's NO_TILE answered every predicate; the Go port answers them from nil
// so an out-of-bounds read behaves the same way.
func TestTilePredicatesTolerateAnOutOfBoundsRead(t *testing.T) {
	grid := NewGrid(2, 2)
	tile := grid.GetXY(5, 5)

	assert.False(t, tile.IsValid())
	assert.False(t, tile.IsPlains())
	assert.False(t, tile.IsWater())
	assert.False(t, tile.IsMountain())
	assert.False(t, tile.IsTown())
	assert.False(t, tile.IsTrack())
	assert.False(t, tile.IsTrackOrTown())
	assert.False(t, tile.CanUseTracks(0))
	assert.False(t, tile.CanUseTrackOrTown(0))
	assert.Equal(t, -1, tile.GetType())
	assert.Equal(t, -1, tile.GetZoneID())
}

// Writes are deliberately not nil-tolerant: an out-of-bounds mutation is a
// bug, and a panic surfaces it instead of it vanishing into a sentinel.
func TestOutOfBoundsWritePanics(t *testing.T) {
	grid := NewGrid(2, 2)

	assert.Panics(t, func() { grid.GetXY(5, 5).SetType(TYPE_WATER) })
	assert.Panics(t, func() { grid.GetXY(5, 5).SetZoneID(0) })
	assert.Panics(t, func() { grid.GetXY(5, 5).Clear() })
}

func TestTilePredicatesReflectTerrainAndOwnership(t *testing.T) {
	grid := NewGrid(3, 1)
	plains, river, mountain := grid.GetXY(0, 0), grid.GetXY(1, 0), grid.GetXY(2, 0)
	river.SetType(TYPE_WATER)
	mountain.SetType(TYPE_MOUNTAIN)

	assert.True(t, plains.IsPlains())
	assert.True(t, river.IsWater())
	assert.True(t, mountain.IsMountain())

	plains.Track = 1
	assert.True(t, plains.IsTrack())
	assert.True(t, plains.CanUseTracks(1))
	assert.False(t, plains.CanUseTracks(0))

	river.Track = TRACK_NEUTRAL
	assert.True(t, river.CanUseTracks(0))
	assert.True(t, river.CanUseTracks(1))

	mountain.TownID = 4
	assert.True(t, mountain.IsTown())
	assert.True(t, mountain.IsTrackOrTown())
	assert.True(t, mountain.CanUseTrackOrTown(0))
}

func TestCanTrainPassOnlyOnTracksAndTowns(t *testing.T) {
	grid := NewGrid(3, 1)
	grid.GetXY(1, 0).Track = 0
	grid.GetXY(2, 0).TownID = 0

	assert.False(t, grid.CanTrainPass(Coord{0, 0}))
	assert.True(t, grid.CanTrainPass(Coord{1, 0}))
	assert.True(t, grid.CanTrainPass(Coord{2, 0}))
	assert.False(t, grid.CanTrainPass(Coord{9, 9}))
}

func TestGridCloneCopiesTerrainAndRegionsOnly(t *testing.T) {
	grid := NewGrid(2, 1)
	grid.GetXY(0, 0).SetType(TYPE_MOUNTAIN)
	grid.GetXY(0, 0).SetZoneID(3)
	grid.GetXY(0, 0).Track = 1

	clone := grid.Clone()

	assert.Equal(t, TYPE_MOUNTAIN, clone.GetXY(0, 0).GetType())
	assert.Equal(t, 3, clone.GetXY(0, 0).GetZoneID())
	assert.Equal(t, TRACK_NONE, clone.GetXY(0, 0).Track)

	clone.GetXY(1, 0).SetType(TYPE_WATER)
	assert.Equal(t, TYPE_GRASS, grid.GetXY(1, 0).GetType())
}

func TestGridOppositeMirrorsOnXAndOptionallyY(t *testing.T) {
	assert.Equal(t, Coord{4, 1}, NewGrid(6, 4).Opposite(Coord{1, 1}))
	assert.Equal(t, Coord{4, 2}, NewGridSym(6, 4, true).Opposite(Coord{1, 1}))
}

func TestRailCostPerTerrain(t *testing.T) {
	grid := NewGrid(4, 1)
	grid.GetXY(1, 0).SetType(TYPE_WATER)
	grid.GetXY(2, 0).SetType(TYPE_MOUNTAIN)
	grid.GetXY(3, 0).SetType(TYPE_POI)

	assert.Equal(t, 1, RailCost(grid.GetXY(0, 0)))
	assert.Equal(t, 2, RailCost(grid.GetXY(1, 0)))
	assert.Equal(t, 3, RailCost(grid.GetXY(2, 0)))
	assert.Equal(t, 3, RailCost(grid.GetXY(3, 0)))
}

func TestClosestTargetsKeepsEveryTieAtTheMinimumDistance(t *testing.T) {
	from := Coord{0, 0}
	targets := []Coord{{3, 0}, {0, 1}, {1, 0}, {5, 5}}

	assert.Equal(t, []Coord{{0, 1}, {1, 0}}, ClosestTargets(from, targets))
}
