// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Seed.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Seed.java:3-32

public class Seed {
    private int owner;
    private int sourceCell;
    private int targetCell;
    // setters/getters
}
*/

// Seed is the in-flight seed action, recorded during the action phase and
// resolved (or refunded) once both players have moved.
type Seed struct {
	Owner      int
	SourceCell int
	TargetCell int
}
