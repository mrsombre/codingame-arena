// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/action/ActionType.java
package engine

import (
	"regexp"
	"strconv"
)

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/action/ActionType.java:9-45

public enum ActionType {
    AUTOPLACE("^AUTOPLACE (?<fromX>\\d+) (?<fromY>\\d+) (?<toX>\\d+) (?<toY>\\d+)", ...),
    PLACE_TRACK("^PLACE_TRACKS? (?<x>\\d+) (?<y>\\d+)", ...),
    DISRUPT("^DISRUPT (?<zoneId>\\d+)", ...),
    DISRUPT_ALT("^DISRUPT (?<x>\\d+) (?<y>\\d+)", ...),
    MESSAGE("^MESSAGE (?<message>[^;]*)", ...),
    WAIT("^WAIT", ActionType::doNothing);

    ActionType(String pattern, BiConsumer<Matcher, Action> consumer) {
        this.pattern = Pattern.compile(pattern, Pattern.CASE_INSENSITIVE);
        this.consumer = consumer;
    }
*/

type ActionType int

// Declaration order is the match order CommandManager walks; DISRUPT is
// tried before DISRUPT_ALT, so `DISRUPT 3` is a region id and `DISRUPT 3 4`
// falls through to the coordinate form.
const (
	ACTION_AUTOPLACE ActionType = iota
	ACTION_PLACE_TRACK
	ACTION_DISRUPT
	ACTION_DISRUPT_ALT
	ACTION_MESSAGE
	ACTION_WAIT
)

var actionTypeNames = [6]string{
	ACTION_AUTOPLACE:   "AUTOPLACE",
	ACTION_PLACE_TRACK: "PLACE_TRACK",
	ACTION_DISRUPT:     "DISRUPT",
	ACTION_DISRUPT_ALT: "DISRUPT_ALT",
	ACTION_MESSAGE:     "MESSAGE",
	ACTION_WAIT:        "WAIT",
}

func (t ActionType) String() string { return actionTypeNames[t] }

// ActionTypes is the values() order the parser iterates.
var ActionTypes = [6]ActionType{
	ACTION_AUTOPLACE,
	ACTION_PLACE_TRACK,
	ACTION_DISRUPT,
	ACTION_DISRUPT_ALT,
	ACTION_MESSAGE,
	ACTION_WAIT,
}

// Java compiles the patterns anchored at the start only but applies them with
// Matcher.matches(), which requires the whole input to match — hence the
// trailing `$` here. (?i) is Pattern.CASE_INSENSITIVE.
var actionTypePatterns = [6]*regexp.Regexp{
	ACTION_AUTOPLACE:   regexp.MustCompile(`(?i)^AUTOPLACE (?P<fromX>\d+) (?P<fromY>\d+) (?P<toX>\d+) (?P<toY>\d+)$`),
	ACTION_PLACE_TRACK: regexp.MustCompile(`(?i)^PLACE_TRACKS? (?P<x>\d+) (?P<y>\d+)$`),
	ACTION_DISRUPT:     regexp.MustCompile(`(?i)^DISRUPT (?P<zoneId>\d+)$`),
	ACTION_DISRUPT_ALT: regexp.MustCompile(`(?i)^DISRUPT (?P<x>\d+) (?P<y>\d+)$`),
	ACTION_MESSAGE:     regexp.MustCompile(`(?i)^MESSAGE (?P<message>[^;]*)$`),
	ACTION_WAIT:        regexp.MustCompile(`(?i)^WAIT$`),
}

func (t ActionType) Pattern() *regexp.Regexp { return actionTypePatterns[t] }

// Apply is the Java BiConsumer<Matcher, Action>: it copies the captured
// groups onto the action. It reports an error where Java would have thrown
// NumberFormatException on an out-of-range integer; the caller turns that
// into the same invalid-input disqualification an unparseable line gets.
func (t ActionType) Apply(match []string, action *Action) error {
	group := func(name string) string {
		for i, n := range actionTypePatterns[t].SubexpNames() {
			if n == name {
				return match[i]
			}
		}
		return ""
	}
	number := func(name string) (int, error) {
		return strconv.Atoi(group(name))
	}

	switch t {
	case ACTION_AUTOPLACE:
		fromX, err := number("fromX")
		if err != nil {
			return err
		}
		fromY, err := number("fromY")
		if err != nil {
			return err
		}
		toX, err := number("toX")
		if err != nil {
			return err
		}
		toY, err := number("toY")
		if err != nil {
			return err
		}
		action.From = Coord{fromX, fromY}
		action.To = Coord{toX, toY}
	case ACTION_PLACE_TRACK, ACTION_DISRUPT_ALT:
		x, err := number("x")
		if err != nil {
			return err
		}
		y, err := number("y")
		if err != nil {
			return err
		}
		action.Coord = Coord{x, y}
	case ACTION_DISRUPT:
		zoneID, err := number("zoneId")
		if err != nil {
			return err
		}
		action.ZoneID = zoneID
	case ACTION_MESSAGE:
		action.Message = group("message")
	case ACTION_WAIT:
	}
	return nil
}
