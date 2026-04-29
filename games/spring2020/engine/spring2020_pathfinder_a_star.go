// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java
package engine

import (
	"container/heap"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java:19-38

public class AStar {
    Map<Coord, PathItem> closedList = new HashMap<>();
    PriorityQueue<PathItem> openList = new PriorityQueue<PathItem>(Comparator.comparingInt(PathItem::getTotalPrevisionalLength));
    List<PathItem> path = new ArrayList<PathItem>();
    Grid grid;
    Coord from;
    Coord target;
    Coord nearest;
    private Function<Coord, Integer> weightFunction;

    public AStar(Grid grid, Coord from, Coord target, Function<Coord, Integer> weightFunction) {
        this.grid = grid;
        this.from = from;
        this.target = target;
        this.weightFunction = weightFunction;
        this.nearest = from;
    }
}
*/

type AStar struct {
	Grid    *Grid
	From    Coord
	Target  Coord
	Nearest Coord
	Weight  func(Coord) int
	Closed  map[Coord]*PathItem
	Open    *ItemHeap
}

func NewAStar(g *Grid, from, target Coord, weight func(Coord) int) *AStar {
	return &AStar{
		Grid:    g,
		From:    from,
		Target:  target,
		Nearest: from,
		Weight:  weight,
		Closed:  make(map[Coord]*PathItem),
		Open:    &ItemHeap{},
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java:57-86

PathItem getPathItemLinkedList() {
    PathItem root = new PathItem();
    root.coord = this.from;
    openList.add(root);

    while (openList.size() > 0) {
        PathItem visiting = openList.remove();
        Coord visitingCoord = visiting.coord;

        if (visitingCoord.equals(target)) return visiting;
        if (closedList.containsKey(visitingCoord)) continue;
        closedList.put(visitingCoord, visiting);

        for (Coord neighbor : grid.getNeighbours(visitingCoord)) {
            if (grid.get(neighbor).getType() == CellType.FLOOR) {
                addToOpenList(visiting, visitingCoord, neighbor);
            }
        }

        if (grid.calculateDistance(visitingCoord, target) < grid.calculateDistance(nearest, target)) {
            this.nearest = visitingCoord;
        }
    }
    return null;
}
*/

func (a *AStar) Find() []*PathItem {
	root := &PathItem{Coord: a.From}
	heap.Push(a.Open, root)
	for a.Open.Len() > 0 {
		visiting := heap.Pop(a.Open).(*PathItem)
		if visiting.Coord == a.Target {
			return ReversePath(visiting)
		}
		if _, ok := a.Closed[visiting.Coord]; ok {
			continue
		}
		a.Closed[visiting.Coord] = visiting

		for _, n := range a.Grid.Neighbours(visiting.Coord) {
			cell := a.Grid.Get(n)
			if cell.Type == CellFloor {
				a.AddOpen(visiting, visiting.Coord, n)
			}
		}

		if a.Grid.CalculateDistance(visiting.Coord, a.Target) < a.Grid.CalculateDistance(a.Nearest, a.Target) {
			a.Nearest = visiting.Coord
		}
	}
	return nil
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java:88-99

void addToOpenList(PathItem visiting, Coord fromCoord, Coord toCoord) {
    if (closedList.containsKey(toCoord)) return;
    PathItem pi = new PathItem();
    pi.coord = toCoord;
    pi.cumulativeLength = visiting.cumulativeLength + weightFunction.apply(toCoord);
    int manh = grid.calculateDistance(fromCoord, toCoord);
    pi.totalPrevisionalLength = pi.cumulativeLength + manh;
    pi.precedent = visiting;
    openList.add(pi);
}
*/

func (a *AStar) AddOpen(visiting *PathItem, from, to Coord) {
	if _, ok := a.Closed[to]; ok {
		return
	}
	pi := &PathItem{
		Coord:            to,
		CumulativeLength: visiting.CumulativeLength + a.Weight(to),
		Precedent:        visiting,
	}
	manh := a.Grid.CalculateDistance(from, to)
	pi.TotalPrevisionalLength = pi.CumulativeLength + manh
	heap.Push(a.Open, pi)
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java:49-55

void calculatePath(PathItem item) {
    PathItem i = item;
    while (i != null) {
        path.add(0, i);
        i = i.precedent;
    }
}
*/

// ReversePath walks the precedent chain and returns it from start to end.
func ReversePath(end *PathItem) []*PathItem {
	var out []*PathItem
	for p := end; p != nil; p = p.Precedent {
		out = append(out, p)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ItemHeap orders by TotalPrevisionalLength only, mirroring Java's
// PriorityQueue with Comparator.comparingInt(...). Equal-priority items
// resolve in heap-internal order (which Go's container/heap matches because
// both implementations are array-backed binary heaps with the same sift
// rules — adding an explicit FIFO tiebreaker would diverge from Java).
type ItemHeap []*PathItem

func (h ItemHeap) Len() int { return len(h) }
func (h ItemHeap) Less(i, j int) bool {
	return h[i].TotalPrevisionalLength < h[j].TotalPrevisionalLength
}
func (h ItemHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *ItemHeap) Push(x any)   { *h = append(*h, x.(*PathItem)) }
func (h *ItemHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
