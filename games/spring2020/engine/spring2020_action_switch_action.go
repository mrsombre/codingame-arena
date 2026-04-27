// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SwitchAction.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SwitchAction.java:5-26

public class SwitchAction implements Action {
    private PacmanType type;
    public SwitchAction(PacmanType type) { this.type = type; }
    public PacmanType getNewType() { return type; }
    @Override public ActionType getActionType() { return ActionType.SWITCH; }
}
*/

// NewSwitchAction builds a SWITCH action to the given pac type.
func NewSwitchAction(t PacmanType) Action {
	return Action{Type: ActionSwitch, NewType: t}
}
