// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java
package engine

import (
	"sort"

	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:23-36

public class TetrisBasedMapGenerator {
    class TetrisPiece {
        Set<Coord> blocks;
        int maxX, maxY;
        public TetrisPiece(Set<Coord> cells) {
            this.blocks = cells;
            maxX = cells.stream().mapToInt(Coord::getX).max().getAsInt();
            maxY = cells.stream().mapToInt(Coord::getY).max().getAsInt();
        }
    }
    List<TetrisPiece> pieces;
}
*/

// TetrisBasedMapGenerator lays out walls and floors on a Grid using shuffled tetris Pieces.
type TetrisBasedMapGenerator struct {
	Pieces []*TetrisPiece
}

type TetrisPiece struct {
	Blocks map[Coord]struct{}
	MaxX   int
	MaxY   int
}

func NewTetrisPiece(cells []Coord) *TetrisPiece {
	Blocks := make(map[Coord]struct{}, len(cells))
	MaxX, MaxY := 0, 0
	for _, c := range cells {
		Blocks[c] = struct{}{}
		if c.X > MaxX {
			MaxX = c.X
		}
		if c.Y > MaxY {
			MaxY = c.Y
		}
	}
	return &TetrisPiece{Blocks: Blocks, MaxX: MaxX, MaxY: MaxY}
}

// NewTetrisBasedMapGenerator constructs a generator with all tetromino variants populated.
func NewTetrisBasedMapGenerator() *TetrisBasedMapGenerator {
	g := &TetrisBasedMapGenerator{}
	g.Init()
	return g
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:38-108

public void init() {
    pieces = new ArrayList<>();
    // 2x2 square
    pieces.add(new TetrisPiece(setOf(0,0; 1,0; 0,1; 1,1)));
    // corner: t + flipX + flipY + transpose
    t = new TetrisPiece(setOf(0,0; 0,1; 1,1));
    pieces.addAll(t, flipX(t), flipY(t), transpose(t));
    // T with tail
    t = new TetrisPiece(setOf(0,0; 0,1; 1,1; 0,2));
    pieces.addAll(t, flipX(t), transpose(t), flipY(transpose(t)));
    // plus
    pieces.add(new TetrisPiece(setOf(1,0; 1,1; 2,1; 1,2; 0,1)));
    // L with extra: t and 7 reflections/rotations
    t = new TetrisPiece(setOf(0,0; 0,1; 1,1; 2,1));
    pieces.addAll(t, flipX(t), flipY(t), flipX(flipY(t)),
        flipX(flipY(transpose(t))), transpose(t), flipY(transpose(t)), flipX(transpose(t)));
}
*/

func (g *TetrisBasedMapGenerator) Init() {
	g.Pieces = nil

	// 2x2 square
	g.Pieces = append(g.Pieces, NewTetrisPiece([]Coord{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1},
	}))

	// corner
	t := NewTetrisPiece([]Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}})
	g.Pieces = append(g.Pieces, t)
	g.Pieces = append(g.Pieces, g.FlipX(t))
	g.Pieces = append(g.Pieces, g.FlipY(t))
	g.Pieces = append(g.Pieces, g.Transpose(t))

	// T with tail
	t = NewTetrisPiece([]Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 0, Y: 2}})
	g.Pieces = append(g.Pieces, t)
	g.Pieces = append(g.Pieces, g.FlipX(t))
	g.Pieces = append(g.Pieces, g.Transpose(t))
	g.Pieces = append(g.Pieces, g.FlipY(g.Transpose(t)))

	// plus
	t = NewTetrisPiece([]Coord{{X: 1, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 1}})
	g.Pieces = append(g.Pieces, t)

	// L with extra
	t = NewTetrisPiece([]Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}})
	g.Pieces = append(g.Pieces, t)
	g.Pieces = append(g.Pieces, g.FlipX(t))
	g.Pieces = append(g.Pieces, g.FlipY(t))
	g.Pieces = append(g.Pieces, g.FlipX(g.FlipY(t)))
	g.Pieces = append(g.Pieces, g.FlipX(g.FlipY(g.Transpose(t))))
	g.Pieces = append(g.Pieces, g.Transpose(t))
	g.Pieces = append(g.Pieces, g.FlipY(g.Transpose(t)))
	g.Pieces = append(g.Pieces, g.FlipX(g.Transpose(t)))
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:118-128

private TetrisPiece flipX(TetrisPiece t) {
    return flip(t, coord -> new Coord(t.maxX - coord.getX(), coord.getY()));
}
private TetrisPiece flipY(TetrisPiece t) {
    return flip(t, coord -> new Coord(coord.getX(), t.maxY - coord.getY()));
}
private TetrisPiece transpose(TetrisPiece t) {
    return flip(t, coord -> new Coord(coord.getY(), coord.getX()));
}
*/

func (g *TetrisBasedMapGenerator) FlipX(p *TetrisPiece) *TetrisPiece {
	return g.MapPiece(p, func(c Coord) Coord { return Coord{X: p.MaxX - c.X, Y: c.Y} })
}
func (g *TetrisBasedMapGenerator) FlipY(p *TetrisPiece) *TetrisPiece {
	return g.MapPiece(p, func(c Coord) Coord { return Coord{X: c.X, Y: p.MaxY - c.Y} })
}
func (g *TetrisBasedMapGenerator) Transpose(p *TetrisPiece) *TetrisPiece {
	return g.MapPiece(p, func(c Coord) Coord { return Coord{X: c.Y, Y: c.X} })
}

func (g *TetrisBasedMapGenerator) MapPiece(p *TetrisPiece, fn func(Coord) Coord) *TetrisPiece {
	cells := make([]Coord, 0, len(p.Blocks))
	for c := range p.Blocks {
		cells = append(cells, fn(c))
	}
	return NewTetrisPiece(cells)
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:130-166

public void generateWithHorizontalSymetry(Grid grid, Random random) {
    int w = grid.getWidth() / 2 + 1;
    int h = grid.getHeight();
    Grid miniGrid = new Grid(w, h);
    generate(miniGrid, random);

    for (Map.Entry<Coord, Cell> entry : miniGrid.getCells().entrySet()) {
        grid.get(sourceCoord).copy(sourceCell);
        Coord pos = new Coord(grid.getWidth() - sourceCoord.getX() - 1, sourceCoord.getY());
        grid.get(pos).copy(sourceCell);
    }

    LinkedList<List<Coord>> islands = detectIslands(generatedFloors, grid);
    islands.stream().sorted((a, b) -> b.size() - a.size()).skip(1)
        .forEach(list -> list.forEach(coord -> grid.get(coord).setType(CellType.WALL)));
}
*/

// GenerateWithHorizontalSymmetry mirrors generateWithHorizontalSymetry. Fills the
// left half via generate, mirrors to the right half, then walls off all islands
// smaller than the largest to guarantee a single connected component.
func (g *TetrisBasedMapGenerator) GenerateWithHorizontalSymmetry(board *Grid, r *javarand.Random) {
	w := board.Width/2 + 1
	h := board.Height
	mini := NewGrid(w, h, board.MapWraps)
	g.Generate(mini, r)

	for _, cell := range mini.Cells {
		board.Get(cell.Coord).Copy(cell)

		rightPos := Coord{X: board.Width - cell.Coord.X - 1, Y: cell.Coord.Y}
		board.Get(rightPos).Copy(cell)
	}

	var generatedFloors []Coord
	for _, cell := range board.Cells {
		if cell.IsFloor() {
			generatedFloors = append(generatedFloors, cell.Coord)
		}
	}

	islands := DetectIslands(generatedFloors, board)
	sort.SliceStable(islands, func(i, j int) bool {
		return len(islands[i]) > len(islands[j])
	})
	for i := 1; i < len(islands); i++ {
		for _, c := range islands[i] {
			board.Get(c).Type = CellWall
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:189-253

public void generate(Grid grid, Random random) {
    int generateWidth = grid.getWidth() / 2 + 1;
    int generateHeight = grid.getHeight() / 2 + 1;

    Map<Coord, TetrisPiece> generatedPieces = new HashMap<>();
    Map<Coord, Coord> blockOrigin = new HashMap<>();
    Set<Coord> occupied = new HashSet<>();
    for (int y = 0; y < generateHeight; ++y) {
        for (int x = 0; x < generateWidth; ++x) {
            Coord pos = new Coord(x, y);
            if (!occupied.contains(pos)) {
                Collections.shuffle(pieces, random);
                // Take the first available piece, unless it is the only one.
                Optional<TetrisPiece> pieceOpt = pieces.stream()
                    .filter(p -> pieceFits(p, occupied, pos))
                    .skip(1)
                    .findFirst();
                if (pieceOpt.isPresent()) {
                    placePiece(generatedPieces, blockOrigin, occupied, pos, pieceOpt.get());
                }
            }
        }
    }

    grid.getCells().values().stream().forEach(c -> c.setType(CellType.WALL));

    for (int y = 1; y < generateHeight; ++y) {
        for (int x = 1; x < generateWidth; ++x) {
            Coord origin = blockOrigin.get(pos);
            if (origin == null) continue;
            Coord gridPos = new Coord(x * 2 - 1, y * 2 - 1);
            // For each adjacency direction not already inside the piece,
            // carve a 3-cell strip of FLOOR through the doubled grid.
        }
    }
}
*/

// Generate lays tetris pieces over the top-left quadrant of board and carves
// floors adjacent to cells outside the chosen pieces.
func (g *TetrisBasedMapGenerator) Generate(board *Grid, r *javarand.Random) {
	genW := board.Width/2 + 1
	genH := board.Height/2 + 1

	generatedPieces := make(map[Coord]*TetrisPiece)
	blockOrigin := make(map[Coord]Coord)
	occupied := make(map[Coord]struct{})

	for y := 0; y < genH; y++ {
		for x := 0; x < genW; x++ {
			pos := Coord{X: x, Y: y}
			if _, ok := occupied[pos]; ok {
				continue
			}
			Shuffle(g.Pieces, r)

			// "Take the first available piece, unless it is the only one."
			// skip(1).findFirst => the second fitting piece.
			found := 0
			var chosen *TetrisPiece
			for _, p := range g.Pieces {
				if PieceFits(p, occupied, pos) {
					found++
					if found == 2 {
						chosen = p
						break
					}
				}
			}
			if chosen != nil {
				PlacePiece(generatedPieces, blockOrigin, occupied, pos, chosen)
			}
		}
	}

	for _, cell := range board.Cells {
		cell.Type = CellWall
	}

	adjacency := [4]Coord{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}
	for y := 1; y < genH; y++ {
		for x := 1; x < genW; x++ {
			pos := Coord{X: x, Y: y}
			origin, ok := blockOrigin[pos]
			if !ok {
				continue
			}
			gridPos := Coord{X: x*2 - 1, Y: y*2 - 1}
			pc := generatedPieces[origin]
			block := Coord{X: pos.X - origin.X, Y: pos.Y - origin.Y}
			for _, delta := range adjacency {
				adj := Coord{X: block.X + delta.X, Y: block.Y + delta.Y}
				if _, hasBlock := pc.Blocks[adj]; hasBlock {
					continue
				}
				for i := 0; i < 3; i++ {
					var cellPos Coord
					if delta.X == 0 {
						cellPos = Coord{X: gridPos.X - 1 + i, Y: gridPos.Y + delta.Y}
					} else {
						cellPos = Coord{X: gridPos.X + delta.X, Y: gridPos.Y - 1 + i}
					}
					cell := board.Get(cellPos)
					if cell.IsValid() {
						cell.Type = CellFloor
					}
				}
			}
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:288-297,314-322

private void placePiece(Map<Coord, TetrisPiece> generatedPieces, Map<Coord, Coord> blockOrigin, Set<Coord> occupied, Coord pos, TetrisPiece tetrisPiece) {
    generatedPieces.put(pos, tetrisPiece);
    for (Coord coord : tetrisPiece.blocks) {
        Coord d = new Coord(pos.getX() + coord.getX(), pos.getY() + coord.getY());
        blockOrigin.put(d, pos);
        occupied.add(d);
    }
}

private boolean pieceFits(TetrisPiece piece, Set<Coord> occupied, Coord pos) {
    for (Coord coord : piece.blocks) {
        Coord d = new Coord(pos.getX() + coord.getX(), pos.getY() + coord.getY());
        if (occupied.contains(d)) return false;
    }
    return true;
}
*/

func PieceFits(p *TetrisPiece, occupied map[Coord]struct{}, pos Coord) bool {
	for c := range p.Blocks {
		d := Coord{X: pos.X + c.X, Y: pos.Y + c.Y}
		if _, ok := occupied[d]; ok {
			return false
		}
	}
	return true
}

func PlacePiece(generatedPieces map[Coord]*TetrisPiece, blockOrigin map[Coord]Coord, occupied map[Coord]struct{}, pos Coord, p *TetrisPiece) {
	generatedPieces[pos] = p
	for c := range p.Blocks {
		d := Coord{X: pos.X + c.X, Y: pos.Y + c.Y}
		blockOrigin[d] = pos
		occupied[d] = struct{}{}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/maps/TetrisBasedMapGenerator.java:255-286

private LinkedList<List<Coord>> detectIslands(List<Coord> generatedFloors, Grid grid) {
    LinkedList<List<Coord>> islands = new LinkedList<>();
    Queue<Coord> fifo = new LinkedList<>();
    Set<Coord> computed = new HashSet<>();
    for (Coord first : generatedFloors) {
        if (computed.contains(first)) continue;
        fifo.add(first);
        island = new LinkedList<>();
        computed.add(first);
        island.add(first);
        while (!fifo.isEmpty()) {
            Coord e = fifo.poll();
            for (Coord n : grid.getNeighbours(e)) {
                if (!computed.contains(n) && grid.get(n).isFloor()) {
                    fifo.add(n);
                    computed.add(n);
                    island.add(n);
                }
            }
        }
        islands.add(island);
    }
    return islands;
}
*/

func DetectIslands(generatedFloors []Coord, board *Grid) [][]Coord {
	var islands [][]Coord
	computed := make(map[Coord]struct{})
	for _, first := range generatedFloors {
		if _, ok := computed[first]; ok {
			continue
		}
		fifo := []Coord{first}
		computed[first] = struct{}{}
		island := []Coord{first}
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

// Shuffle is Java's Collections.shuffle(list, random).
// Iterates from the end, swapping with a random index in [0, i).
func Shuffle[T any](list []T, r *javarand.Random) {
	for i := len(list); i > 1; i-- {
		j := r.NextInt(i)
		list[i-1], list[j] = list[j], list[i-1]
	}
}
