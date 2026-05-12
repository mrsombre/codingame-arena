// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/TrainTask.java
package engine

import (
	"regexp"
	"strconv"
)

var trainRe = regexp.MustCompile(`(?i)^\s*(TRAIN)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)

type TrainTask struct {
	TaskBase
	Talents [4]int
	League  int
}

func maxOfPlantHealth() int {
	m := PLANT_FINAL_HEALTH[0]
	for _, v := range PLANT_FINAL_HEALTH {
		if v > m {
			m = v
		}
	}
	return m
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/TrainTask.java:14-39

public TrainTask(Player player, Board board, String command, int league) {
    super(player, board);
    this.league = league;
    Matcher matcher = pattern.matcher(command);
    matcher.matches();
    // parse 4 talents
    if (movementSpeed < 1 || movementSpeed > board.getWidth() * board.getHeight())
        addParsingError(..., INVALID_SKILL, false);
    if (carryCapacity < 0 || carryCapacity > 1000) addParsingError(...);
    if (harvestPower < 0 || harvestPower > Constants.PLANT_MAX_RESOURCES) addParsingError(...);
    if (league < 3 && chopPower > 0) addParsingError("chop power is not available in this league", NOT_AVAILABLE, false);
    if (chopPower < 0 || chopPower > Arrays.stream(Constants.PLANT_FINAL_HEALTH).max().getAsInt()) addParsingError(...);
    if (!Unit.canTrain(player, talents, league)) addCantAfford();
}
*/

func newTrainTask(player *Player, board *Board, m []string, league int) Task {
	t := &TrainTask{TaskBase: TaskBase{Player: player, Board: board}, League: league}
	ms, _ := strconv.Atoi(m[2])
	cc, _ := strconv.Atoi(m[3])
	hp, _ := strconv.Atoi(m[4])
	cp, _ := strconv.Atoi(m[5])
	t.Talents = [4]int{ms, cc, hp, cp}
	if ms < 1 || ms > board.Width*board.Height {
		t.addParsingError("invalid movement speed: "+itoa(ms), ErrInvalidSkill, false)
	}
	if cc < 0 || cc > 1000 {
		t.addParsingError("invalid carry capacity: "+itoa(cc), ErrInvalidSkill, false)
	}
	if hp < 0 || hp > PLANT_MAX_RESOURCES {
		t.addParsingError("invalid harvest power: "+itoa(hp), ErrInvalidSkill, false)
	}
	if league < 3 && cp > 0 {
		t.addParsingError("chop power is not available in this league", ErrNotAvailable, false)
	}
	if cp < 0 || cp > maxOfPlantHealth() {
		t.addParsingError("invalid chop power: "+itoa(cp), ErrInvalidSkill, false)
	}
	if !CanTrain(player, t.Talents, league) {
		t.addCantAfford()
	}
	return t
}

func (t *TrainTask) addCantAfford() {
	costs := GetTrainingCosts(t.Player, t.Talents, t.League)
	t.addParsingError(
		"can't afford unit training, costs: "+
			itoa(costs[ItemPLUM])+" "+ItemPLUM.String()+", "+
			itoa(costs[ItemLEMON])+" "+ItemLEMON.String()+", "+
			itoa(costs[ItemAPPLE])+" "+ItemAPPLE.String()+", "+
			itoa(costs[ItemIRON])+" "+ItemIRON.String(),
		ErrCantAfford, false)
}

func (t *TrainTask) GetTaskPriority() int   { return 6 }
func (t *TrainTask) GetRequiredLeague() int { return 2 }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/TrainTask.java:59-70

public void apply(Board board, ArrayList<Task> concurrentTasks) {
    if (!Unit.canTrain(player, talents, league)) { addCantAfford(); return; }
    if (board.getUnitsByCell(player.getShack()).count() > 0) {
        addParsingError("can't train unit, cell blocked", InputError.MOVE_BLOCKED, false);
        return;
    }
    Unit unit = new Unit(player, talents, league);
    board.addUnit(unit);
    applied = true;
}
*/

func (t *TrainTask) Apply(board *Board, concurrent []Task) {
	if !CanTrain(t.Player, t.Talents, t.League) {
		t.addCantAfford()
		return
	}
	if len(board.GetUnitsByCell(t.Player.Shack)) > 0 {
		t.addParsingError("can't train unit, cell blocked", ErrMoveBlocked, false)
		return
	}
	u := NewUnit(t.Player, t.Talents, t.League)
	board.AddUnit(u)
	t.Applied = true
	t.Player.AddSummary("trained a troll")
}
