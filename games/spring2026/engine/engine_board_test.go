package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Board-level tests for BFS distance / neighbour wiring / cell lookups. The
// full map-gen pipeline is covered by engine_board_seed_test.go; here we
// focus on the BFS and lookup primitives in isolation.

func TestBoardGetDistancesBFSAndUnreachable(t *testing.T) {
	// Rules: BFS over walkable cells only. The wall column on x=2 splits the
	// board into two regions; cells right of the wall are unreachable from
	// (0,0). Shack cells (non-walkable) get -1 too, including the start
	// cell's BFS-stop neighbours.
	board, _, _ := loadScenario(t, 4, []string{
		"0.#.",
		"..#.",
		"..#1",
	})
	dist := board.GetDistances(board.GetCell(0, 0))

	assert.Equal(t, 0, dist[0][0], "start has distance 0")
	assert.Equal(t, 1, dist[1][0], "right neighbour")
	assert.Equal(t, 2, dist[1][1])
	assert.Equal(t, -1, dist[2][0], "wall is unreachable")
	assert.Equal(t, -1, dist[3][0], "blocked by wall column")
	assert.Equal(t, -1, dist[3][2], "opp shack non-walkable")
}

func TestBoardGetNextCellReturnsTargetWhenReachable(t *testing.T) {
	// When sourceDist[target] is within movementSpeed, GetNextCell returns
	// the target itself (no need to look up alternates).
	board, p0, _ := loadScenario(t, 4, []string{
		"0.......",
		"........",
		".......1",
	})
	u := spawnUnit(board, p0, [4]int{3, 1, 1, 0}, 0, 0)
	got := board.GetNextCell(u, board.GetCell(3, 0))
	assert.Equal(t, board.GetCell(3, 0), got)
}

func TestBoardGetUnitsByPlayerAndCell(t *testing.T) {
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	a := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	b := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 2, 0)
	c := spawnUnit(board, p1, [4]int{1, 1, 1, 0}, 1, 0)

	p0Units := board.GetUnitsByPlayerID(0)
	assert.Equal(t, []*Unit{a, b}, p0Units, "filtered + ordered by insertion")

	atCell := board.GetUnitsByCell(board.GetCell(1, 0))
	assert.Len(t, atCell, 2, "stacked own-team and opponent at the same cell")
	assert.Contains(t, atCell, a)
	assert.Contains(t, atCell, c)
}

func TestBoardGetUnitLookup(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	assert.Same(t, u, board.GetUnit(u.ID))
	assert.Nil(t, board.GetUnit(999), "unknown id returns nil")
}
