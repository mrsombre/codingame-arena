// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:15-27

public static final Coord[] ADJACENCY = new Coord[] {
    Direction.NORTH.coord, Direction.EAST.coord,
    Direction.SOUTH.coord, Direction.WEST.coord };

public static final Coord[] ADJACENCY_8 = new Coord[] {
    Direction.NORTH.coord, Direction.EAST.coord,
    Direction.SOUTH.coord, Direction.WEST.coord,
    new Coord(-1, -1), new Coord(1, 1),
    new Coord(1, -1), new Coord(-1, 1) };
*/

// Adjacency4 is the 4-directional neighbor deltas.
var Adjacency4 = [4]Coord{
	DirNorth.Coord(),
	DirEast.Coord(),
	DirSouth.Coord(),
	DirWest.Coord(),
}

// Adjacency8 is the 8-directional neighbor deltas.
var Adjacency8 = [8]Coord{
	{X: 0, Y: -1},
	{X: 1, Y: 0},
	{X: 0, Y: 1},
	{X: -1, Y: 0},
	{X: -1, Y: -1},
	{X: 1, Y: 1},
	{X: 1, Y: -1},
	{X: -1, Y: 1},
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:29-54

public int width, height;
public LinkedHashMap<Coord, Tile> cells;
boolean ySymetry;
public List<Coord> spawns;
public List<Coord> apples;

public Grid(int width, int height) { this(width, height, false); }

public Grid(int width, int height, boolean ySymetry) {
    this.width = width;
    this.height = height;
    this.ySymetry = ySymetry;
    spawns = new ArrayList<>();
    apples = new ArrayList<>();
    cells = new LinkedHashMap<>();
    for (int y = 0; y < height; ++y) {
        for (int x = 0; x < width; ++x) {
            cells.put(new Coord(x, y), new Tile(new Coord(x, y)));
        }
    }
}
*/

// Grid holds the 2D tile map and game objects.
type Grid struct {
	Width     int
	Height    int
	YSymmetry bool
	Spawns    []Coord
	Apples    []Coord
	Cells     []Tile // row-major: Cells[y*Width+x]
}

// NewGrid creates a grid with all tiles initialized as empty.
func NewGrid(width, height int) *Grid {
	return NewGridSym(width, height, false)
}

// NewGridSym creates a grid with optional Y-axis symmetry.
func NewGridSym(width, height int, ySymmetry bool) *Grid {
	g := &Grid{
		Width:     width,
		Height:    height,
		YSymmetry: ySymmetry,
		Cells:     make([]Tile, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: y}
			g.Cells[y*width+x] = NewTile(c)
		}
	}
	return g
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:56-58

public Tile get(int x, int y) {
    return cells.getOrDefault(new Coord(x, y), Tile.NO_TILE);
}
*/

// Get returns the tile at coord c, or nil if out of bounds.
func (g *Grid) Get(c Coord) *Tile {
	return g.GetXY(c.X, c.Y)
}

// GetXY returns the tile at (x, y), or nil if out of bounds.
// Tile.IsValid/IsWall/IsEmpty tolerate nil receivers.
func (g *Grid) GetXY(x, y int) *Tile {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return nil
	}
	return &g.Cells[y*g.Width+x]
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:96-98

public List<Coord> getCoords() {
    return cells.keySet().stream().toList();
}
*/

// Coords returns all coordinates in Java LinkedHashMap insertion order.
func (g *Grid) Coords() []Coord {
	coords := make([]Coord, 0, len(g.Cells))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			coords = append(coords, Coord{X: x, Y: y})
		}
	}
	return coords
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:79-94

public <T extends Coord> List<T> getClosestTargets(Coord from, List<T> targets) {
    List<T> closest = new ArrayList<>();
    int closestBy = 0;
    for (T targ : targets) {
        int distance = from.manhattanTo(targ);
        if (closest.isEmpty() || closestBy > distance) {
            closest.clear();
            closest.add(targ);
            closestBy = distance;
        } else if (!closest.isEmpty() && closestBy == distance) {
            closest.add(targ);
        }
    }
    return closest;
}
*/

// ClosestTargets returns all targets with the smallest Manhattan distance.
func (g *Grid) ClosestTargets(from Coord, targets []Coord) []Coord {
	closest := make([]Coord, 0)
	closestBy := 0
	for _, target := range targets {
		distance := from.ManhattanTo(target)
		if len(closest) == 0 || closestBy > distance {
			closest = closest[:0]
			closest = append(closest, target)
			closestBy = distance
		} else if closestBy == distance {
			closest = append(closest, target)
		}
	}
	return closest
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:60-68

public List<Coord> getNeighbours(Coord pos, Coord[] adjacency) {
    List<Coord> neighs = new ArrayList<>();
    for (Coord delta : adjacency) {
        Coord n = new Coord(pos.getX() + delta.getX(), pos.getY() + delta.getY());
        if (get(n) != Tile.NO_TILE) neighs.add(n);
    }
    return neighs;
}
*/

// Neighbours returns in-bounds neighbor coords using the given deltas.
func (g *Grid) Neighbours(pos Coord, adjacency []Coord) []Coord {
	neighs := make([]Coord, 0, len(adjacency))
	for _, delta := range adjacency {
		n := pos.Add(delta)
		if g.Get(n).IsValid() {
			neighs = append(neighs, n)
		}
	}
	return neighs
}

// Neighbours4 returns the 4-directional in-bounds neighbors.
func (g *Grid) Neighbours4(pos Coord) []Coord {
	return g.Neighbours(pos, Adjacency4[:])
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:100-102

public Coord opposite(Coord c) {
    return new Coord(width - c.x - 1, ySymetry ? (height - c.y - 1) : c.y);
}
*/

// Opposite returns the point-symmetric coord.
// X is always mirrored; Y is mirrored only if YSymmetry is set.
func (g *Grid) Opposite(c Coord) Coord {
	y := c.Y
	if g.YSymmetry {
		y = g.Height - c.Y - 1
	}
	return Coord{X: g.Width - c.X - 1, Y: y}
}

// HasApple returns true if there is an apple at coord c.
func (g *Grid) HasApple(c Coord) bool {
	for _, a := range g.Apples {
		if a == c {
			return true
		}
	}
	return false
}

// RemoveApple removes the first apple at coord c.
func (g *Grid) RemoveApple(c Coord) {
	for i, a := range g.Apples {
		if a == c {
			g.Apples = append(g.Apples[:i], g.Apples[i+1:]...)
			return
		}
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:147-181

public List<Set<Coord>> detectAirPockets() {
    // BFS flood-fill: group all non-wall cells into connected islands.
    // Islands with size < 10 will be filled in by GridMaker.
    for (Coord p : cells.keySet()) {
        if (curCell.getType() == Tile.TYPE_WALL) { computed.add(p); continue; }
        if (!computed.contains(p)) {
            // BFS: expand from p over non-wall cells
            islands.add(island); current.clear();
        }
    }
    return islands;
}
*/

// DetectAirPockets finds all non-wall connected components via BFS.
func (g *Grid) DetectAirPockets() []map[Coord]struct{} {
	var islands []map[Coord]struct{}
	computed := make(map[Coord]struct{})

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			p := Coord{X: x, Y: y}
			tile := g.Get(p)
			if tile.IsWall() {
				computed[p] = struct{}{}
				continue
			}
			if _, done := computed[p]; done {
				continue
			}

			island := make(map[Coord]struct{})
			fifo := []Coord{p}
			computed[p] = struct{}{}

			for len(fifo) > 0 {
				e := fifo[0]
				fifo = fifo[1:]
				island[e] = struct{}{}
				for _, delta := range Adjacency4 {
					n := e.Add(delta)
					cell := g.Get(n)
					if cell.IsValid() && !cell.IsWall() {
						if _, done := computed[n]; !done {
							fifo = append(fifo, n)
							computed[n] = struct{}{}
						}
					}
				}
			}
			islands = append(islands, island)
		}
	}
	return islands
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:183-212

public List<Set<Coord>> detectSpawnIslands() {
    for (Coord p : spawns) {
        if (!computed.contains(p)) {
            // BFS: expand from spawn p to adjacent spawns only
            islands.add(island); current.clear();
        }
    }
    return islands;
}
*/

// DetectSpawnIslands finds connected groups among spawn points via BFS.
func (g *Grid) DetectSpawnIslands() []map[Coord]struct{} {
	spawnSet := make(map[Coord]struct{}, len(g.Spawns))
	for _, s := range g.Spawns {
		spawnSet[s] = struct{}{}
	}

	var islands []map[Coord]struct{}
	computed := make(map[Coord]struct{})

	for _, p := range g.Spawns {
		if _, done := computed[p]; done {
			continue
		}

		island := make(map[Coord]struct{})
		fifo := []Coord{p}
		computed[p] = struct{}{}

		for len(fifo) > 0 {
			e := fifo[0]
			fifo = fifo[1:]
			island[e] = struct{}{}
			for _, delta := range Adjacency4 {
				n := e.Add(delta)
				cell := g.Get(n)
				if _, isSpawn := spawnSet[n]; cell.IsValid() && isSpawn {
					if _, done := computed[n]; !done {
						fifo = append(fifo, n)
						computed[n] = struct{}{}
					}
				}
			}
		}
		islands = append(islands, island)
	}
	return islands
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java:214-240

public List<Coord> detectLowestIsland() {
    Coord start = new Coord(0, height - 1);
    if (get(start).getType() != Tile.TYPE_WALL) return Collections.emptyList();
    // BFS flood-fill wall tiles from bottom-left corner
    while (!fifo.isEmpty()) {
        Coord e = fifo.poll();
        for (Coord delta : ADJACENCY) {
            Coord n = e.add(delta);
            if (cell.isValid() && !computed.contains(n) && cell.getType() == Tile.TYPE_WALL) {
                fifo.add(n); lowest.add(n);
            }
        }
    }
    return lowest;
}
*/

// DetectLowestIsland flood-fills wall tiles from the bottom-left corner.
func (g *Grid) DetectLowestIsland() []Coord {
	start := Coord{X: 0, Y: g.Height - 1}
	if !g.Get(start).IsWall() {
		return nil
	}

	computed := make(map[Coord]struct{})
	fifo := []Coord{start}
	computed[start] = struct{}{}
	lowest := []Coord{start}

	for len(fifo) > 0 {
		e := fifo[0]
		fifo = fifo[1:]
		for _, delta := range Adjacency4 {
			n := e.Add(delta)
			cell := g.Get(n)
			if cell.IsWall() {
				if _, done := computed[n]; !done {
					fifo = append(fifo, n)
					computed[n] = struct{}{}
					lowest = append(lowest, n)
				}
			}
		}
	}
	return lowest
}
