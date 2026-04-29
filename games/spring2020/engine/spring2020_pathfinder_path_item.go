// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathItem.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathItem.java:5-10

public class PathItem {
    public int cumulativeLength = 0;
    int totalPrevisionalLength = 0;
    PathItem precedent = null;
    Coord coord;
}
*/

type PathItem struct {
	Coord                  Coord
	CumulativeLength       int
	TotalPrevisionalLength int
	Precedent              *PathItem
}
