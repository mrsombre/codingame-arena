// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Direction.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Direction.java:3-16

public enum Direction {
    NORTH(0, -1, "N"),
    EAST(1, 0, "E"),
    SOUTH(0, 1, "S"),
    WEST(-1, 0, "W"),
    UNSET(0, 0, "X");

    public Coord coord;
    public String alias;
*/

type Direction int

const (
	NORTH Direction = iota
	EAST
	SOUTH
	WEST
	UNSET
)

// Declaration order is the connection tie-break priority: a shortest path is
// explored NORTH, EAST, SOUTH, WEST from the requesting town.
var directionCoords = [5]Coord{
	NORTH: {0, -1},
	EAST:  {1, 0},
	SOUTH: {0, 1},
	WEST:  {-1, 0},
	UNSET: {0, 0},
}

var directionAliases = [5]string{
	NORTH: "N",
	EAST:  "E",
	SOUTH: "S",
	WEST:  "W",
	UNSET: "X",
}

func (d Direction) Coord() Coord { return directionCoords[d] }

func (d Direction) Alias() string { return directionAliases[d] }

func (d Direction) String() string { return d.Alias() }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Direction.java:18-32

public static Direction fromCoord(Coord coord) {
    if (coord.equals(Direction.NORTH.coord)) return Direction.NORTH;
    ...
    return Direction.UNSET;
}
*/

func DirectionFromCoord(coord Coord) Direction {
	for _, d := range []Direction{NORTH, EAST, SOUTH, WEST} {
		if directionCoords[d] == coord {
			return d
		}
	}
	return UNSET
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Direction.java:35-41

public Direction opposite() {
    return this.alias.equals("N") ? Direction.SOUTH : (...);
}
*/

func (d Direction) Opposite() Direction {
	switch d {
	case NORTH:
		return SOUTH
	case EAST:
		return WEST
	case SOUTH:
		return NORTH
	case WEST:
		return EAST
	default:
		return UNSET
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Direction.java:49-61

public static Direction fromAlias(String alias) {
    switch (alias) { case "N": return NORTH; ... }
    throw new RuntimeException(alias + " is not a direction alias");
}
*/

// DirectionFromAlias panics on an unknown alias, matching the Java
// RuntimeException — the only callers are hardcoded fixtures.
func DirectionFromAlias(alias string) Direction {
	switch alias {
	case "N":
		return NORTH
	case "E":
		return EAST
	case "S":
		return SOUTH
	case "W":
		return WEST
	}
	panic(fmt.Sprintf("%s is not a direction alias", alias))
}
