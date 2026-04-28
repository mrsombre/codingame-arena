// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Cell.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Cell.java:3-35

public class Cell {
    public static final Cell NO_CELL = new Cell() {
        @Override public boolean isValid() { return false; }
        @Override public void copy(Cell other) { throw new RuntimeException("Invalid cell"); }
        @Override public void setType(CellType type) { throw new RuntimeException("Invalid cell"); }
    };

    private CellType type;
    private Ability powerUp;
    private boolean hasPellet;
    private boolean hasCherry;

    public Cell(CellType type) { this.setType(type); }
}
*/

// Cell represents a single tile in the grid.
//
// Out-of-bounds lookups return a nil *Cell instead of a Java-style NO_CELL
// sentinel; IsValid/IsFloor/IsWall tolerate nil receivers and return false,
// matching Java's NO_CELL behaviour.
type Cell struct {
	Coord     Coord
	Type      CellType
	HasPellet bool
	HasCherry bool
}

func NewCell(coord Coord) *Cell {
	return &Cell{Coord: coord}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Cell.java:37-39,57-63

public boolean isValid() { return true; }
public boolean isFloor() { return type == CellType.FLOOR; }
public boolean isWall()  { return type == CellType.WALL; }
*/

func (c *Cell) IsValid() bool { return c != nil }
func (c *Cell) IsFloor() bool { return c != nil && c.Type == CellFloor }
func (c *Cell) IsWall() bool  { return c != nil && c.Type == CellWall }

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Cell.java:81-84

public void copy(Cell source) {
    setType(source.type);
    setHasPellet(source.hasPellet);
}
*/

// Copy copies type and HasPellet from source, like Java's Cell.copy.
func (c *Cell) Copy(source *Cell) {
	c.Type = source.Type
	c.HasPellet = source.HasPellet
}
