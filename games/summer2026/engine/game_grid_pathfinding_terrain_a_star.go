// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TerrainAStar.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/TerrainAStar.java:9-48

public class TerrainAStar extends AbstractAStar<Coord> {
    Grid grid;
    Town from, to;

    protected Coord getInitialState() { return from.coord; }
    protected double heuristic(Coord state) { return state.manhattanTo(to.coord); }
    protected double cost(Coord from, Coord to) { return 1; }
    protected List<Coord> getSuccessors(Coord state) {
        return grid.getNeighbours(state).stream()
            .filter(coord -> !grid.zones.get(grid.get(coord).zoneId).inked)
            .toList();
    }
    protected boolean isGoal(Coord state) { return state.equals(to.coord); }
}
*/

// TerrainAStar answers "could a route between these two towns ever exist",
// which is what the early game-over check asks. It walks terrain, not track:
// rivers and mountains cost paint but are traversable, so the only thing that
// can cut a route is an inked region, which no rail can ever be laid in
// again.
//
// It inherits AbstractAStar's zero tie-break, so among equal-f nodes the
// earlier-queued one wins.
type TerrainAStar struct {
	Grid *Grid
	From *Town
	To   *Town
}

func NewTerrainAStar(grid *Grid, from, to *Town) *TerrainAStar {
	return &TerrainAStar{Grid: grid, From: from, To: to}
}

func (a *TerrainAStar) InitialState() Coord { return a.From.Coord }

func (a *TerrainAStar) IsGoal(state Coord) bool { return state == a.To.Coord }

func (a *TerrainAStar) Heuristic(state Coord) float64 {
	return float64(state.ManhattanTo(a.To.Coord))
}

func (a *TerrainAStar) Cost(_, _ Coord) float64 { return 1 }

// Successors keeps every in-bounds neighbour whose region is not inked. The
// origin town's own cell is never re-tested, so a search starting inside an
// inked region still leaves it.
func (a *TerrainAStar) Successors(state Coord) []Coord {
	neighbours := a.Grid.Neighbours(state)
	out := make([]Coord, 0, len(neighbours))
	for _, coord := range neighbours {
		if !a.Grid.Zones[a.Grid.Get(coord).ZoneID].Inked {
			out = append(out, coord)
		}
	}
	return out
}

func (a *TerrainAStar) TieBreaker(_, _ Coord) int { return 0 }

func (a *TerrainAStar) StateKey(state Coord) Coord { return state }

// Search returns the cells from the origin town to the destination town
// inclusive, and false when the two are separated by inked regions.
func (a *TerrainAStar) Search() ([]Coord, bool) {
	return AStarSearch[Coord, Coord](a)
}
