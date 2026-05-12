// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/MoveTask.java
package engine

import (
	"regexp"
	"sort"
	"strconv"
)

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/MoveTask.java:12-12

protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>MOVE)\\s+(?<id>\\d+)\\s+(?<x>-?\\d+)\\s+(?<y>-?\\d+)\\s*$", Pattern.CASE_INSENSITIVE);
*/

var moveRe = regexp.MustCompile(`(?i)^\s*(MOVE)\s+(\d+)\s+(-?\d+)\s+(-?\d+)\s*$`)

// MoveTask resolves a per-troll movement order. The target Cell is computed
// at parse time via Board.GetNextCell, mirroring Java MoveTask's constructor.
type MoveTask struct {
	TaskBase
	Target *Cell
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/MoveTask.java:14-21

public MoveTask(Player player, Board board, String command) {
    super(player, board);
    Matcher matcher = pattern.matcher(command);
    matcher.matches();
    parseUnit(matcher);
    if (unit == null) return;
    int x = Integer.parseInt(matcher.group("x"));
    int y = Integer.parseInt(matcher.group("y"));
    if (x < 0 || x >= board.getWidth() || y < 0 || y >= board.getHeight())
        addParsingError("(" + x + ", " + y + ") is outside of the board", InputError.OUT_OF_BOARD, false);
    else target = board.getNextCell(unit, board.getCell(x, y));
}
*/

func newMoveTask(player *Player, board *Board, m []string, league int) Task {
	t := &MoveTask{TaskBase: TaskBase{Player: player, Board: board}}
	t.parseUnit(m[2])
	if t.Unit == nil {
		return t
	}
	x, _ := strconv.Atoi(m[3])
	y, _ := strconv.Atoi(m[4])
	if x < 0 || x >= board.Width || y < 0 || y >= board.Height {
		t.addParsingError("("+itoa(x)+", "+itoa(y)+") is outside of the board", ErrOutOfBoard, false)
		return t
	}
	t.Target = board.GetNextCell(t.Unit, board.GetCell(x, y))
	return t
}

func (t *MoveTask) GetTaskPriority() int   { return 1 }
func (t *MoveTask) GetRequiredLeague() int { return 1 }
func (t *MoveTask) GetTarget() *Cell       { return t.Target }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/MoveTask.java:316-414

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (this != concurrentTasks.get(0)) return;
    // 1. group by player, then per-player:
    //    - collect each unit's current target (own cell if not moving)
    //    - remove stationary units (their cells stay marked occupied)
    //    - iteratively move units whose target is uniquely free
    //    - resolve circular swaps via cycle detection on the remaining set
}
*/

func (t *MoveTask) Apply(board *Board, concurrent []Task) {
	if concurrent[0] != t {
		return
	}
	for _, moves := range groupByPlayer(concurrent) {
		player := moves[0].GetPlayer()
		// all of the player's units, in board insertion order
		units := board.GetUnitsByPlayerID(player.GetIndex())
		targets := make([]*Cell, len(units))
		for i, u := range units {
			targets[i] = u.Cell
		}
		// each MoveTask overrides the target for its unit
		for _, mv := range moves {
			m := mv.(*MoveTask)
			for i, u := range units {
				if u == m.Unit {
					targets[i] = m.Target
					break
				}
			}
		}

		occupied := make([][]bool, board.Width)
		for x := 0; x < board.Width; x++ {
			occupied[x] = make([]bool, board.Height)
		}
		for i := len(units) - 1; i >= 0; i-- {
			occupied[units[i].Cell.X][units[i].Cell.Y] = true
			if units[i].Cell == targets[i] {
				units = append(units[:i], units[i+1:]...)
				targets = append(targets[:i], targets[i+1:]...)
			}
		}

		madeMove := true
		resolveBlocking := false
		for madeMove {
			madeMove = false

			// targetFreq[cell] = number of units pointing at it
			freq := make(map[*Cell]int)
			for _, c := range targets {
				freq[c]++
			}
			for i := len(units) - 1; i >= 0; i-- {
				cell := targets[i]
				canMove := resolveBlocking || freq[cell] == 1
				if canMove && !occupied[cell.X][cell.Y] {
					occupied[cell.X][cell.Y] = true
					occupied[units[i].Cell.X][units[i].Cell.Y] = false
					units[i].SetCell(cell)
					units = append(units[:i], units[i+1:]...)
					targets = append(targets[:i], targets[i+1:]...)
					madeMove = true
					resolveBlocking = false
				}
			}

			if madeMove {
				continue
			}

			// detect circular swaps: a->b, b->a, or longer cycles
			for startIndex := 0; startIndex < len(units); startIndex++ {
				path := []int{startIndex}
				looped := false
				for iter := 0; iter < len(units)+1; iter++ {
					tgt := targets[path[len(path)-1]]
					blockingIdx := -1
					for j, u := range units {
						if u.Cell == tgt {
							blockingIdx = j
							break
						}
					}
					if blockingIdx < 0 {
						break
					}
					if blockingIdx == path[0] {
						looped = true
						break
					}
					path = append(path, blockingIdx)
				}
				if looped {
					sort.Ints(path)
					for i := len(path) - 1; i >= 0; i-- {
						idx := path[i]
						units[idx].SetCell(targets[idx])
						units = append(units[:idx], units[idx+1:]...)
						targets = append(targets[:idx], targets[idx+1:]...)
						madeMove = true
					}
				}
			}

			if !madeMove && !resolveBlocking {
				resolveBlocking = true
				madeMove = true
			}
		}

		// any unit still in `units` failed to move
		stuck := make(map[*Unit]struct{}, len(units))
		for _, u := range units {
			stuck[u] = struct{}{}
		}
		for _, mv := range moves {
			m := mv.(*MoveTask)
			if _, isStuck := stuck[m.Unit]; isStuck {
				m.Player.AddError(NewInputError(
					"troll "+itoa(m.Unit.ID)+" can't move to ("+itoa(m.Target.X)+", "+itoa(m.Target.Y)+"), target blocked",
					ErrMoveBlocked, false))
			} else {
				m.Applied = true
				m.Player.AddSummary("troll " + itoa(m.Unit.ID) +
					" moved to (" + itoa(m.Unit.Cell.X) + ", " + itoa(m.Unit.Cell.Y) + ")")
			}
		}
	}
}
