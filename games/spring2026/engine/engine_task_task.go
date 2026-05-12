// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/Task.java
package engine

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/Task.java:11-19

public abstract class Task {
    protected Player player;
    protected Board board;
    protected Unit unit;
    private int unitInitialCarry;
    protected boolean failedParsing = false;
    protected boolean applied;
*/

// Task is the interface every command type satisfies. Methods mirror the
// abstract methods on the Java engine.task.Task super-class.
type Task interface {
	GetTaskPriority() int
	GetRequiredLeague() int
	Apply(board *Board, concurrent []Task)
	GetCell() *Cell

	// Accessors common to all tasks (provided by TaskBase).
	GetPlayer() *Player
	GetUnit() *Unit
	HasFailedParsing() bool
	WasApplied() bool
	GetDeltaCarry() int
	markApplied()
	failParse()
	base() *TaskBase
}

// TaskBase carries the shared fields from Java's abstract Task. Specific task
// structs embed it so they only need to declare their unique fields.
type TaskBase struct {
	Player        *Player
	Board         *Board
	Unit          *Unit
	InitialCarry  int
	FailedParsing bool
	Applied       bool
}

func (t *TaskBase) GetPlayer() *Player    { return t.Player }
func (t *TaskBase) GetUnit() *Unit        { return t.Unit }
func (t *TaskBase) HasFailedParsing() bool { return t.FailedParsing }
func (t *TaskBase) WasApplied() bool      { return t.Applied }
func (t *TaskBase) GetCell() *Cell        { return nil } // default; tasks override

// GetDeltaCarry mirrors Java Task.getDeltaCarry: current carry total minus the
// snapshot captured by parseUnit. Used for game-summary text only.
func (t *TaskBase) GetDeltaCarry() int {
	return t.Unit.Inv.GetTotal() - t.InitialCarry
}

func (t *TaskBase) markApplied() { t.Applied = true }
func (t *TaskBase) failParse()   { t.FailedParsing = true }
func (t *TaskBase) base() *TaskBase { return t }

// addParsingError mirrors Java Task.addParsingError: the first error per task
// marks failedParsing and surfaces on the player error stream.
func (t *TaskBase) addParsingError(message string, code int, critical bool) {
	if t.FailedParsing {
		return
	}
	t.FailedParsing = true
	t.Player.AddError(NewInputError(message, code, critical))
}

// parseUnit reads the unit id from the regex match and validates ownership.
// On failure the task is marked as having failedParsing and t.Unit stays nil
// — caller should early-return in that case, mirroring Java's `if (unit ==
// null) return;`.
func (t *TaskBase) parseUnit(idText string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		t.addParsingError("invalid troll id: "+idText, ErrUnitNotFound, false)
		return
	}
	u := t.Board.GetUnit(id)
	if u == nil {
		t.addParsingError("Troll "+itoa(id)+" does not exist", ErrUnitNotFound, false)
		return
	}
	t.Unit = u
	t.InitialCarry = u.Inv.GetTotal()
	if u.Player != t.Player {
		t.addParsingError("You don't own troll "+itoa(id), ErrUnitNotOwned, false)
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/Task.java:586-626

public static Task parseTask(Player player, Board board, String command, int league, HashSet<Unit> usedUnits) {
    if (command.trim().equals("")) return null;
    if (command.trim().toUpperCase().equals("WAIT")) return null;
    Task task = null;
    try {
        if (MoveTask.pattern.matcher(command).matches()) task = new MoveTask(...);
        ...
    } catch (Exception ex) { player.addError(new InputError("Unknown command: " + command, ...)); return null; }
    if (task == null) { player.addError(...); return null; }
    if (task.unit != null && !usedUnits.add(task.unit)) {
        task.addParsingError("Troll " + task.unit.getId() + " already used", ...);
        return null;
    }
    if (task.getRequiredLeague() > league) { task.addParsingError("Command not available in your league: " + command, ...); return null; }
    return task;
}
*/

// taskCtor is the factory for one concrete task type. It returns nil and may
// register parse errors on the player; the caller validates further.
type taskCtor struct {
	re   *regexp.Regexp
	make func(*Player, *Board, []string, int) Task
}

// taskPatterns is the regex/constructor table used by ParseTask. Order
// matters only insofar as Java's `if-cascade` was lattice-shaped (each
// matcher is tested independently) — we mirror that exactly by iterating
// through them and using the last match (the same way Java falls through).
var taskPatterns = []taskCtor{
	{re: moveRe, make: newMoveTask},
	{re: harvestRe, make: newHarvestTask},
	{re: plantRe, make: newPlantTask},
	{re: chopRe, make: newChopTask},
	{re: pickRe, make: newPickTask},
	{re: trainRe, make: newTrainTask},
	{re: dropRe, make: newDropTask},
	{re: mineRe, make: newMineTask},
	{re: waitRe, make: newWaitTask},
}

// ParseTask returns a Task for the command, or nil when the command is empty,
// a WAIT, or invalid. Side-effect: registers errors on player.
func ParseTask(player *Player, board *Board, command string, league int, usedUnits map[*Unit]struct{}) Task {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return nil
	}
	if strings.EqualFold(trimmed, "WAIT") {
		board.tracePlayer(player.GetIndex(), arena.TurnTrace{Type: TraceWait})
		return nil
	}

	var task Task
	for _, ctor := range taskPatterns {
		m := ctor.re.FindStringSubmatch(command)
		if m == nil {
			continue
		}
		task = ctor.make(player, board, m, league)
	}
	if task == nil {
		player.AddError(NewInputError("Unknown command: "+command, ErrUnknownCommand, true))
		return nil
	}
	if u := task.GetUnit(); u != nil {
		if _, used := usedUnits[u]; used {
			task.base().addParsingError("Troll "+itoa(u.ID)+" already used", ErrAlreadyUsed, false)
			return nil
		}
		usedUnits[u] = struct{}{}
	}
	if task.GetRequiredLeague() > league {
		task.base().addParsingError("Command not available in your league: "+command, ErrNotAvailable, false)
		return nil
	}
	return task
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/Task.java:669-679

protected ArrayList<ArrayList<Task>> groupByCell(ArrayList<Task> tasks) {
    Hashtable<Cell, ArrayList<Task>> byCell = new Hashtable<>();
    ...
    for (ArrayList<Task> list : byCell.values().stream().sorted(Comparator.comparing(task -> task.get(0).getCell().getId())).collect(Collectors.toList()))
        result.add(list);
    return result;
}
*/

// groupByCell groups tasks by their target Cell, then orders the groups by
// the cell's Java-style id (x + (y<<16)). Within a group, original insertion
// order is preserved.
func groupByCell(tasks []Task) [][]Task {
	byCell := make(map[*Cell][]Task)
	cells := make([]*Cell, 0)
	for _, t := range tasks {
		c := t.GetCell()
		if _, ok := byCell[c]; !ok {
			cells = append(cells, c)
		}
		byCell[c] = append(byCell[c], t)
	}
	sort.Slice(cells, func(i, j int) bool {
		return cells[i].GetID() < cells[j].GetID()
	})
	result := make([][]Task, 0, len(cells))
	for _, c := range cells {
		result = append(result, byCell[c])
	}
	return result
}

// groupByPlayer mirrors Java Task.groupByPlayer, but iterates deterministically
// by player index so simulations stay reproducible (Java's Hashtable order is
// implementation-dependent).
func groupByPlayer(tasks []Task) [][]Task {
	byPlayer := make(map[*Player][]Task)
	players := make([]*Player, 0)
	for _, t := range tasks {
		p := t.GetPlayer()
		if _, ok := byPlayer[p]; !ok {
			players = append(players, p)
		}
		byPlayer[p] = append(byPlayer[p], t)
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].GetIndex() < players[j].GetIndex()
	})
	result := make([][]Task, 0, len(players))
	for _, p := range players {
		result = append(result, byPlayer[p])
	}
	return result
}
