// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/action/SeedAction.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/action/SeedAction.java:3-14

public class SeedAction extends Action {
    public SeedAction(int sourceId, int targetId) {
        this.sourceId = sourceId;
        this.targetId = targetId;
    }
    @Override public boolean isSeed() { return true; }
}
*/

func NewSeedAction(sourceID, targetID int) Action {
	return Action{Kind: ActionSeed, SourceID: sourceID, TargetID: targetID}
}
