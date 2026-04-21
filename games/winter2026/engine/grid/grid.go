// Package grid
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Grid.java
package grid

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

// Grid holds the 2D tile map and game objects.
type Grid struct {
	Width     int
	Height    int
	YSymmetry bool
	Spawns    []Coord
	Apples    []Coord
	cells     []Tile // row-major: cells[y*Width+x]
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
		cells:     make([]Tile, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: y}
			g.cells[y*width+x] = NewTile(c)
		}
	}
	return g
}

// Get returns a pointer to the tile at coord c.
// Returns a pointer to NoTile if out of bounds.
func (g *Grid) Get(c Coord) *Tile {
	return g.GetXY(c.X, c.Y)
}

// GetXY returns a pointer to the tile at (x, y).
// Returns a pointer to NoTile if out of bounds.
func (g *Grid) GetXY(x, y int) *Tile {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		noTile := NoTile
		return &noTile
	}
	return &g.cells[y*g.Width+x]
}

// Coords returns all coordinates in Java LinkedHashMap insertion order.
func (g *Grid) Coords() []Coord {
	coords := make([]Coord, 0, len(g.cells))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			coords = append(coords, Coord{X: x, Y: y})
		}
	}
	return coords
}

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

// DetectAirPockets finds all non-wall connected components via BFS.
func (g *Grid) DetectAirPockets() []map[Coord]struct{} {
	var islands []map[Coord]struct{}
	computed := make(map[Coord]struct{})

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			p := Coord{X: x, Y: y}
			tile := g.Get(p)
			if tile.Type == TileWall {
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
					if cell.IsValid() && cell.Type != TileWall {
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

// DetectLowestIsland flood-fills wall tiles from the bottom-left corner.
func (g *Grid) DetectLowestIsland() []Coord {
	start := Coord{X: 0, Y: g.Height - 1}
	if g.Get(start).Type != TileWall {
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
			if cell.IsValid() && cell.Type == TileWall {
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
