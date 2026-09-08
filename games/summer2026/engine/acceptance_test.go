package engine

import (
	"strconv"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Acceptance tests exercise the rules of Back Track King on hand-built grids,
// bypassing map generation entirely, so rule coverage does not depend on the
// generator being correct first.

// townSpec places one town and names the towns it wants connected to it.
// DesiredConnections are unilateral, so the ids listed here are exactly what
// the initialization block reports for this town.
type townSpec struct {
	ID      int
	X, Y    int
	Desires []int
}

// scenario is the input to loadScenario. Terrain is required; Regions is an
// optional parallel layer where each distinct rune is one region, numbered by
// first appearance in row-major order. With Regions omitted every cell lands
// in region 0. Tracks is an optional parallel layer laying rails down
// directly, bypassing the paint economy.
type scenario struct {
	Terrain []string
	Regions []string
	Tracks  []string
	Towns   []townSpec
	// League selects the tutorial win condition; zero means the full game.
	League int
}

// loadScenario builds a Game on a hand-built grid. Terrain characters:
// '.' plains, '~' river, '^' mountain.
func loadScenario(t *testing.T, sc scenario) (*Game, *Player, *Player) {
	t.Helper()
	require.NotEmpty(t, sc.Terrain, "scenario needs terrain")

	width := len(sc.Terrain[0])
	height := len(sc.Terrain)
	grid := NewGrid(width, height)

	for y, row := range sc.Terrain {
		require.Equalf(t, width, len(row), "terrain row %d length mismatch", y)
		for x, ch := range row {
			tile := grid.GetXY(x, y)
			switch ch {
			case '.':
				tile.Type = TYPE_GRASS
			case '~':
				tile.Type = TYPE_WATER
			case '^':
				tile.Type = TYPE_MOUNTAIN
			default:
				t.Fatalf("unknown terrain char %q at %d,%d", ch, x, y)
			}
		}
	}

	assignRegions(t, grid, sc.Regions)
	assignTracks(t, grid, sc.Tracks)

	byID := make(map[int]*Town, len(sc.Towns))
	for _, spec := range sc.Towns {
		town := NewTown(spec.ID, Coord{spec.X, spec.Y})
		tile := grid.Get(town.Coord)
		require.Truef(t, tile.IsValid(), "town %d is off the grid", spec.ID)
		tile.TownID = spec.ID
		grid.Towns = append(grid.Towns, town)
		grid.Zones[tile.ZoneID].AddTown(town)
		byID[spec.ID] = town
	}
	for _, spec := range sc.Towns {
		for _, id := range spec.Desires {
			other, ok := byID[id]
			require.Truef(t, ok, "town %d desires unknown town %d", spec.ID, id)
			byID[spec.ID].DesiredConnections = append(byID[spec.ID].DesiredConnections, other)
		}
	}

	league := sc.League
	if league == 0 {
		league = DEFAULT_LEAGUE
	}

	p0, p1 := NewPlayer(0), NewPlayer(1)
	game := NewGame(nil, league)
	game.Init([]*Player{p0, p1})
	game.Grid = grid

	return game, p0, p1
}

// assignRegions numbers each distinct rune of the region layer in row-major
// order of first appearance, and builds one Zone per number. A nil layer puts
// every cell in region 0.
func assignRegions(t *testing.T, grid *Grid, regions []string) {
	t.Helper()
	if regions == nil {
		coords := grid.Coords()
		for _, tile := range grid.Cells {
			tile.ZoneID = 0
		}
		grid.Zones = []*Zone{NewZone(0, coords)}
		return
	}

	require.Equal(t, grid.Height, len(regions), "region layer height mismatch")
	ids := map[rune]int{}
	coordsByZone := map[int][]Coord{}
	for y, row := range regions {
		require.Equalf(t, grid.Width, len(row), "region row %d length mismatch", y)
		for x, ch := range row {
			id, seen := ids[ch]
			if !seen {
				id = len(ids)
				ids[ch] = id
			}
			grid.GetXY(x, y).ZoneID = id
			coordsByZone[id] = append(coordsByZone[id], Coord{x, y})
		}
	}
	grid.Zones = make([]*Zone, len(ids))
	for id, coords := range coordsByZone {
		grid.Zones[id] = NewZone(id, coords)
	}
}

// assignTracks lays rails from the optional track layer. Characters: '.' no
// track, '0' and '1' a track owned by that player, 'n' a neutral one.
func assignTracks(t *testing.T, grid *Grid, tracks []string) {
	t.Helper()
	if tracks == nil {
		return
	}

	require.Equal(t, grid.Height, len(tracks), "track layer height mismatch")
	for y, row := range tracks {
		require.Equalf(t, grid.Width, len(row), "track row %d length mismatch", y)
		for x, ch := range row {
			switch ch {
			case '.':
			case '0':
				grid.GetXY(x, y).Track = 0
			case '1':
				grid.GetXY(x, y).Track = 1
			case 'n':
				grid.GetXY(x, y).Track = TRACK_NEUTRAL
			default:
				t.Fatalf("unknown track char %q at %d,%d", ch, x, y)
			}
		}
	}
}

// runTurn parses both players' output lines and runs one game update, in the
// same order the referee does.
func runTurn(game *Game, p0Cmd, p1Cmd string) {
	cm := NewCommandManager(game)
	game.ResetGameTurnData()
	for i, cmd := range []string{p0Cmd, p1Cmd} {
		player := game.Players[i]
		if player.IsDeactivated() {
			continue
		}
		cm.ParseCommands(player, []string{cmd})
	}
	game.PerformGameUpdate(game.Turn)
}

// ——— initialization serialization ————————————————————————————————————————

func TestGlobalInfoDescribesGridRegionsAndTowns(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{
			"..~",
			".^.",
		},
		Regions: []string{
			"aab",
			"aab",
		},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 1},
		},
	})

	assert.Equal(t, []string{
		"0",
		"3", "2",
		"0 0", "0 0", "1 1",
		"0 0", "0 2", "1 0",
		"2",
		"0 0 0 1",
		"1 2 1 x",
	}, SerializeGlobalInfoFor(p0, game))

	// Only the leading player-id line differs between the two sides.
	assert.Equal(t, "1", SerializeGlobalInfoFor(p1, game)[0])
	assert.Equal(t,
		SerializeGlobalInfoFor(p0, game)[1:],
		SerializeGlobalInfoFor(p1, game)[1:],
	)
}

func TestGlobalInfoUsesXForATownWithNoDesiredConnections(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0},
			{ID: 1, X: 1, Y: 0, Desires: []int{0}},
		},
	})

	lines := SerializeGlobalInfoFor(p0, game)
	assert.Equal(t, "0 0 0 x", lines[len(lines)-2])
	assert.Equal(t, "1 1 0 0", lines[len(lines)-1])
}

func TestGlobalInfoListsMultipleDesiredConnectionsCommaSeparated(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1, 2}},
			{ID: 1, X: 1, Y: 0},
			{ID: 2, X: 2, Y: 0},
		},
	})

	assert.Equal(t, "0 0 0 1,2", SerializeGlobalInfoFor(p0, game)[7])
}

// ——— per-turn serialization ——————————————————————————————————————————————

func TestFrameInfoReportsOwnScoreFirstThenTheOpponents(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{Terrain: []string{"."}})
	p0.SetScore(7)
	p1.SetScore(3)

	assert.Equal(t, []string{"7", "3", "-1 0 0 x"}, SerializeFrameInfoFor(p0, game))
	assert.Equal(t, []string{"3", "7", "-1 0 0 x"}, SerializeFrameInfoFor(p1, game))
}

func TestFrameInfoReportsTrackOwnerInstabilityAndInked(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Regions: []string{"ab"},
	})
	game.Grid.GetXY(0, 0).Track = 1
	game.Grid.GetXY(1, 0).Track = TRACK_NEUTRAL
	game.Grid.Zones[0].Instability = 3
	game.Grid.Zones[1].Inked = true

	lines := SerializeFrameInfoFor(p0, game)
	assert.Equal(t, "1 3 0 x", lines[2])
	assert.Equal(t, "2 0 1 x", lines[3])
}

func TestFrameInfoSortsActiveConnectionsAsStrings(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})
	game.Grid.GetXY(0, 0).ActiveConnections = []ScheduleStep{
		{FromTownID: 2, ToTownID: 3},
		{FromTownID: 10, ToTownID: 2},
		{FromTownID: 1, ToTownID: 2},
	}

	// Java sorts the rendered pairs with String.compareTo, so "10-2" sorts
	// before "2-3" rather than after it.
	assert.Equal(t, "-1 0 0 1-2,10-2,2-3", SerializeFrameInfoFor(p0, game)[2])
}

// ——— command parsing ——————————————————————————————————————————————————————

func TestWaitIsAcceptedAsANoOpIntent(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})
	NewCommandManager(game).ParseCommands(p0, []string{"WAIT"})

	assert.False(t, p0.IsDeactivated())
	assert.Len(t, p0.Intents, 1)
	assert.Equal(t, ACTION_WAIT, p0.Intents[0].Type)
}

func TestUnparseableCommandDisqualifiesWithTheExpectedSyntax(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})
	NewCommandManager(game).ParseCommands(p0, []string{"WAIT PLEASE"})

	assert.True(t, p0.IsDeactivated())
	assert.Equal(t, -1, p0.GetScore())
	assert.Equal(t,
		"Invalid Input: Expected WAIT but got 'WAIT PLEASE'",
		p0.DeactivationReason(),
	)
}

func TestEmptyOutputDisqualifies(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})
	NewCommandManager(game).ParseCommands(p0, nil)

	assert.True(t, p0.IsDeactivated())
	assert.Equal(t,
		"Invalid Input: Expected AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT but got ''",
		p0.DeactivationReason(),
	)
}

// ——— turn economy ————————————————————————————————————————————————————————

func TestPaintPointsResetEachTurnAndDoNotAccumulate(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})

	runTurn(game, "WAIT", "WAIT")
	assert.Equal(t, PASSIVE_INCOME, p0.Dosh)
	assert.Equal(t, BLOT_POINTS_PER_TURN, p0.BlotPoints)

	runTurn(game, "WAIT", "WAIT")
	assert.Equal(t, PASSIVE_INCOME, p0.Dosh)
	assert.Equal(t, BLOT_POINTS_PER_TURN, p0.BlotPoints)
}

func TestTrackCostIsChargedPerTerrain(t *testing.T) {
	cases := map[string]struct {
		terrain string
		cost    int
	}{
		"plains":   {".", GRASS_COST_MULTIPLIER * BASE_RAIL_COST},
		"river":    {"~", RIVER_COST_MULTIPLIER * BASE_RAIL_COST},
		"mountain": {"^", MOUNTAIN_COST_MULTIPLIER * BASE_RAIL_COST},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			game, p0, _ := loadScenario(t, scenario{Terrain: []string{tc.terrain}})

			runTurn(game, "PLACE_TRACK 0 0", "WAIT")

			assert.Equal(t, PASSIVE_INCOME-tc.cost, p0.Dosh)
			assert.Equal(t, 0, game.Grid.GetXY(0, 0).Track)
			assert.Empty(t, game.Summary)
		})
	}
}

// Income assigns rather than accumulates, so a turn spent waiting buys nothing
// on the next one: three plains tracks a turn is the ceiling however long the
// player saves.
func TestUnspentPaintDoesNotCarryIntoTheNextTurn(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"...."}})

	runTurn(game, "WAIT", "WAIT")
	runTurn(game, "PLACE_TRACK 0 0;PLACE_TRACK 1 0;PLACE_TRACK 2 0;PLACE_TRACK 3 0", "WAIT")

	assert.Equal(t, 0, p0.Dosh)
	assert.Equal(t, 0, game.Grid.GetXY(2, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(3, 0).Track)
	assert.False(t, p0.IsDeactivated())
	assert.Equal(t, []string{
		"¤RED¤Player 0 Not enough track points to build a track at (3, 0).§RED§",
	}, game.Summary)
}

// ——— placement rejection ——————————————————————————————————————————————————

// Every rejection below is reported and skipped. None of them deactivates the
// player: only unparseable output does that, which
// TestUnparseableCommandDisqualifiesWithTheExpectedSyntax covers.
func TestIllegalPlacementsAreSkippedWithAnErrorAndDoNotDisqualify(t *testing.T) {
	cases := map[string]struct {
		setup   func(game *Game)
		command string
		summary string
	}{
		"off grid": {
			command: "PLACE_TRACK 9 9",
			summary: "¤RED¤Player 0 Not part of grid: (9, 9)§RED§",
		},
		"on a town": {
			command: "PLACE_TRACK 0 0",
			summary: "¤RED¤Player 0 Cannot place tracks on a town at (0, 0)§RED§",
		},
		"on an existing track": {
			setup:   func(game *Game) { game.Grid.GetXY(1, 0).Track = 1 },
			command: "PLACE_TRACK 1 0",
			summary: "¤RED¤Player 0 Cannot place tracks on existing tracks at (1, 0)§RED§",
		},
		"in an inked region": {
			setup:   func(game *Game) { game.Grid.Zones[0].Inked = true },
			command: "PLACE_TRACK 1 0",
			summary: "¤RED¤Player 0 Cannot build in region 0§RED§",
		},
		"without enough paint": {
			command: "PLACE_TRACK 1 0;PLACE_TRACK 2 0",
			summary: "¤RED¤Player 0 Not enough track points to build a track at (2, 0).§RED§",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			game, p0, _ := loadScenario(t, scenario{
				Terrain: []string{".^^"},
				Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
			})
			if tc.setup != nil {
				tc.setup(game)
			}
			doshBefore := PASSIVE_INCOME

			runTurn(game, tc.command, "WAIT")

			assert.False(t, p0.IsDeactivated())
			assert.NotEqual(t, -1, p0.GetScore())
			assert.Contains(t, game.Summary, tc.summary)
			assert.GreaterOrEqual(t, p0.Dosh, 0)
			assert.LessOrEqual(t, p0.Dosh, doshBefore)
		})
	}
}

// The same player naming one cell twice pays once; the second attempt reads as
// a placement onto a track it has already bought this turn.
func TestPlacingTwiceOnOneCellInATurnIsRejectedTheSecondTime(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})

	runTurn(game, "PLACE_TRACK 0 0;PLACE_TRACK 0 0", "WAIT")

	assert.Equal(t, PASSIVE_INCOME-1, p0.Dosh)
	assert.Equal(t, 0, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, []string{
		"¤RED¤Player 0 Cannot place tracks on existing tracks at (0, 0)§RED§",
	}, game.Summary)
}

// ——— contested placement ——————————————————————————————————————————————————

func TestBothPlayersClaimingOneCellYieldsNeutralOwnershipAndBothPay(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{Terrain: []string{".."}})

	runTurn(game, "PLACE_TRACK 0 0", "PLACE_TRACK 0 0")

	assert.Equal(t, TRACK_NEUTRAL, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, PASSIVE_INCOME-1, p0.Dosh)
	assert.Equal(t, PASSIVE_INCOME-1, p1.Dosh)
	assert.Empty(t, game.Summary)
}

// Ownership is written only after both players have been charged, so player 1
// does not see player 0's claim as an existing track and is not blocked by it.
func TestContestedCellDoesNotBlockTheSecondPlayersOtherPlacements(t *testing.T) {
	game, _, p1 := loadScenario(t, scenario{Terrain: []string{".."}})

	runTurn(game, "PLACE_TRACK 0 0", "PLACE_TRACK 0 0;PLACE_TRACK 1 0")

	assert.Equal(t, TRACK_NEUTRAL, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, 1, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, PASSIVE_INCOME-2, p1.Dosh)
}

// The distinction the engine has to keep straight: an illegal action costs the
// offender that action, an unparseable one costs it the match.
func TestIllegalActionSkipsButUnparseableOutputDisqualifies(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{".."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})

	runTurn(game, "PLACE_TRACK 0 0", "PLACE_TRACK 0")

	assert.False(t, p0.IsDeactivated())
	assert.True(t, p1.IsDeactivated())
	assert.Equal(t, -1, p1.GetScore())
}

func TestPlacementCountersAreTalliedPerPlayerAndTerrain(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{Terrain: []string{".~^"}})

	runTurn(game, "PLACE_TRACK 0 0;PLACE_TRACK 1 0", "PLACE_TRACK 2 0")

	assert.Equal(t, [2]int{2, 1}, game.PlacedTracks)
	assert.Equal(t, [2]int{1, 0}, game.TracksPlacedOnPlains)
	assert.Equal(t, [2]int{1, 0}, game.TracksPlacedOnRiver)
	assert.Equal(t, [2]int{0, 1}, game.TracksPlacedOnMountains)
}

func TestIntentsAreClearedBetweenTurns(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})

	runTurn(game, "WAIT", "WAIT")
	runTurn(game, "WAIT", "WAIT")

	assert.Len(t, p0.Intents, 1)
}

// ——— autobuild expansion ——————————————————————————————————————————————————

// AUTOPLACE is rewritten into placements before DoActions runs, so everything
// downstream — cost, rejection, contested ownership — treats them as if the
// bot had spelled them out itself.
func TestAutoplaceExpandsToTheCheapestTrackSequence(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{
			"...",
			".~.",
		},
		Towns: []townSpec{{ID: 0, X: 0, Y: 1}, {ID: 1, X: 2, Y: 1}},
	})

	// One river crossing costs 2; the dry way round the top costs 3.
	runTurn(game, "AUTOPLACE 0 1 2 1", "WAIT")

	assert.Equal(t, 0, game.Grid.GetXY(1, 1).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(2, 0).Track)
	assert.Equal(t, PASSIVE_INCOME-RIVER_COST_MULTIPLIER, p0.Dosh)
	assert.Equal(t, [2]int{1, 0}, game.AutobuildCalled)
	assert.Empty(t, game.Summary)
}

func TestAutoplaceThatCannotReachTheGoalPlacesNothing(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"..."}})

	runTurn(game, "AUTOPLACE 0 0 9 9", "WAIT")

	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, PASSIVE_INCOME, p0.Dosh)
	assert.Equal(t, [2]int{1, 0}, game.AutobuildCalled)
	assert.Empty(t, game.Summary)
}

func TestOnlyOneAutoplaceIsHonouredPerTurn(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{
			"..",
			"..",
		},
	})

	runTurn(game, "AUTOPLACE 0 0 1 0;AUTOPLACE 0 1 1 1", "WAIT")

	assert.Equal(t, 0, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, 0, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(0, 1).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 1).Track)
	assert.Equal(t, [2]int{1, 0}, game.AutobuildCalled)
	assert.Equal(t, []string{
		"¤RED¤Player 0 Only one autobuild action allowed per turn.§RED§",
	}, game.Summary)
	assert.False(t, p0.IsDeactivated())
}

// The planner ignores the budget, so a plan longer than the turn's paint is
// cut off where the money runs out. What was already laid stays.
func TestAutoplaceStopsAtTheLastAffordableTrack(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"......."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}, {ID: 1, X: 6, Y: 0}},
	})

	runTurn(game, "AUTOPLACE 1 0 6 0", "WAIT")

	assert.Equal(t, 0, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, 0, game.Grid.GetXY(2, 0).Track)
	assert.Equal(t, 0, game.Grid.GetXY(3, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(4, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(5, 0).Track)
	assert.Equal(t, 0, p0.Dosh)
	assert.Equal(t, []string{
		"¤RED¤Player 0 Autobuild interrupted: not enough track points to build a track at (4, 0).§RED§",
	}, game.Summary)
}

// The interrupt only silences the rest of the expansion. A manual placement
// after it is still considered, and still rejected on its own merits.
func TestAutoplaceInterruptDoesNotSilenceLaterManualActions(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"......."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}, {ID: 1, X: 6, Y: 0}},
	})

	runTurn(game, "AUTOPLACE 1 0 6 0;PLACE_TRACK 5 0", "WAIT")

	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(5, 0).Track)
	assert.Equal(t, []string{
		"¤RED¤Player 0 Autobuild interrupted: not enough track points to build a track at (4, 0).§RED§",
		"¤RED¤Player 0 Not enough track points to build a track at (5, 0).§RED§",
	}, game.Summary)
}

// Expanded placements go through the same rejection gate as manual ones: the
// planner routes around an inked region rather than into it, and when it
// cannot, nothing is placed.
func TestAutoplaceObeysTheInkedRegionRule(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Regions: []string{"aba"},
	})
	game.Grid.Zones[1].Inked = true

	runTurn(game, "AUTOPLACE 0 0 2 0", "WAIT")

	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(0, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, PASSIVE_INCOME, p0.Dosh)
}

// Two expansions crossing the same cell contest it exactly as two manual
// placements would.
func TestAutoplacedTracksCanBeContestedIntoNeutralOwnership(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}, {ID: 1, X: 2, Y: 0}},
	})

	runTurn(game, "AUTOPLACE 0 0 2 0", "AUTOPLACE 2 0 0 0")

	assert.Equal(t, TRACK_NEUTRAL, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, [2]int{1, 1}, game.AutobuildCalled)
}

// A plan that ends on the opponent's rail block still only pays for the cells
// it lays itself.
func TestAutoplaceRidesExistingTrackWhoeverOwnsIt(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"......"},
		Tracks:  []string{"..111."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}, {ID: 1, X: 5, Y: 0}},
	})

	runTurn(game, "AUTOPLACE 0 0 5 0", "WAIT")

	assert.Equal(t, 0, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, 1, game.Grid.GetXY(2, 0).Track)
	assert.Equal(t, PASSIVE_INCOME-1, p0.Dosh)
	assert.Empty(t, game.Summary)
}

// ——— connections and scoring —————————————————————————————————————————————

func TestConnectionIsDiscoveredWhenTrackLinksATownToItsDesiredCounterpart(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Tracks:  []string{".000."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	town0 := game.Grid.Towns[0]
	require.Len(t, town0.ActiveConnections, 1)
	assert.Equal(t, 1, town0.ActiveConnections[0].ID)
	assert.Equal(t,
		[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}},
		town0.Paths[1],
	)
}

func TestNoConnectionIsDiscoveredAcrossAGapInTheTrack(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Tracks:  []string{".0.0."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.Empty(t, game.Grid.Towns[0].ActiveConnections)
	assert.Empty(t, game.Grid.Towns[0].Paths)
}

// Each owned track on the path pays one point per turn, so an unchanged
// connection keeps paying for as long as it stands.
func TestEachOwnedTrackOnAConnectionScoresOnePointPerTurn(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Tracks:  []string{".001."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")
	assert.Equal(t, 2, p0.GetScore())
	assert.Equal(t, 1, p1.GetScore())

	runTurn(game, "WAIT", "WAIT")
	assert.Equal(t, 4, p0.GetScore())
	assert.Equal(t, 2, p1.GetScore())
}

// A path runs over towns and can run over contested track, and neither is
// owned by a player index, so neither scores for anybody.
func TestNeutralTrackAndTownsScoreForNobody(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Tracks:  []string{".nnn."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	require.Len(t, game.Grid.Towns[0].ActiveConnections, 1)
	assert.Equal(t, 0, p0.GetScore())
	assert.Equal(t, 0, p1.GetScore())
}

// DesiredConnections are unilateral. When both towns want each other there
// are two connections over the same rails, and each pays out separately.
func TestAMutuallyDesiredPairScoresTwicePerTurn(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"...."},
		Tracks:  []string{".00."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 3, Y: 0, Desires: []int{0}},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t, 4, p0.GetScore())
}

// Scoring is recomputed from scratch each turn, so a connection that is cut
// simply stops paying.
func TestABrokenConnectionStopsScoring(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Tracks:  []string{".000."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")
	require.Equal(t, 3, p0.GetScore())

	game.Grid.GetXY(2, 0).Track = TRACK_NONE
	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t, 3, p0.GetScore())
	assert.Empty(t, game.Grid.Towns[0].ActiveConnections)
}

// Both ways round are five cells. NORTH precedes SOUTH in the neighbour scan
// order, so the northern route is the one that scores — the southern rails
// earn their owner nothing.
func TestEqualLengthPathsAreBrokenByNorthEastSouthWestPriority(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{
			"...",
			"...",
			"...",
		},
		Tracks: []string{
			"000",
			"...",
			"111",
		},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 1, Desires: []int{1}},
			{ID: 1, X: 2, Y: 1},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t,
		[]Coord{{0, 1}, {0, 0}, {1, 0}, {2, 0}, {2, 1}},
		game.Grid.Towns[0].Paths[1],
	)
	assert.Equal(t, 3, p0.GetScore())
	assert.Equal(t, 0, p1.GetScore())
}

func TestScoringCountsEveryConnectionATrackParticipatesIn(t *testing.T) {
	// Both towns 1 and 2 are reachable from town 0 over the same middle
	// track, and that one cell is paid for once per connection.
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Tracks:  []string{".0."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0, Desires: []int{0}},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t, 2, p0.GetScore())
}

func TestConnectionMetricsRecordDetourLengthAndOwnershipShare(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{
			"...",
			"...",
		},
		Tracks: []string{
			"00.",
			".0.",
		},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 1},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	// The path is four cells for a Manhattan distance of three, and player 0
	// owns three of the four.
	assert.Equal(t, [2]int{1, 0}, game.ExtraTilesInConnection)
	assert.Equal(t, [2]int{1, 0}, game.TrackOwnershipPercentagePerActiveConnectionTotal)
	assert.Equal(t, float32(0.75), game.TrackOwnershipPercentagePerActiveConnection[0])
}

// ——— connections in the per-turn serialization ————————————————————————————

func TestFrameInfoReportsConnectionsAndScoresAsTheBotSeesThem(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{"...."},
		Tracks:  []string{".01."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 3, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t, []string{
		"1", "1",
		"-1 0 0 0-1",
		"0 0 0 0-1",
		"1 0 0 0-1",
		"-1 0 0 0-1",
	}, SerializeFrameInfoFor(p0, game))

	// Player 1 reads the same board with the score pair swapped.
	assert.Equal(t, []string{"1", "1"}, SerializeFrameInfoFor(p1, game)[:2])
}

func TestFrameInfoReportsCellsOffEveryConnectionWithTheXSentinel(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{
			"...",
			"...",
		},
		Tracks: []string{
			".0.",
			"000",
		},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")

	// The connection runs along the top row, so the bottom row is on no path.
	lines := SerializeFrameInfoFor(p0, game)
	assert.Equal(t, "0 0 0 0-1", lines[3])
	assert.Equal(t, "0 0 0 x", lines[5])
}

// Tile stamps are rebuilt from scratch every turn, so a connection that goes
// away leaves nothing behind on the cells it used to cross.
func TestConnectionStampsAreClearedWhenTheConnectionGoesAway(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"...."},
		Tracks:  []string{".00."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 3, Y: 0},
		},
	})

	runTurn(game, "WAIT", "WAIT")
	require.Equal(t, "0 0 0 0-1", SerializeFrameInfoFor(p0, game)[3])

	game.Grid.GetXY(1, 0).Track = TRACK_NONE
	runTurn(game, "WAIT", "WAIT")

	assert.Equal(t, "0 0 0 x", SerializeFrameInfoFor(p0, game)[4])
}

// ——— disruption, instability and inking ——————————————————————————————————

// Both DISRUPT forms end up in the same place: DISRUPT names a region
// directly, DISRUPT x y resolves the region from the cell. Neither is
// case-sensitive.
func TestBothDisruptFormsRaiseInstabilityOnTheSameRegion(t *testing.T) {
	cases := map[string]string{
		"region id":       "DISRUPT 1",
		"coordinate":      "DISRUPT 1 0",
		"lowercase":       "disrupt 1",
		"lowercase coord": "disrupt 1 0",
	}
	for name, command := range cases {
		t.Run(name, func(t *testing.T) {
			game, p0, _ := loadScenario(t, scenario{
				Terrain: []string{".."},
				Regions: []string{"ab"},
			})

			runTurn(game, command, "WAIT")

			assert.False(t, p0.IsDeactivated())
			assert.Equal(t, 0, game.Grid.Zones[0].Instability)
			assert.Equal(t, 1, game.Grid.Zones[1].Instability)
			assert.Equal(t, 0, p0.BlotPoints)
			assert.Empty(t, game.Summary)
		})
	}
}

func TestIllegalDisruptionsAreSkippedWithAnErrorAndDoNotDisqualify(t *testing.T) {
	cases := map[string]struct {
		setup   func(game *Game)
		command string
		summary string
	}{
		"off grid": {
			command: "DISRUPT 9 9",
			summary: "¤RED¤Player 0 Not part of grid: (9, 9)§RED§",
		},
		"unknown region": {
			command: "DISRUPT 7",
			summary: "¤RED¤Player 0 Invalid region id: 7§RED§",
		},
		"region holding a town": {
			command: "DISRUPT 0",
			summary: "¤RED¤Player 0 Cannot disrupt region0. It contains a town.§RED§",
		},
		"already inked region": {
			setup:   func(game *Game) { game.Grid.Zones[1].Inked = true },
			command: "DISRUPT 1",
			summary: "¤RED¤Player 0 Cannot disrupt region1. Already inked out.§RED§",
		},
		"out of disruption points": {
			command: "DISRUPT 1;DISRUPT 1",
			summary: "¤RED¤Player 0 Not enough disruption points.§RED§",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			game, p0, _ := loadScenario(t, scenario{
				Terrain: []string{".."},
				Regions: []string{"ab"},
				Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
			})
			if tc.setup != nil {
				tc.setup(game)
			}

			runTurn(game, tc.command, "WAIT")

			assert.False(t, p0.IsDeactivated())
			assert.NotEqual(t, -1, p0.GetScore())
			assert.Contains(t, game.Summary, tc.summary)
		})
	}
}

// Disruption points reset like paint does, so a region needs as many turns to
// tip over as the threshold demands however the blots are spread.
func TestInstabilityAccumulatesAcrossTurnsUntilTheThresholdInksTheRegion(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Regions: []string{"ab"},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	zone := game.Grid.Zones[1]

	for turn := 1; turn < INSTABILITY_THRESHOLD_BASE; turn++ {
		runTurn(game, "DISRUPT 1", "WAIT")
		require.Equalf(t, turn, zone.Instability, "after %d disruptions", turn)
		require.Falsef(t, zone.Inked, "inked early after %d disruptions", turn)
	}

	runTurn(game, "DISRUPT 1", "WAIT")

	assert.Equal(t, INSTABILITY_THRESHOLD_BASE, zone.Instability)
	assert.True(t, zone.Inked)
}

// Two players blotting the same region get there twice as fast.
func TestBothPlayersDisruptingOneRegionStackInstabilityInTheSameTurn(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Regions: []string{"ab"},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})

	runTurn(game, "DISRUPT 1", "DISRUPT 1")

	assert.Equal(t, 2, game.Grid.Zones[1].Instability)
}

func TestInkingARegionClearsEveryTrackInsideItForBothPlayers(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Regions: []string{"abbbb"},
		Tracks:  []string{".01n."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")

	require.True(t, game.Grid.Zones[1].Inked)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(2, 0).Track)
	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(3, 0).Track)
}

// The blot that tipped the region over is what credits the inking. Slot 2 of
// the tally is the contested track, which counts for neither side.
func TestInkAttributionCountsOwnAndEnemyTracksForTheDisruptingPlayer(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Regions: []string{"abbbb"},
		Tracks:  []string{".01n."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")

	assert.Equal(t, [2]int{1, 0}, game.ZonesInked)
	assert.Equal(t, [2]int{1, 0}, game.OwnTracksInkedOut)
	assert.Equal(t, [2]int{1, 0}, game.EnemyTracksInkedOut)
}

// A region holding a town cannot be disrupted, and the check skips it again
// even when its instability was raised some other way.
func TestARegionHoldingATownIsNeverInked(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Regions: []string{"ab"},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	game.Grid.Zones[0].Instability = INSTABILITY_THRESHOLD_BASE

	runTurn(game, "WAIT", "WAIT")

	assert.False(t, game.Grid.Zones[0].Inked)
}

// The rejection path from the placement rules becomes reachable through play:
// once a region inks, nothing can be built in it again.
func TestPlacementIntoARegionInkedByPlayIsSkipped(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Regions: []string{"abb"},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")
	require.True(t, game.Grid.Zones[1].Inked)

	runTurn(game, "PLACE_TRACK 1 0", "WAIT")

	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, PASSIVE_INCOME, p0.Dosh)
	assert.Contains(t, game.Summary, "¤RED¤Player 0 Cannot build in region 1§RED§")
}

// Disruptions run after placements, so a track laid this turn into a region
// that inks this turn is paid for and then immediately wiped.
func TestATrackPlacedIntoARegionThatInksThisTurnIsClearedImmediately(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Regions: []string{"abb"},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "PLACE_TRACK 1 0;DISRUPT 1", "WAIT")

	assert.Equal(t, TRACK_NONE, game.Grid.GetXY(1, 0).Track)
	assert.Equal(t, PASSIVE_INCOME-1, p0.Dosh)
}

// ——— game end ————————————————————————————————————————————————————————————

func TestGameEndsOnTurn100(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})

	for turn := 0; turn < MAX_TURNS-1; turn++ {
		runTurn(game, "WAIT", "WAIT")
		require.Falsef(t, game.Ended(), "ended early on turn %d", game.Turn)
	}
	runTurn(game, "WAIT", "WAIT")

	assert.True(t, game.Ended())
	assert.Equal(t, MAX_TURNS, game.Turn)
}

// The early end is checked over terrain, so it fires the moment inking walls
// off the last desired pair — long before either side could have built the
// route.
func TestGameEndsWhenInkingCutsTheLastRouteBetweenDesiredTowns(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Regions: []string{"aba"},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")

	assert.True(t, game.Grid.Zones[1].Inked)
	assert.True(t, game.Ended())
	assert.Equal(t, 1, game.Turn)
}

// A second desired pair that is still routable keeps the match alive, so the
// end condition is "no connection at all", not "any connection lost".
func TestGameContinuesWhileOneDesiredPairIsStillRoutable(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{
			"...",
			"...",
		},
		Regions: []string{
			"aba",
			"ccc",
		},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")

	assert.True(t, game.Grid.Zones[1].Inked)
	assert.False(t, game.Ended())
}

// A grid where nobody wants anything has no possible connection, so the
// referee ends it on the first turn.
func TestGameEndsOnTurnOneWhenNoTownWantsAConnection(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{".."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0}, {ID: 1, X: 1, Y: 0}},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.True(t, game.Ended())
	assert.Equal(t, 1, game.Turn)
}

// ——— leagues ——————————————————————————————————————————————————————————————

func TestResolveLeagueDefaultsToTheHighest(t *testing.T) {
	factory := NewFactory().(arena.LeagueResolver)

	assert.Equal(t, DEFAULT_LEAGUE, factory.ResolveLeague(nil))
	assert.Equal(t, DEFAULT_LEAGUE, factory.ResolveLeague(viper.New()))
}

func TestResolveLeagueHonoursTheLeagueOption(t *testing.T) {
	factory := NewFactory().(arena.LeagueResolver)
	options := viper.New()
	options.Set("league", "2")

	assert.Equal(t, 2, factory.ResolveLeague(options))
}

// League 1 asks player 0 for a single point. The match stops the turn it
// arrives, and the earned points are replaced by the verdict.
func TestTutorialLeagueOneEndsAsSoonAsPlayerZeroScores(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		League:  1,
		Terrain: []string{"..."},
		Tracks:  []string{".0."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})
	require.True(t, game.InTutorial)

	runTurn(game, "WAIT", "WAIT")
	require.True(t, game.Ended())
	require.Equal(t, 1, p0.GetScore())

	game.OnEnd()

	assert.Equal(t, 0, p0.GetScore())
	assert.Equal(t, -1, p1.GetScore())
}

// Failing the league 1 objective runs the full 100 turns and reverses the
// verdict.
func TestTutorialLeagueOneFailsWhenPlayerZeroNeverScores(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		League:  1,
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 2, Y: 0},
		},
	})

	for game.Turn < MAX_TURNS {
		runTurn(game, "WAIT", "WAIT")
	}
	require.True(t, game.Ended())

	game.OnEnd()

	assert.Equal(t, -1, p0.GetScore())
	assert.Equal(t, 0, p1.GetScore())
}

// League 2 asks player 0 to blot a region and take an opposing track down
// with it. Inking a region holding only its own tracks does not count.
func TestTutorialLeagueTwoEndsWhenPlayerZeroInksAnEnemyTrack(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		League:  2,
		Terrain: []string{"..."},
		Regions: []string{"aba"},
		Tracks:  []string{".1."},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")
	require.True(t, game.Ended())

	game.OnEnd()

	assert.Equal(t, 0, p0.GetScore())
	assert.Equal(t, -1, p1.GetScore())
}

func TestTutorialLeagueTwoIgnoresAnInkingThatSparesTheOpponent(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		League:  2,
		Terrain: []string{"..."},
		Regions: []string{"aba"},
		Tracks:  []string{".0."},
	})
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "DISRUPT 1", "WAIT")

	assert.True(t, game.Grid.Zones[1].Inked)
	assert.False(t, game.Ended())
}

// The tutorial replaces the end condition rather than adding to it: a grid
// with no routable connection would end the full game on turn 1, but a
// tutorial keeps running until its objective resolves.
func TestTutorialLeagueDoesNotUseTheNoRouteEndCondition(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		League:  1,
		Terrain: []string{".."},
	})

	runTurn(game, "WAIT", "WAIT")

	assert.False(t, game.Ended())
}

// Above the tutorials every league runs the same rules, so the objective
// machinery stays out of the way entirely.
func TestLeagueThreeAndAboveRunTheFullGame(t *testing.T) {
	for _, league := range []int{3, 4, 5} {
		game, _, _ := loadScenario(t, scenario{
			League:  league,
			Terrain: []string{".."},
		})

		assert.Falsef(t, game.InTutorial, "league %d", league)

		runTurn(game, "WAIT", "WAIT")
		assert.Truef(t, game.Ended(), "league %d", league)
	}
}

// ——— full match through the arena Referee contract ————————————————————————

// waitBots drives the referee exactly as arena.Runner does, answering "WAIT"
// on every turn for both sides.
func waitBots(t *testing.T) (arena.Referee, []arena.Player, int) {
	t.Helper()
	factory := NewFactory()
	referee, players := factory.NewGame(42, viper.New())
	referee.Init(players)

	for _, player := range players {
		for _, line := range referee.GlobalInfoFor(player) {
			player.SendInputLine(line)
		}
		player.ConsumeInputLines()
	}

	turn := 0
	for ; !referee.Ended() && turn < factory.MaxTurns(); turn++ {
		referee.ResetGameTurnData()
		for _, player := range players {
			if player.IsDeactivated() || referee.ShouldSkipPlayerTurn(player) {
				continue
			}
			for _, line := range referee.FrameInfoFor(player) {
				player.SendInputLine(line)
			}
			player.ConsumeInputLines()
			player.SetOutputs([]string{"WAIT"})
		}
		referee.ParsePlayerOutputs(players)
		referee.PerformGameUpdate(turn)
	}
	if !referee.Ended() {
		referee.EndGame()
	}
	referee.OnEnd()
	return referee, players, turn
}

func TestTwoWaitBotsPlayAFullMatchAndFinishScorelessDraw(t *testing.T) {
	_, players, turns := waitBots(t)

	assert.Equal(t, MAX_TURNS, turns)
	assert.False(t, players[0].IsDeactivated())
	assert.False(t, players[1].IsDeactivated())
	assert.Equal(t, 0, players[0].GetScore())
	assert.Equal(t, 0, players[1].GetScore())
}

func TestFrameInfoLineCountMatchesTheGridEveryTurn(t *testing.T) {
	factory := NewFactory()
	referee, players := factory.NewGame(1, viper.New())
	referee.Init(players)

	global := referee.GlobalInfoFor(players[0])
	width := atoiOrFail(t, global[1])
	height := atoiOrFail(t, global[2])

	assert.Len(t, global, 3+width*height+1+countTowns(t, global, width, height))
	assert.Len(t, referee.FrameInfoFor(players[0]), 2+width*height)
}

func countTowns(t *testing.T, global []string, width, height int) int {
	t.Helper()
	return atoiOrFail(t, global[3+width*height])
}

func atoiOrFail(t *testing.T, s string) int {
	t.Helper()
	v, err := strconv.Atoi(s)
	require.NoErrorf(t, err, "not an integer: %q", s)
	return v
}
