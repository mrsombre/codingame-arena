// Package maps
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java
package maps

import (
	"sort"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

// Generator lays out walls and floors on a Grid using shuffled tetris pieces.
type Generator struct {
	pieces []*piece
}

type piece struct {
	blocks map[grid.Coord]struct{}
	maxX   int
	maxY   int
}

func newPiece(cells []grid.Coord) *piece {
	blocks := make(map[grid.Coord]struct{}, len(cells))
	maxX, maxY := 0, 0
	for _, c := range cells {
		blocks[c] = struct{}{}
		if c.X > maxX {
			maxX = c.X
		}
		if c.Y > maxY {
			maxY = c.Y
		}
	}
	return &piece{blocks: blocks, maxX: maxX, maxY: maxY}
}

// NewGenerator constructs a Generator with all tetromino variants populated,
// matching TetrisBasedMapGenerator.init().
func NewGenerator() *Generator {
	g := &Generator{}
	g.initPieces()
	return g
}

func (g *Generator) initPieces() {
	g.pieces = nil

	// 2x2 square
	g.pieces = append(g.pieces, newPiece([]grid.Coord{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1},
	}))

	// corner
	t := newPiece([]grid.Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}})
	g.pieces = append(g.pieces, t)
	g.pieces = append(g.pieces, g.flipX(t))
	g.pieces = append(g.pieces, g.flipY(t))
	g.pieces = append(g.pieces, g.transpose(t))

	// T with tail
	t = newPiece([]grid.Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 0, Y: 2}})
	g.pieces = append(g.pieces, t)
	g.pieces = append(g.pieces, g.flipX(t))
	g.pieces = append(g.pieces, g.transpose(t))
	g.pieces = append(g.pieces, g.flipY(g.transpose(t)))

	// plus
	t = newPiece([]grid.Coord{{X: 1, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 1}})
	g.pieces = append(g.pieces, t)

	// L with extra
	t = newPiece([]grid.Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}})
	g.pieces = append(g.pieces, t)
	g.pieces = append(g.pieces, g.flipX(t))
	g.pieces = append(g.pieces, g.flipY(t))
	g.pieces = append(g.pieces, g.flipX(g.flipY(t)))
	g.pieces = append(g.pieces, g.flipX(g.flipY(g.transpose(t))))
	g.pieces = append(g.pieces, g.transpose(t))
	g.pieces = append(g.pieces, g.flipY(g.transpose(t)))
	g.pieces = append(g.pieces, g.flipX(g.transpose(t)))
}

func (g *Generator) flipX(p *piece) *piece {
	return g.mapPiece(p, func(c grid.Coord) grid.Coord { return grid.Coord{X: p.maxX - c.X, Y: c.Y} })
}
func (g *Generator) flipY(p *piece) *piece {
	return g.mapPiece(p, func(c grid.Coord) grid.Coord { return grid.Coord{X: c.X, Y: p.maxY - c.Y} })
}
func (g *Generator) transpose(p *piece) *piece {
	return g.mapPiece(p, func(c grid.Coord) grid.Coord { return grid.Coord{X: c.Y, Y: c.X} })
}

func (g *Generator) mapPiece(p *piece, fn func(grid.Coord) grid.Coord) *piece {
	cells := make([]grid.Coord, 0, len(p.blocks))
	for c := range p.blocks {
		cells = append(cells, fn(c))
	}
	return newPiece(cells)
}

// GenerateWithHorizontalSymmetry mirrors generateWithHorizontalSymetry. Fills the
// left half via generate, mirrors to the right half, then walls off all islands
// smaller than the largest to guarantee a single connected component.
func (g *Generator) GenerateWithHorizontalSymmetry(board *grid.Grid, r *javarand.Random) {
	w := board.Width/2 + 1
	h := board.Height
	mini := grid.NewGrid(w, h, board.MapWraps)
	g.Generate(mini, r)

	for _, cell := range mini.Cells() {
		board.Get(cell.Coord).Copy(cell)

		rightPos := grid.Coord{X: board.Width - cell.Coord.X - 1, Y: cell.Coord.Y}
		board.Get(rightPos).Copy(cell)
	}

	var generatedFloors []grid.Coord
	for _, cell := range board.Cells() {
		if cell.IsFloor() {
			generatedFloors = append(generatedFloors, cell.Coord)
		}
	}

	islands := detectIslands(generatedFloors, board)
	sort.SliceStable(islands, func(i, j int) bool {
		return len(islands[i]) > len(islands[j])
	})
	for i := 1; i < len(islands); i++ {
		for _, c := range islands[i] {
			board.Get(c).Type = grid.CellWall
		}
	}
}

// Generate lays tetris pieces over the top-left quadrant of board and carves
// floors adjacent to cells outside the chosen pieces.
func (g *Generator) Generate(board *grid.Grid, r *javarand.Random) {
	genW := board.Width/2 + 1
	genH := board.Height/2 + 1

	generatedPieces := make(map[grid.Coord]*piece)
	blockOrigin := make(map[grid.Coord]grid.Coord)
	occupied := make(map[grid.Coord]struct{})

	for y := 0; y < genH; y++ {
		for x := 0; x < genW; x++ {
			pos := grid.Coord{X: x, Y: y}
			if _, ok := occupied[pos]; ok {
				continue
			}
			shuffle(g.pieces, r)

			// "Take the first available piece, unless it is the only one."
			// skip(1).findFirst => the second fitting piece.
			found := 0
			var chosen *piece
			for _, p := range g.pieces {
				if pieceFits(p, occupied, pos) {
					found++
					if found == 2 {
						chosen = p
						break
					}
				}
			}
			if chosen != nil {
				placePiece(generatedPieces, blockOrigin, occupied, pos, chosen)
			}
		}
	}

	for _, cell := range board.Cells() {
		cell.Type = grid.CellWall
	}

	adjacency := [4]grid.Coord{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}
	for y := 1; y < genH; y++ {
		for x := 1; x < genW; x++ {
			pos := grid.Coord{X: x, Y: y}
			origin, ok := blockOrigin[pos]
			if !ok {
				continue
			}
			gridPos := grid.Coord{X: x*2 - 1, Y: y*2 - 1}
			pc := generatedPieces[origin]
			block := grid.Coord{X: pos.X - origin.X, Y: pos.Y - origin.Y}
			for _, delta := range adjacency {
				adj := grid.Coord{X: block.X + delta.X, Y: block.Y + delta.Y}
				if _, hasBlock := pc.blocks[adj]; hasBlock {
					continue
				}
				for i := 0; i < 3; i++ {
					var cellPos grid.Coord
					if delta.X == 0 {
						cellPos = grid.Coord{X: gridPos.X - 1 + i, Y: gridPos.Y + delta.Y}
					} else {
						cellPos = grid.Coord{X: gridPos.X + delta.X, Y: gridPos.Y - 1 + i}
					}
					cell := board.Get(cellPos)
					if cell.IsValid() {
						cell.Type = grid.CellFloor
					}
				}
			}
		}
	}
}

func pieceFits(p *piece, occupied map[grid.Coord]struct{}, pos grid.Coord) bool {
	for c := range p.blocks {
		d := grid.Coord{X: pos.X + c.X, Y: pos.Y + c.Y}
		if _, ok := occupied[d]; ok {
			return false
		}
	}
	return true
}

func placePiece(generatedPieces map[grid.Coord]*piece, blockOrigin map[grid.Coord]grid.Coord, occupied map[grid.Coord]struct{}, pos grid.Coord, p *piece) {
	generatedPieces[pos] = p
	for c := range p.blocks {
		d := grid.Coord{X: pos.X + c.X, Y: pos.Y + c.Y}
		blockOrigin[d] = pos
		occupied[d] = struct{}{}
	}
}

func detectIslands(generatedFloors []grid.Coord, board *grid.Grid) [][]grid.Coord {
	var islands [][]grid.Coord
	computed := make(map[grid.Coord]struct{})
	for _, first := range generatedFloors {
		if _, ok := computed[first]; ok {
			continue
		}
		fifo := []grid.Coord{first}
		computed[first] = struct{}{}
		island := []grid.Coord{first}
		for len(fifo) > 0 {
			e := fifo[0]
			fifo = fifo[1:]
			for _, n := range board.Neighbours(e) {
				if _, done := computed[n]; done {
					continue
				}
				if board.Get(n).IsFloor() {
					fifo = append(fifo, n)
					computed[n] = struct{}{}
					island = append(island, n)
				}
			}
		}
		islands = append(islands, island)
	}
	return islands
}

// shuffle is Java's Collections.shuffle(list, random).
// Iterates from the end, swapping with a random index in [0, i).
func shuffle[T any](list []T, r *javarand.Random) {
	for i := len(list); i > 1; i-- {
		j := r.NextInt(i)
		list[i-1], list[j] = list[j], list[i-1]
	}
}
