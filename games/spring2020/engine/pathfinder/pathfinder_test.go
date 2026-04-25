package pathfinder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

func TestFindPathStraightCorridor(t *testing.T) {
	g := grid.NewGridFromRows([]string{
		"#######",
		"#     #",
		"#######",
	}, false)

	r := FindPath(g, grid.Coord{X: 1, Y: 1}, grid.Coord{X: 5, Y: 1}, nil)
	assert.False(t, r.IsNearest)
	assert.Equal(t, 4, r.WeightedLength)
	assert.Equal(t,
		[]grid.Coord{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1}, {X: 5, Y: 1}},
		r.Path,
	)
}

func TestFindPathAroundWall(t *testing.T) {
	// U-shaped corridor: path must detour down and back up.
	g := grid.NewGridFromRows([]string{
		"#####",
		"# # #",
		"# # #",
		"#   #",
		"#####",
	}, false)

	r := FindPath(g, grid.Coord{X: 1, Y: 1}, grid.Coord{X: 3, Y: 1}, nil)
	assert.False(t, r.IsNearest)
	assert.Equal(t, 6, r.WeightedLength)
	assert.Equal(t, grid.Coord{X: 1, Y: 1}, r.Path[0])
	assert.Equal(t, grid.Coord{X: 3, Y: 1}, r.Path[len(r.Path)-1])
}

func TestFindPathUnreachableReturnsNearest(t *testing.T) {
	g := grid.NewGridFromRows([]string{
		"#####",
		"#   #",
		"#####",
		"#   #",
		"#####",
	}, false)

	r := FindPath(g, grid.Coord{X: 1, Y: 1}, grid.Coord{X: 1, Y: 3}, nil)
	assert.True(t, r.IsNearest)
	// Nearest reachable cell closest to target — must end within the start's island.
	assert.NotEmpty(t, r.Path)
	assert.Equal(t, grid.Coord{X: 1, Y: 1}, r.Path[0])
}

func TestFindPathSamePoint(t *testing.T) {
	g := grid.NewGridFromRows([]string{"#####", "#   #", "#####"}, false)
	r := FindPath(g, grid.Coord{X: 2, Y: 1}, grid.Coord{X: 2, Y: 1}, nil)
	assert.Equal(t, 0, r.WeightedLength)
	assert.Equal(t, []grid.Coord{{X: 2, Y: 1}}, r.Path)
	assert.False(t, r.IsNearest)
}

func TestFindPathUsesWrappingForShorterRoute(t *testing.T) {
	// A single straight row that wraps horizontally. Shorter via wrap.
	rows := make([]string, 0, 3)
	rows = append(rows, "#########")
	rows = append(rows, "         ")
	rows = append(rows, "#########")
	g := grid.NewGridFromRows(rows, true)

	r := FindPath(g, grid.Coord{X: 0, Y: 1}, grid.Coord{X: 8, Y: 1}, nil)
	// Going via wrap: 0 → 8 is one step; direct path would be 8.
	assert.Equal(t, 1, r.WeightedLength)
	assert.Equal(t, 2, len(r.Path))
}

func TestFindPathWeightedLengthMatchesStepCount(t *testing.T) {
	g := grid.NewGridFromRows([]string{
		"#####",
		"#   #",
		"#####",
	}, false)
	r := FindPath(g, grid.Coord{X: 1, Y: 1}, grid.Coord{X: 3, Y: 1}, nil)
	assert.Equal(t, 2, r.WeightedLength)
	assert.Equal(t, 3, len(r.Path))
}

func TestResultHasNoPath(t *testing.T) {
	r := Result{WeightedLength: -1}
	assert.True(t, r.HasNoPath())

	r = Result{WeightedLength: 0, Path: []grid.Coord{{}}}
	assert.False(t, r.HasNoPath())
}
