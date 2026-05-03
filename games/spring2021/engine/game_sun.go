// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Sun.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Sun.java:3-20

public class Sun {
    private int orientation;
    public int getOrientation() { return orientation; }
    public void setOrientation(int orientation) { this.orientation = (orientation) % 6; }
    public void move() { orientation = (orientation + 1) % 6; }
}
*/

type Sun struct {
	Orientation int
}

func (s *Sun) SetOrientation(orientation int) {
	s.Orientation = orientation % 6
}

func (s *Sun) Move() {
	s.Orientation = (s.Orientation + 1) % 6
}
