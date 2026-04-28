// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java
package engine

import (
	"fmt"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java:6-16

public class Action {
    final ActionType type;
    private Direction direction;
    private Integer birdId;
    private Coord coord;
    private String message;

    public Action(ActionType type) {
        this.type = type;
    }
*/

// Action holds a parsed player command.
type Action struct {
	Type       ActionType
	Direction  Direction
	BirdID     int
	HasBirdID  bool
	Coord      Coord
	HasCoord   bool
	Message    string
	HasMessage bool
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java:35-41

public boolean isMove() {
    return direction != null;
}

public boolean isMark() {
    return coord != null;
}
*/

func (a Action) IsMove() bool { return a.HasBirdID && a.Direction != DirUnset }
func (a Action) IsMark() bool { return a.HasCoord }

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java:31-34

@Override
public String toString() {
    return "Action [type=" + type + ", direction=" + direction + ", birdId=" + birdId + "]";
}
*/

func (a Action) String() string {
	return fmt.Sprintf("Action [type=%d, direction=%s, birdId=%d]", a.Type, a.Direction, a.BirdID)
}

// ActionError is returned when an action string cannot be parsed.
type ActionError struct {
	Message string
}

func (e *ActionError) Error() string { return e.Message }

// Parse tries each pattern against a single action string.
// Returns an ActionError if no pattern matches.
func Parse(raw string) (Action, error) {
	s := raw
	for _, p := range patterns {
		match := p.re.FindStringSubmatch(s)
		if match == nil || match[0] != s {
			continue
		}
		a := Action{Type: p.actionType}
		p.populate(match, subexpMap(p.re), &a)
		return a, nil
	}
	return Action{}, &ActionError{Message: fmt.Sprintf("invalid action: %s", raw)}
}
