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

// TestFindPathMatchesJavaTieBreaking locks in parity with the upstream Java
// AStar / PriorityQueue when multiple paths share the same total length. Each
// expected path was captured by running tmp/AStarTest.java against the same
// grid; if Go's heap ever drifts from Java's PriorityQueue, this fails.
func TestFindPathMatchesJavaTieBreaking(t *testing.T) {
	cases := []struct {
		name         string
		rows         []string
		wraps        bool
		from, to     Coord
		expectedPath []Coord
	}{
		{
			name: "rect_3x6",
			rows: []string{
				"######",
				"#    #",
				"#    #",
				"#    #",
				"######",
			},
			from: Coord{X: 1, Y: 2}, to: Coord{X: 4, Y: 2},
			expectedPath: []Coord{{X: 1, Y: 2}, {X: 2, Y: 2}, {X: 3, Y: 2}, {X: 4, Y: 2}},
		},
		{
			name: "two_detours",
			rows: []string{
				"#######",
				"#     #",
				"# ### #",
				"#     #",
				"#######",
			},
			from: Coord{X: 1, Y: 2}, to: Coord{X: 5, Y: 2},
			expectedPath: []Coord{
				{X: 1, Y: 2}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1},
				{X: 4, Y: 1}, {X: 5, Y: 1}, {X: 5, Y: 2},
			},
		},
		{
			name: "triple_branch",
			rows: []string{
				"#####",
				"#   #",
				"# # #",
				"#   #",
				"#####",
			},
			from: Coord{X: 2, Y: 1}, to: Coord{X: 2, Y: 3},
			expectedPath: []Coord{
				{X: 2, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 3}, {X: 2, Y: 3},
			},
		},
		{
			name: "wrap_tie",
			rows: []string{
				"########",
				"        ",
				"########",
			},
			wraps: true,
			from:  Coord{X: 0, Y: 1}, to: Coord{X: 4, Y: 1},
			expectedPath: []Coord{
				{X: 0, Y: 1}, {X: 7, Y: 1}, {X: 6, Y: 1}, {X: 5, Y: 1}, {X: 4, Y: 1},
			},
		},
		{
			name: "long_corridor_obstacle",
			rows: []string{
				"##########",
				"#        #",
				"#  ####  #",
				"#        #",
				"##########",
			},
			from: Coord{X: 1, Y: 2}, to: Coord{X: 8, Y: 2},
			expectedPath: []Coord{
				{X: 1, Y: 2}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 4, Y: 1},
				{X: 5, Y: 1}, {X: 6, Y: 1}, {X: 7, Y: 1}, {X: 8, Y: 1}, {X: 8, Y: 2},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := NewGridFromRows(c.rows, c.wraps)
			r := FindPath(g, c.from, c.to, nil)
			assert.False(t, r.IsNearest)
			assert.Equal(t, c.expectedPath, r.Path)
		})
	}
}
