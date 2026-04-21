// Package grid
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java
package grid

// Tile types.
const (
	TileEmpty = 0
	TileWall  = 1
)

// Tile represents a single cell in the grid.
type Tile struct {
	Coord Coord
	Type  int
	valid bool
}

// NoTile is the sentinel for out-of-bounds lookups.
var NoTile = Tile{Coord: Coord{X: -1, Y: -1}, Type: -1, valid: false}

func NewTile(coord Coord) Tile {
	return Tile{Coord: coord, valid: true}
}

func NewTileWithType(coord Coord, tileType int) Tile {
	return Tile{Coord: coord, Type: tileType, valid: true}
}

func (t *Tile) SetType(tileType int) {
	t.Type = tileType
}

func (t *Tile) Clear() {
	t.Type = TileEmpty
}

func (t *Tile) IsValid() bool {
	return t.valid
}
