package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

// Acceptance tests exercise the full rules of Troll Farm on hand-crafted
// boards, bypassing random map generation. Each scenario asserts a single
// rule observable through the public Board / TaskManager API so any
// divergence from the Java reference is caught here without depending on
// seed parity.

// loadScenario builds a Board from an ASCII map. Recognised characters:
// '.' GRASS, '~' WATER, '#' ROCK, '+' IRON, '0' p0 shack, '1' p1 shack.
// Plants, units and inventories are NOT placed — callers wire those up via
// spawnUnit / plantAt / setInventory below.
func loadScenario(t *testing.T, league int, rows []string) (*Board, *Player, *Player) {
	t.Helper()
	width := len(rows[0])
	height := len(rows)
	board := newBoard(width, height, javarand.New(1))
	board.League = league

	p0 := NewPlayer(0)
	p1 := NewPlayer(1)
	board.Players = []*Player{p0, p1}

	for y, row := range rows {
		require.Equalf(t, width, len(row), "row %d length mismatch", y)
		for x, ch := range row {
			c := board.GetCell(x, y)
			switch ch {
			case '.':
				c.Type = CellGRASS
			case '~':
				c.Type = CellWATER
			case '#':
				c.Type = CellROCK
			case '+':
				c.Type = CellIRON
			case '0':
				c.Type = CellSHACK
				p0.Shack = c
			case '1':
				c.Type = CellSHACK
				p1.Shack = c
			default:
				t.Fatalf("unknown char %q at %d,%d", ch, x, y)
			}
		}
	}
	require.NotNil(t, p0.Shack, "p0 shack missing")
	require.NotNil(t, p1.Shack, "p1 shack missing")
	return board, p0, p1
}

// spawnUnit places a fresh troll for the given player on (x, y) with the
// supplied talents. Bypasses TrainTask cost accounting — tests pay no fruits.
func spawnUnit(board *Board, player *Player, talents [4]int, x, y int) *Unit {
	u := &Unit{
		ID:            board.AllocateUnitID(),
		Player:        player,
		Cell:          board.GetCell(x, y),
		MovementSpeed: talents[0],
		CarryCapacity: talents[1],
		HarvestPower:  talents[2],
		ChopPower:     talents[3],
		Inv:           NewInventory(),
	}
	player.AddUnit(u)
	board.AddUnit(u)
	return u
}

func TestUnitIDAllocatorIsPerBoard(t *testing.T) {
	boardA, p0A, _ := loadScenario(t, 4, []string{
		"0.",
		".1",
	})
	boardB, p0B, _ := loadScenario(t, 4, []string{
		"0.",
		".1",
	})

	a0 := spawnUnit(boardA, p0A, [4]int{1, 1, 1, 1}, 0, 0)
	b0 := spawnUnit(boardB, p0B, [4]int{1, 1, 1, 1}, 0, 0)
	a1 := spawnUnit(boardA, p0A, [4]int{1, 1, 1, 1}, 0, 1)

	assert.Equal(t, 0, a0.ID)
	assert.Equal(t, 0, b0.ID)
	assert.Equal(t, 1, a1.ID)
}

// plantAt drops a plant of the given kind on (x, y) with the requested
// size/resources/health/cooldown. Use 0 to fall back to the post-NewPlant
// default (size 0, full sapling health).
func plantAt(board *Board, kind Item, x, y, size, resources, health, cooldown int) *Plant {
	cell := board.GetCell(x, y)
	p := NewPlant(cell, kind)
	p.Size = size
	p.Resources = resources
	if health > 0 {
		p.Health = health
	}
	p.Cooldown = cooldown
	cell.SetPlant(p)
	board.AddPlant(p)
	return p
}

// runTurn parses both players' command lines, then ticks one game turn. The
// turn counter advances by one. Errors and summaries land on the players for
// the caller to inspect via PopErrors / PopSummaries.
func runTurn(board *Board, p0Cmd, p1Cmd string) {
	mgr := NewTaskManager()
	mgr.ParseTasks(board.Players[0], board, p0Cmd, board.League)
	mgr.ParseTasks(board.Players[1], board, p1Cmd, board.League)
	board.Turn++
	board.Tick(board.Turn, mgr)
}

// hasErrorCode returns true if the player accumulated an error with the given
// code without draining the queue.
func hasErrorCode(p *Player, code int) bool {
	for _, e := range p.errors {
		if e.ErrorCode == code {
			return true
		}
	}
	return false
}

// ——— input serialization ——————————————————————————————————————————————————

func TestGetInitialInputsRendersTerrainAndShackOwnerSwap(t *testing.T) {
	// Rules: initial input gives "width height" then height ASCII rows; '0'
	// is the recipient's shack, '1' is the opponent's. Both shacks swap when
	// addressed from the opponent's side.
	board, _, _ := loadScenario(t, 4, []string{
		"..0...",
		"......",
		"...1..",
	})

	inputs := board.GetInitialInputs(0)
	require.Equal(t, "6 3", inputs[0])
	assert.Equal(t, "..0...", inputs[1])
	assert.Equal(t, "...1..", inputs[3])

	flipped := board.GetInitialInputs(1)
	assert.Equal(t, "..1...", flipped[1], "from p1's view, p0's shack reads as '1'")
	assert.Equal(t, "...0..", flipped[3], "from p1's view, p1's shack reads as '0'")
}

func TestGetTurnInputsRendersOwnInventoryFirstAndPlayerSwapOnTrolls(t *testing.T) {
	// Rules: per-turn input is recipient inventory, opponent inventory, tree
	// list, troll list; troll lines report the recipient as player=0 and the
	// opponent as player=1 regardless of engine-side ownership.
	board, p0, p1 := loadScenario(t, 4, []string{
		"..0..",
		".....",
		"..1..",
	})
	p0.Inv.SetItem(ItemPLUM, 3)
	p1.Inv.SetItem(ItemPLUM, 7)
	spawnUnit(board, p0, [4]int{1, 1, 1, 1}, 1, 0)
	spawnUnit(board, p1, [4]int{1, 1, 1, 1}, 3, 2)

	out := board.GetTurnInputs(1) // p1's view
	assert.Equal(t, "7 0 0 0 0 0", out[0], "recipient inv first")
	assert.Equal(t, "3 0 0 0 0 0", out[1], "opponent inv second")
	assert.Equal(t, "0", out[2], "tree count")
	assert.Equal(t, "2", out[3], "troll count")
	// p0's troll (id 0) should appear as player=1 from p1's stdin.
	assert.True(t, strings.HasPrefix(out[4], "0 1 "), "p0 troll seen as player=1")
	assert.True(t, strings.HasPrefix(out[5], "1 0 "), "p1 troll seen as player=0")
}

// ——— MOVE —————————————————————————————————————————————————————————————————

func TestMoveAdvancesUpToMovementSpeed(t *testing.T) {
	// Rules: MOVE id x y advances the troll up to movementSpeed cells along
	// a shortest walkable path.
	board, p0, _ := loadScenario(t, 4, []string{
		"0.......",
		"........",
		".......1",
	})
	u := spawnUnit(board, p0, [4]int{3, 1, 1, 0}, 0, 0)

	runTurn(board, "MOVE 0 7 0", "WAIT")
	assert.Equal(t, board.GetCell(3, 0), u.Cell, "moved 3 steps east")
}

func TestMoveStopsAtUnreachableTargetNearestApproach(t *testing.T) {
	// Rules: if the target is unreachable, the troll moves to the nearest
	// reachable cell to it.
	board, p0, _ := loadScenario(t, 4, []string{
		"0..#....",
		"...#....",
		"...#...1",
	})
	u := spawnUnit(board, p0, [4]int{5, 1, 1, 0}, 0, 0)

	runTurn(board, "MOVE 0 7 0", "WAIT")
	assert.Equal(t, board.GetCell(2, 0), u.Cell,
		"stops at column 2, hugging the wall on the reachable side")
}

func TestMoveBlockedByOwnTroll(t *testing.T) {
	// Rules: each cell holds at most one troll per team. A troll can't move
	// onto a cell another own-team troll is staying put on.
	board, p0, _ := loadScenario(t, 4, []string{
		"0.......",
		"........",
		".......1",
	})
	a := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 0, 0)
	b := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)

	// Only `a` moves; `b` stands its ground (no MoveTask for unit 1).
	runTurn(board, "MOVE 0 1 0", "WAIT")
	assert.Equal(t, board.GetCell(0, 0), a.Cell, "a couldn't enter b's cell")
	assert.Equal(t, board.GetCell(1, 0), b.Cell, "b stayed put")
	assert.True(t, hasErrorCode(p0, ErrMoveBlocked), "blocked move surfaces MOVE_BLOCKED")
}

func TestMoveResolvesCircularSwap(t *testing.T) {
	// Rules / engine: two trolls whose targets are each other's cells must
	// be resolved as a simultaneous swap (the circular-cycle path in
	// MoveTask.apply).
	board, p0, _ := loadScenario(t, 4, []string{
		"0.....",
		"......",
		".....1",
	})
	a := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	b := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 2, 0)

	runTurn(board, "MOVE 0 2 0;MOVE 1 1 0", "WAIT")
	assert.Equal(t, board.GetCell(2, 0), a.Cell, "a took b's old cell")
	assert.Equal(t, board.GetCell(1, 0), b.Cell, "b took a's old cell")
}

// ——— HARVEST ——————————————————————————————————————————————————————————————

func TestHarvestPullsFruitsUpToCapacityAndPower(t *testing.T) {
	// Rules: HARVEST takes fruits limited by harvestPower and free capacity.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 10, 2, 0}, 1, 1)
	plantAt(board, ItemAPPLE, 1, 1, 4, 3, 20, 9)

	runTurn(board, "HARVEST 0", "WAIT")
	assert.Equal(t, 2, u.Inv.GetItemCount(ItemAPPLE),
		"harvestPower=2 caps the take at 2 fruits")
}

func TestHarvestSplitsAlternatingWithLastFruitDuplicated(t *testing.T) {
	// Rules: two trolls (one per team) on the same plant take fruits one at
	// a time, alternating. The last fruit may be duplicated so both
	// receive it.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	a := spawnUnit(board, p0, [4]int{1, 10, 5, 0}, 1, 1)
	b := spawnUnit(board, p1, [4]int{1, 10, 5, 0}, 1, 1)
	plant := plantAt(board, ItemAPPLE, 1, 1, 4, 1, 20, 9)

	runTurn(board, "HARVEST 0", "HARVEST 1")
	assert.Equal(t, 0, plant.Resources, "tree drained")
	assert.Equal(t, 1, a.Inv.GetItemCount(ItemAPPLE), "p0 troll took the fruit")
	assert.Equal(t, 1, b.Inv.GetItemCount(ItemAPPLE),
		"p1 troll received the duplicated last fruit")
}

func TestHarvestErrorsOutOfCapacity(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 1, 3, 0}, 1, 1)
	u.Inv.SetItem(ItemBANANA, 1) // already at carryCapacity=1
	plantAt(board, ItemAPPLE, 1, 1, 4, 3, 20, 9)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "HARVEST 0", board.League)
	assert.True(t, hasErrorCode(p0, ErrNoCapacity), "raises NO_CAPACITY at parse time")
}

// ——— PLANT ————————————————————————————————————————————————————————————————

func TestPlantConsumesSeedAndCreatesTree(t *testing.T) {
	// Rules: PLANT id type consumes one seed of `type` and creates a tree.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 1)
	u.Inv.SetItem(ItemAPPLE, 1)

	runTurn(board, "PLANT 0 APPLE", "WAIT")
	assert.Equal(t, 0, u.Inv.GetItemCount(ItemAPPLE), "seed consumed")
	require.NotNil(t, board.GetCell(1, 1).Plant, "tree planted on cell")
	assert.Equal(t, ItemAPPLE, board.GetCell(1, 1).Plant.Type)
}

func TestPlantSameTypeConcurrentSpawnsOneTreeBothPay(t *testing.T) {
	// Rules: when two trolls plant the same type on the same cell, both
	// lose a seed and the tree is planted (once).
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	a := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 1)
	b := spawnUnit(board, p1, [4]int{1, 5, 1, 0}, 1, 1)
	a.Inv.SetItem(ItemAPPLE, 1)
	b.Inv.SetItem(ItemAPPLE, 1)

	runTurn(board, "PLANT 0 APPLE", "PLANT 1 APPLE")
	assert.Equal(t, 0, a.Inv.GetItemCount(ItemAPPLE))
	assert.Equal(t, 0, b.Inv.GetItemCount(ItemAPPLE))
	plant := board.GetCell(1, 1).Plant
	require.NotNil(t, plant)
	assert.Equal(t, 1, len(board.Plants), "only one tree exists")
	assert.Equal(t, ItemAPPLE, plant.Type)
}

func TestPlantDifferentTypesContradictNoneAccepted(t *testing.T) {
	// Rules: when the two trolls plant different types on the same cell,
	// nothing happens. Both keep their seeds.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	a := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 1)
	b := spawnUnit(board, p1, [4]int{1, 5, 1, 0}, 1, 1)
	a.Inv.SetItem(ItemAPPLE, 1)
	b.Inv.SetItem(ItemPLUM, 1)

	runTurn(board, "PLANT 0 APPLE", "PLANT 1 PLUM")
	assert.Nil(t, board.GetCell(1, 1).Plant, "contradicting plant — nothing planted")
	assert.Equal(t, 1, a.Inv.GetItemCount(ItemAPPLE), "p0 retains seed")
	assert.Equal(t, 1, b.Inv.GetItemCount(ItemPLUM), "p1 retains seed")
}

// ——— CHOP —————————————————————————————————————————————————————————————————

func TestChopReducesHealthAndDropsWoodOnKill(t *testing.T) {
	// Rules: CHOP cuts the tree's health by chopPower. When health <= 0 the
	// troll collects wood = tree.size (capped by free capacity).
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 10, 1, 5}, 1, 1)
	plantAt(board, ItemBANANA, 1, 1, 3, 0, 5, 0) // size 3, health 5

	runTurn(board, "CHOP 0", "WAIT")
	assert.Equal(t, 3, u.Inv.GetItemCount(ItemWOOD), "wood == tree.size on kill")
	assert.Equal(t, 0, len(board.Plants), "dead plant removed at end of tick")
}

func TestChopDoesNotKillWhenHealthSurvives(t *testing.T) {
	// Rules: when health remains > 0 after chop, no wood is collected.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 10, 1, 2}, 1, 1)
	plantAt(board, ItemAPPLE, 1, 1, 4, 0, 20, 0)

	runTurn(board, "CHOP 0", "WAIT")
	assert.Equal(t, 0, u.Inv.GetItemCount(ItemWOOD), "tree survived, no wood")
	assert.Equal(t, 18, board.GetCell(1, 1).Plant.Health, "chopPower=2 → 20-2=18")
}

// ——— PICK / DROP —————————————————————————————————————————————————————————

func TestPickTransfersOneItemFromShackToTroll(t *testing.T) {
	// Rules: PICK id type takes one item of `type` from the shack inventory.
	// Must be adjacent to own shack.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 0)
	p0.Inv.SetItem(ItemAPPLE, 3)

	runTurn(board, "PICK 0 APPLE", "WAIT")
	assert.Equal(t, 2, p0.Inv.GetItemCount(ItemAPPLE), "one removed from shack")
	assert.Equal(t, 1, u.Inv.GetItemCount(ItemAPPLE), "one added to troll")
}

func TestDropMovesAllCarriedItemsToShack(t *testing.T) {
	// Rules: DROP id transfers all carried items to the shack.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 10, 1, 0}, 1, 0)
	u.Inv.SetItem(ItemAPPLE, 2)
	u.Inv.SetItem(ItemWOOD, 1)

	runTurn(board, "DROP 0", "WAIT")
	assert.Equal(t, 0, u.Inv.GetTotal(), "troll cleared")
	assert.Equal(t, 2, p0.Inv.GetItemCount(ItemAPPLE), "shack received apples")
	assert.Equal(t, 1, p0.Inv.GetItemCount(ItemWOOD), "shack received wood")
}

// ——— MINE —————————————————————————————————————————————————————————————————

func TestMineCollectsUpToChopPowerFromAdjacentIron(t *testing.T) {
	// Rules: MINE requires adjacency to an IRON cell, takes up to chopPower
	// iron limited by free capacity.
	board, p0, _ := loadScenario(t, 4, []string{
		"0..+...",
		".......",
		"......1",
	})
	u := spawnUnit(board, p0, [4]int{1, 10, 1, 4}, 2, 0)

	runTurn(board, "MINE 0", "WAIT")
	assert.Equal(t, 4, u.Inv.GetItemCount(ItemIRON), "chopPower=4 → 4 iron")
	assert.Equal(t, CellIRON, board.GetCell(3, 0).Type, "iron cell never depletes")
}

func TestMineCappedByFreeCapacity(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0..+",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 2, 1, 5}, 2, 0)

	runTurn(board, "MINE 0", "WAIT")
	assert.Equal(t, 2, u.Inv.GetItemCount(ItemIRON),
		"carryCapacity=2 caps the take regardless of chopPower")
}

// ——— TRAIN ————————————————————————————————————————————————————————————————

func TestTrainSpawnsAtShackAndChargesQuadraticCost(t *testing.T) {
	// Rules: TRAIN cost per attribute = existingTrolls + attribute².
	// With 1 existing troll and TRAIN 2 1 1 1 in league 3: cost vector is
	// [1+4, 1+1, 1+1, 1+1] = [5,2,2,2] PLUM/LEMON/APPLE/IRON.
	board, p0, _ := loadScenario(t, 3, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 1}, 0, 1)
	p0.Inv.SetItem(ItemPLUM, 5)
	p0.Inv.SetItem(ItemLEMON, 2)
	p0.Inv.SetItem(ItemAPPLE, 2)
	p0.Inv.SetItem(ItemIRON, 2)

	runTurn(board, "TRAIN 2 1 1 1", "WAIT")
	require.Equal(t, 2, len(p0.Units), "new troll spawned")
	assert.Equal(t, 0, p0.Inv.GetItemCount(ItemPLUM), "PLUM drained")
	assert.Equal(t, 0, p0.Inv.GetItemCount(ItemLEMON), "LEMON drained")
	assert.Equal(t, 0, p0.Inv.GetItemCount(ItemAPPLE), "APPLE drained")
	assert.Equal(t, 0, p0.Inv.GetItemCount(ItemIRON), "IRON drained")
	new := p0.Units[1]
	assert.Equal(t, p0.Shack, new.Cell, "new troll on shack cell")
	assert.Equal(t, 2, new.MovementSpeed)
}

func TestTrainFailsWhenCantAfford(t *testing.T) {
	board, p0, _ := loadScenario(t, 3, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 1}, 0, 1)
	p0.Inv.SetItem(ItemPLUM, 1) // not enough for any TRAIN

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "TRAIN 1 1 1 1", board.League)
	assert.True(t, hasErrorCode(p0, ErrCantAfford), "raises CANT_AFFORD")
	assert.Equal(t, 1, len(p0.Units), "no new troll")
}

func TestTrainRejectedInLeagueOne(t *testing.T) {
	// Rules: TRAIN is unavailable in league 1. The check fires in parseTask
	// *after* the constructor, so the player must be able to afford the
	// command — otherwise CANT_AFFORD in the constructor sets failedParsing
	// and the NOT_AVAILABLE error is silently dropped (matches Java).
	board, p0, _ := loadScenario(t, 1, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 0, 1)
	// 1 existing troll, talents [1,1,1,0] → costs [2,2,2,0]. Fund it.
	p0.Inv.SetItem(ItemPLUM, 2)
	p0.Inv.SetItem(ItemLEMON, 2)
	p0.Inv.SetItem(ItemAPPLE, 2)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "TRAIN 1 1 1 0", board.League)
	assert.True(t, hasErrorCode(p0, ErrNotAvailable), "league 1 rejects TRAIN")
}

// ——— priority / action ordering ——————————————————————————————————————————

func TestPriorityOrderMoveBeforeDrop(t *testing.T) {
	// Engine rule (TaskManager.peekTasks): tasks are applied lowest-priority
	// first. MOVE=1 fires before DROP=7, so a MOVE-then-DROP on different
	// trolls leaves the carrier next to the shack and the second action
	// transfers items the same turn.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	moveU := spawnUnit(board, p0, [4]int{2, 5, 1, 0}, 2, 0)
	moveU.Inv.SetItem(ItemAPPLE, 2)
	carrier := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 0)
	carrier.Inv.SetItem(ItemPLUM, 1)

	runTurn(board, "MOVE 0 1 0;DROP 1", "WAIT")
	// MOVE applies before DROP would have applied. Carrier never moved,
	// it just dropped. moveU moves freely toward shack.
	assert.Equal(t, 1, p0.Inv.GetItemCount(ItemPLUM), "DROP delivered carrier's items")
	assert.Equal(t, 0, carrier.Inv.GetTotal(), "carrier emptied")
}

// ——— plant growth ———————————————————————————————————————————————————————

func TestPlantTickGrowsAndProducesFruit(t *testing.T) {
	// Rules: every turn cooldown drops; on 0, the plant grows (size+=1) until
	// PLANT_MAX_SIZE then produces a fruit (resources+=1) until
	// PLANT_MAX_RESOURCES.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 0, 1)
	// Tree at max size, full health, ready to fruit next tick.
	plantAt(board, ItemBANANA, 1, 1, 4, 0, 6, 1)

	runTurn(board, "WAIT", "WAIT")
	assert.Equal(t, 1, board.GetCell(1, 1).Plant.Resources, "first fruit produced")
}

func TestPlantCooldownReducedNearWater(t *testing.T) {
	// Rules: a tree is "near water" if any 4-neighbour is WATER. Cooldown
	// drops by PLANT_WATER_COOLDOWN_BOOST[type].
	board, p0, _ := loadScenario(t, 4, []string{
		"0.~.",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 0, 1)
	plant := plantAt(board, ItemAPPLE, 2, 1, 0, 0, 8, 0)

	// Trigger growth (cooldown==0 → grow), then check next cooldown.
	plant.Tick(true)
	assert.Equal(t, PLANT_COOLDOWN[ItemAPPLE]-PLANT_WATER_COOLDOWN_BOOST[ItemAPPLE],
		plant.Cooldown, "cooldown reduced by water boost")
}

// ——— league differences —————————————————————————————————————————————————

func TestLeagueOneRejectsChop(t *testing.T) {
	board, p0, _ := loadScenario(t, 1, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 1}, 1, 1)
	plantAt(board, ItemBANANA, 1, 1, 1, 0, 3, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "CHOP 0", board.League)
	assert.True(t, hasErrorCode(p0, ErrNotAvailable), "league 1 rejects CHOP")
}

func TestLeagueTwoRejectsMine(t *testing.T) {
	// Rules: MINE is league 3+ only.
	board, p0, _ := loadScenario(t, 2, []string{
		"0..+",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 5, 1, 1}, 2, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "MINE 0", board.League)
	assert.True(t, hasErrorCode(p0, ErrNotAvailable), "league 2 rejects MINE")
}

// ——— scoring / game end —————————————————————————————————————————————————

func TestRecomputeScoreCountsFruitsAndWood(t *testing.T) {
	// Rules: each fruit in shack = 1 pt, wood = 4 pts, iron = 0 pts.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	p0.Inv.SetItem(ItemPLUM, 3)
	p0.Inv.SetItem(ItemAPPLE, 2)
	p0.Inv.SetItem(ItemWOOD, 4)
	p0.Inv.SetItem(ItemIRON, 100)

	board.Players = []*Player{p0}
	p0.RecomputeScore()
	assert.Equal(t, 3+2+4*4, p0.GetScore(), "5 fruits + 16 wood = 21")
}

func TestHasStalledFiresAfterTenTreelessTurns(t *testing.T) {
	// Rules: when no trees remain on the map for STALL_LIMIT (10)
	// consecutive turns, the game ends — even if neither player is "stuck".
	// Both sides need at least one productive item to keep the
	// stuck-leader short-circuit from firing first.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	p0.Inv.SetItem(ItemAPPLE, 1)
	p1.Inv.SetItem(ItemAPPLE, 1)

	for i := 0; i < STALL_LIMIT-1; i++ {
		assert.Falsef(t, board.HasStalled(), "turn %d still pending", i)
	}
	assert.True(t, board.HasStalled(), "10th treeless turn triggers stall")
}

func TestHasStalledStuckLeaderWins(t *testing.T) {
	// Rules: a player can no longer make progress (no productive items) AND
	// their score <= opponent's → game ends so the leader can force a win.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	// p0 is "stuck" (no fruits, only iron carried), p1 leads on score.
	stuck := spawnUnit(board, p0, [4]int{1, 5, 1, 1}, 1, 1)
	stuck.Inv.SetItem(ItemIRON, 3)
	p1.Inv.SetItem(ItemAPPLE, 5)
	p1.RecomputeScore()
	p0.RecomputeScore()

	assert.True(t, board.HasStalled(), "stuck p0 + leading p1 ends the game immediately")
}
