// Package grid
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Cell.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/CellType.java
package grid

// Cell types.
type CellType int

const (
	CellWall CellType = iota
	CellFloor
)

// Cell represents a single tile in the grid.
type Cell struct {
	Coord     Coord
	Type      CellType
	HasPellet bool
	HasCherry bool
	valid     bool
}

// NoCell is the sentinel for out-of-bounds lookups.
var NoCell = Cell{Coord: Coord{X: -1, Y: -1}, valid: false}

func NewCell(coord Coord) *Cell {
	return &Cell{Coord: coord, valid: true}
}

func (c *Cell) IsValid() bool { return c.valid }
func (c *Cell) IsFloor() bool { return c.Type == CellFloor }
func (c *Cell) IsWall() bool  { return c.Type == CellWall }

// Copy copies type and HasPellet from source, like Java's Cell.copy.
func (c *Cell) Copy(source *Cell) {
	c.Type = source.Type
	c.HasPellet = source.HasPellet
}
