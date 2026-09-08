// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainAStar.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TrainAStar.java:9-53

public class TrainAStar extends AbstractAStar<TrainState> {
    Grid grid;
    Town from, to;

    protected TrainState getInitialState() { return new TrainState(from.coord); }
    protected double heuristic(TrainState state) { return state.coord.manhattanTo(to.coord); }
    protected double cost(TrainState from, TrainState to) { return 1; }
    protected List<TrainState> getSuccessors(TrainState state) {
        return grid.getNeighbours(state.coord).stream()
            .filter(coord -> grid.canTrainPass(coord))
            .map(TrainState::new)
            .toList();
    }
    protected boolean isGoal(TrainState state) { return state.coord.equals(to.coord); }
    protected int tieBreaker(TrainState from, TrainState to) {
        return Direction.fromCoord(to.coord.sub(from.coord)).ordinal();
    }
}
*/

// TrainAStar finds a route between two towns over track, the same problem
// TrainBFS solves. Nothing in the game source calls it — moveTrains uses
// TrainBFS — so it is ported for completeness and is not on any scoring path.
//
// Its tie-break is the real one: for a step between adjacent cells the
// direction ordinal is NORTH 0, EAST 1, SOUTH 2, WEST 3, so among equal-f
// nodes the one reached by heading north wins.
type TrainAStar struct {
	Grid *Grid
	From *Town
	To   *Town
}

func NewTrainAStar(grid *Grid, from, to *Town) *TrainAStar {
	return &TrainAStar{Grid: grid, From: from, To: to}
}

func (a *TrainAStar) InitialState() *TrainState { return NewTrainState(a.From.Coord) }

func (a *TrainAStar) IsGoal(state *TrainState) bool { return state.Coord == a.To.Coord }

func (a *TrainAStar) Heuristic(state *TrainState) float64 {
	return float64(state.Coord.ManhattanTo(a.To.Coord))
}

func (a *TrainAStar) Cost(_, _ *TrainState) float64 { return 1 }

func (a *TrainAStar) Successors(state *TrainState) []*TrainState {
	neighbours := a.Grid.Neighbours(state.Coord)
	out := make([]*TrainState, 0, len(neighbours))
	for _, coord := range neighbours {
		if a.Grid.CanTrainPass(coord) {
			out = append(out, NewTrainState(coord))
		}
	}
	return out
}

func (a *TrainAStar) TieBreaker(from, to *TrainState) int {
	return int(DirectionFromCoord(to.Coord.Sub(from.Coord)))
}

// StateKey is TrainState.equals/hashCode: the coord alone.
func (a *TrainAStar) StateKey(state *TrainState) Coord { return state.Coord }

// Search returns the cells from the origin town to the destination town
// inclusive, and false when no route over track exists.
func (a *TrainAStar) Search() ([]Coord, bool) {
	states, ok := AStarSearch[*TrainState, Coord](a)
	if !ok {
		return nil, false
	}
	path := make([]Coord, len(states))
	for i, s := range states {
		path[i] = s.Coord
	}
	return path, true
}
