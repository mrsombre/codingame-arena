package grid

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

const (
	testArenaPositiveSeed = int64(468706172918629800)
	testArenaNegativeSeed = int64(-468706172918629800)
)

var testArenaExpectedInitialInputPositive = strings.Join([]string{
	"0",
	"32",
	"18",
	"...............##...............",
	"................................",
	"...............##...............",
	"................................",
	"..........#..........#..........",
	"...........#...##...#...........",
	"................................",
	"................................",
	"......#..................#......",
	".......#.......##.......#.......",
	"....#......................#....",
	"....#...##............##...#....",
	".....##....##......##....##.....",
	"...........##......##...........",
	"...##....#............#....##...",
	".##.....#....######....#.....##.",
	".##...###...########...###...##.",
	"################################",
	"4",
	"0",
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"26",
	"2 3",
	"29 3",
	"1 5",
	"30 5",
	"15 7",
	"16 7",
	"3 11",
	"28 11",
	"15 12",
	"16 12",
	"11 1",
	"20 1",
	"0 2",
	"31 2",
	"13 3",
	"18 3",
	"5 5",
	"26 5",
	"7 5",
	"24 5",
	"3 6",
	"28 6",
	"9 7",
	"22 7",
	"1 13",
	"30 13",
	"8",
	"0 7,13:7,14:7,15",
	"1 25,5:25,6:25,7",
	"2 21,1:21,2:21,3",
	"3 20,9:20,10:20,11",
	"4 24,13:24,14:24,15",
	"5 6,5:6,6:6,7",
	"6 10,1:10,2:10,3",
	"7 11,9:11,10:11,11",
}, "\n")

var testArenaExpectedInitialInputNegative = strings.Join([]string{
	"0",
	"36",
	"20",
	"....................................",
	"....................................",
	"....................................",
	"....................................",
	"....................................",
	"...#............................#...",
	"..#..............................#..",
	"....................................",
	".........#.....#....#.....#.........",
	".......#..##....#..#....##..#.......",
	"#....##.....#..........#.....##....#",
	"#.....#..#................#..#.....#",
	".#......##................##......#.",
	"....#......#....#..#....#......#....",
	".....#......#..##..##..#......#.....",
	"........#......##..##......#........",
	"..#..##..##...###..###...##..##..#..",
	"..#####..........##..........#####..",
	"##########.......##.......##########",
	"####################################",
	"4",
	"0",
	"1",
	"2",
	"3",
	"4",
	"5",
	"6",
	"7",
	"40",
	"7 5",
	"28 5",
	"8 7",
	"27 7",
	"2 11",
	"33 11",
	"7 14",
	"28 14",
	"6 15",
	"29 15",
	"13 17",
	"22 17",
	"13 18",
	"22 18",
	"3 0",
	"32 0",
	"7 2",
	"28 2",
	"3 3",
	"32 3",
	"1 4",
	"34 4",
	"6 4",
	"29 4",
	"13 4",
	"22 4",
	"0 6",
	"35 6",
	"6 6",
	"29 6",
	"11 7",
	"24 7",
	"1 8",
	"34 8",
	"3 10",
	"32 10",
	"0 15",
	"35 15",
	"12 17",
	"23 17",
	"8",
	"0 30,11:30,12:30,13",
	"1 30,7:30,8:30,9",
	"2 26,13:26,14:26,15",
	"3 16,6:16,7:16,8",
	"4 5,11:5,12:5,13",
	"5 5,7:5,8:5,9",
	"6 9,13:9,14:9,15",
	"7 19,6:19,7:19,8",
}, "\n")

func buildInitialInput(seed int64, leagueLevel int) string {
	rng := sha1prng.New(seed)
	gm := NewGridMaker(rng, leagueLevel)
	g := gm.Make()

	var lines []string
	lines = append(lines, "0")
	lines = append(lines, fmt.Sprintf("%d", g.Width))
	lines = append(lines, fmt.Sprintf("%d", g.Height))
	for y := 0; y < g.Height; y++ {
		var row strings.Builder
		for x := 0; x < g.Width; x++ {
			if g.GetXY(x, y).Type == TileWall {
				row.WriteByte('#')
			} else {
				row.WriteByte('.')
			}
		}
		lines = append(lines, row.String())
	}
	spawnIslands := g.DetectSpawnIslands()
	birdBodies := make([][]Coord, len(spawnIslands))
	for i, island := range spawnIslands {
		birdBodies[i] = SortedCoordsFromSet(island)
	}
	birdsPerPlayer := len(birdBodies)
	lines = append(lines, fmt.Sprintf("%d", birdsPerPlayer))
	for i := 0; i < birdsPerPlayer; i++ {
		lines = append(lines, fmt.Sprintf("%d", i))
	}
	for i := 0; i < birdsPerPlayer; i++ {
		lines = append(lines, fmt.Sprintf("%d", birdsPerPlayer+i))
	}
	lines = append(lines, fmt.Sprintf("%d", len(g.Apples)))
	for _, c := range g.Apples {
		lines = append(lines, c.ToIntString())
	}
	totalBirds := birdsPerPlayer * 2
	lines = append(lines, fmt.Sprintf("%d", totalBirds))
	for i, body := range birdBodies {
		parts := make([]string, len(body))
		for j, c := range body {
			parts[j] = fmt.Sprintf("%d,%d", c.X, c.Y)
		}
		lines = append(lines, fmt.Sprintf("%d %s", i, strings.Join(parts, ":")))
	}
	for i, body := range birdBodies {
		parts := make([]string, len(body))
		for j, c := range body {
			opp := g.Opposite(c)
			parts[j] = fmt.Sprintf("%d,%d", opp.X, opp.Y)
		}
		lines = append(lines, fmt.Sprintf("%d %s", birdsPerPlayer+i, strings.Join(parts, ":")))
	}
	return strings.Join(lines, "\n")
}

// TestGridMakerSeedDeterminism verifies the GridMaker is deterministic for a given seed.
func TestGridMakerSeedDeterminism(t *testing.T) {
	a := buildInitialInput(testArenaPositiveSeed, 1)
	b := buildInitialInput(testArenaPositiveSeed, 1)
	assert.Equal(t, a, b, "GridMaker should be deterministic for the same seed")
}

func TestGridMakerPositiveParityCheck(t *testing.T) {
	got := buildInitialInput(testArenaPositiveSeed, 4)
	assert.Equalf(t, testArenaExpectedInitialInputPositive, got, "arena parity mismatch for seed=%d", testArenaPositiveSeed)
}

func TestGridMakerNegativeParityCheck(t *testing.T) {
	got := buildInitialInput(testArenaNegativeSeed, 4)
	assert.Equalf(t, testArenaExpectedInitialInputNegative, got, "arena parity mismatch for seed=%d", testArenaNegativeSeed)
}
