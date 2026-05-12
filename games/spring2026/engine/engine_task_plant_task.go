// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/PlantTask.java
package engine

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

var plantRe = regexp.MustCompile(`(?i)^\s*(PLANT)\s+(\d+)\s+(\w+)\s*$`)

type PlantTask struct {
	TaskBase
	Type Item
}

// resolveItem parses the user-facing token: matches an Item name (case
// insensitive) or a numeric ordinal. Returns -1 when neither parses.
func resolveItem(text string) Item {
	upper := strings.ToUpper(text)
	if v := ItemFromName(upper); v >= 0 {
		return v
	}
	if n, err := strconv.Atoi(text); err == nil {
		return Item(n)
	}
	return -1
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/PlantTask.java:16-33

public PlantTask(Player player, Board board, String command) {
    super(player, board);
    Matcher matcher = pattern.matcher(command);
    matcher.matches();
    parseUnit(matcher);
    if (unit == null) return;
    String typeText = matcher.group("type");
    // resolve type from name or ordinal
    if (type < 0 || type >= Item.values().length || !Item.values()[type].isPlant())
        addParsingError(type + " is not a plant", InputError.INVALID_PLANT, false);
    if (unit.getCell().getType() != Cell.Type.GRASS)
        addParsingError("troll " + unit.getId() + " is not on grass. How did you even get there?", InputError.NO_GRASS, false);
    if (unit.getCell().getPlant() != null)
        addParsingError("troll " + unit.getId() + " can't plant on top of existing plant", InputError.EXISTING_PLANT, false);
    if (unit.getInventory().getItemCount(type) == 0)
        addParsingError("troll " + unit.getId() + " has no " + Item.values()[type] + " seed to plant", InputError.NO_SEEDS, false);
}
*/

func newPlantTask(player *Player, board *Board, m []string, league int) Task {
	t := &PlantTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	typeText := m[3]
	t.Type = resolveItem(typeText)
	if t.Type < 0 || int(t.Type) >= ItemsCount || !t.Type.IsPlant() {
		// Java prints `type` here (the int post-parsing) when name failed,
		// otherwise the int again. We match by formatting the int even when
		// the user passed a bare numeric token — keeps parity for replays.
		t.addParsingError(itoa(int(t.Type))+" is not a plant", ErrInvalidPlant, false)
	}
	if t.Unit.Cell.Type != CellGRASS {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" is not on grass. How did you even get there?", ErrNoGrass, false)
	}
	if t.Unit.Cell.Plant != nil {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" can't plant on top of existing plant", ErrExistingPlant, false)
	}
	if t.Type >= 0 && int(t.Type) < ItemsCount && t.Unit.Inv.GetItemCount(t.Type) == 0 {
		t.addParsingError("troll "+itoa(t.Unit.ID)+" has no "+t.Type.String()+" seed to plant", ErrNoSeeds, false)
	}
	return t
}

func (t *PlantTask) GetTaskPriority() int   { return 3 }
func (t *PlantTask) GetRequiredLeague() int { return 2 }
func (t *PlantTask) GetCell() *Cell         { return t.Unit.Cell }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/PlantTask.java:54-72

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (this != concurrentTasks.get(0)) return;
    for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
        HashSet<Integer> types = new HashSet<>();
        for (Task t : byCell) types.add(((PlantTask) t).type);
        if (types.size() == 1) { // all the same — proceed
            for (Task t : byCell) {
                int type = ((PlantTask) t).type;
                t.unit.getInventory().decrementItem(type);
                if (t.getCell().getPlant() == null) {
                    Plant newPlant = new Plant(t.unit.getCell(), Item.values()[type]);
                    t.unit.getCell().setPlant(newPlant);
                    board.addPlant(newPlant);
                }
                t.applied = true;
            }
        } else { // mismatch — nobody plants
            // OPPONENT_BLOCKING error per task
        }
    }
}
*/

func (t *PlantTask) Apply(board *Board, concurrent []Task) {
	if concurrent[0] != t {
		return
	}
	for _, byCell := range groupByCell(concurrent) {
		types := make(map[Item]struct{})
		for _, task := range byCell {
			types[task.(*PlantTask).Type] = struct{}{}
		}
		if len(types) == 1 {
			for _, task := range byCell {
				pt := task.(*PlantTask)
				pt.Unit.Inv.DecrementItem(pt.Type)
				if pt.GetCell().Plant == nil {
					p := NewPlant(pt.Unit.Cell, pt.Type)
					pt.Unit.Cell.SetPlant(p)
					board.AddPlant(p)
				}
				pt.Applied = true
				pt.Player.AddSummary("troll " + itoa(pt.Unit.ID) + " planted a " + pt.Type.String())
				cell := pt.Unit.Cell
				board.tracePlayer(pt.Player.GetIndex(), arena.MakeTurnTrace(TracePlant, PlantData{
					Unit: pt.Unit.ID,
					Cell: [2]int{cell.X, cell.Y},
					Type: pt.Type.String(),
				}))
			}
		} else {
			for _, task := range byCell {
				pt := task.(*PlantTask)
				pt.addParsingError("troll "+itoa(pt.Unit.ID)+" can't plant "+pt.Type.String()+", contradicting opponent planting", ErrOpponentBlocking, false)
			}
		}
	}
}
