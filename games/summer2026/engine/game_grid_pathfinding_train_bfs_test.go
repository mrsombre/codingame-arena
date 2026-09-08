package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pathGrid builds a bare grid for pathfinder tests, with no zones, players or
// game around it. Characters: '.' an empty cell, '#' a cell carrying player
// 0's track, and a digit a town with that id. Terrain is irrelevant to the
// train pathfinders — only track and towns are — so it is all plains.
func pathGrid(t *testing.T, rows ...string) *Grid {
	t.Helper()
	require.NotEmpty(t, rows)

	grid := NewGrid(len(rows[0]), len(rows))
	for y, row := range rows {
		require.Equalf(t, grid.Width, len(row), "row %d length mismatch", y)
		for x, ch := range row {
			tile := grid.GetXY(x, y)
			switch {
			case ch == '.':
			case ch == '#':
				tile.Track = 0
			case ch >= '0' && ch <= '9':
				id := int(ch - '0')
				tile.TownID = id
				grid.Towns = append(grid.Towns, NewTown(id, Coord{x, y}))
			default:
				t.Fatalf("unknown path grid char %q at %d,%d", ch, x, y)
			}
		}
	}
	return grid
}

// town looks a town up by id rather than by position in grid.Towns, so a
// fixture can name towns in any layout order.
func town(t *testing.T, grid *Grid, id int) *Town {
	t.Helper()
	for _, tn := range grid.Towns {
		if tn.ID == id {
			return tn
		}
	}
	t.Fatalf("no town %d in fixture", id)
	return nil
}

func TestTrainBFSReturnsBothTownsAndTheTrackBetweenThem(t *testing.T) {
	grid := pathGrid(t, "0###1")

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Equal(t, []Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}, path)
}

func TestTrainBFSFindsNoPathAcrossAGapInTheTrack(t *testing.T) {
	grid := pathGrid(t, "0#.#1")

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Empty(t, path)
}

// A third town is a cell a train can pass, so a route may run straight
// through one.
func TestTrainBFSRoutesThroughAnInterveningTown(t *testing.T) {
	grid := pathGrid(t, "0#2#1")

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Len(t, path, 5)
	assert.Contains(t, path, Coord{2, 0})
}

// Track ownership is irrelevant to reachability: a connection can run over
// the opponent's rails or over contested neutral ones.
func TestTrainBFSIgnoresTrackOwnership(t *testing.T) {
	grid := pathGrid(t, "0###1")
	grid.GetXY(1, 0).Track = 1
	grid.GetXY(2, 0).Track = TRACK_NEUTRAL

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Len(t, path, 5)
}

// Both routes below are five cells long. The tie is broken by the order
// neighbours are enqueued, which is Grid.ADJACENCY's: NORTH, EAST, SOUTH,
// WEST.
func TestTrainBFSBreaksEqualLengthTiesNorthBeforeSouth(t *testing.T) {
	grid := pathGrid(t,
		"###",
		"0.1",
		"###",
	)

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Equal(t, []Coord{{0, 1}, {0, 0}, {1, 0}, {2, 0}, {2, 1}}, path)
}

func TestTrainBFSBreaksEqualLengthTiesEastBeforeSouth(t *testing.T) {
	grid := pathGrid(t,
		"0##",
		"###",
		"##1",
	)

	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Equal(t, []Coord{{0, 0}, {1, 0}, {2, 0}, {2, 1}, {2, 2}}, path)
}

func TestTrainBFSBreaksEqualLengthTiesEastBeforeWest(t *testing.T) {
	grid := pathGrid(t,
		"#0#",
		"#.#",
		"#1#",
	)
	// The centre carries no track, so the two towns are joined only by going
	// round one side or the other, and both sides are five cells.
	path := NewTrainBFS(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.Equal(t, []Coord{{1, 0}, {2, 0}, {2, 1}, {2, 2}, {1, 2}}, path)
}

// The upstream comparator asks for the direction between two neighbours of
// the same centre cell. That difference is never a unit vector, so the answer
// is always UNSET and the sort never reorders anything — verified against
// OpenJDK 17 for all fifteen non-empty neighbour subsets.
func TestTrainBFSNeighbourComparatorIsAlwaysUnset(t *testing.T) {
	grid := NewGrid(11, 11)
	centre := Coord{5, 5}

	neighbours := grid.Neighbours(centre)
	require.Len(t, neighbours, 4)

	for _, a := range neighbours {
		for _, b := range neighbours {
			if a == b {
				continue
			}
			assert.Equalf(t, int(UNSET), compareTrainBFSNeighbours(a, b), "%s vs %s", a, b)
		}
	}

	before := append([]Coord(nil), neighbours...)
	javaListSort(neighbours, compareTrainBFSNeighbours)
	assert.Equal(t, before, neighbours)
}
