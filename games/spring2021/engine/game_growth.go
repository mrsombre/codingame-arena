// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Growth.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Growth.java:1-7

public class Growth {
    public int targetId;
    public Player player;
    public int cost;
}
*/

// Growth is a tiny data carrier used by the Java engine but never referenced
// from any simulation path that matters for the arena port. Kept for source
// parity; consumers are free to ignore it.
type Growth struct {
	TargetID int
	Player   *Player
	Cost     int
}
