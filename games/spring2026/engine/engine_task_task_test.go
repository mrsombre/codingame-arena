package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Per-file tests for the abstract Task layer: regex cascade, usedUnits
// deduplication, league rejection, and the deterministic group-by helpers.

func TestParseTaskRecognisesEachCommandShape(t *testing.T) {
	// Every priority bucket has a regex; ParseTask should route the right
	// command to the right Task type.
	board, p0, _ := loadScenario(t, 4, []string{
		"0..+",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 5, 1, 1}, 1, 0)

	cases := map[string]string{
		"MOVE 0 2 0":    "*engine.MoveTask",
		"HARVEST 0":     "*engine.HarvestTask",
		"PLANT 0 PLUM":  "*engine.PlantTask",
		"CHOP 0":        "*engine.ChopTask",
		"PICK 0 APPLE":  "*engine.PickTask",
		"TRAIN 1 1 1 1": "*engine.TrainTask",
		"DROP 0":        "*engine.DropTask",
		"MINE 0":        "*engine.MineTask",
	}
	for cmd, want := range cases {
		used := make(map[*Unit]struct{})
		got := ParseTask(p0, board, cmd, board.League, used)
		require.NotNilf(t, got, "ParseTask returned nil for %q", cmd)
		// reflect-free type discrimination via type assertion
		switch got.(type) {
		case *MoveTask:
			assert.Equalf(t, "*engine.MoveTask", want, cmd)
		case *HarvestTask:
			assert.Equalf(t, "*engine.HarvestTask", want, cmd)
		case *PlantTask:
			assert.Equalf(t, "*engine.PlantTask", want, cmd)
		case *ChopTask:
			assert.Equalf(t, "*engine.ChopTask", want, cmd)
		case *PickTask:
			assert.Equalf(t, "*engine.PickTask", want, cmd)
		case *TrainTask:
			assert.Equalf(t, "*engine.TrainTask", want, cmd)
		case *DropTask:
			assert.Equalf(t, "*engine.DropTask", want, cmd)
		case *MineTask:
			assert.Equalf(t, "*engine.MineTask", want, cmd)
		default:
			t.Fatalf("unexpected task type for %q", cmd)
		}
	}
}

func TestParseTaskIsCaseInsensitive(t *testing.T) {
	// Pattern regexes carry the (?i) flag — mixed case should still parse.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	used := make(map[*Unit]struct{})
	got := ParseTask(p0, board, "move 0 2 0", board.League, used)
	_, ok := got.(*MoveTask)
	assert.True(t, ok, "lowercase MOVE recognised")
}

func TestParseTaskUnknownEmitsCriticalError(t *testing.T) {
	// Java rule: any command that doesn't match any pattern adds an
	// UNKNOWN_COMMAND error with critical=true → triggers deactivation.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	used := make(map[*Unit]struct{})
	got := ParseTask(p0, board, "DANCE 0", board.League, used)
	assert.Nil(t, got)
	require.NotEmpty(t, p0.errors)
	assert.Equal(t, ErrUnknownCommand, p0.errors[0].ErrorCode)
	assert.True(t, p0.errors[0].Critical, "unknown commands are critical")
}

func TestParseTaskWaitReturnsNilWithoutError(t *testing.T) {
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	used := make(map[*Unit]struct{})
	got := ParseTask(p0, board, "WAIT", board.League, used)
	assert.Nil(t, got, "WAIT short-circuits before regex matching")
	assert.Empty(t, p0.errors)
}

func TestParseTaskUnitReuseRaisesAlreadyUsed(t *testing.T) {
	// Java rule: usedUnits dedups across the same player's commands; the
	// second appearance of a unit id raises ALREADY_USED.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	used := make(map[*Unit]struct{})

	first := ParseTask(p0, board, "MOVE 0 2 0", board.League, used)
	require.NotNil(t, first)
	require.Contains(t, used, u)

	second := ParseTask(p0, board, "MOVE 0 0 0", board.League, used)
	assert.Nil(t, second)
	assert.True(t, hasErrorCode(p0, ErrAlreadyUsed))
}

func TestGroupByCellSortsByCellID(t *testing.T) {
	// groupByCell orders buckets by Cell.GetID() = x + (y<<16). Two cells on
	// the same row should sort left-to-right.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u0 := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 2, 0)
	u1 := spawnUnit(board, p0, [4]int{1, 5, 1, 0}, 1, 0)

	// HARVEST has a getCell; build two tasks with mismatched ids to force a
	// re-sort. HARVEST requires a plant; spawn lightweight ones inline.
	plantAt(board, ItemAPPLE, 1, 0, 4, 1, 20, 0)
	plantAt(board, ItemAPPLE, 2, 0, 4, 1, 20, 0)
	used := make(map[*Unit]struct{})
	t1 := ParseTask(p0, board, "HARVEST 0", board.League, used).(*HarvestTask)
	t2 := ParseTask(p0, board, "HARVEST 1", board.League, used).(*HarvestTask)
	// In input order, t1 is at x=2 (u0) and t2 is at x=1 (u1). After
	// groupByCell they swap into ascending GetID order.
	_ = u0
	_ = u1

	groups := groupByCell([]Task{t1, t2})
	require.Equal(t, 2, len(groups))
	assert.Equal(t, 1, groups[0][0].GetCell().X, "leftmost cell first")
	assert.Equal(t, 2, groups[1][0].GetCell().X, "rightmost cell second")
}

func TestGroupByPlayerSortsByPlayerIndex(t *testing.T) {
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	u0 := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	u1 := spawnUnit(board, p1, [4]int{1, 1, 1, 0}, 2, 0)

	// Two MoveTasks, build them by hand so they share a single slice.
	used0 := make(map[*Unit]struct{})
	used1 := make(map[*Unit]struct{})
	tp1 := ParseTask(p1, board, "MOVE 1 1 0", board.League, used1).(*MoveTask)
	tp0 := ParseTask(p0, board, "MOVE 0 2 0", board.League, used0).(*MoveTask)
	_ = u0
	_ = u1

	// Feed in p1's task first; groupByPlayer must re-order p0 ahead of p1.
	groups := groupByPlayer([]Task{tp1, tp0})
	require.Equal(t, 2, len(groups))
	assert.Equal(t, 0, groups[0][0].GetPlayer().GetIndex())
	assert.Equal(t, 1, groups[1][0].GetPlayer().GetIndex())
}
