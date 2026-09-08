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
// in region 0.
type scenario struct {
	Terrain []string
	Regions []string
	Towns   []townSpec
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

	p0, p1 := NewPlayer(0), NewPlayer(1)
	game := NewGame(nil, DEFAULT_LEAGUE)
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

func TestIntentsAreClearedBetweenTurns(t *testing.T) {
	game, p0, _ := loadScenario(t, scenario{Terrain: []string{"."}})

	runTurn(game, "WAIT", "WAIT")
	runTurn(game, "WAIT", "WAIT")

	assert.Len(t, p0.Intents, 1)
}

// ——— game end ————————————————————————————————————————————————————————————

func TestGameEndsOnTurn100(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{Terrain: []string{"."}})

	for turn := 0; turn < MAX_TURNS-1; turn++ {
		runTurn(game, "WAIT", "WAIT")
		require.Falsef(t, game.Ended(), "ended early on turn %d", game.Turn)
	}
	runTurn(game, "WAIT", "WAIT")

	assert.True(t, game.Ended())
	assert.Equal(t, MAX_TURNS, game.Turn)
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
