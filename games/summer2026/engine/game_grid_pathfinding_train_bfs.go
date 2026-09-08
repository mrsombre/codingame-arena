// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainBFS.java
package engine

import "slices"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainBFS.java:15-25

public class TrainBFS {
    Grid grid;
    Town from, to;

    public TrainBFS(Grid grid, Town from, Town to) { this.grid = grid; this.from = from; this.to = to; }
}
*/

// TrainBFS finds the shortest route from one town to another over cells a
// train can occupy — track of any owner, or another town. It is the one
// pathfinder that decides connections and therefore scoring.
type TrainBFS struct {
	Grid *Grid
	From *Town
	To   *Town
}

func NewTrainBFS(grid *Grid, from, to *Town) *TrainBFS {
	return &TrainBFS{Grid: grid, From: from, To: to}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainBFS.java:27-60

public List<Coord> search() {
    Deque<TrainState> fifo = new LinkedList<>();
    fifo.add(new TrainState(from.coord));
    Set<Coord> visited = new LinkedHashSet<>();
    visited.add(from.coord);

    while (!fifo.isEmpty()) {
        TrainState current = fifo.poll();
        if (isGoal(current)) {
            List<Coord> path = new LinkedList<>();
            TrainState state = current;
            while (state != null) { path.add(0, state.coord); state = state.prev; }
            return path;
        }

        List<Coord> sortedNeighs = grid.getNeighbours(current.coord).stream().sorted((a, b) -> {
            return Direction.fromCoord(b.sub(a)).ordinal();
        }).toList();

        for (Coord neighbour : sortedNeighs) {
            if (grid.canTrainPass(neighbour) && !visited.contains(neighbour)) {
                visited.add(neighbour);
                fifo.add(new TrainState(neighbour, current));
            }
        }
    }
    return Collections.emptyList();
}
*/

// Search returns the cells from the origin town to the destination town
// inclusive, or an empty path when no route over track exists. Both endpoints
// are towns, so a returned path always has at least two entries.
//
// Breadth-first with the queue seeded from the origin means every returned
// path is of minimum length; which minimum-length path is returned is decided
// entirely by the order neighbours are enqueued, and that order is
// compareTrainBFSNeighbours below.
func (b *TrainBFS) Search() []Coord {
	fifo := []*TrainState{NewTrainState(b.From.Coord)}
	visited := map[Coord]bool{b.From.Coord: true}

	for len(fifo) > 0 {
		current := fifo[0]
		fifo = fifo[1:]

		if current.Coord == b.To.Coord {
			var path []Coord
			for state := current; state != nil; state = state.Prev {
				path = append(path, state.Coord)
			}
			slices.Reverse(path)
			return path
		}

		neighbours := b.Grid.Neighbours(current.Coord)
		javaListSort(neighbours, compareTrainBFSNeighbours)

		for _, neighbour := range neighbours {
			if b.Grid.CanTrainPass(neighbour) && !visited[neighbour] {
				visited[neighbour] = true
				fifo = append(fifo, NewTrainStateFrom(neighbour, current))
			}
		}
	}

	return nil
}

// compareTrainBFSNeighbours is the upstream comparator, ported verbatim
// quirks and all.
//
// It reads as "order by the direction from a to b", but a and b are two
// neighbours of the same centre cell, never a centre and its neighbour. Their
// difference is therefore always diagonal or two cells apart — (±1,±1),
// (0,±2) or (±2,0) — never a unit vector, so fromCoord always answers UNSET
// and the comparator always returns 4.
//
// A constant positive comparator makes Arrays.sort see one ascending run of
// the full length and leave the array untouched, so the effective enqueue
// order is Grid.ADJACENCY's: NORTH, EAST, SOUTH, WEST. That is the tie-break
// between equal-length paths. javaListSort is used rather than that
// conclusion so the behaviour holds even if a future grid hands the sort a
// neighbour set this argument does not cover.
func compareTrainBFSNeighbours(a, b Coord) int {
	return int(DirectionFromCoord(b.Sub(a)))
}
