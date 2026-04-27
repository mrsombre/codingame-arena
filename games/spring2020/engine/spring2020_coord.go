// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java
package engine

import (
	"fmt"
	"math"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java:3-10

public class Coord {
    protected final int x;
    protected final int y;
    public Coord(int x, int y) { this.x = x; this.y = y; }
}
*/

// Coord is an immutable 2D coordinate.
type Coord struct {
	X int
	Y int
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java:64-70

public Coord add(Coord d)      { return new Coord(x + d.x, y + d.y); }
public Coord subtract(Coord d) { return new Coord(x - d.x, y - d.y); }
*/

func (c Coord) Add(other Coord) Coord {
	return Coord{X: c.X + other.X, Y: c.Y + other.Y}
}

func (c Coord) AddXY(x, y int) Coord {
	return Coord{X: c.X + x, Y: c.Y + y}
}

func (c Coord) Subtract(other Coord) Coord {
	return Coord{X: c.X - other.X, Y: c.Y - other.Y}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java:12-14,52-62

public int manhattanTo(int x, int y)        { return Math.abs(x - this.x) + Math.abs(y - this.y); }
public int manhattanTo(Coord other)         { return manhattanTo(other.x, other.y); }
public int chebyshevTo(int x, int y)        { return Math.max(Math.abs(x - this.x), Math.abs(y - this.y)); }
public double euclideanTo(int x, int y)     { return Math.sqrt(sqrEuclideanTo(x, y)); }
*/

func (c Coord) ManhattanTo(other Coord) int {
	return Abs(c.X-other.X) + Abs(c.Y-other.Y)
}

func (c Coord) ManhattanToXY(x, y int) int {
	return Abs(c.X-x) + Abs(c.Y-y)
}

func (c Coord) ChebyshevToXY(x, y int) int {
	dx := Abs(c.X - x)
	dy := Abs(c.Y - y)
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

// Less compares coords by X then Y. Go-only helper for sorts.
func (c Coord) Less(other Coord) bool {
	if c.X != other.X {
		return c.X < other.X
	}
	return c.Y < other.Y
}

func (c Coord) String() string {
	return fmt.Sprintf("(%d, %d)", c.X, c.Y)
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Coord.java:76-80

public Coord getUnitVector() {
    int newX = this.x == 0 ? 0 : this.x / Math.abs(this.x);
    int newY = this.y == 0 ? 0 : this.y / Math.abs(this.y);
    return new Coord(newX, newY);
}
*/

func (c Coord) UnitVector() Coord {
	ux, uy := 0, 0
	if c.X != 0 {
		ux = c.X / Abs(c.X)
	}
	if c.Y != 0 {
		uy = c.Y / Abs(c.Y)
	}
	return Coord{X: ux, Y: uy}
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
