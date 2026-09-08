package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// autobuildGrid builds a grid for autobuild tests. Unlike pathGrid it needs
// terrain, because rail cost drives the search, and it needs a Zone for every
// cell, because the successor filter reads one. Characters: '.' plains,
// '~' river, '^' mountain, '#' plains carrying player 0's track, and a digit
// a town on plains with that id.
func autobuildGrid(t *testing.T, rows ...string) *Grid {
	t.Helper()
	require.NotEmpty(t, rows)

	grid := NewGrid(len(rows[0]), len(rows))
	for y, row := range rows {
		require.Equalf(t, grid.Width, len(row), "row %d length mismatch", y)
		for x, ch := range row {
			tile := grid.GetXY(x, y)
			tile.ZoneID = 0
			switch {
			case ch == '.':
			case ch == '~':
				tile.Type = TYPE_WATER
			case ch == '^':
				tile.Type = TYPE_MOUNTAIN
			case ch == '#':
				tile.Track = 0
			case ch >= '0' && ch <= '9':
				id := int(ch - '0')
				tile.TownID = id
				grid.Towns = append(grid.Towns, NewTown(id, Coord{x, y}))
			default:
				t.Fatalf("unknown autobuild grid char %q at %d,%d", ch, x, y)
			}
		}
	}
	grid.Zones = []*Zone{NewZone(0, grid.Coords())}
	return grid
}

// plan runs the planner and returns the cells it would build, in order.
func plan(t *testing.T, grid *Grid, from, to Coord) ([]Coord, bool) {
	t.Helper()

	states, ok := NewAutobuildAStar(grid, NewPlayer(0), from, to).Search()
	if !ok {
		return nil, false
	}
	var built []Coord
	for _, s := range states {
		if s.Action != nil {
			require.Equal(t, ACTION_PLACE_TRACK, s.Action.Type)
			require.True(t, s.Action.GeneratedByAutobuild)
			built = append(built, s.Action.Coord)
		}
	}
	return built, true
}

func TestAutobuildPavesTheStraightLineBetweenTwoTowns(t *testing.T) {
	grid := autobuildGrid(t, "0...1")

	built, ok := plan(t, grid, Coord{1, 0}, Coord{3, 0})

	require.True(t, ok)
	assert.Equal(t, []Coord{{1, 0}, {2, 0}, {3, 0}}, built)
}

func TestAutobuildDetoursAroundExpensiveTerrain(t *testing.T) {
	grid := autobuildGrid(t,
		".....",
		".^^..",
		".^^..",
	)

	built, ok := plan(t, grid, Coord{0, 1}, Coord{3, 1})

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 1}, {0, 0}, {1, 0}, {2, 0}, {3, 0}, {3, 1}}, built)
}

func TestAutobuildCrossesARiverWhenTheDetourIsLonger(t *testing.T) {
	grid := autobuildGrid(t,
		"~~~~~",
		".~~~.",
		"~~~~~",
	)

	built, ok := plan(t, grid, Coord{0, 1}, Coord{4, 1})

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 1}}, built)
}

func TestAutobuildStopsAtTheGoalWithoutPavingIt(t *testing.T) {
	grid := autobuildGrid(t, "0..1")

	// The goal cell is a town, so the last step moves the cursor onto it for
	// free rather than trying to build there.
	built, ok := plan(t, grid, Coord{1, 0}, Coord{3, 0})

	require.True(t, ok)
	assert.Equal(t, []Coord{{1, 0}, {2, 0}}, built)
}

func TestAutobuildBuildsNothingWhenTheStartAlreadyReachesTheGoal(t *testing.T) {
	grid := autobuildGrid(t, "0###1")

	built, ok := plan(t, grid, Coord{0, 0}, Coord{4, 0})

	require.True(t, ok)
	assert.Empty(t, built)
}

func TestAutobuildRidesAnExistingRailBlockForFree(t *testing.T) {
	grid := autobuildGrid(t,
		"0####.",
		"......",
	)

	// Starting on the town at (0,0), the whole rail block is free to traverse,
	// so only the gap between its far end and the goal is paid for.
	built, ok := plan(t, grid, Coord{0, 0}, Coord{5, 0})

	require.True(t, ok)
	assert.Equal(t, []Coord{{5, 0}}, built)
}

func TestAutobuildTreatsAnyOwnersTrackAsPartOfTheBlock(t *testing.T) {
	grid := autobuildGrid(t, "0##..1")
	grid.GetXY(1, 0).Track = 1
	grid.GetXY(2, 0).Track = TRACK_NEUTRAL

	built, ok := plan(t, grid, Coord{0, 0}, Coord{5, 0})

	require.True(t, ok)
	assert.Equal(t, []Coord{{3, 0}, {4, 0}}, built)
}

func TestAutobuildFindsNoPlanFromAnOffGridStart(t *testing.T) {
	grid := autobuildGrid(t, "0...1")

	_, ok := plan(t, grid, Coord{-1, 0}, Coord{3, 0})

	assert.False(t, ok)
}

func TestAutobuildFindsNoPlanToAnOffGridGoal(t *testing.T) {
	grid := autobuildGrid(t, "0...1")

	_, ok := plan(t, grid, Coord{1, 0}, Coord{9, 9})

	assert.False(t, ok)
}

func TestAutobuildWillNotRouteThroughAnInkedRegion(t *testing.T) {
	grid := autobuildGrid(t,
		".....",
		".....",
	)
	for x := 1; x <= 3; x++ {
		grid.GetXY(x, 0).ZoneID = 1
		grid.GetXY(x, 1).ZoneID = 1
	}
	inked := NewZone(1, []Coord{{1, 0}, {2, 0}, {3, 0}, {1, 1}, {2, 1}, {3, 1}})
	inked.Inked = true
	grid.Zones = append(grid.Zones, inked)

	_, ok := plan(t, grid, Coord{0, 0}, Coord{4, 0})

	assert.False(t, ok)
}

func TestAutobuildStillReachesATownInsideAnInkedRegion(t *testing.T) {
	grid := autobuildGrid(t, ".0")
	grid.GetXY(1, 0).ZoneID = 1
	inked := NewZone(1, []Coord{{1, 0}})
	inked.Inked = true
	grid.Zones = append(grid.Zones, inked)

	built, ok := plan(t, grid, Coord{0, 0}, Coord{1, 0})

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 0}}, built)
}
