// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java:3-19

public class Tile {
    public static final Tile NO_TILE = new Tile(new Coord(-1, -1), -1);

    public static int TYPE_EMPTY = 0;
    public static int TYPE_WALL = 1;

    private int type;
    Coord coord;

    public Tile(Coord coord) { this.coord = coord; }

    public Tile(Coord coord, int type) {
        this.coord = coord;
        this.type = type;
    }
}

// Java's NO_TILE sentinel is replaced by a nil *Tile in Go; see Grid.GetXY
// and the nil-tolerant Tile methods (IsValid/IsWall/IsEmpty) below.
*/

// Tile types.
const (
	TileEmpty = 0
	TileWall  = 1
)

// Tile represents a single cell in the grid.
//
// Out-of-bounds lookups return a nil *Tile instead of a Java-style NO_TILE
// sentinel; IsValid/IsWall/IsEmpty tolerate nil receivers and return false,
// matching Java's NO_TILE behaviour.
type Tile struct {
	Coord Coord
	Type  int
}

func NewTile(coord Coord) Tile {
	return Tile{Coord: coord}
}

func NewTileWithType(coord Coord, tileType int) Tile {
	return Tile{Coord: coord, Type: tileType}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java:21-26

public void setType(int type) {
    this.type = type;
}
*/

func (t *Tile) SetType(tileType int) {
	t.Type = tileType
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java:33-35

public void clear() {
    type = TYPE_EMPTY;
}
*/

func (t *Tile) Clear() {
	t.Type = TileEmpty
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/Tile.java:37-39

public boolean isValid() {
    return this != NO_TILE;
}
*/

func (t *Tile) IsValid() bool { return t != nil }
func (t *Tile) IsWall() bool  { return t != nil && t.Type == TileWall }
func (t *Tile) IsEmpty() bool { return t != nil && t.Type == TileEmpty }
