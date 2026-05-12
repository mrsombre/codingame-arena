// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/DropTask.java
package engine

import "regexp"

var dropRe = regexp.MustCompile(`(?i)^\s*(DROP)\s+(\d+)\s*$`)

type DropTask struct{ TaskBase }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/DropTask.java:13-23

public DropTask(Player player, Board board, String command) {
    super(player, board);
    parseUnit(matcher);
    if (unit == null) return;
    if (unit.getInventory().getTotal() == 0)
        addParsingError("troll " + unit.getId() + " has nothing to drop", InputError.NO_SEEDS, false);
    if (!unit.isNearShack())
        addParsingError("troll " + unit.getId() + " isn't next to shack", InputError.NO_SHACK, false);
}
*/

func newDropTask(player *Player, board *Board, m []string, league int) Task {
	t := &DropTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	if t.Unit.Inv.GetTotal() == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has nothing to drop", ErrNoSeeds, false)
	}
	if !t.Unit.IsNearShack() {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" isn't next to shack", ErrNoShack, false)
	}
	return t
}

func (t *DropTask) GetTaskPriority() int   { return 7 }
func (t *DropTask) GetRequiredLeague() int { return 1 }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/DropTask.java:39-49

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    for (int i = 0; i < unit.getInventory().getItemsLength(); i++) {
        player.getInventory().setItem(i, player.getInventory().getItemCount(i) + unit.getInventory().getItemCount(i));
        unit.getInventory().setItem(i, 0);
    }
    applied = true;
}
*/

func (t *DropTask) Apply(board *Board, concurrent []Task) {
	for i := 0; i < ItemsCount; i++ {
		item := Item(i)
		t.Player.Inv.SetItem(item, t.Player.Inv.GetItemCount(item)+t.Unit.Inv.GetItemCount(item))
		t.Unit.Inv.SetItem(item, 0)
	}
	t.Applied = true
	itemText := "item"
	if t.GetDeltaCarry() < -1 {
		itemText += "s"
	}
	t.Player.AddSummary("troll " + itoa(t.Unit.ID) + " dropped " + itoa(-t.GetDeltaCarry()) + " " + itemText + " to the shack")
}
