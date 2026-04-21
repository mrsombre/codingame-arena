// Package grid
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java
package grid

import (
	"fmt"
	"math"
)

// Coord is an immutable 2D coordinate.
type Coord struct {
	X int
	Y int
}

func (c Coord) Add(other Coord) Coord {
	return Coord{X: c.X + other.X, Y: c.Y + other.Y}
}

func (c Coord) AddXY(x, y int) Coord {
	return Coord{X: c.X + x, Y: c.Y + y}
}

func (c Coord) ManhattanTo(other Coord) int {
	return abs(c.X-other.X) + abs(c.Y-other.Y)
}

func (c Coord) ManhattanToXY(x, y int) int {
	return abs(c.X-x) + abs(c.Y-y)
}

func (c Coord) ChebyshevTo(other Coord) int {
	dx := abs(c.X - other.X)
	dy := abs(c.Y - other.Y)
	if dx > dy {
		return dx
	}
	return dy
}

func (c Coord) ChebyshevToXY(x, y int) int {
	dx := abs(c.X - x)
	dy := abs(c.Y - y)
	if dx > dy {
		return dx
	}
	return dy
}

func (c Coord) EuclideanTo(other Coord) float64 {
	return math.Sqrt(c.SqrEuclideanTo(other))
}

func (c Coord) EuclideanToXY(x, y int) float64 {
	return math.Sqrt(c.SqrEuclideanToXY(float64(x), float64(y)))
}

func (c Coord) SqrEuclideanTo(other Coord) float64 {
	return c.SqrEuclideanToXY(float64(other.X), float64(other.Y))
}

func (c Coord) SqrEuclideanToXY(x, y float64) float64 {
	dx := x - float64(c.X)
	dy := y - float64(c.Y)
	return dx*dx + dy*dy
}

// Less compares coords by X then Y (equivalent to Java's compareTo).
func (c Coord) Less(other Coord) bool {
	if c.X != other.X {
		return c.X < other.X
	}
	return c.Y < other.Y
}

func (c Coord) String() string {
	return fmt.Sprintf("(%d, %d)", c.X, c.Y)
}

func (c Coord) ToIntString() string {
	return fmt.Sprintf("%d %d", c.X, c.Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
