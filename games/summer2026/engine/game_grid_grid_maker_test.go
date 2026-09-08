package engine

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// javaRoundFloat is floor(x + 0.5), which rounds every half towards positive
// infinity. Go's math.Round rounds halves away from zero, so the two disagree
// on every negative half — the reason the port cannot use math.Round.
func TestJavaRoundFloatRoundsHalvesTowardsPositiveInfinity(t *testing.T) {
	cases := []struct {
		in   float32
		want int
	}{
		{2.5, 3},
		{2.4999998, 2},
		{-2.5, -2},
		{-2.5000005, -3},
		{0.5, 1},
		{-0.5, 0},
		{-0.5000001, -1},
		{0, 0},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, javaRoundFloat(tc.in), "javaRoundFloat(%v)", tc.in)
	}

	assert.NotEqual(t, int(math.Round(-2.5)), javaRoundFloat(-2.5),
		"math.Round is the wrong rounding for this port")
}

// Grid width is Math.round(h * ASPECT_RATIO) over a float32 expression. These
// are the only seven heights the generator can draw.
func TestGridWidthForEveryDrawableHeight(t *testing.T) {
	want := map[int]int{14: 21, 15: 23, 16: 24, 17: 26, 18: 27, 19: 29, 20: 30}

	for h := MIN_GRID_HEIGHT; h <= MAX_GRID_HEIGHT; h++ {
		assert.Equalf(t, want[h], javaRoundFloat(float32(h)*ASPECT_RATIO), "width for height %d", h)
	}
}

// The generator only ever draws a height in [MIN_GRID_HEIGHT, MAX_GRID_HEIGHT]
// and pairs it with the matching width.
func TestInitDrawsDimensionsInRange(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		m := NewGridMaker()
		m.Init(sha1prng.New(seed), false)

		require.GreaterOrEqual(t, m.h, MIN_GRID_HEIGHT)
		require.LessOrEqual(t, m.h, MAX_GRID_HEIGHT)
		require.Equal(t, javaRoundFloat(float32(m.h)*ASPECT_RATIO), m.w)
	}
}

// The river and mountain counts are float32 expressions too. Keeping them at
// float32 is what stops a later refactor from widening them silently.
func TestRiverAndMountainCountExpressionsStayFloat32(t *testing.T) {
	const w, h = 21, 14

	assert.Equal(t, javaRoundFloat(float32(w*h)*RIVER_TO_LAND_MIN_RATIO), 21)
	assert.IsType(t, float32(0), float32(w*h)*RIVER_TO_LAND_MIN_RATIO)
	assert.IsType(t, float32(0), float32(w*h)*MOUNTAIN_TO_CELL_RATIO)
}

// Region and mountain growth pick a candidate by index out of what Java calls
// a HashSet<Coord>, so the candidate list has to come back in HashMap bucket
// order, not in the order the neighbours were visited. The expectation below
// is the bucket order for these ten coords on a 16-slot table; walking the
// seeds in order would instead give (3,2) (4,3) (3,4) (2,3) (3,3) (4,4) (3,5)
// (2,4) (2,2) (1,3).
func TestAvailableNeighboursComeBackInJavaHashOrder(t *testing.T) {
	grid := NewGrid(7, 7)
	for _, tile := range grid.Cells {
		tile.ZoneID = -1
	}
	maker := &GridMaker{grid: grid, w: 7, h: 7}

	got := maker.getAvailableNeighbours(grid, []Coord{{3, 3}, {3, 4}, {2, 3}}, func(tile *Tile) bool {
		return tile.GetZoneID() == -1
	})

	assert.Equal(t, []Coord{
		{3, 2}, {4, 3},
		{3, 3}, {4, 4}, {2, 2},
		{3, 4}, {2, 3},
		{3, 5}, {2, 4}, {1, 3},
	}, got)
}

// Every generated map has to be usable by the rest of the engine: each cell
// belongs to a real zone, each town sits on a plains tile carrying its own id,
// and no two towns share a cell.
func TestGeneratedMapsAreWellFormed(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		maker := NewGridMaker()
		maker.Init(sha1prng.New(seed), false)
		grid := maker.Make()

		require.Equal(t, grid.Width*grid.Height, len(grid.Cells))
		for _, tile := range grid.Cells {
			require.GreaterOrEqualf(t, tile.ZoneID, 0, "seed %d: cell %v has no region", seed, tile.Coord)
			require.Lessf(t, tile.ZoneID, len(grid.Zones), "seed %d: cell %v region out of range", seed, tile.Coord)
		}

		seen := map[Coord]bool{}
		for id, town := range grid.Towns {
			require.Equalf(t, id, town.ID, "seed %d: towns are renumbered from 0", seed)
			require.Falsef(t, seen[town.Coord], "seed %d: two towns on %v", seed, town.Coord)
			seen[town.Coord] = true

			tile := grid.Get(town.Coord)
			require.Equalf(t, town.ID, tile.TownID, "seed %d: town %d not marked on its tile", seed, town.ID)
			require.Equalf(t, TYPE_GRASS, tile.Type, "seed %d: town %d is not on plains", seed, town.ID)
		}
	}
}

// A desired connection is unilateral by construction: makeTownConnections
// strips any connection whose target already asks for this town.
func TestDesiredConnectionsAreNeverReciprocal(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		maker := NewGridMaker()
		maker.Init(sha1prng.New(seed), false)
		grid := maker.Make()

		for _, town := range grid.Towns {
			for _, other := range town.DesiredConnections {
				require.NotSamef(t, town, other, "seed %d: town %d wants itself", seed, town.ID)
				require.Falsef(t, contains(other.DesiredConnections, town),
					"seed %d: towns %d and %d both want each other", seed, town.ID, other.ID)
			}
		}
	}
}

func TestPOIsAreNotPlacedWhileTheSideQuestIsDisabled(t *testing.T) {
	maker := NewGridMaker()
	maker.Init(sha1prng.New(1), false)
	grid := maker.Make()

	assert.False(t, grid.HasPOI)
	for _, tile := range grid.Cells {
		assert.NotEqual(t, TYPE_POI, tile.Type)
	}
}
