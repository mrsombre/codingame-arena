package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrainAStarReturnsBothTownsAndTheTrackBetweenThem(t *testing.T) {
	grid := pathGrid(t, "0###1")

	path, ok := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	require.True(t, ok)
	assert.Equal(t, []Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}, path)
}

func TestTrainAStarFindsNoPathAcrossAGapInTheTrack(t *testing.T) {
	grid := pathGrid(t, "0#.#1")

	path, ok := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	assert.False(t, ok)
	assert.Empty(t, path)
}

// The route has to bend round the unbuilt centre, and every shortest way
// round is seven cells.
func TestTrainAStarFindsAShortestPathAroundAnObstacle(t *testing.T) {
	grid := pathGrid(t,
		"0####",
		"##.##",
		"####1",
	)

	path, ok := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 1)).Search()

	require.True(t, ok)
	assert.Len(t, path, 7)
	assert.Equal(t, Coord{0, 0}, path[0])
	assert.Equal(t, Coord{4, 2}, path[len(path)-1])
	assert.NotContains(t, path, Coord{2, 1})
}

// Unlike TrainBFS's, this tie-break is the real thing: the ordinal of the
// direction of the step being taken, so among equal-f nodes north wins, then
// east, then south, then west.
func TestTrainAStarTieBreakerIsTheDirectionOrdinalOfTheStep(t *testing.T) {
	grid := pathGrid(t, "0#1")
	a := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 1))
	from := NewTrainState(Coord{5, 5})

	assert.Equal(t, int(NORTH), a.TieBreaker(from, NewTrainState(Coord{5, 4})))
	assert.Equal(t, int(EAST), a.TieBreaker(from, NewTrainState(Coord{6, 5})))
	assert.Equal(t, int(SOUTH), a.TieBreaker(from, NewTrainState(Coord{5, 6})))
	assert.Equal(t, int(WEST), a.TieBreaker(from, NewTrainState(Coord{4, 5})))
	assert.Equal(t, int(UNSET), a.TieBreaker(from, NewTrainState(Coord{7, 7})))
}

func TestTrainAStarSuccessorsAreOnlyCellsATrainCanPass(t *testing.T) {
	grid := pathGrid(t,
		".#.",
		"#0#",
		"...",
	)
	grid.GetXY(2, 1).Track = TRACK_NONE
	a := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 0))

	successors := a.Successors(NewTrainState(Coord{1, 1}))

	require.Len(t, successors, 2)
	assert.Equal(t, Coord{1, 0}, successors[0].Coord)
	assert.Equal(t, Coord{0, 1}, successors[1].Coord)
}

// StateKey is TrainState's equals/hashCode, which ignores the back-pointer:
// two states on the same cell are one state to the search.
func TestTrainAStarStateKeyIgnoresThePreviousState(t *testing.T) {
	grid := pathGrid(t, "0#1")
	a := NewTrainAStar(grid, town(t, grid, 0), town(t, grid, 1))

	bare := NewTrainState(Coord{1, 0})
	linked := NewTrainStateFrom(Coord{1, 0}, NewTrainState(Coord{0, 0}))

	assert.Equal(t, a.StateKey(bare), a.StateKey(linked))
}
