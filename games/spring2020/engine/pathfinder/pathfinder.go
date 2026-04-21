// Package pathfinder
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathFinder.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/AStar.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/pathfinder/PathItem.java
package pathfinder

import (
	"container/heap"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

// Result mirrors PathFinderResult.
type Result struct {
	Path           []grid.Coord
	WeightedLength int
	IsNearest      bool
}

// HasNoPath reports whether the path is empty or missing.
func (r *Result) HasNoPath() bool { return r.WeightedLength == -1 }

// FindPath performs A* from "from" to "to" on the grid.
// If the exact target is unreachable, it returns a path to the grid cell that
// came closest to the target (by grid distance), with IsNearest=true.
// Weight of each step is determined by weight(coord).
func FindPath(g *grid.Grid, from, to grid.Coord, weight func(grid.Coord) int) Result {
	if weight == nil {
		weight = func(grid.Coord) int { return 1 }
	}
	a := newAStar(g, from, to, weight)
	items := a.find()
	if len(items) == 0 {
		a2 := newAStar(g, from, a.nearest, weight)
		items = a2.find()
		if len(items) == 0 {
			return Result{WeightedLength: -1}
		}
		return Result{Path: pathCoords(items), WeightedLength: items[len(items)-1].cumulativeLength, IsNearest: true}
	}
	return Result{Path: pathCoords(items), WeightedLength: items[len(items)-1].cumulativeLength}
}

func pathCoords(items []*pathItem) []grid.Coord {
	out := make([]grid.Coord, len(items))
	for i, it := range items {
		out[i] = it.coord
	}
	return out
}

type pathItem struct {
	coord                  grid.Coord
	cumulativeLength       int
	totalPrevisionalLength int
	precedent              *pathItem
	seq                    int // insertion order tie-breaker for stable priority
}

type astar struct {
	grid    *grid.Grid
	from    grid.Coord
	target  grid.Coord
	nearest grid.Coord
	weight  func(grid.Coord) int
	closed  map[grid.Coord]*pathItem
	open    *itemHeap
	seq     int
}

func newAStar(g *grid.Grid, from, target grid.Coord, weight func(grid.Coord) int) *astar {
	return &astar{
		grid:    g,
		from:    from,
		target:  target,
		nearest: from,
		weight:  weight,
		closed:  make(map[grid.Coord]*pathItem),
		open:    &itemHeap{},
	}
}

func (a *astar) find() []*pathItem {
	root := &pathItem{coord: a.from}
	a.push(root)
	for a.open.Len() > 0 {
		visiting := a.pop()
		if visiting.coord == a.target {
			return reversePath(visiting)
		}
		if _, ok := a.closed[visiting.coord]; ok {
			continue
		}
		a.closed[visiting.coord] = visiting

		for _, n := range a.grid.Neighbours(visiting.coord) {
			cell := a.grid.Get(n)
			if cell.Type == grid.CellFloor {
				a.addOpen(visiting, visiting.coord, n)
			}
		}

		if a.grid.CalculateDistance(visiting.coord, a.target) < a.grid.CalculateDistance(a.nearest, a.target) {
			a.nearest = visiting.coord
		}
	}
	return nil
}

func (a *astar) addOpen(visiting *pathItem, from, to grid.Coord) {
	if _, ok := a.closed[to]; ok {
		return
	}
	pi := &pathItem{
		coord:            to,
		cumulativeLength: visiting.cumulativeLength + a.weight(to),
		precedent:        visiting,
	}
	manh := a.grid.CalculateDistance(from, to)
	pi.totalPrevisionalLength = pi.cumulativeLength + manh
	a.push(pi)
}

func (a *astar) push(pi *pathItem) {
	pi.seq = a.seq
	a.seq++
	heap.Push(a.open, pi)
}

func (a *astar) pop() *pathItem {
	return heap.Pop(a.open).(*pathItem)
}

func reversePath(end *pathItem) []*pathItem {
	var out []*pathItem
	for p := end; p != nil; p = p.precedent {
		out = append(out, p)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

type itemHeap []*pathItem

func (h itemHeap) Len() int { return len(h) }
func (h itemHeap) Less(i, j int) bool {
	if h[i].totalPrevisionalLength != h[j].totalPrevisionalLength {
		return h[i].totalPrevisionalLength < h[j].totalPrevisionalLength
	}
	return h[i].seq < h[j].seq
}
func (h itemHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *itemHeap) Push(x any)   { *h = append(*h, x.(*pathItem)) }
func (h *itemHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
