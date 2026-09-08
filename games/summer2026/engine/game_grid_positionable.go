// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Positionable.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Positionable.java:3-6

public interface Positionable {
    Coord getPosition();
    double getDistanceMultiplier();
}
*/

// Positionable is anything Grid.ClosestTargets can measure a distance to.
type Positionable interface {
	Position() Coord
	DistanceMultiplier() float64
}
