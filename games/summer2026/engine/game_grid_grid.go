// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:8-19

public static final Coord[] ADJACENCY = new Coord[] { Direction.NORTH.coord, Direction.EAST.coord, Direction.SOUTH.coord, Direction.WEST.coord };

public static final Coord[] ADJACENCY_8 = new Coord[] {
    Direction.NORTH.coord, Direction.EAST.coord, Direction.SOUTH.coord, Direction.WEST.coord,
    new Coord(-1, -1), new Coord(1, 1), new Coord(1, -1), new Coord(-1, 1)
};
*/

// ADJACENCY is the N/E/S/W scan order every neighbour walk uses. Path
// tie-breaks fall out of it, so the order is load-bearing.
var ADJACENCY = []Coord{
	NORTH.Coord(), EAST.Coord(), SOUTH.Coord(), WEST.Coord(),
}

var ADJACENCY_8 = []Coord{
	NORTH.Coord(), EAST.Coord(), SOUTH.Coord(), WEST.Coord(),
	{-1, -1}, {1, 1}, {1, -1}, {-1, 1},
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:21-49

public int width, height;
public LinkedHashMap<Coord, Tile> cells;
boolean ySymetry;
public List<Town> towns;
public List<Coord> rails;
public List<Zone> zones;
private Coord poi;

public Grid(int width, int height, boolean ySymetry) {
    ...
    cells = new LinkedHashMap<>();
    for (int y = 0; y < height; ++y)
        for (int x = 0; x < width; ++x)
            cells.put(new Coord(x, y), new Tile(new Coord(x, y)));
}
*/

// Grid holds the map. Java's row-major LinkedHashMap<Coord, Tile> becomes a
// flat row-major slice indexed Cells[y*Width+x]; iteration order is
// identical and lookups are O(1) without hashing.
type Grid struct {
	Width, Height int
	Cells         []*Tile
	YSymetry      bool
	Towns         []*Town
	Rails         []Coord
	Zones         []*Zone

	POI    Coord
	HasPOI bool
}

func NewGrid(width, height int) *Grid {
	return NewGridSym(width, height, false)
}

func NewGridSym(width, height int, ySymetry bool) *Grid {
	g := &Grid{
		Width:    width,
		Height:   height,
		YSymetry: ySymetry,
		Cells:    make([]*Tile, width*height),
		Rails:    []Coord{},
		Towns:    []*Town{},
		Zones:    []*Zone{},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g.Cells[y*width+x] = NewTile(Coord{x, y})
		}
	}
	return g
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:51-63

public Grid clone() {
    Grid newGrid = new Grid(width, height, ySymetry);
    for (int y = 0; ...) for (int x = 0; ...) {
        Tile newTile = new Tile(coord, tile.getType());
        newTile.setZoneId(tile.getZoneId());
        newGrid.cells.put(coord, newTile);
    }
    return newGrid;
}
*/

// Clone copies terrain and region ids only — tracks, towns and zones are
// intentionally left out, as in Java.
func (g *Grid) Clone() *Grid {
	out := NewGridSym(g.Width, g.Height, g.YSymetry)
	for i, tile := range g.Cells {
		out.Cells[i].Type = tile.Type
		out.Cells[i].ZoneID = tile.ZoneID
	}
	return out
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:65-67,84-86

public Tile get(int x, int y) { return cells.getOrDefault(new Coord(x, y), Tile.NO_TILE); }
public Tile get(Coord n) { return get(n.getX(), n.getY()); }
*/

// GetXY returns nil out of bounds. Every Tile predicate is nil-tolerant, so
// callers reading state behave as they did against Java's NO_TILE; callers
// writing state must guard with IsValid first or they will panic.
func (g *Grid) GetXY(x, y int) *Tile {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return nil
	}
	return g.Cells[y*g.Width+x]
}

func (g *Grid) Get(c Coord) *Tile { return g.GetXY(c.X, c.Y) }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:69-82

public List<Coord> getNeighbours(Coord pos, Coord[] adjacency) {
    List<Coord> neighs = new ArrayList<>();
    for (Coord delta : adjacency) {
        Coord n = new Coord(pos.getX() + delta.getX(), pos.getY() + delta.getY());
        if (get(n) != Tile.NO_TILE) neighs.add(n);
    }
    return neighs;
}

public List<Coord> getNeighbours(Coord pos) { return getNeighbours(pos, ADJACENCY); }
*/

func (g *Grid) NeighboursWith(pos Coord, adjacency []Coord) []Coord {
	neighs := make([]Coord, 0, len(adjacency))
	for _, delta := range adjacency {
		n := pos.Add(delta)
		if g.Get(n).IsValid() {
			neighs = append(neighs, n)
		}
	}
	return neighs
}

// Neighbours walks NORTH, EAST, SOUTH, WEST in that order, skipping cells
// off the grid.
func (g *Grid) Neighbours(pos Coord) []Coord {
	return g.NeighboursWith(pos, ADJACENCY)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:88-104

public <T extends Positionable> List<T> getClosestTargets(Coord from, List<T> targets) {
    List<T> closest = new ArrayList<>();
    int closestBy = 0;
    for (T targ : targets) {
        int distance = (int) (from.manhattanTo(targ.getPosition()) * targ.getDistanceMultiplier());
        if (closest.isEmpty() || closestBy > distance) { closest.clear(); closest.add(targ); closestBy = distance; }
        else if (closestBy == distance) closest.add(targ);
    }
    return closest;
}
*/

func ClosestTargets[T Positionable](from Coord, targets []T) []T {
	closest := make([]T, 0, len(targets))
	closestBy := 0
	for _, targ := range targets {
		distance := int(float64(from.ManhattanTo(targ.Position())) * targ.DistanceMultiplier())
		switch {
		case len(closest) == 0 || closestBy > distance:
			closest = closest[:0]
			closest = append(closest, targ)
			closestBy = distance
		case closestBy == distance:
			closest = append(closest, targ)
		}
	}
	return closest
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Grid.java:106-124

public List<Coord> getCoords() { return cells.keySet().stream().toList(); }
public Coord opposite(Coord c) { return new Coord(width - c.x - 1, ySymetry ? (height - c.y - 1) : c.y); }
public boolean canTrainPass(Coord coord) {
    Tile tile = get(coord);
    if (!tile.isValid()) return false;
    return tile.isTown() || tile.track != Tile.TRACK_NONE;
}
*/

// Coords lists every cell in row-major order, matching LinkedHashMap key
// iteration.
func (g *Grid) Coords() []Coord {
	out := make([]Coord, 0, len(g.Cells))
	for _, tile := range g.Cells {
		out = append(out, tile.Coord)
	}
	return out
}

func (g *Grid) Opposite(c Coord) Coord {
	y := c.Y
	if g.YSymetry {
		y = g.Height - c.Y - 1
	}
	return Coord{g.Width - c.X - 1, y}
}

func (g *Grid) CanTrainPass(coord Coord) bool {
	tile := g.Get(coord)
	if !tile.IsValid() {
		return false
	}
	return tile.IsTown() || tile.Track != TRACK_NONE
}

func (g *Grid) SetPOI(poi Coord) {
	g.POI = poi
	g.HasPOI = true
}
