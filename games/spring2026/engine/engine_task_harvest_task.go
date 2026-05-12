// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/HarvestTask.java
package engine

import "regexp"

var harvestRe = regexp.MustCompile(`(?i)^\s*(HARVEST)\s+(\d+)\s*$`)

type HarvestTask struct{ TaskBase }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/HarvestTask.java:14-21

public HarvestTask(Player player, Board board, String command) {
    super(player, board);
    Matcher matcher = pattern.matcher(command);
    matcher.matches();
    parseUnit(matcher);
    if (unit == null) return;
    Plant plant = unit.getCell().getPlant();
    if (plant == null) addParsingError("troll " + unit.getId() + " is not at a plant", InputError.NO_PLANT, false);
    else if (plant.getResources() == 0)
        addParsingError("troll " + unit.getId() + " has no fruits to harvest", InputError.NO_FRUIT, false);
    if (unit.getFreeCarryCapacity() == 0)
        addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
    if (unit.getHarvestPower() == 0)
        addParsingError("troll " + unit.getId() + " has no harvest power", InputError.NO_HARVEST, false);
}
*/

func newHarvestTask(player *Player, board *Board, m []string, league int) Task {
	t := &HarvestTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	plant := t.Unit.Cell.Plant
	if plant == nil {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" is not at a plant", ErrNoPlant, false)
	} else if plant.Resources == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no fruits to harvest", ErrNoFruit, false)
	}
	if t.Unit.GetFreeCarryCapacity() == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no capacity", ErrNoCapacity, false)
	}
	if t.Unit.HarvestPower == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no harvest power", ErrNoHarvest, false)
	}
	return t
}

func (t *HarvestTask) GetTaskPriority() int   { return 2 }
func (t *HarvestTask) GetRequiredLeague() int { return 1 }
func (t *HarvestTask) GetCell() *Cell         { return t.Unit.Cell }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/HarvestTask.java:46-65

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (this != concurrentTasks.get(0)) return;
    for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
        Plant plant = byCell.get(0).getCell().getPlant();
        for (int i = 1; i <= Constants.PLANT_MAX_RESOURCES; i++) {
            if (plant.getResources() == 0) break;
            for (Task t : byCell) { t.unit.harvest(i); t.applied = true; }
        }
        // summary text omitted
    }
}
*/

func (t *HarvestTask) Apply(board *Board, concurrent []Task) {
	if concurrent[0] != t {
		return
	}
	for _, byCell := range groupByCell(concurrent) {
		plant := byCell[0].GetCell().Plant
		for i := 1; i <= PLANT_MAX_RESOURCES; i++ {
			if plant.Resources == 0 {
				break
			}
			for _, task := range byCell {
				task.GetUnit().Harvest(i)
				task.base().Applied = true
			}
		}
		for _, task := range byCell {
			itemText := plant.Type.String()
			if task.GetDeltaCarry() > 1 {
				itemText += "s"
			}
			task.GetPlayer().AddSummary("troll " + itoa(task.GetUnit().ID) +
				" harvested " + itoa(task.GetDeltaCarry()) + " " + itemText)
		}
	}
}
