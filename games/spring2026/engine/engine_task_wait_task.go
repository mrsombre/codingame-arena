// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/WaitTask.java
package engine

import "regexp"

var waitRe = regexp.MustCompile(`(?i)^\s*(WAIT)\s*$`)

type WaitTask struct{ TaskBase }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/WaitTask.java:11-26

public WaitTask(Player player, Board board, String command) {
    super(player, board);
}
@Override public int getTaskPriority() { return 10; }
@Override public int getRequiredLeague() { return 1; }
@Override public void apply(Board board, ArrayList<Task> concurrentTasks) { applied = true; }
*/

// WaitTask exists only so that the regex cascade in ParseTask has something to
// match on for explicit "WAIT" commands. The TaskManager filter in parseTasks
// drops the bare command before it reaches the priority queue, so Apply is
// effectively dead — but kept for parity.
func newWaitTask(player *Player, board *Board, m []string, league int) Task {
	return &WaitTask{TaskBase: TaskBase{Player: player, Board: board}}
}

func (t *WaitTask) GetTaskPriority() int   { return 10 }
func (t *WaitTask) GetRequiredLeague() int { return 1 }
func (t *WaitTask) Apply(board *Board, concurrent []Task) {
	t.Applied = true
}
