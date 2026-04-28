// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java
package engine

import (
	"fmt"
	"math"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:3-9

public class Coord implements Comparable<Coord> {
    protected final int x;
    protected final int y;

    public Coord(int x, int y) {
        this.x = x;
        this.y = y;
    }
*/

// Coord is an immutable 2D coordinate.
type Coord struct {
	X int
	Y int
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:20-26

public Coord add(int x, int y) {
    return new Coord(this.x + x, this.y + y);
}

public Coord add(Coord c) {
    return add(c.x, c.y);
}
*/

func (c Coord) Add(other Coord) Coord {
	return Coord{X: c.X + other.X, Y: c.Y + other.Y}
}

func (c Coord) AddXY(x, y int) Coord {
	return Coord{X: c.X + x, Y: c.Y + y}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:64-74

public int manhattanTo(Coord other) {
    return manhattanTo(other.x, other.y);
}

public int manhattanTo(int x, int y) {
    return Math.abs(x - this.x) + Math.abs(y - this.y);
}
*/

func (c Coord) ManhattanTo(other Coord) int {
	return abs(c.X-other.X) + abs(c.Y-other.Y)
}

func (c Coord) ManhattanToXY(x, y int) int {
	return abs(c.X-x) + abs(c.Y-y)
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:68-70

public int chebyshevTo(int x, int y) {
    return Math.max(Math.abs(x - this.x), Math.abs(y - this.y));
}
*/

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

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:12-18

public double euclideanTo(int x, int y) {
    return Math.sqrt(sqrEuclideanTo(x, y));
}

public double sqrEuclideanTo(double x, double y) {
    return Math.pow(x - this.x, 2) + Math.pow(y - this.y, 2);
}
*/

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

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:84-91

@Override
public int compareTo(Coord o) {
    int cmp = Integer.compare(this.x, o.x);
    if (cmp == 0) {
        cmp = Integer.compare(this.y, o.y);
    }
    return cmp;
}
*/

// Less compares coords by X then Y (equivalent to Java's compareTo).
func (c Coord) Less(other Coord) bool {
	if c.X != other.X {
		return c.X < other.X
	}
	return c.Y < other.Y
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Coord.java:47-54

@Override
public String toString() {
    return "(" + x + ", " + y + ")";
}

public String toIntString() {
    return x + " " + y;
}
*/

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
