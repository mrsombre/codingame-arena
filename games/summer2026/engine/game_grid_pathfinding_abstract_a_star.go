// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AbstractAStar.java
package engine

import "slices"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AbstractAStar.java:5-19

public abstract class AbstractAStar<T> {
    protected abstract T getInitialState();
    protected abstract boolean isGoal(T state);
    protected abstract double heuristic(T state);
    protected abstract List<T> getSuccessors(T state);
    protected abstract double cost(T from, T to);
    protected int tieBreaker(T from, T to) { return 0; }
}
*/

// AStarProblem is the Java abstract class turned inside out: Go has no
// abstract methods, so each concrete searcher supplies the same six hooks as
// an interface.
//
// StateKey has no Java counterpart. Java identifies states through
// equals/hashCode — TrainState by coord, AutobuildState by cursor plus cursor
// state — and the search's two maps are keyed by it. K is that identity made
// explicit.
type AStarProblem[S any, K comparable] interface {
	InitialState() S
	IsGoal(state S) bool
	Heuristic(state S) float64
	Successors(state S) []S
	Cost(from, to S) float64
	TieBreaker(from, to S) int
	StateKey(state S) K
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AbstractAStar.java:21-61

public Optional<List<T>> search() {
    PriorityQueue<Node> open = new PriorityQueue<>(
        Comparator.comparingDouble((Node n) -> n.f).thenComparingInt(n -> n.tieBreaker)
    );
    Map<T, Double> gScores = new LinkedHashMap<>();
    Map<T, T> cameFrom = new LinkedHashMap<>();

    T start = getInitialState();
    gScores.put(start, 0.0);
    open.add(new Node(start, heuristic(start), 0));

    while (!open.isEmpty()) {
        Node current = open.poll();
        if (isGoal(current.state)) return Optional.of(reconstructPath(cameFrom, current.state));

        for (T neighbor : getSuccessors(current.state)) {
            double tentativeG = gScores.get(current.state) + cost(current.state, neighbor);
            if (tentativeG < gScores.getOrDefault(neighbor, Double.POSITIVE_INFINITY)) {
                cameFrom.put(neighbor, current.state);
                gScores.put(neighbor, tentativeG);
                int tieBreakerValue = tieBreaker(current.state, neighbor);
                open.add(new Node(neighbor, tentativeG + heuristic(neighbor), tieBreakerValue));
            }
        }
    }
    return Optional.empty();
}

private List<T> reconstructPath(Map<T, T> cameFrom, T current) {
    List<T> path = new ArrayList<>();
    while (current != null) { path.add(current); current = cameFrom.get(current); }
    Collections.reverse(path);
    return path;
}
*/

// aStarNode is AbstractAStar.Node: the queued state with its f score and the
// tie-break value recorded when it was queued.
type aStarNode[S any] struct {
	State      S
	F          float64
	TieBreaker int
}

// AStarSearch returns the path from the initial state to the first goal state
// popped, and false when the open set drains without reaching one.
//
// The search is stale-entry tolerant only by accident: a state improved after
// being queued is queued again and the older, worse entry is simply popped
// later. That is Java's behaviour and it is preserved.
func AStarSearch[S any, K comparable](p AStarProblem[S, K]) ([]S, bool) {
	open := newJavaPriorityQueue(func(a, b aStarNode[S]) int {
		if c := javaCompareDouble(a.F, b.F); c != 0 {
			return c
		}
		return compareInt(a.TieBreaker, b.TieBreaker)
	})

	gScores := map[K]float64{}
	cameFrom := map[K]S{}

	start := p.InitialState()
	gScores[p.StateKey(start)] = 0
	open.Add(aStarNode[S]{State: start, F: p.Heuristic(start)})

	for {
		current, ok := open.Poll()
		if !ok {
			return nil, false
		}
		if p.IsGoal(current.State) {
			return aStarReconstructPath(p, cameFrom, current.State), true
		}

		currentG := gScores[p.StateKey(current.State)]
		for _, neighbor := range p.Successors(current.State) {
			key := p.StateKey(neighbor)
			tentativeG := currentG + p.Cost(current.State, neighbor)
			if best, seen := gScores[key]; seen && tentativeG >= best {
				continue
			}
			cameFrom[key] = current.State
			gScores[key] = tentativeG
			open.Add(aStarNode[S]{
				State:      neighbor,
				F:          tentativeG + p.Heuristic(neighbor),
				TieBreaker: p.TieBreaker(current.State, neighbor),
			})
		}
	}
}

// aStarReconstructPath walks the back-pointers to the start, which is the one
// state with no cameFrom entry, and returns the path start-first.
func aStarReconstructPath[S any, K comparable](p AStarProblem[S, K], cameFrom map[K]S, current S) []S {
	path := []S{current}
	for {
		prev, ok := cameFrom[p.StateKey(current)]
		if !ok {
			break
		}
		path = append(path, prev)
		current = prev
	}
	slices.Reverse(path)
	return path
}
