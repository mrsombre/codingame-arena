// Package grid
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Grid.java
package grid

// Adjacency4 is the 4-directional neighbor deltas: left, right, up, down.
// Order matches Java Config.ADJACENCY: {-1,0}, {1,0}, {0,-1}, {0,1}.
var Adjacency4 = [4]Coord{
	{X: -1, Y: 0},
	{X: 1, Y: 0},
	{X: 0, Y: -1},
	{X: 0, Y: 1},
}

// Grid holds a 2D grid of cells.
// Wraps horizontally when MapWraps is true.
type Grid struct {
	Width    int
	Height   int
	MapWraps bool
	cells    []*Cell
}

func NewGrid(width, height int, mapWraps bool) *Grid {
	g := &Grid{
		Width:    width,
		Height:   height,
		MapWraps: mapWraps,
		cells:    make([]*Cell, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g.cells[y*width+x] = NewCell(Coord{X: x, Y: y})
		}
	}
	return g
}

// NewGridFromRows builds a grid from char rows, matching Java's Grid(String[]).
func NewGridFromRows(rows []string, mapWraps bool) *Grid {
	height := len(rows)
	width := len(rows[0])
	g := NewGrid(width, height, mapWraps)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := rows[y][x]
			g.cells[y*width+x].Type = cellTypeFromChar(c)
			g.cells[y*width+x].HasPellet = cellHasPelletFromChar(c)
		}
	}
	return g
}

func cellTypeFromChar(c byte) CellType {
	switch c {
	case '#', 'x':
		return CellWall
	case ' ', '.', 'o':
		return CellFloor
	}
	panic("unrecognised cell type: " + string(c))
}

func cellHasPelletFromChar(c byte) bool {
	switch c {
	case '#', 'x', ' ':
		return false
	case '.', 'o':
		return true
	}
	panic("unrecognised cell type: " + string(c))
}

// Get returns a pointer to the cell at coord. Returns &NoCell for out-of-bounds.
func (g *Grid) Get(c Coord) *Cell {
	return g.GetXY(c.X, c.Y)
}

// GetXY returns a pointer to the cell at (x, y). Returns &NoCell for out-of-bounds.
func (g *Grid) GetXY(x, y int) *Cell {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		no := NoCell
		return &no
	}
	return g.cells[y*g.Width+x]
}

// Coords returns all coords in insertion order (y-major, x-minor).
func (g *Grid) Coords() []Coord {
	coords := make([]Coord, 0, g.Width*g.Height)
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			coords = append(coords, Coord{X: x, Y: y})
		}
	}
	return coords
}

// Cells returns all cells in insertion order (y-major, x-minor).
func (g *Grid) Cells() []*Cell {
	return g.cells
}

// GetCoordNeighbour returns the neighbour of pos shifted by delta.
// If the grid wraps horizontally, the X coordinate is taken modulo width.
// The returned bool is false if the cell is out of bounds.
func (g *Grid) GetCoordNeighbour(pos, delta Coord) (Coord, bool) {
	n := pos.Add(delta)
	if g.MapWraps {
		n = Coord{X: ((n.X % g.Width) + g.Width) % g.Width, Y: n.Y}
	}
	if g.Get(n).IsValid() {
		return n, true
	}
	return Coord{}, false
}

// Neighbours returns all 4-directional valid neighbours.
func (g *Grid) Neighbours(pos Coord) []Coord {
	out := make([]Coord, 0, 4)
	for _, delta := range Adjacency4 {
		if n, ok := g.GetCoordNeighbour(pos, delta); ok {
			out = append(out, n)
		}
	}
	return out
}

// CalculateDistance returns the wrap-aware manhattan distance.
func (g *Grid) CalculateDistance(a, b Coord) int {
	dv := abs(a.Y - b.Y)
	dh := abs(a.X - b.X)
	if w1 := a.X + g.Width - b.X; w1 < dh {
		dh = w1
	}
	if w2 := b.X + g.Width - a.X; w2 < dh {
		dh = w2
	}
	return dv + dh
}

// AllPellets returns all coords with pellets in insertion order.
func (g *Grid) AllPellets() []Coord {
	out := make([]Coord, 0)
	for _, c := range g.cells {
		if c.HasPellet {
			out = append(out, c.Coord)
		}
	}
	return out
}

// AllCherries returns all coords with cherries in insertion order.
func (g *Grid) AllCherries() []Coord {
	out := make([]Coord, 0)
	for _, c := range g.cells {
		if c.HasCherry {
			out = append(out, c.Coord)
		}
	}
	return out
}
