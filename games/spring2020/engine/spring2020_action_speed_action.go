// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SpeedAction.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SpeedAction.java:5-15

public class SpeedAction implements Action {
    @Override public ActionType getActionType() { return ActionType.SPEED; }
}
*/

// NewSpeedAction builds a SPEED action.
func NewSpeedAction() Action {
	return Action{Type: ActionSpeed}
}
