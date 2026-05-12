package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskManagerParseTasksHandlesMSGAndSplit(t *testing.T) {
	// Rules: commands are split on ';'. A `MSG <text>` segment sets the
	// player's message without becoming a Task.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "MSG hello world;MOVE 0 2 0", board.League)
	assert.Equal(t, "hello world", p0.GetMessage())
	require.Len(t, mgr.Tasks, 1, "MSG is consumed, MOVE produces a task")
	_, ok := mgr.Tasks[0].(*MoveTask)
	assert.True(t, ok)
}

func TestTaskManagerParseTasksClearsMessageEachTurn(t *testing.T) {
	// Java rule: player.setMessage("") fires at the top of ParseTasks so a
	// silent turn drops any previous bubble.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	p0.SetMessage("stale")
	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "WAIT", board.League)
	assert.Equal(t, "", p0.GetMessage(), "message reset on each parse")
}

func TestTaskManagerPopTasksReturnsLowestPriorityBucket(t *testing.T) {
	// peekTasks scans the queue, keeping only the lowest-priority items in
	// insertion order. MOVE=1 should fire before HARVEST=2 even though
	// HARVEST was queued first.
	board, p0, _ := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	a := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	b := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 2, 0)
	plantAt(board, ItemAPPLE, 1, 0, 4, 1, 20, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "HARVEST 0;MOVE 1 2 0", board.League)
	require.Len(t, mgr.Tasks, 2)

	first := mgr.PopTasks()
	require.Len(t, first, 1, "MOVE bucket on its own")
	_, ok := first[0].(*MoveTask)
	assert.True(t, ok)

	second := mgr.PopTasks()
	require.Len(t, second, 1, "then HARVEST")
	_, ok = second[0].(*HarvestTask)
	assert.True(t, ok)
	_ = a
	_ = b
}

func TestTaskManagerPopTasksKeepsInsertionOrderInBucket(t *testing.T) {
	// Tasks within the same priority bucket retain insertion order — that
	// drives the "concurrent[0] != this return" guard inside Apply.
	board, p0, _ := loadScenario(t, 4, []string{
		"0.......",
		"........",
		".......1",
	})
	a := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	b := spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 3, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "MOVE 1 4 0;MOVE 0 2 0", board.League)
	bucket := mgr.PopTasks()
	require.Len(t, bucket, 2)
	assert.Same(t, b, bucket[0].GetUnit(), "MOVE 1 first")
	assert.Same(t, a, bucket[1].GetUnit(), "MOVE 0 second")
}

func TestTaskManagerMergesPlayers(t *testing.T) {
	// Both players push into the same TaskManager; tasks coexist and pop
	// together by priority.
	board, p0, p1 := loadScenario(t, 4, []string{
		"0...",
		"....",
		"...1",
	})
	spawnUnit(board, p0, [4]int{1, 1, 1, 0}, 1, 0)
	spawnUnit(board, p1, [4]int{1, 1, 1, 0}, 2, 0)

	mgr := NewTaskManager()
	mgr.ParseTasks(p0, board, "MOVE 0 2 0", board.League)
	mgr.ParseTasks(p1, board, "MOVE 1 1 0", board.League)

	pop := mgr.PopTasks()
	assert.Len(t, pop, 2, "both players' MOVE bucket pops together")
}
