// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/MineTask.java
package engine

import "regexp"

var mineRe = regexp.MustCompile(`(?i)^\s*(MINE)\s+(\d+)\s*$`)

type MineTask struct{ TaskBase }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/MineTask.java:14-22

public MineTask(Player player, Board board, String command) {
    super(player, board);
    parseUnit(matcher);
    if (unit == null) return;
    if (!unit.getCell().isNearIron())
        addParsingError("troll " + unit.getId() + " is not next to iron", InputError.NO_IRON, false);
    if (unit.getFreeCarryCapacity() == 0)
        addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
    if (unit.getChopPower() == 0)
        addParsingError("troll " + unit.getId() + " has no chopping power", InputError.NO_CHOP, false);
}
*/

func newMineTask(player *Player, board *Board, m []string, league int) Task {
	t := &MineTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	if !t.Unit.Cell.IsNearIron() {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" is not next to iron", ErrNoIron, false)
	}
	if t.Unit.GetFreeCarryCapacity() == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no capacity", ErrNoCapacity, false)
	}
	if t.Unit.ChopPower == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no chopping power", ErrNoChop, false)
	}
	return t
}

func (t *MineTask) GetTaskPriority() int   { return 8 }
func (t *MineTask) GetRequiredLeague() int { return 3 }

func (t *MineTask) Apply(board *Board, concurrent []Task) {
	t.Unit.Mine()
	t.Applied = true
	t.Player.AddSummary("troll " + itoa(t.Unit.ID) + " collected " + itoa(t.GetDeltaCarry()) + " " + ItemIRON.String())
}
