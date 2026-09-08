// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Player.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Player.java:9-27

public class Player extends AbstractMultiplayerPlayer {
    List<Action> intents;
    int dosh;
    public int blotPoints;
    private String message;

    public void init() {
        intents = new ArrayList<>();
        dosh = Game.STARTING_DOSH;
    }

    public void reset() {
        intents.clear();
        message = null;
    }
*/

// Player is the engine-side handle for one agent; it satisfies arena.Player.
// Dosh is the paint budget and BlotPoints the disruption budget, both reset
// every turn by Game.DoIncome.
type Player struct {
	index      int
	score      int
	Intents    []*Action
	Dosh       int
	BlotPoints int
	Message    string

	deactivated        bool
	deactivationReason string
	timedOut           bool
	inputLines         []string
	outputs            []string
	outputError        error
	executeFunc        func() error
}

func NewPlayer(index int) *Player {
	return &Player{index: index, Intents: []*Action{}}
}

func (p *Player) Init() {
	p.Intents = []*Action{}
	p.Dosh = STARTING_DOSH
}

func (p *Player) Reset() {
	p.Intents = p.Intents[:0]
	p.Message = ""
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Player.java:29-50

public int getDosh() { return dosh; }
@Override public int getExpectedOutputLines() { return 1; }
public void addScore(int points) { setScore(getScore() + points); }
public void pay(int cost) { dosh -= cost; }
public void addDosh(int amount) { dosh += amount; }
*/

func (p *Player) GetDosh() int { return p.Dosh }

func (p *Player) AddScore(points int) { p.SetScore(p.GetScore() + points) }

func (p *Player) Pay(cost int) { p.Dosh -= cost }

func (p *Player) AddDosh(amount int) { p.Dosh += amount }

func (p *Player) GetMessage() string { return p.Message }

func (p *Player) SetMessage(message string) { p.Message = message }

// --- arena.Player implementation ---

func (p *Player) GetIndex() int { return p.index }

func (p *Player) GetScore() int { return p.score }

func (p *Player) SetScore(score int) { p.score = score }

func (p *Player) IsDeactivated() bool { return p.deactivated }

func (p *Player) Deactivate(reason string) {
	p.deactivated = true
	p.deactivationReason = reason
}

func (p *Player) DeactivationReason() string { return p.deactivationReason }

func (p *Player) IsTimedOut() bool { return p.timedOut }

func (p *Player) SetTimedOut(timedOut bool) { p.timedOut = timedOut }

// GetExpectedOutputLines: Back Track King bots emit one semicolon-joined
// line per turn.
func (p *Player) GetExpectedOutputLines() int { return 1 }

func (p *Player) SendInputLine(line string) {
	p.inputLines = append(p.inputLines, line)
}

func (p *Player) ConsumeInputLines() []string {
	lines := append([]string(nil), p.inputLines...)
	p.inputLines = p.inputLines[:0]
	return lines
}

func (p *Player) GetOutputs() []string { return p.outputs }

func (p *Player) SetOutputs(outputs []string) {
	p.outputs = outputs
	p.outputError = nil
}

func (p *Player) GetOutputError() error { return p.outputError }

func (p *Player) SetExecuteFunc(fn func() error) { p.executeFunc = fn }

func (p *Player) Execute() error {
	if p.executeFunc == nil {
		return nil
	}
	err := p.executeFunc()
	p.outputError = err
	return err
}
