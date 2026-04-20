// Package grid
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Direction.java
package grid

import "fmt"

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

// Coord returns the delta for this direction.
func (d Direction) Coord() Coord {
	return dirCoords[d]
}

func (d Direction) String() string {
	return dirAliases[d]
}

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

// DirectionFromCoord returns the direction matching a delta coord.
func DirectionFromCoord(c Coord) Direction {
	for i, dc := range dirCoords {
		if dc == c {
			return Direction(i)
		}
	}
	return DirUnset
}

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
