// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildState.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildState.java:9-28

public class AutobuildState {
    public enum CursorState {
        UNBOUGHT,
        BUILDABLE,
        BUILT,
        // For invalid coords
        BROKEN;
    }

    // For history
    public Action action;

    // For pathfinding
    Set<Integer> ownedZones;
    Set<Coord> tracks;
    int moneySpent;

    Coord cursor;
    CursorState cursorState;
}
*/

type CursorState int

const (
	CURSOR_UNBOUGHT CursorState = iota
	CURSOR_BUILDABLE
	CURSOR_BUILT
	// CURSOR_BROKEN marks a cursor on an invalid coord.
	CURSOR_BROKEN
)

var cursorStateNames = [4]string{
	CURSOR_UNBOUGHT:  "UNBOUGHT",
	CURSOR_BUILDABLE: "BUILDABLE",
	CURSOR_BUILT:     "BUILT",
	CURSOR_BROKEN:    "BROKEN",
}

func (s CursorState) String() string { return cursorStateNames[s] }

// AutobuildState is one node of the autobuild search: where the cursor sits,
// which cells the plan has already committed to building, and what it has
// cost so far.
//
// Action is the placement this step implies, nil for a step that only moves
// the cursor over existing track. Game.resolveAutobuild reads it back off the
// winning path — it is the search's output, not part of its state.
//
// OwnedZones exists in the source and is never written to, so it is always
// empty; the successor filter that consults it therefore falls through to the
// inked / town checks.
type AutobuildState struct {
	Action *Action

	OwnedZones map[int]struct{}
	Tracks     map[Coord]struct{}
	MoneySpent int

	Cursor      Coord
	CursorState CursorState
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildState.java:30-55

@Override public String toString() { return "State{" + cursorState + ", action='" + action + '\'' + ...; }
@Override public int hashCode() { return Objects.hash(cursor, cursorState); }
@Override public boolean equals(Object obj) {
    ...
    return Objects.equals(cursor, other.cursor) && Objects.equals(cursorState, other.cursorState);
}
*/

// AutobuildStateKey is AutobuildState.equals/hashCode: the cursor and its
// state, nothing else. Two plans that reach the same cell the same way are
// one node however differently they got there or however much they spent.
type AutobuildStateKey struct {
	Cursor      Coord
	CursorState CursorState
}

func (s *AutobuildState) Key() AutobuildStateKey {
	return AutobuildStateKey{Cursor: s.Cursor, CursorState: s.CursorState}
}

func (s *AutobuildState) String() string {
	return fmt.Sprintf("State{%s, action='%v', moneySpent=%d, cursor=%s}",
		s.CursorState, s.Action, s.MoneySpent, s.Cursor)
}
