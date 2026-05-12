package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// traceTypes returns the event type tags for one player's bucket, preserving
// emission order.
func traceTypes(traces []arena.TurnTrace) []string {
	out := make([]string, len(traces))
	for i, t := range traces {
		out[i] = t.Type
	}
	return out
}

// decodeTraceState calls Board.DecorateTraceTurn and decodes the payload
// into TraceTurnState. Tests use this to assert on typed fields without
// re-stringifying the JSON.
func decodeTraceState(t *testing.T, b *Board, turn int) TraceTurnState {
	t.Helper()
	raw := b.DecorateTraceTurn(turn)
	require.NotEmpty(t, raw, "DecorateTraceTurn returned empty payload")
	var state TraceTurnState
	require.NoError(t, json.Unmarshal(raw, &state))
	return state
}

func TestTraceMoveEmitsOnSuccessfulStep(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)

	runTurn(board, "MOVE 0 3 0", "")

	require.Equal(t, []string{TraceMove}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[MoveData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, 0, data.Unit)
	assert.Equal(t, [2]int{2, 0}, data.To, "single-step move toward (3,0) lands on (2,0)")
}

func TestTraceMoveSuppressedWhenBlocked(t *testing.T) {
	// Two trolls on adjacent cells, each requesting the other's cell with
	// move speed 1: the engine resolves by swap, both succeed. We assert the
	// non-blocked baseline first to keep this test honest.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	spawnUnit(board, p1, [4]int{1, 1, 1, 0}, 1, 1) // adjacent below

	runTurn(board, "MOVE 0 1 1", "MOVE 1 1 0")

	// Both swap, so both emit MOVE.
	assert.Equal(t, []string{TraceMove}, traceTypes(board.Traces[0]))
	assert.Equal(t, []string{TraceMove}, traceTypes(board.Traces[1]))
}

func TestTraceHarvestEmitsWithAmountAndType(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 3, 1, 0}, 1, 0)
	plantAt(board, ItemPLUM, 1, 0, PLANT_MAX_SIZE, 2, 12, 5)

	runTurn(board, "HARVEST 0", "")

	require.Equal(t, []string{TraceHarvest}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[HarvestData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, u.ID, data.Unit)
	assert.Equal(t, [2]int{1, 0}, data.Cell)
	assert.Equal(t, "PLUM", data.Type)
	assert.Equal(t, 1, data.Amount, "harvestPower 1 takes one fruit per turn")
}

func TestTracePlantEmitsOnSuccessfulSeed(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 2, 0)
	u.Inv.SetItem(ItemAPPLE, 1)

	runTurn(board, "PLANT 0 APPLE", "")

	require.Equal(t, []string{TracePlant}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[PlantData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, u.ID, data.Unit)
	assert.Equal(t, [2]int{2, 0}, data.Cell)
	assert.Equal(t, "APPLE", data.Type)
}

func TestTracePickEmitsOnePerToken(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 0) // adjacent to shack at (0,0)
	p0.Inv.SetItem(ItemLEMON, 3)

	runTurn(board, "PICK 0 LEMON", "")

	require.Equal(t, []string{TracePick}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[PickData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, u.ID, data.Unit)
	assert.Equal(t, "LEMON", data.Type)
}

func TestTraceDropEmitsItemsBreakdown(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 0)
	u.Inv.SetItem(ItemPLUM, 2)
	u.Inv.SetItem(ItemAPPLE, 1)

	runTurn(board, "DROP 0", "")

	require.Equal(t, []string{TraceDrop}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[DropData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, u.ID, data.Unit)
	assert.Equal(t, [ItemsCount]int{2, 0, 1, 0, 0, 0}, data.Items)
}

func TestTraceTrainEmitsTalentsAndUnitID(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	// Initial unit already on shack (created during InitForGame would, but
	// loadScenario doesn't call it). Move it out so the TRAIN spawn cell is free.
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	for i := 0; i <= int(ItemAPPLE); i++ {
		p0.Inv.SetItem(Item(i), 5) // base cost = 1 (existing units) + talent^2; 5 covers (1+1)
	}
	// At league 2, chopPower talent must be 0.
	board.League = 2

	runTurn(board, "TRAIN 1 1 1 0", "")

	require.Equal(t, []string{TraceTrain}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[TrainData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, [4]int{1, 1, 1, 0}, data.Talents)
	// Spawned troll is the second unit registered after our spawnUnit (id 0),
	// so its id is 1.
	assert.Equal(t, 1, data.Unit)
}

func TestTraceChopReportsDamageAndKill(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 4}, 1, 0)
	// Small banana plant: health 4 - 1*4 = 4 at size 4. chopPower 4 kills it.
	p := plantAt(board, ItemBANANA, 1, 0, 0, 0, 4, 5)

	runTurn(board, "CHOP 0", "")

	require.Equal(t, []string{TraceChop}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[ChopData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, u.ID, data.Unit)
	assert.Equal(t, [2]int{1, 0}, data.Cell)
	assert.Equal(t, 4, data.Damage)
	assert.True(t, data.Killed, "plant with health=4 should die from chopPower=4")
	assert.True(t, p.IsDead())
}

func TestTraceMineEmitsOnSuccess(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0+..",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 5, 1, 2}, 2, 0) // (2,0) is grass adjacent to iron at (1,0)
	_ = u
	// (2,0) must be adjacent to iron for MINE; iron is at (1,0).
	runTurn(board, "MINE 0", "")

	require.Equal(t, []string{TraceMine}, traceTypes(board.Traces[0]))
	data, err := arena.DecodeData[MineData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, [2]int{2, 0}, data.Cell)
	assert.Equal(t, 2, data.Iron, "chopPower 2 yields 2 iron when carry capacity allows")
}

func TestTraceMessageEmittedAtParseTime(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)

	runTurn(board, "MSG hello there;MOVE 0 2 0", "")

	// MSG emits during parse, MOVE during Tick — MSG appears first.
	assert.Equal(t, []string{TraceMessage, TraceMove}, traceTypes(board.Traces[0]))
	msg, err := arena.DecodeData[MessageData](board.Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, "hello there", msg.Text)
}

func TestTraceWaitEmittedOnBareToken(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)

	runTurn(board, "WAIT", "")

	require.Equal(t, []string{TraceWait}, traceTypes(board.Traces[0]))
	assert.Empty(t, board.Traces[0][0].Data, "WAIT carries no data payload")
}

func TestRefereeTurnTracesReturnsCopyAndClearsOnReset(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	r := NewReferee(board)

	board.Traces[0] = append(board.Traces[0], arena.TurnTrace{Type: TraceWait})
	out := r.TurnTraces(0, []arena.Player{board.Players[0], board.Players[1]})
	require.Len(t, out[0], 1)
	assert.Equal(t, TraceWait, out[0][0].Type)

	r.ResetGameTurnData()
	assert.Empty(t, board.Traces[0], "ResetGameTurnData clears the buffer")
	// The copy returned earlier must not be mutated by reset.
	require.Len(t, out[0], 1)
}

func TestDecorateTraceTurnCapturesInventoriesUnitsAndPlants(t *testing.T) {
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u0 := spawnUnit(board, p0, [4]int{1, 3, 1, 0}, 1, 0)
	u0.Inv.SetItem(ItemPLUM, 1)
	u1 := spawnUnit(board, p1, [4]int{2, 4, 1, 1}, 3, 2)
	u1.Inv.SetItem(ItemWOOD, 2)
	p0.Inv.SetItem(ItemAPPLE, 5)
	p1.Inv.SetItem(ItemIRON, 3)
	plantAt(board, ItemLEMON, 2, 1, 2, 1, 8, 3)

	state := decodeTraceState(t, board, 7)

	assert.Equal(t, 7, state.Turn)
	assert.Equal(t, [2][ItemsCount]int{
		{0, 0, 5, 0, 0, 0},
		{0, 0, 0, 0, 3, 0},
	}, state.Inventories)

	require.Len(t, state.Units[0], 1)
	assert.Equal(t, u0.ID, state.Units[0][0].ID)
	assert.Equal(t, [2]int{1, 0}, state.Units[0][0].Pos)
	assert.Equal(t, [ItemsCount]int{1, 0, 0, 0, 0, 0}, state.Units[0][0].Carry)

	require.Len(t, state.Units[1], 1)
	assert.Equal(t, [ItemsCount]int{0, 0, 0, 0, 0, 2}, state.Units[1][0].Carry)
	assert.Equal(t, 2, state.Units[1][0].MoveSpeed)
	assert.Equal(t, 1, state.Units[1][0].ChopPower)

	require.Len(t, state.Plants, 1)
	assert.Equal(t, "LEMON", state.Plants[0].Type)
	assert.Equal(t, [2]int{2, 1}, state.Plants[0].Pos)
	assert.Equal(t, 2, state.Plants[0].Size)
	assert.Equal(t, 1, state.Plants[0].Resources)
}

// On-disk JSON keys for the per-turn state payload must be camelCase to match
// the cross-game trace envelope.
func TestDecorateTraceTurnUsesCamelCaseJSONKeys(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	_ = u

	raw := board.DecorateTraceTurn(0)
	var top map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &top))
	for _, k := range []string{"turn", "inventories", "units"} {
		_, ok := top[k]
		assert.Truef(t, ok, "expected key %q in state payload", k)
	}

	var units struct {
		Units [2][]map[string]json.RawMessage `json:"units"`
	}
	require.NoError(t, json.Unmarshal(raw, &units))
	require.NotEmpty(t, units.Units[0])
	for _, k := range []string{"moveSpeed", "carryCapacity", "harvestPower", "chopPower"} {
		_, ok := units.Units[0][0][k]
		assert.Truef(t, ok, "expected camelCase key %q in unit entry", k)
	}
	for _, k := range []string{"move_speed", "carry_capacity"} {
		_, ok := units.Units[0][0][k]
		assert.Falsef(t, ok, "snake_case key %q must not appear", k)
	}
}
