// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/ChopTask.java
package engine

import (
	"regexp"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

var chopRe = regexp.MustCompile(`(?i)^\s*(CHOP)\s+(\d+)\s*$`)

type ChopTask struct{ TaskBase }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/ChopTask.java:14-24

public ChopTask(Player player, Board board, String command) {
    super(player, board);
    parseUnit(matcher);
    if (unit == null) return;
    if (unit.getCell().getPlant() == null)
        addParsingError("troll " + unit.getId() + " is not at a plant", InputError.NO_PLANT, false);
    if (unit.getChopPower() == 0)
        addParsingError("troll " + unit.getId() + " has no chopping power", InputError.NO_CHOP, false);
}
*/

func newChopTask(player *Player, board *Board, m []string, league int) Task {
	t := &ChopTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	if t.Unit.Cell.Plant == nil {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" is not at a plant", ErrNoPlant, false)
	}
	if t.Unit.ChopPower == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no chopping power", ErrNoChop, false)
	}
	return t
}

func (t *ChopTask) GetTaskPriority() int   { return 4 }
func (t *ChopTask) GetRequiredLeague() int { return 3 }
func (t *ChopTask) GetCell() *Cell         { return t.Unit.Cell }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/ChopTask.java:42-67

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (this != concurrentTasks.get(0)) return;
    for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
        Plant plant = byCell.get(0).getCell().getPlant();
        for (Task t : byCell) { plant.damage(t.unit.getChopPower()); t.applied = true; }
        if (plant.isDead()) {
            int remainingWood = plant.getSize();
            for (int i = 0; i < plant.getSize() && remainingWood > 0; i++) {
                for (Task t : byCell) {
                    if (t.unit.getFreeCarryCapacity() > 0) { t.unit.getInventory().incrementItem(Item.WOOD); remainingWood--; }
                }
            }
        }
        // summary
    }
}
*/

func (t *ChopTask) Apply(board *Board, concurrent []Task) {
	if concurrent[0] != t {
		return
	}
	for _, byCell := range groupByCell(concurrent) {
		plant := byCell[0].GetCell().Plant
		for _, task := range byCell {
			plant.Damage(task.GetUnit().ChopPower)
			task.base().Applied = true
		}
		if plant.IsDead() {
			remaining := plant.GetSize()
			for i := 0; i < plant.GetSize() && remaining > 0; i++ {
				for _, task := range byCell {
					if task.GetUnit().GetFreeCarryCapacity() > 0 {
						task.GetUnit().Inv.IncrementItem(ItemWOOD)
						remaining--
					}
				}
			}
		}
		for _, task := range byCell {
			if task.GetDeltaCarry() > 0 {
				task.GetPlayer().AddSummary("troll " + itoa(task.GetUnit().ID) +
					" collected " + itoa(task.GetDeltaCarry()) + " " + ItemWOOD.String())
			} else {
				task.GetPlayer().AddSummary("troll " + itoa(task.GetUnit().ID) + " damaged a tree")
			}
			cell := task.GetCell()
			board.tracePlayer(task.GetPlayer().GetIndex(), arena.MakeTurnTrace(TraceChop, ChopData{
				Unit:   task.GetUnit().ID,
				Cell:   [2]int{cell.X, cell.Y},
				Damage: task.GetUnit().ChopPower,
				Wood:   task.GetDeltaCarry(),
				Killed: plant.IsDead(),
			}))
		}
	}
}
