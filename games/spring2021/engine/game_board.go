// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Board.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Board.java:8-23

public class Board {
    public final Map<CubeCoord, Cell> map;
    public final List<CubeCoord> coords;

    public Board(Map<CubeCoord, Cell> map) {
        this.map = map;
        coords = map.entrySet().stream()
            .sorted((a, b) -> a.getValue().getIndex() - b.getValue().getIndex())
            .map(Entry::getKey)
            .collect(Collectors.toList());
    }
}
*/

// Board mirrors Java's Map<CubeCoord,Cell> + index-ordered coords list. The
// Java engine relies on the coords list being sorted by Cell index, which is
// also the cell's position inside `Cells` here — so the slice doubles as the
// id→coord lookup the simulation needs.
type Board struct {
	Map    map[CubeCoord]*Cell
	Coords []CubeCoord
	Cells  []*Cell
}

// NewBoard sorts the cells by their `index` field — same ordering Java
// produces with its Comparator on Map.Entry::getValue.
func NewBoard(m map[CubeCoord]*Cell) *Board {
	coords := make([]CubeCoord, len(m))
	cells := make([]*Cell, len(m))
	for coord, cell := range m {
		coords[cell.Index] = coord
		cells[cell.Index] = cell
	}
	return &Board{Map: m, Coords: coords, Cells: cells}
}

// CellByIndex returns the Cell with the given index, or nil if the index is
// out of range. Mirrors Java's stream lookup that throws CellNotFoundException
// when the id is missing — callers wrap that into a GameError.
func (b *Board) CellByIndex(index int) *Cell {
	if index < 0 || index >= len(b.Cells) {
		return nil
	}
	return b.Cells[index]
}

// CoordByIndex returns the cube coord of the cell with the given index.
func (b *Board) CoordByIndex(index int) (CubeCoord, bool) {
	if index < 0 || index >= len(b.Coords) {
		return CubeCoord{}, false
	}
	return b.Coords[index], true
}

// CellAt returns the cell at the given coord, or nil for off-board coords.
// Mirrors Java's `map.getOrDefault(coord, Cell.NO_CELL)` callers.
func (b *Board) CellAt(c CubeCoord) *Cell {
	return b.Map[c]
}
