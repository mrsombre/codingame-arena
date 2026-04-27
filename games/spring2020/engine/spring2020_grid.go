// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:10-26

public class Grid {
    int width, height;
    Map<Coord, Cell> cells;

    public Grid(int width, int height) {
        this.width = width;
        this.height = height;
        cells = new LinkedHashMap<>();
        for (int y = 0; y < height; ++y) {
            for (int x = 0; x < width; ++x) {
                cells.put(new Coord(x, y), new Cell());
            }
        }
    }
}
*/

// Grid holds a 2D grid of cells.
// Wraps horizontally when MapWraps is true.
type Grid struct {
	Width    int
	Height   int
	MapWraps bool
	Cells    []*Cell
}

func NewGrid(width, height int, mapWraps bool) *Grid {
	g := &Grid{
		Width:    width,
		Height:   height,
		MapWraps: mapWraps,
		Cells:    make([]*Cell, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g.Cells[y*width+x] = NewCell(Coord{X: x, Y: y})
		}
	}
	return g
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:28-43

public Grid(String[] rows) {
    this.width = rows[0].length();
    this.height = rows.length;
    cells = new LinkedHashMap<>();
    for (int y = 0; y < height; ++y) {
        for (int x = 0; x < width; ++x) {
            char cellChar = rows[y].charAt(x);
            CellType type = getCellTypeFromChar(cellChar);
            Cell cell = new Cell(type);
            cell.setHasPellet(cellHasPellet(cellChar));
            cells.put(new Coord(x, y), cell);
        }
    }
}
*/

// NewGridFromRows builds a grid from char rows, matching Java's Grid(String[]).
func NewGridFromRows(rows []string, mapWraps bool) *Grid {
	height := len(rows)
	width := len(rows[0])
	g := NewGrid(width, height, mapWraps)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := rows[y][x]
			g.Cells[y*width+x].Type = CellTypeFromChar(c)
			g.Cells[y*width+x].HasPellet = CellHasPelletFromChar(c)
		}
	}
	return g
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:45-71

private CellType getCellTypeFromChar(char cellChar) {
    switch (cellChar) {
    case '#': case 'x':           return CellType.WALL;
    case ' ': case '.': case 'o': return CellType.FLOOR;
    default: throw new RuntimeException("Unrecognised cell type: " + cellChar);
    }
}

private boolean cellHasPellet(char cellChar) {
    switch (cellChar) {
    case '#': case 'x': case ' ': return false;
    case '.': case 'o':           return true;
    default: throw new RuntimeException("Unrecognised cell type: " + cellChar);
    }
}
*/

func CellTypeFromChar(c byte) CellType {
	switch c {
	case '#', 'x':
		return CellWall
	case ' ', '.', 'o':
		return CellFloor
	}
	panic("unrecognised cell type: " + string(c))
}

func CellHasPelletFromChar(c byte) bool {
	switch c {
	case '#', 'x', ' ':
		return false
	case '.', 'o':
		return true
	}
	panic("unrecognised cell type: " + string(c))
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:73-79

public Cell get(Coord coord) { return get(coord.x, coord.y); }
public Cell get(int x, int y) {
    return cells.getOrDefault(new Coord(x, y), Cell.NO_CELL);
}
*/

// Get returns the cell at coord, or nil if out-of-bounds.
func (g *Grid) Get(c Coord) *Cell {
	return g.GetXY(c.X, c.Y)
}

// GetXY returns the cell at (x, y), or nil if out-of-bounds.
// Cell.IsValid/IsFloor/IsWall tolerate nil receivers.
func (g *Grid) GetXY(x, y int) *Cell {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return nil
	}
	return g.Cells[y*g.Width+x]
}

// Coords returns all coords in insertion order (y-major, x-minor).
// Mirrors LinkedHashMap iteration order in the Java grid.
func (g *Grid) Coords() []Coord {
	coords := make([]Coord, 0, g.Width*g.Height)
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			coords = append(coords, Coord{X: x, Y: y})
		}
	}
	return coords
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:90-100

public Optional<Coord> getCoordNeighbour(Coord pos, Coord delta) {
    Coord n = pos.add(delta);
    if (Config.MAP_WRAPS) {
        n = new Coord((n.x + width) % width, n.y);
    }
    if (get(n) != Cell.NO_CELL) {
        return Optional.of(n);
    }
    return Optional.empty();
}
*/

// CoordNeighbour returns the neighbour of pos shifted by delta.
// If the grid wraps horizontally, the X coordinate is taken modulo width.
// The returned bool is false if the cell is out of bounds.
func (g *Grid) CoordNeighbour(pos, delta Coord) (Coord, bool) {
	n := pos.Add(delta)
	if g.MapWraps {
		n = Coord{X: ((n.X % g.Width) + g.Width) % g.Width, Y: n.Y}
	}
	if g.Get(n).IsValid() {
		return n, true
	}
	return Coord{}, false
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:81-88

public List<Coord> getNeighbours(Coord pos) {
    return Arrays
        .stream(Config.ADJACENCY)
        .map(delta -> getCoordNeighbour(pos, delta))
        .filter(Optional::isPresent)
        .map(Optional::get)
        .collect(Collectors.toList());
}
*/

// Neighbours returns all 4-directional valid neighbours.
func (g *Grid) Neighbours(pos Coord) []Coord {
	out := make([]Coord, 0, 4)
	for _, delta := range ADJACENCY {
		if n, ok := g.CoordNeighbour(pos, delta); ok {
			out = append(out, n)
		}
	}
	return out
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:102-109

public int calculateDistance(Coord a, Coord b) {
    int dv = Math.abs(a.y - b.y);
    int dh = Math.min(Math.abs(a.x - b.x),
        Math.min(a.x + width - b.x, b.x + width - a.x));
    return dv + dh;
}
*/

// CalculateDistance returns the wrap-aware manhattan distance.
func (g *Grid) CalculateDistance(a, b Coord) int {
	dv := Abs(a.Y - b.Y)
	dh := Abs(a.X - b.X)
	if w1 := a.X + g.Width - b.X; w1 < dh {
		dh = w1
	}
	if w2 := b.X + g.Width - a.X; w2 < dh {
		dh = w2
	}
	return dv + dh
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java:123-129

public List<Coord> getAllPellets() {
    return cells.entrySet().stream().filter(e -> e.getValue().hasPellet()).map(e -> e.getKey()).collect(Collectors.toList());
}
public List<Coord> getAllCherries() {
    return cells.entrySet().stream().filter(e -> e.getValue().hasCherry()).map(e -> e.getKey()).collect(Collectors.toList());
}
*/

// AllPellets returns all coords with pellets in insertion order.
func (g *Grid) AllPellets() []Coord {
	out := make([]Coord, 0)
	for _, c := range g.Cells {
		if c.HasPellet {
			out = append(out, c.Coord)
		}
	}
	return out
}

// AllCherries returns all coords with cherries in insertion order.
func (g *Grid) AllCherries() []Coord {
	out := make([]Coord, 0)
	for _, c := range g.Cells {
		if c.HasCherry {
			out = append(out, c.Coord)
		}
	}
	return out
}
