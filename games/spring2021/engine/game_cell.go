// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Cell.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Cell.java:3-39

public class Cell {
    public static final Cell NO_CELL = new Cell(-1) {
        @Override public boolean isValid() { return false; }
        @Override public int getIndex() { return -1; }
    };
    private int richness;
    private int index;

    public Cell(int index) { this.index = index; }
    public int getIndex() { return index; }
    public boolean isValid() { return true; }
    public void setRichness(int richness) { this.richness = richness; }
    public int getRichness() { return richness; }
}
*/

// Cell represents one hex on the board. The Java NO_CELL sentinel is
// represented by a (*Cell)(nil); IsValid(), GetIndex(), and GetRichness() are
// nil-tolerant so out-of-bounds lookups behave the same as the sentinel.
type Cell struct {
	Index    int
	Richness int
}

func NewCell(index int) *Cell {
	return &Cell{Index: index}
}

func (c *Cell) IsValid() bool {
	return c != nil
}

func (c *Cell) GetIndex() int {
	if c == nil {
		return -1
	}
	return c.Index
}

func (c *Cell) GetRichness() int {
	if c == nil {
		return 0
	}
	return c.Richness
}

func (c *Cell) SetRichness(r int) {
	c.Richness = r
}
