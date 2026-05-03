// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Player.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Player.java:6-17

public class Player extends AbstractMultiplayerPlayer {
    private String message;
    private Action action;
    private int sun;
    private boolean waiting = false;
    private int bonusScore = 0;

    public Player() {
        sun = Config.STARTING_SUN;
        action = Action.NO_ACTION;
    }
}
*/

// Player carries both the simulation state Java's Player has (sun, action,
// message, waiting, bonus) and the I/O bookkeeping the arena needs. We keep
// it as a single type because Java's Player extends AbstractMultiplayerPlayer
// and the arena interface methods (Deactivate, Outputs, etc.) sit on the same
// instance.
type Player struct {
	index int

	score      int
	Sun        int
	Action     Action
	Message    string
	HasMessage bool
	Waiting    bool
	BonusScore int

	deactivated        bool
	deactivationReason string
	timedOut           bool
	inputLines         []string
	outputs            []string
	outputError        error
	executeFunc        func() error
}

func NewPlayer(index int) *Player {
	return &Player{
		index:  index,
		Action: NoAction,
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Player.java:24-31

public void addScore(int score) { setScore(getScore() + score); }
public void reset() {
    message = null;
    action = Action.NO_ACTION;
}
*/

func (p *Player) AddScore(score int) {
	p.SetScore(p.GetScore() + score)
}

func (p *Player) Reset() {
	p.Message = ""
	p.HasMessage = false
	p.Action = NoAction
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Player.java:33-72

public String getMessage()        { return message; }
public void setMessage(String m)  { this.message = m; }
public void setAction(Action a)   { this.action = a; }
public Action getAction()         { return action; }
public int getSun()               { return sun; }
public void setSun(int sun)       { this.sun = sun; }
public void addSun(int sun)       { this.sun += sun; }
public void removeSun(int amount) { this.sun = Math.max(0, this.sun - amount); }
public boolean isWaiting()        { return waiting; }
public void setWaiting(boolean w) { this.waiting = w; }
*/

func (p *Player) SetMessage(message string) {
	p.Message = message
	p.HasMessage = true
}

func (p *Player) SetAction(a Action) { p.Action = a }
func (p *Player) GetAction() Action  { return p.Action }
func (p *Player) GetSun() int        { return p.Sun }
func (p *Player) SetSun(s int)       { p.Sun = s }
func (p *Player) AddSun(s int)       { p.Sun += s }
func (p *Player) RemoveSun(amount int) {
	v := p.Sun - amount
	if v < 0 {
		v = 0
	}
	p.Sun = v
}
func (p *Player) IsWaiting() bool   { return p.Waiting }
func (p *Player) SetWaiting(w bool) { p.Waiting = w }

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Player.java:73-87

public String getBonusScore() {
    if (bonusScore > 0) {
        return String.format("%d points and %d trees", getScore() - bonusScore, bonusScore);
    } else {
        return "";
    }
}
public void addBonusScore(int bonusScore) { this.bonusScore += bonusScore; }
*/

func (p *Player) GetBonusScoreText() string {
	if p.BonusScore > 0 {
		return fmt.Sprintf("%d points and %d trees", p.GetScore()-p.BonusScore, p.BonusScore)
	}
	return ""
}

func (p *Player) AddBonusScore(b int) {
	p.BonusScore += b
}

// arena.Player interface implementation.

func (p *Player) GetIndex() int       { return p.index }
func (p *Player) GetScore() int       { return p.score }
func (p *Player) SetScore(score int)  { p.score = score }
func (p *Player) IsDeactivated() bool { return p.deactivated }
func (p *Player) Deactivate(reason string) {
	p.deactivated = true
	p.deactivationReason = reason
}
func (p *Player) DeactivationReason() string { return p.deactivationReason }
func (p *Player) IsTimedOut() bool           { return p.timedOut }
func (p *Player) SetTimedOut(v bool)         { p.timedOut = v }

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Player.java:19-22

@Override
public int getExpectedOutputLines() { return 1; }
*/

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
func (p *Player) SetOutputs(lines []string) {
	p.outputs = lines
	p.outputError = nil
}
func (p *Player) GetOutputError() error      { return p.outputError }
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

// NicknameToken returns the placeholder the Java engine substitutes into
// summary lines (e.g. "$0", "$1"). Mirrors Player::getNicknameToken.
func (p *Player) NicknameToken() string {
	return fmt.Sprintf("$%d", p.index)
}

// IsActive mirrors Java's `player.isActive()` — true while the player has not
// been deactivated.
func (p *Player) IsActive() bool {
	return !p.deactivated
}
