// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java:8-22

public class Bird {
    int id;
    LinkedList<Coord> body;
    Player owner;
    boolean alive;
    Direction direction;
    public String message;

    public Bird(int id, Player owner) {
        this.id = id;
        this.owner = owner;
        this.alive = true;
        this.message = null;
        body = new LinkedList<>();
    }
*/

type Bird struct {
	ID        int
	Body      []Coord
	Owner     *Player
	Alive     bool
	Direction Direction
	HasMove   bool
	Message   string
	HasMsg    bool
}

func NewBird(id int, owner *Player) *Bird {
	return &Bird{
		ID:        id,
		Owner:     owner,
		Alive:     true,
		Direction: DirUnset,
		Body:      make([]Coord, 0),
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java:24-26

public Coord getHeadPos() {
    return body.get(0);
}
*/

func (b *Bird) HeadPos() Coord {
	return b.Body[0]
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java:28-36

public Direction getFacing() {
    if (body.size() < 2) {
        return Direction.UNSET;
    }
    return Direction.fromCoord(
        new Coord(body.get(0).getX() - body.get(1).getX(),
                  body.get(0).getY() - body.get(1).getY())
    );
}
*/

func (b *Bird) Facing() Direction {
	if len(b.Body) < 2 {
		return DirUnset
	}
	return DirectionFromCoord(Coord{
		X: b.Body[0].X - b.Body[1].X,
		Y: b.Body[0].Y - b.Body[1].Y,
	})
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java:38-40

public boolean isAlive() {
    return alive;
}
*/

func (b *Bird) IsAlive() bool {
	return b.Alive
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Bird.java:42-48

public void setMessage(String message) {
    this.message = message;
    if (message != null && message.length() > 48) {
        this.message = message.substring(0, 46) + "...";
    }
}

public boolean hasMessage() {
    return message != null;
}
*/

func (b *Bird) SetMessage(message string) {
	b.Message = message
	b.HasMsg = true
	if len(message) > 48 {
		b.Message = message[:46] + "..."
	}
}

func (b *Bird) HasMessage() bool {
	return b.HasMsg
}
