// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/Action.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/Action.java:5-22

public interface Action {

    Action NO_ACTION = new Action() {
        @Override public PacmanType getType() { return null; }
        @Override public ActionType getActionType() { return ActionType.WAIT; }
    };

    public PacmanType getType();
    public ActionType getActionType();
}
*/

// Action is a tagged-union port of the Java Action hierarchy.
//   - MoveAction:   Type=ActionMove, Target set
//   - SpeedAction:  Type=ActionSpeed
//   - SwitchAction: Type=ActionSwitch, NewType set
//   - NO_ACTION:    Type=ActionWait (use NoAction to check)
type Action struct {
	Type    ActionType
	Target  Coord
	NewType PacmanType
}

// NoAction sentinel (equivalent to Java's Action.NO_ACTION).
var NoAction = Action{Type: ActionWait}

// IsNoAction reports whether a is the sentinel no-op action.
func (a Action) IsNoAction() bool {
	return a.Type == ActionWait
}
