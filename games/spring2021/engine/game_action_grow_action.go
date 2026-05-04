// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/action/GrowAction.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/action/GrowAction.java:3-13

public class GrowAction extends Action {
    public GrowAction(int targetId) {
        this.targetId = targetId;
    }
    @Override public boolean isGrow() { return true; }
}
*/

func NewGrowAction(targetID int) Action {
	return Action{Kind: ActionGrow, TargetID: targetID}
}
