// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/MoveAction.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/MoveAction.java:6-27

public class MoveAction implements Action {
    private Coord destination;
    public MoveAction(Coord destination, boolean activateSpeed) {
        this.destination = destination;
    }
    public Coord getTarget() { return destination; }
    @Override public ActionType getActionType() { return ActionType.MOVE; }
}
*/

// NewMoveAction builds a MOVE action targeting coord.
func NewMoveAction(target Coord) Action {
	return Action{Type: ActionMove, Target: target}
}
