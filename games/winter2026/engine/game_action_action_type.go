// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionType.java
package engine

import (
	"regexp"
	"strconv"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionType.java:10-42

public enum ActionType {

    MOVE_UP("^(?<birdId>\\d+) UP( (?<message>[^;]*))?", (match, action) -> {
        action.setBirdId(Integer.valueOf(match.group("birdId")));
        action.setDirection(Direction.NORTH);
        action.setMessage(match.group("message"));
    }),
    MOVE_DOWN("^(?<birdId>\\d+) DOWN( (?<message>[^;]*))?", ...Direction.SOUTH...),
    MOVE_LEFT("^(?<birdId>\\d+) LEFT( (?<message>[^;]*))?", ...Direction.WEST...),
    MOVE_RIGHT("^(?<birdId>\\d+) RIGHT( (?<message>[^;]*))?", ...Direction.EAST...),
    MARK("MARK (?<x>\\d+) (?<y>\\d+)", (match, action) -> {
        action.setCoord(new Coord(...));
    }),
    WAIT("WAIT", ActionType::doNothing);
*/

// ActionType identifies the kind of action.
type ActionType int

const (
	TypeMoveUp ActionType = iota
	TypeMoveDown
	TypeMoveLeft
	TypeMoveRight
	TypeMark
	TypeWait
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionType.java:47-64

private final Pattern pattern;
private final BiConsumer<Matcher, Action> consumer;

ActionType(String pattern, BiConsumer<Matcher, Action> consumer) {
    this.pattern = Pattern.compile(pattern, Pattern.CASE_INSENSITIVE);
    this.consumer = consumer;
}

public Pattern getPattern() { return pattern; }
public BiConsumer<Matcher, Action> getConsumer() { return consumer; }
*/

// pattern holds a compiled regex and a populate function for one action type.
type pattern struct {
	actionType ActionType
	re         *regexp.Regexp
	populate   func(match []string, groups map[string]int, a *Action)
}

var patterns []pattern

func init() {
	movePopulate := func(dir Direction) func([]string, map[string]int, *Action) {
		return func(match []string, groups map[string]int, a *Action) {
			a.BirdID, _ = strconv.Atoi(match[groups["birdId"]])
			a.HasBirdID = true
			a.Direction = dir
			if idx, ok := groups["messageGroup"]; ok && idx < len(match) && match[idx] != "" {
				messageIdx := groups["message"]
				a.Message = match[messageIdx]
				a.HasMessage = true
			}
		}
	}

	movePattern := func(dirName string) string {
		return `(?i)^(?P<birdId>\d+) ` + dirName + `(?P<messageGroup> (?P<message>[^;]*))?`
	}

	patterns = []pattern{
		{TypeMoveUp, regexp.MustCompile(movePattern("UP")), movePopulate(DirNorth)},
		{TypeMoveDown, regexp.MustCompile(movePattern("DOWN")), movePopulate(DirSouth)},
		{TypeMoveLeft, regexp.MustCompile(movePattern("LEFT")), movePopulate(DirWest)},
		{TypeMoveRight, regexp.MustCompile(movePattern("RIGHT")), movePopulate(DirEast)},
		{TypeMark, regexp.MustCompile(`(?i)MARK (?P<x>\d+) (?P<y>\d+)`), func(match []string, groups map[string]int, a *Action) {
			x, _ := strconv.Atoi(match[groups["x"]])
			y, _ := strconv.Atoi(match[groups["y"]])
			a.Coord = Coord{X: x, Y: y}
			a.HasCoord = true
		}},
		{TypeWait, regexp.MustCompile(`(?i)WAIT`), func([]string, map[string]int, *Action) {}},
	}
}

// subexpMap builds a name→index map for a compiled regexp.
func subexpMap(re *regexp.Regexp) map[string]int {
	m := make(map[string]int)
	for i, name := range re.SubexpNames() {
		if name != "" {
			m[name] = i
		}
	}
	return m
}
