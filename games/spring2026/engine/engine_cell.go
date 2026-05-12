// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Cell.java
package engine

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Cell.java:9-13

public class Cell {
    public enum Type {
        GRASS, WATER, ROCK, IRON, SHACK
    }
*/

// CellType enumerates the possible terrain types for a cell.
type CellType int

const (
	CellGRASS CellType = iota
	CellWATER
	CellROCK
	CellIRON
	CellSHACK
)

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Cell.java:14-72

private int x, y;
private Type type;
private Plant plant;
private Cell[] neighbors = new Cell[4];

public Cell(int x, int y) {
    this.x = x;
    this.y = y;
    this.type = Type.GRASS;
}

private static final int[] dx = {0, 1, 0, -1};
private static final int[] dy = {1, 0, -1, 0};

public void initNeighbors(Board board) {
    for (int dir = 0; dir < 4; dir++) {
        int x_ = x + dx[dir];
        int y_ = y + dy[dir];
        if (x_ < 0 || x_ >= board.getWidth() || y_ < 0 || y_ >= board.getHeight()) continue;
        neighbors[dir] = board.getCell(x_, y_);
    }
}
*/

// Cell mirrors Java engine.Cell. Neighbours [up, right, down, left] use Java's
// dx={0,1,0,-1}, dy={1,0,-1,0} order. Out-of-board neighbours are nil — the
// same shape Java carries as null entries.
type Cell struct {
	X, Y      int
	Type      CellType
	Plant     *Plant
	Neighbors [4]*Cell
}

// Java Cell.dx / Cell.dy.
var cellDX = [4]int{0, 1, 0, -1}
var cellDY = [4]int{1, 0, -1, 0}

func NewCell(x, y int) *Cell {
	return &Cell{X: x, Y: y, Type: CellGRASS}
}

// GetID mirrors Java Cell.getId(): x + (y << 16). Used as a map / sort key for
// deterministic iteration in tasks (groupByCell).
func (c *Cell) GetID() int { return c.X + (c.Y << 16) }

func (c *Cell) GetX() int { return c.X }
func (c *Cell) GetY() int { return c.Y }

func (c *Cell) GetType() CellType    { return c.Type }
func (c *Cell) SetType(t CellType)   { c.Type = t }
func (c *Cell) GetPlant() *Plant     { return c.Plant }
func (c *Cell) SetPlant(p *Plant)    { c.Plant = p }
func (c *Cell) GetNeighbor(i int) *Cell { return c.Neighbors[i] }
func (c *Cell) GetNeighbors() [4]*Cell  { return c.Neighbors }

func (c *Cell) InitNeighbors(board *Board) {
	for dir := 0; dir < 4; dir++ {
		nx := c.X + cellDX[dir]
		ny := c.Y + cellDY[dir]
		if nx < 0 || nx >= board.Width || ny < 0 || ny >= board.Height {
			continue
		}
		c.Neighbors[dir] = board.GetCell(nx, ny)
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Cell.java:74-103

public boolean isWalkable() { return type == Type.GRASS; }
public boolean isNearWater() { return isNearType(Type.WATER); }
public boolean isNearIron() { return isNearType(Type.IRON); }
public boolean isNearShack(Player player) {
    return this == player.getShack() || Arrays.stream(neighbors).anyMatch(c -> c == player.getShack());
}
public boolean isNearEdge() { return Arrays.stream(neighbors).anyMatch(n -> n == null); }
public int manhattan(Cell cell) { return Math.abs(x - cell.x) + Math.abs(y - cell.y); }

private boolean isNearType(Type type) {
    for (Cell neighbor : neighbors) {
        if (neighbor != null && neighbor.type == type) return true;
    }
    return false;
}
*/

func (c *Cell) IsWalkable() bool  { return c.Type == CellGRASS }
func (c *Cell) IsNearWater() bool { return c.isNearType(CellWATER) }
func (c *Cell) IsNearIron() bool  { return c.isNearType(CellIRON) }

func (c *Cell) IsNearShack(p *Player) bool {
	if c == p.Shack {
		return true
	}
	for _, n := range c.Neighbors {
		if n == p.Shack && n != nil {
			return true
		}
	}
	return false
}

func (c *Cell) IsNearEdge() bool {
	for _, n := range c.Neighbors {
		if n == nil {
			return true
		}
	}
	return false
}

func (c *Cell) Manhattan(other *Cell) int {
	dx := c.X - other.X
	if dx < 0 {
		dx = -dx
	}
	dy := c.Y - other.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (c *Cell) isNearType(t CellType) bool {
	for _, n := range c.Neighbors {
		if n != nil && n.Type == t {
			return true
		}
	}
	return false
}
