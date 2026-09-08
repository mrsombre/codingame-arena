// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainState.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainState.java:7-40

public class TrainState {
    public Coord coord;
    public TrainState prev;

    public TrainState(Coord coord, TrainState prev) { this.coord = coord; this.prev = prev; }
    public TrainState(Coord coord) { this(coord, null); }

    @Override public int hashCode() { return Objects.hash(coord); }
    @Override public boolean equals(Object obj) {
        ...
        return Objects.equals(coord, other.coord);
    }
}
*/

// TrainState is one step of a train path. Prev is the back-pointer TrainBFS
// reconstructs from; equality deliberately ignores it, which is why StateKey
// below is the coord alone.
type TrainState struct {
	Coord Coord
	Prev  *TrainState
}

func NewTrainState(coord Coord) *TrainState { return &TrainState{Coord: coord} }

func NewTrainStateFrom(coord Coord, prev *TrainState) *TrainState {
	return &TrainState{Coord: coord, Prev: prev}
}

func (s *TrainState) String() string { return fmt.Sprintf("State{, coord=%s}", s.Coord) }
