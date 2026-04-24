// Package grid
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java
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

func (c Coord) Subtract(other Coord) Coord {
	return Coord{X: c.X - other.X, Y: c.Y - other.Y}
}

func (c Coord) ManhattanTo(other Coord) int {
	return abs(c.X-other.X) + abs(c.Y-other.Y)
}

func (c Coord) ManhattanToXY(x, y int) int {
	return abs(c.X-x) + abs(c.Y-y)
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
	dx := float64(c.X - other.X)
	dy := float64(c.Y - other.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// Less compares coords by X then Y.
func (c Coord) Less(other Coord) bool {
	if c.X != other.X {
		return c.X < other.X
	}
	return c.Y < other.Y
}

func (c Coord) String() string {
	return fmt.Sprintf("(%d, %d)", c.X, c.Y)
}

func (c Coord) GetUnitVector() Coord {
	ux, uy := 0, 0
	if c.X != 0 {
		ux = c.X / abs(c.X)
	}
	if c.Y != 0 {
		uy = c.Y / abs(c.Y)
	}
	return Coord{X: ux, Y: uy}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
