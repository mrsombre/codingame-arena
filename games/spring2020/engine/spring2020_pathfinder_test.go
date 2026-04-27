package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPathStraightCorridor(t *testing.T) {
	g := NewGridFromRows([]string{
		"#######",
		"#     #",
		"#######",
	}, false)

	r := FindPath(g, Coord{X: 1, Y: 1}, Coord{X: 5, Y: 1}, nil)
	assert.False(t, r.IsNearest)
	assert.Equal(t, 4, r.WeightedLength)
	assert.Equal(t,
		[]Coord{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1}, {X: 5, Y: 1}},
		r.Path,
	)
}

func TestFindPathAroundWall(t *testing.T) {
	// U-shaped corridor: path must detour down and back up.
	g := NewGridFromRows([]string{
		"#####",
		"# # #",
		"# # #",
		"#   #",
		"#####",
	}, false)

	r := FindPath(g, Coord{X: 1, Y: 1}, Coord{X: 3, Y: 1}, nil)
	assert.False(t, r.IsNearest)
	assert.Equal(t, 6, r.WeightedLength)
	assert.Equal(t, Coord{X: 1, Y: 1}, r.Path[0])
	assert.Equal(t, Coord{X: 3, Y: 1}, r.Path[len(r.Path)-1])
}

func TestFindPathUnreachableReturnsNearest(t *testing.T) {
	g := NewGridFromRows([]string{
		"#####",
		"#   #",
		"#####",
		"#   #",
		"#####",
	}, false)

	r := FindPath(g, Coord{X: 1, Y: 1}, Coord{X: 1, Y: 3}, nil)
	assert.True(t, r.IsNearest)
	assert.NotEmpty(t, r.Path)
	assert.Equal(t, Coord{X: 1, Y: 1}, r.Path[0])
}

func TestFindPathSamePoint(t *testing.T) {
	g := NewGridFromRows([]string{"#####", "#   #", "#####"}, false)
	r := FindPath(g, Coord{X: 2, Y: 1}, Coord{X: 2, Y: 1}, nil)
	assert.Equal(t, 0, r.WeightedLength)
	assert.Equal(t, []Coord{{X: 2, Y: 1}}, r.Path)
	assert.False(t, r.IsNearest)
}

func TestFindPathUsesWrappingForShorterRoute(t *testing.T) {
	rows := make([]string, 0, 3)
	rows = append(rows, "#########")
	rows = append(rows, "         ")
	rows = append(rows, "#########")
	g := NewGridFromRows(rows, true)

	r := FindPath(g, Coord{X: 0, Y: 1}, Coord{X: 8, Y: 1}, nil)
	assert.Equal(t, 1, r.WeightedLength)
	assert.Equal(t, 2, len(r.Path))
}

func TestFindPathWeightedLengthMatchesStepCount(t *testing.T) {
	g := NewGridFromRows([]string{
		"#####",
		"#   #",
		"#####",
	}, false)
	r := FindPath(g, Coord{X: 1, Y: 1}, Coord{X: 3, Y: 1}, nil)
	assert.Equal(t, 2, r.WeightedLength)
	assert.Equal(t, 3, len(r.Path))
}

func TestPathResultHasNoPath(t *testing.T) {
	r := PathResult{WeightedLength: -1}
	assert.True(t, r.HasNoPath())

	r = PathResult{WeightedLength: 0, Path: []Coord{{}}}
	assert.False(t, r.HasNoPath())
}
