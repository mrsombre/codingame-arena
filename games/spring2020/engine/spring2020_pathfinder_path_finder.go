// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathFinder.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathFinder.java:16-33

public static class PathFinderResult {
    public static final PathFinderResult NO_PATH = new PathFinderResult();
    public List<Coord> path = new ArrayList<>();
    public int weightedLength = -1;
    public boolean isNearest = false;

    public boolean hasNextCoord() { return path.size() > 1; }
    public Coord getNextCoord()   { return path.get(1); }
    public boolean hasNoPath()    { return weightedLength == -1; }
}
*/

// PathResult mirrors PathFinderResult.
type PathResult struct {
	Path           []Coord
	WeightedLength int
	IsNearest      bool
}

// HasNoPath reports whether the path is empty or missing.
func (r *PathResult) HasNoPath() bool { return r.WeightedLength == -1 }

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathFinder.java:60-79

public PathFinderResult findPath() {
    if (from == null || to == null) return new PathFinderResult();

    AStar a = new AStar(grid, from, to, weightFunction);
    List<PathItem> pathItems = a.find();
    PathFinderResult pfr = new PathFinderResult();

    if (pathItems.isEmpty()) {
        pfr.isNearest = true;
        pathItems = new AStar(grid, from, a.getNearest(), weightFunction).find();
    }

    pfr.path = pathItems.stream().map(item -> item.coord).collect(Collectors.toList());
    pfr.weightedLength = pathItems.get(pathItems.size() - 1).cumulativeLength;
    return pfr;
}
*/

// FindPath performs A* from "from" to "to" on the grid.
// If the exact target is unreachable, it returns a path to the grid cell that
// came closest to the target (by grid distance), with IsNearest=true.
// Weight of each step is determined by weight(coord).
func FindPath(g *Grid, from, to Coord, weight func(Coord) int) PathResult {
	if weight == nil {
		weight = func(Coord) int { return 1 }
	}
	a := NewAStar(g, from, to, weight)
	items := a.Find()
	if len(items) == 0 {
		a2 := NewAStar(g, from, a.Nearest, weight)
		items = a2.Find()
		if len(items) == 0 {
			return PathResult{WeightedLength: -1}
		}
		return PathResult{Path: PathCoords(items), WeightedLength: items[len(items)-1].CumulativeLength, IsNearest: true}
	}
	return PathResult{Path: PathCoords(items), WeightedLength: items[len(items)-1].CumulativeLength}
}

func PathCoords(items []*PathItem) []Coord {
	out := make([]Coord, len(items))
	for i, it := range items {
		out[i] = it.Coord
	}
	return out
}
