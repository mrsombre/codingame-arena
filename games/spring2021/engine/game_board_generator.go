// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/BoardGenerator.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/BoardGenerator.java:9-63

public class BoardGenerator {
    static Map<CubeCoord, Cell> board;
    static int index;

    public static void generateCell(CubeCoord coord, int richness) {
        Cell cell = new Cell(index++);
        cell.setRichness(richness);
        board.put(coord, cell);
    }

    public static Board generate(Random random) {
        board = new HashMap<>();
        index = 0;
        CubeCoord centre = new CubeCoord(0, 0, 0);
        generateCell(centre, Constants.RICHNESS_LUSH);
        CubeCoord coord = centre.neighbor(0);
        for (int distance = 1; distance <= Config.MAP_RING_COUNT; distance++) {
            for (int orientation = 0; orientation < 6; orientation++) {
                for (int count = 0; count < distance; count++) {
                    if (distance == Config.MAP_RING_COUNT)            generateCell(coord, RICHNESS_POOR);
                    else if (distance == Config.MAP_RING_COUNT - 1)   generateCell(coord, RICHNESS_OK);
                    else                                              generateCell(coord, RICHNESS_LUSH);
                    coord = coord.neighbor((orientation + 2) % 6);
                }
            }
            coord = coord.neighbor(0);
        }
        // Optionally NULL out cell pairs (with their opposite) until wantedEmptyCells reached.
        return new Board(board);
    }
}
*/

// BoardGenerator builds the 37-cell hex map and applies the optional empty-
// cell holes. The Java version uses static fields; we keep matching state on
// a per-Game owner.
type BoardGenerator struct {
	board map[CubeCoord]*Cell
	index int
}

func NewBoardGenerator() *BoardGenerator {
	return &BoardGenerator{}
}

func (g *BoardGenerator) generateCell(coord CubeCoord, richness int) {
	cell := NewCell(g.index)
	g.index++
	cell.SetRichness(richness)
	g.board[coord] = cell
}

// Generate builds a Board, pulling random ints from r when ENABLE_HOLES is true.
// Callers pass Config and the ENABLE_HOLES flag because Java reads them off
// per-match Config + Game statics; bundling them keeps the function pure.
func (g *BoardGenerator) Generate(r *javarand.Random, cfg Config, enableHoles bool) *Board {
	g.board = make(map[CubeCoord]*Cell)
	g.index = 0

	centre := NewCubeCoord(0, 0, 0)
	g.generateCell(centre, RICHNESS_LUSH)

	coord := centre.Neighbor(0)
	for distance := 1; distance <= cfg.MAP_RING_COUNT; distance++ {
		for orientation := 0; orientation < 6; orientation++ {
			for count := 0; count < distance; count++ {
				switch distance {
				case cfg.MAP_RING_COUNT:
					g.generateCell(coord, RICHNESS_POOR)
				case cfg.MAP_RING_COUNT - 1:
					g.generateCell(coord, RICHNESS_OK)
				default:
					g.generateCell(coord, RICHNESS_LUSH)
				}
				coord = coord.Neighbor((orientation + 2) % 6)
			}
		}
		coord = coord.Neighbor(0)
	}

	// Java keeps insertion order via `new ArrayList<>(board.keySet())` over a
	// HashMap. That order is undefined across JVMs, but the consumer of this
	// list (the empty-cell loop) only uses it as a sampling pool, so the
	// resulting set of disabled cells will still be deterministic from r.
	// We sort by cell index for stable Go iteration.
	coordList := make([]CubeCoord, len(g.board))
	for c, cell := range g.board {
		coordList[cell.Index] = c
	}
	coordListSize := len(coordList)

	wantedEmptyCells := 0
	if enableHoles {
		wantedEmptyCells = r.NextInt(cfg.MAX_EMPTY_CELLS + 1)
	}
	actualEmptyCells := 0

	for actualEmptyCells < wantedEmptyCells-1 {
		randIndex := r.NextInt(coordListSize)
		randCoord := coordList[randIndex]
		if g.board[randCoord].GetRichness() != RICHNESS_NULL {
			g.board[randCoord].SetRichness(RICHNESS_NULL)
			actualEmptyCells++
			opp := randCoord.Opposite()
			if randCoord != opp {
				g.board[opp].SetRichness(RICHNESS_NULL)
				actualEmptyCells++
			}
		}
	}

	return NewBoard(g.board)
}
