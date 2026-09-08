// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:21-35

public class AutobuildAStar extends AbstractAStar<AutobuildState> {
    Grid grid;
    Player player;
    Coord from, to;
    List<Zone> zones;

    public AutobuildAStar(Grid grid, Player player, Coord from, Coord to) {
        this.grid = grid;
        this.player = player;
        this.from = from;
        this.to = to;
        this.zones = grid.zones;
    }
}
*/

// AutobuildAStar plans the track sequence an AUTOPLACE expands to: the
// cheapest way to get a rail block containing From to touch To.
//
// Player is stored and never read, exactly as in Java — the plan does not
// depend on who asked for it, and in particular does not consider whether the
// player can afford it. Budget is enforced later, when Game.DoActions walks
// the expanded placements and interrupts on the first unaffordable one.
type AutobuildAStar struct {
	Grid   *Grid
	Player *Player
	From   Coord
	To     Coord
	Zones  []*Zone
}

func NewAutobuildAStar(grid *Grid, player *Player, from, to Coord) *AutobuildAStar {
	return &AutobuildAStar{
		Grid:   grid,
		Player: player,
		From:   from,
		To:     to,
		Zones:  grid.Zones,
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:37-62

protected AutobuildState getInitialState() {
    AutobuildState s = new AutobuildState();
    s.action = null;
    s.ownedZones = new HashSet<>();
    s.moneySpent = 0;
    s.tracks = new HashSet<>();
    s.cursor = from;
    Tile t = grid.get(from);
    if (!t.isValid()) {
        s.cursorState = CursorState.BROKEN;
        return s;
    }
    Zone zone = zones.get(t.zoneId);
    boolean usableTrack = t.isTrackOrTown();
    s.cursorState = usableTrack ? CursorState.BUILT : CursorState.BUILDABLE;
    return s;
}
*/

// InitialState starts the cursor on From. An off-grid From yields BROKEN,
// which Successors turns into a dead end rather than an error.
//
// Java's unused `Zone zone = zones.get(t.zoneId)` is dropped: it can only
// change behaviour by throwing on a cell with no region, which the generator
// never produces.
func (a *AutobuildAStar) InitialState() *AutobuildState {
	s := &AutobuildState{
		OwnedZones: map[int]struct{}{},
		Tracks:     map[Coord]struct{}{},
		Cursor:     a.From,
	}

	t := a.Grid.Get(a.From)
	if !t.IsValid() {
		s.CursorState = CURSOR_BROKEN
		return s
	}

	if t.IsTrackOrTown() {
		s.CursorState = CURSOR_BUILT
	} else {
		s.CursorState = CURSOR_BUILDABLE
	}
	return s
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:64-67,106-114

protected int tieBreaker(AutobuildState from, AutobuildState to) {
    return Direction.fromCoord(to.cursor.sub(from.cursor)).ordinal();
}

protected double heuristic(AutobuildState state) {
    return state.cursor.manhattanTo(state.cursor) * Game.BASE_RAIL_COST * Game.GRASS_COST_MULTIPLIER;
}

protected double cost(AutobuildState from, AutobuildState to) {
    return (to.moneySpent - from.moneySpent);
}
*/

// TieBreaker orders equal-cost nodes by the direction of the step that
// reached them: NORTH 0, EAST 1, SOUTH 2, WEST 3. A jump across a rail block
// is not a single step, so it scores UNSET's 4 and loses every tie.
func (a *AutobuildAStar) TieBreaker(from, to *AutobuildState) int {
	return int(DirectionFromCoord(to.Cursor.Sub(from.Cursor)))
}

// Heuristic measures the cursor's distance to itself, so it is always zero
// and the search is a plain Dijkstra. This is upstream's bug — the intended
// `to` is missing — and it is reproduced because fixing it would change which
// of several equal-cost plans wins.
func (a *AutobuildAStar) Heuristic(state *AutobuildState) float64 {
	return float64(state.Cursor.ManhattanTo(state.Cursor) * BASE_RAIL_COST * GRASS_COST_MULTIPLIER)
}

// Cost is what the step added to the running bill: zero for moving over
// existing track, the cell's rail cost for building on it.
func (a *AutobuildAStar) Cost(from, to *AutobuildState) float64 {
	return float64(to.MoneySpent - from.MoneySpent)
}

func (a *AutobuildAStar) StateKey(state *AutobuildState) AutobuildStateKey { return state.Key() }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:69-104

protected boolean isGoal(AutobuildState state) {
    if (state.cursor.equals(to)) return true;
    Tile t = grid.get(state.cursor);
    if (t.isTrackOrTown()) return isPartOfRailBlock(state.cursor, to);
    return false;
}

private boolean isPartOfRailBlock(Coord coord, Coord goal) {
    LinkedList<Coord> fifo = new LinkedList<>();
    Set<Coord> visited = new HashSet<>();
    fifo.add(coord);
    visited.add(coord);
    while (!fifo.isEmpty()) {
        Coord current = fifo.poll();
        if (current.equals(goal)) return true;
        List<Coord> neighbours = grid.getNeighbours(current);
        for (Coord n : neighbours) {
            Tile t = grid.get(n);
            if (!visited.contains(n) && t.isTrackOrTown()) { fifo.add(n); visited.add(n); }
        }
    }
    return false;
}
*/

// IsGoal accepts the cursor landing on To, or the cursor standing on a
// pre-existing track or town whose rail block already reaches To — so an
// AUTOPLACE that hits the far end's own network stops there rather than
// paving the last stretch again.
//
// The rail block is read off the grid, so the cells this plan intends to
// build do not count towards it.
func (a *AutobuildAStar) IsGoal(state *AutobuildState) bool {
	if state.Cursor == a.To {
		return true
	}
	if a.Grid.Get(state.Cursor).IsTrackOrTown() {
		return a.isPartOfRailBlock(state.Cursor, a.To)
	}
	return false
}

func (a *AutobuildAStar) isPartOfRailBlock(coord, goal Coord) bool {
	fifo := []Coord{coord}
	visited := map[Coord]struct{}{coord: {}}
	for len(fifo) > 0 {
		current := fifo[0]
		fifo = fifo[1:]
		if current == goal {
			return true
		}
		for _, n := range a.Grid.Neighbours(current) {
			if _, seen := visited[n]; seen {
				continue
			}
			if a.Grid.Get(n).IsTrackOrTown() {
				fifo = append(fifo, n)
				visited[n] = struct{}{}
			}
		}
	}
	return false
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:116-190

protected List<AutobuildState> getSuccessors(AutobuildState state) {
    if (state.cursorState == CursorState.BROKEN) return List.of();
    if (!grid.get(to).isValid()) return List.of();

    Set<AutobuildState> result = new LinkedHashSet<>();
    List<Coord> scope = new ArrayList<>(4);

    Tile currentTile = grid.get(state.cursor);
    boolean usableTrack = state.tracks.contains(state.cursor) || currentTile.isTrackOrTown();

    if (usableTrack) {
        if (state.tracks.contains(state.cursor)) {
            scope.addAll(grid.getNeighbours(state.cursor));
        } else {
            scope.addAll(getAllNeighboursOfRailBlock(state.cursor));
        }
    } else {
        scope.add(state.cursor);
    }

    for (Coord neigh : scope) {
        Tile t = grid.get(neigh);
        Zone zone = zones.get(t.zoneId);
        boolean isTown = t.townId != Tile.TOWN_NONE;
        boolean buildableZone = !zone.inked;
        if (state.ownedZones.contains(t.zoneId) || buildableZone || isTown) {
            if (state.tracks.contains(neigh)) continue;
            if (t.isTrackOrTown()) {
                if (neigh.equals(state.cursor)) continue;
                AutobuildState next = new AutobuildState();
                next.action = null;
                next.cursor = neigh;
                next.moneySpent = state.moneySpent;
                next.ownedZones = state.ownedZones;
                next.tracks = state.tracks;
                next.cursorState = CursorState.BUILT;
                result.add(next);
            } else {
                int railCost = Game.getRailCost(t);
                AutobuildState next = new AutobuildState();
                next.action = new Action(ActionType.PLACE_TRACK, true);
                next.action.setCoord(neigh);
                next.cursor = neigh;
                next.moneySpent = state.moneySpent + railCost;
                next.ownedZones = state.ownedZones;
                next.tracks = new HashSet<>(state.tracks);
                next.tracks.add(neigh);
                next.cursorState = CursorState.BUILT;
                result.add(next);
            }
        }
    }
    return result.stream().toList();
}
*/

// Successors expands one node. The cursor's own cell is the candidate when it
// is still bare — the plan has to pave where it stands before it can move on.
// Once the cursor stands on track it steps to a neighbour, and when that
// track was already on the grid it may jump to any cell touching the whole
// connected rail block, for free.
//
// A cell is a candidate when its region is not inked, or it holds a town —
// towns stay reachable through an inked region.
//
// Every successor carries CursorState BUILT, so the LinkedHashSet dedups
// purely by cursor and the first candidate to reach a cell is the one kept.
func (a *AutobuildAStar) Successors(state *AutobuildState) []*AutobuildState {
	// Invalid start point.
	if state.CursorState == CURSOR_BROKEN {
		return nil
	}
	// Invalid end point.
	if !a.Grid.Get(a.To).IsValid() {
		return nil
	}

	var scope []Coord
	_, cursorPlanned := state.Tracks[state.Cursor]
	switch {
	case cursorPlanned:
		// Track this plan intends to build: only its own neighbours are
		// reachable, because it is not on the grid for the block walk to find.
		scope = a.Grid.Neighbours(state.Cursor)
	case a.Grid.Get(state.Cursor).IsTrackOrTown():
		scope = a.allNeighboursOfRailBlock(state.Cursor)
	default:
		scope = []Coord{state.Cursor}
	}

	result := make([]*AutobuildState, 0, len(scope))
	seen := make(map[AutobuildStateKey]struct{}, len(scope))
	add := func(next *AutobuildState) {
		if _, dup := seen[next.Key()]; dup {
			return
		}
		seen[next.Key()] = struct{}{}
		result = append(result, next)
	}

	for _, neigh := range scope {
		t := a.Grid.Get(neigh)
		zone := a.Zones[t.ZoneID]
		isTown := t.TownID != TOWN_NONE
		_, ownsZone := state.OwnedZones[t.ZoneID]
		if !ownsZone && zone.Inked && !isTown {
			continue
		}

		// No point in looping back onto a cell this plan already builds.
		if _, planned := state.Tracks[neigh]; planned {
			continue
		}

		if t.IsTrackOrTown() {
			if neigh == state.Cursor {
				continue
			}
			add(&AutobuildState{
				Cursor:      neigh,
				MoneySpent:  state.MoneySpent,
				OwnedZones:  state.OwnedZones,
				Tracks:      state.Tracks,
				CursorState: CURSOR_BUILT,
			})
			continue
		}

		action := NewAutobuildAction(ACTION_PLACE_TRACK, true)
		action.Coord = neigh

		tracks := make(map[Coord]struct{}, len(state.Tracks)+1)
		for c := range state.Tracks {
			tracks[c] = struct{}{}
		}
		tracks[neigh] = struct{}{}

		add(&AutobuildState{
			Action:      action,
			Cursor:      neigh,
			MoneySpent:  state.MoneySpent + RailCost(t),
			OwnedZones:  state.OwnedZones,
			Tracks:      tracks,
			CursorState: CURSOR_BUILT,
		})
	}

	return result
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/pathfinding/AutobuildAStar.java:192-215

private List<Coord> getAllNeighboursOfRailBlock(Coord cursor) {
    List<Coord> neighs = new ArrayList<>();
    LinkedList<Coord> fifo = new LinkedList<>();
    fifo.add(cursor);
    Set<Coord> visited = new HashSet<>();
    visited.add(cursor);
    while (!fifo.isEmpty()) {
        Coord current = fifo.poll();
        boolean usableTrack = grid.get(current).isTrackOrTown();
        if (!usableTrack) { neighs.add(current); continue; }
        List<Coord> neighbours = grid.getNeighbours(current);
        for (Coord n : neighbours) {
            if (!visited.contains(n)) { fifo.add(n); visited.add(n); }
        }
    }
    return neighs;
}
*/

// allNeighboursOfRailBlock floods over the connected run of track and towns
// containing cursor and returns the bare cells ringing it, in breadth-first
// N/E/S/W discovery order. That order is what the LinkedHashSet dedup and the
// priority queue's tie-break both see.
func (a *AutobuildAStar) allNeighboursOfRailBlock(cursor Coord) []Coord {
	var neighs []Coord
	fifo := []Coord{cursor}
	visited := map[Coord]struct{}{cursor: {}}

	for len(fifo) > 0 {
		current := fifo[0]
		fifo = fifo[1:]
		if !a.Grid.Get(current).IsTrackOrTown() {
			neighs = append(neighs, current)
			continue
		}
		for _, n := range a.Grid.Neighbours(current) {
			if _, seen := visited[n]; seen {
				continue
			}
			fifo = append(fifo, n)
			visited[n] = struct{}{}
		}
	}
	return neighs
}

// Search returns the winning plan's states, start first, and false when no
// plan exists.
func (a *AutobuildAStar) Search() ([]*AutobuildState, bool) {
	return AStarSearch[*AutobuildState, AutobuildStateKey](a)
}
