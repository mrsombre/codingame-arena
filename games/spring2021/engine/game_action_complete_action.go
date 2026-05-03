// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/action/CompleteAction.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/action/CompleteAction.java:3-13

public class CompleteAction extends Action {
    public CompleteAction(int targetId) {
        this.targetId = targetId;
    }
    @Override public boolean isComplete() { return true; }
}
*/

func NewCompleteAction(targetID int) Action {
	return Action{Kind: ActionComplete, TargetID: targetID}
}
