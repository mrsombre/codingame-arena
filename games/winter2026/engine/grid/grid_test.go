package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGridDefaultsAndGet(t *testing.T) {
	g := NewGrid(4, 3)
	assert.Equal(t, 4, g.Width)
	assert.Equal(t, 3, g.Height)
	assert.False(t, g.YSymmetry)

	tile := g.GetXY(2, 2)
	assert.True(t, tile.IsValid())
	assert.Equal(t, TileEmpty, tile.Type)
	assert.Equal(t, Coord{X: 2, Y: 2}, tile.Coord)
}

func TestGridGetOutOfBoundsReturnsNoTile(t *testing.T) {
	g := NewGrid(3, 2)
	assert.False(t, g.GetXY(-1, 0).IsValid())
	assert.False(t, g.GetXY(0, 5).IsValid())
	assert.False(t, g.Get(Coord{X: 5, Y: 1}).IsValid())
}

func TestGridCoordsInsertionOrder(t *testing.T) {
	g := NewGrid(2, 2)
	assert.Equal(t,
		[]Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
		g.Coords(),
	)
}

func TestGridNeighbours4(t *testing.T) {
	g := NewGrid(3, 3)
	corner := g.Neighbours4(Coord{X: 0, Y: 0})
	assert.ElementsMatch(t, []Coord{{X: 1, Y: 0}, {X: 0, Y: 1}}, corner)

	center := g.Neighbours4(Coord{X: 1, Y: 1})
	assert.ElementsMatch(t,
		[]Coord{{X: 0, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 2}},
		center)
}

func TestGridOppositeXOnly(t *testing.T) {
	g := NewGrid(10, 5)
	assert.False(t, g.YSymmetry)
	assert.Equal(t, Coord{X: 6, Y: 2}, g.Opposite(Coord{X: 3, Y: 2}))
	assert.Equal(t, Coord{X: 0, Y: 4}, g.Opposite(Coord{X: 9, Y: 4}))
}

func TestGridOppositeXAndY(t *testing.T) {
	g := NewGridSym(10, 5, true)
	assert.True(t, g.YSymmetry)
	// height=5, so Y mirrored across 4.
	assert.Equal(t, Coord{X: 6, Y: 2}, g.Opposite(Coord{X: 3, Y: 2}))
	assert.Equal(t, Coord{X: 9, Y: 4}, g.Opposite(Coord{X: 0, Y: 0}))
}

func TestGridHasAppleAndRemoveApple(t *testing.T) {
	g := NewGrid(3, 3)
	g.Apples = []Coord{{X: 1, Y: 1}, {X: 2, Y: 2}}

	assert.True(t, g.HasApple(Coord{X: 1, Y: 1}))
	assert.True(t, g.HasApple(Coord{X: 2, Y: 2}))
	assert.False(t, g.HasApple(Coord{X: 0, Y: 0}))

	g.RemoveApple(Coord{X: 1, Y: 1})
	assert.False(t, g.HasApple(Coord{X: 1, Y: 1}))
	assert.True(t, g.HasApple(Coord{X: 2, Y: 2}))

	// Removing a missing apple is a no-op.
	g.RemoveApple(Coord{X: 9, Y: 9})
	assert.Len(t, g.Apples, 1)
}

func TestGridClosestTargets(t *testing.T) {
	g := NewGrid(10, 10)
	from := Coord{X: 0, Y: 0}
	targets := []Coord{
		{X: 5, Y: 5}, // dist 10
		{X: 1, Y: 1}, // dist 2
		{X: 2, Y: 0}, // dist 2
		{X: 0, Y: 4}, // dist 4
	}
	got := g.ClosestTargets(from, targets)
	assert.ElementsMatch(t, []Coord{{X: 1, Y: 1}, {X: 2, Y: 0}}, got)

	assert.Empty(t, g.ClosestTargets(from, nil))
}

func TestDetectAirPocketsCountsConnectedComponents(t *testing.T) {
	g := NewGrid(5, 5)
	// Two rooms separated by a wall column.
	for y := 0; y < 5; y++ {
		g.GetXY(2, y).SetType(TileWall)
	}
	islands := g.DetectAirPockets()
	assert.Len(t, islands, 2)
	totalFloors := 0
	for _, island := range islands {
		totalFloors += len(island)
	}
	assert.Equal(t, 5*5-5, totalFloors, "all non-wall cells belong to some island")
}

func TestDetectSpawnIslandsGroupsAdjacentSpawns(t *testing.T) {
	g := NewGrid(10, 10)
	// Two vertical columns: spawn group A at x=1, spawn group B at x=5.
	g.Spawns = []Coord{
		{X: 1, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 3},
		{X: 5, Y: 1}, {X: 5, Y: 2}, {X: 5, Y: 3},
	}
	islands := g.DetectSpawnIslands()
	assert.Len(t, islands, 2)
}

func TestDetectLowestIslandFloodFillsWallsFromBottomLeft(t *testing.T) {
	g := NewGrid(4, 4)
	// Floor everywhere; no wall at (0, height-1) → no lowest island.
	assert.Nil(t, g.DetectLowestIsland())

	// Fill bottom row with walls.
	for x := 0; x < 4; x++ {
		g.GetXY(x, 3).SetType(TileWall)
	}
	// Add an isolated wall that shouldn't be in the lowest island.
	g.GetXY(1, 1).SetType(TileWall)

	lowest := g.DetectLowestIsland()
	assert.Len(t, lowest, 4, "only bottom row is connected to (0, height-1)")
	for _, c := range lowest {
		assert.Equal(t, 3, c.Y)
	}
}
