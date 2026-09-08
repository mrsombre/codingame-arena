// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/ScheduleStep.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/ScheduleStep.java:3-5

public record ScheduleStep(int fromTownId, int toTownId) {}
*/

// ScheduleStep names one directed town-to-town connection. Tiles on an active
// connection's path carry one per connection they belong to.
type ScheduleStep struct {
	FromTownID int
	ToTownID   int
}
