// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java
package engine

import (
	"fmt"
	"math"
)

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:3-10

public class Coord implements Positionable, Comparable<Coord> {
    public final int x;
    public final int y;

    public Coord(int x, int y) {
        this.x = x;
        this.y = y;
    }
*/

// Coord is an immutable grid position. It is a Go value type, so `==`
// replaces Java's equals/hashCode for comparison and map keys.
type Coord struct {
	X, Y int
}

func NewCoord(x, y int) Coord { return Coord{X: x, Y: y} }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:20-34

public Coord add(int x, int y) { return new Coord(this.x + x, this.y + y); }
public Coord add(Coord c) { return add(c.x, c.y); }
public Coord sub(int x, int y) { return new Coord(this.x - x, this.y - y); }
public Coord sub(Coord c) { return sub(c.x, c.y); }
*/

func (c Coord) Add(other Coord) Coord { return Coord{c.X + other.X, c.Y + other.Y} }
func (c Coord) AddXY(x, y int) Coord  { return Coord{c.X + x, c.Y + y} }
func (c Coord) Sub(other Coord) Coord { return Coord{c.X - other.X, c.Y - other.Y} }
func (c Coord) SubXY(x, y int) Coord  { return Coord{c.X - x, c.Y - y} }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:36-43

@Override
public int hashCode() {
    final int prime = 31;
    int result = 1;
    result = prime * result + x;
    result = prime * result + y;
    return result;
}
*/

// JavaHash reproduces Coord.hashCode() with 32-bit wraparound. GridMaker
// picks candidates out of a HashSet<Coord> by index, so map generation
// depends on this value feeding javahash's bucket ordering.
func (c Coord) JavaHash() int32 {
	return 31*(31+int32(c.X)) + int32(c.Y)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:12-18,72-90

public double euclideanTo(int x, int y) { return Math.sqrt(sqrEuclideanTo(x, y)); }
public double sqrEuclideanTo(double x, double y) { return Math.pow(x - this.x, 2) + Math.pow(y - this.y, 2); }
public int manhattanTo(int x, int y) { return Math.abs(x - this.x) + Math.abs(y - this.y); }
public int chebyshevTo(int x, int y) { return Math.max(Math.abs(x - this.x), Math.abs(y - this.y)); }
*/

func (c Coord) SqrEuclideanTo(x, y float64) float64 {
	return math.Pow(x-float64(c.X), 2) + math.Pow(y-float64(c.Y), 2)
}

func (c Coord) EuclideanTo(other Coord) float64 {
	return math.Sqrt(c.SqrEuclideanTo(float64(other.X), float64(other.Y)))
}

func (c Coord) ManhattanTo(other Coord) int {
	return abs(other.X-c.X) + abs(other.Y-c.Y)
}

func (c Coord) ChebyshevTo(other Coord) int {
	return max(abs(other.X-c.X), abs(other.Y-c.Y))
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:55-62,92-100

@Override public String toString() { return "(" + x + ", " + y + ")"; }
public String toIntString() { return x + " " + y; }
@Override public Coord getPosition() { return this; }
@Override public double getDistanceMultiplier() { return 1; }
*/

func (c Coord) String() string { return fmt.Sprintf("(%d, %d)", c.X, c.Y) }

func (c Coord) ToIntString() string { return fmt.Sprintf("%d %d", c.X, c.Y) }

func (c Coord) Position() Coord { return c }

func (c Coord) DistanceMultiplier() float64 { return 1 }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Coord.java:102-114

public Coord intNormalize() {
    return new Coord((int) Math.signum(x), (int) Math.signum(y));
}

@Override
public int compareTo(Coord o) {
    return this.x != o.x ? Integer.compare(this.x, o.x) : Integer.compare(this.y, o.y);
}
*/

// IntNormalize truncates to a cardinal direction.
func (c Coord) IntNormalize() Coord {
	return Coord{signum(c.X), signum(c.Y)}
}

// CompareTo orders by x then y, matching Comparable<Coord>. Java's TreeMap
// keyed by Coord (Game.doActions' tracksPlaced) iterates in this order.
func (c Coord) CompareTo(other Coord) int {
	if c.X != other.X {
		return compareInt(c.X, other.X)
	}
	return compareInt(c.Y, other.Y)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func signum(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
