// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:10-16

public class Player extends AbstractMultiplayerPlayer {

    int points;
    List<Bird> birds;
    List<Coord> marks;
*/

type Player struct {
	index              int
	score              int
	Points             int
	Birds              []*Bird
	Marks              []Coord
	deactivated        bool
	deactivationReason string
	timedOut           bool
	inputLines         []string
	outputs            []string
	outputError        error
	executeFunc        func() error
}

func NewPlayer(index int) *Player {
	return &Player{index: index}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:28-31

public void init() {
    birds = new ArrayList<>();
    marks = new ArrayList<>();
}
*/

func (p *Player) Init() {
	p.Birds = make([]*Bird, 0)
	p.Marks = make([]Coord, 0)
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:20-26

public void reset() {
    birds.stream().forEach(bird -> {
        bird.direction = null;
        bird.message = null;
    });
    marks.clear();
}
*/

func (p *Player) Reset() {
	for _, bird := range p.Birds {
		bird.Direction = DirUnset
		bird.HasMove = false
		bird.Message = ""
		bird.HasMsg = false
	}
	p.Marks = p.Marks[:0]
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:38-41

public void addScore(int points) {
    this.points += points;
    setScore(this.points);
}
*/

func (p *Player) AddScore(points int) {
	p.Points += points
	p.SetScore(p.Points)
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:43-49

public Bird getBirdById(int id) {
    for (Bird bird : birds) {
        if (bird.id == id) {
            return bird;
        }
    }
    return null;
}
*/

func (p *Player) BirdByID(id int) *Bird {
	for _, bird := range p.Birds {
		if bird.ID == id {
			return bird
		}
	}
	return nil
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:51-57

public boolean addMark(Coord coord) {
    if (marks.size() < 4) {
        marks.add(coord);
        return true;
    }
    return false;
}
*/

func (p *Player) AddMark(coord Coord) bool {
	if len(p.Marks) < 4 {
		p.Marks = append(p.Marks, coord)
		return true
	}
	return false
}

func (p *Player) GetIndex() int { return p.index }
func (p *Player) GetScore() int { return p.score }
func (p *Player) SetScore(score int) {
	p.score = score
}
func (p *Player) IsDeactivated() bool { return p.deactivated }
func (p *Player) Deactivate(reason string) {
	p.deactivated = true
	p.deactivationReason = reason
}
func (p *Player) DeactivationReason() string { return p.deactivationReason }
func (p *Player) IsTimedOut() bool           { return p.timedOut }
func (p *Player) SetTimedOut(timedOut bool)  { p.timedOut = timedOut }
/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java:33-36

@Override
public int getExpectedOutputLines() {
    return 1;
}
*/

func (p *Player) GetExpectedOutputLines() int {
	return 1
}
func (p *Player) SendInputLine(line string) {
	p.inputLines = append(p.inputLines, line)
}
func (p *Player) ConsumeInputLines() []string {
	lines := append([]string(nil), p.inputLines...)
	p.inputLines = p.inputLines[:0]
	return lines
}
func (p *Player) GetOutputs() []string {
	return p.outputs
}
func (p *Player) SetOutputs(outputs []string) {
	p.outputs = outputs
	p.outputError = nil
}
func (p *Player) GetOutputError() error {
	return p.outputError
}
func (p *Player) SetExecuteFunc(fn func() error) {
	p.executeFunc = fn
}
func (p *Player) Execute() error {
	if p.executeFunc == nil {
		return nil
	}
	err := p.executeFunc()
	p.outputError = err
	return err
}
