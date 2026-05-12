// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/PickTask.java
package engine

import (
	"regexp"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

var pickRe = regexp.MustCompile(`(?i)^\s*(PICK)\s+(\d+)\s+(\w+)\s*$`)

type PickTask struct {
	TaskBase
	Type Item
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/PickTask.java:14-30

public PickTask(Player player, Board board, String command) {
    super(player, board);
    parseUnit(matcher);
    if (unit == null) return;
    if (unit.getFreeCarryCapacity() == 0)
        addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
    String typeText = matcher.group("type");
    // resolve item
    if (type < 0 || type >= Item.values().length || !Item.values()[type].isPlant())
        addParsingError(typeText + " is not a plant", InputError.INVALID_PLANT, false);
    else if (player.getInventory().getItemCount(type) == 0)
        addParsingError(Item.values()[type] + " is out of stock", InputError.OUT_OF_STOCK, false);
    if (!unit.isNearShack())
        addParsingError("troll " + unit.getId() + " isn't next to shack", InputError.NO_SHACK, false);
}
*/

func newPickTask(player *Player, board *Board, m []string, league int) Task {
	t := &PickTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	if t.Unit.GetFreeCarryCapacity() == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no capacity", ErrNoCapacity, false)
	}
	typeText := m[3]
	t.Type = resolveItem(typeText)
	if t.Type < 0 || int(t.Type) >= ItemsCount || !t.Type.IsPlant() {
		t.addParsingError(typeText+" is not a plant", ErrInvalidPlant, false)
	} else if player.Inv.GetItemCount(t.Type) == 0 {
		t.addParsingError(t.Type.String()+" is out of stock", ErrOutOfStock, false)
	}
	if !t.Unit.IsNearShack() {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" isn't next to shack", ErrNoShack, false)
	}
	return t
}

func (t *PickTask) GetTaskPriority() int   { return 5 }
func (t *PickTask) GetRequiredLeague() int { return 2 }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/PickTask.java:50-60

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (player.getInventory().getItemCount(type) > 0) {
        player.getInventory().decrementItem(type);
        unit.getInventory().incrementItem(type);
        applied = true;
    } else
        player.addError(new InputError("troll " + unit.getId() + " can't pick " + Item.values()[type] + ", out of stock", InputError.CANT_AFFORD, false));
}
*/

func (t *PickTask) Apply(board *Board, concurrent []Task) {
	if t.Player.Inv.GetItemCount(t.Type) > 0 {
		t.Player.Inv.DecrementItem(t.Type)
		t.Unit.Inv.IncrementItem(t.Type)
		t.Applied = true
		t.Player.AddSummary("troll " + itoa(t.Unit.ID) + " picked 1 " + t.Type.String())
		board.tracePlayer(t.Player.GetIndex(), arena.MakeTurnTrace(TracePick, PickData{
			Unit: t.Unit.ID,
			Type: t.Type.String(),
		}))
	} else {
		t.Player.AddError(NewInputError(
			"troll "+itoa(t.Unit.ID)+" can't pick "+t.Type.String()+", out of stock",
			ErrCantAfford, false))
	}
}
