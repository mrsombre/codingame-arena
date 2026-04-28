// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java
package engine

import "fmt"

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java:3-9

public enum Direction {
    NORTH(0, -1, "N"),
    EAST(1, 0, "E"),
    SOUTH(0, 1, "S"),
    WEST(-1, 0, "W"),
    UNSET(0, 0, "X");
*/

// Direction represents a cardinal direction or unset.
type Direction int

const (
	DirNorth Direction = iota
	DirEast
	DirSouth
	DirWest
	DirUnset
)

var dirCoords = [5]Coord{
	DirNorth: {X: 0, Y: -1},
	DirEast:  {X: 1, Y: 0},
	DirSouth: {X: 0, Y: 1},
	DirWest:  {X: -1, Y: 0},
	DirUnset: {X: 0, Y: 0},
}

var dirAliases = [5]string{
	DirNorth: "N",
	DirEast:  "E",
	DirSouth: "S",
	DirWest:  "W",
	DirUnset: "X",
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java:11-14

Direction(int x, int y, String alias) {
    this.coord = new Coord(x, y);
    this.alias = alias;
}
*/

// Coord returns the delta for this direction.
func (d Direction) Coord() Coord {
	return dirCoords[d]
}

func (d Direction) String() string {
	return dirAliases[d]
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java:46-59

public Direction opposite() {
    switch (this) {
    case NORTH: return SOUTH;
    case EAST:  return WEST;
    case SOUTH: return NORTH;
    case WEST:  return EAST;
    default:    return UNSET;
    }
}
*/

// Opposite returns the reverse direction.
func (d Direction) Opposite() Direction {
	switch d {
	case DirNorth:
		return DirSouth
	case DirEast:
		return DirWest
	case DirSouth:
		return DirNorth
	case DirWest:
		return DirEast
	default:
		return DirUnset
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java:23-30

public static Direction fromCoord(Coord coord) {
    for (Direction dir : Direction.values()) {
        if (dir.coord.equals(coord)) {
            return dir;
        }
    }
    return UNSET;
}
*/

// DirectionFromCoord returns the direction matching a delta coord.
func DirectionFromCoord(c Coord) Direction {
	for i, dc := range dirCoords {
		if dc == c {
			return Direction(i)
		}
	}
	return DirUnset
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java:32-44

public static Direction fromAlias(String alias) {
    switch (alias) {
    case "N": return NORTH;
    case "E": return EAST;
    case "S": return SOUTH;
    case "W": return WEST;
    }
    throw new RuntimeException(alias + " is not a direction alias");
}
*/

// DirectionFromAlias returns the direction for a single-char alias.
func DirectionFromAlias(alias string) Direction {
	switch alias {
	case "N":
		return DirNorth
	case "E":
		return DirEast
	case "S":
		return DirSouth
	case "W":
		return DirWest
	}
	panic(fmt.Sprintf("%s is not a direction alias", alias))
}
