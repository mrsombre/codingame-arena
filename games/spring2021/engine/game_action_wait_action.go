// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/action/WaitAction.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/action/WaitAction.java:3-9

public class WaitAction extends Action {
    @Override public boolean isWait() { return true; }
}
*/

func NewWaitAction() Action {
	return Action{Kind: ActionWait}
}
