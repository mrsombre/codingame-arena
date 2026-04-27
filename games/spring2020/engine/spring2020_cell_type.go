// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/CellType.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/CellType.java:3-5

public enum CellType {
    WALL, FLOOR;
}
*/

// CellType matches Java CellType enum values.
type CellType int

const (
	CellWall CellType = iota
	CellFloor
)
