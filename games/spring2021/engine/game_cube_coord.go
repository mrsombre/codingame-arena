// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/CubeCoord.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CubeCoord.java:5-13

public class CubeCoord {
    static int[][] directions = new int[][] {
        { 1, -1, 0 }, { +1, 0, -1 }, { 0, +1, -1 },
        { -1, +1, 0 }, { -1, 0, +1 }, { 0, -1, +1 }
    };
    int x, y, z;
}
*/

// CubeCoord is an immutable hex cube coordinate (axial trio x+y+z=0).
type CubeCoord struct {
	X, Y, Z int
}

// CubeDirections lists the six unit hex direction vectors. Indexed by
// orientation 0..5 the same way the Java referee numbers them — orientation
// is sent verbatim to the bot in the global input lines.
var CubeDirections = [6][3]int{
	{1, -1, 0},
	{1, 0, -1},
	{0, 1, -1},
	{-1, 1, 0},
	{-1, 0, 1},
	{0, -1, 1},
}

func NewCubeCoord(x, y, z int) CubeCoord {
	return CubeCoord{X: x, Y: y, Z: z}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CubeCoord.java:55-65

CubeCoord neighbor(int orientation, int distance) {
    int nx = this.x + directions[orientation][0] * distance;
    int ny = this.y + directions[orientation][1] * distance;
    int nz = this.z + directions[orientation][2] * distance;
    return new CubeCoord(nx, ny, nz);
}
*/

func (c CubeCoord) Neighbor(orientation int) CubeCoord {
	return c.NeighborAt(orientation, 1)
}

func (c CubeCoord) NeighborAt(orientation, distance int) CubeCoord {
	d := CubeDirections[orientation]
	return CubeCoord{
		X: c.X + d[0]*distance,
		Y: c.Y + d[1]*distance,
		Z: c.Z + d[2]*distance,
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CubeCoord.java:67-69

int distanceTo(CubeCoord dst) {
    return (Math.abs(x - dst.x) + Math.abs(y - dst.y) + Math.abs(z - dst.z)) / 2;
}
*/

func (c CubeCoord) DistanceTo(dst CubeCoord) int {
	return (absInt(c.X-dst.X) + absInt(c.Y-dst.Y) + absInt(c.Z-dst.Z)) / 2
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CubeCoord.java:76-79

public CubeCoord getOpposite() {
    CubeCoord oppositeCoord = new CubeCoord(-this.x, -this.y, -this.z);
    return oppositeCoord;
}
*/

func (c CubeCoord) Opposite() CubeCoord {
	return CubeCoord{X: -c.X, Y: -c.Y, Z: -c.Z}
}

func (c CubeCoord) String() string {
	return fmt.Sprintf("%d %d %d", c.X, c.Y, c.Z)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
