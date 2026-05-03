// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Tree.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Tree.java:3-49

public class Tree {
    private int size;
    private Player owner;
    private int fatherIndex = -1;
    private boolean isDormant;
    public void grow() { size++; }
    public void reset() { this.isDormant = false; }
}
*/

type Tree struct {
	Size        int
	Owner       *Player
	FatherIndex int
	Dormant     bool
}

func NewTree() *Tree {
	return &Tree{FatherIndex: -1}
}

func (t *Tree) Grow() {
	t.Size++
}

func (t *Tree) SetDormant() {
	t.Dormant = true
}

func (t *Tree) Reset() {
	t.Dormant = false
}
