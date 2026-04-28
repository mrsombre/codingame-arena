// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Player.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Player.java:10-18

public class Player extends AbstractMultiplayerPlayer {
    private List<Pacman> pacmen = new ArrayList<>();
    public int pellets = 0;
    private boolean timeout;
    public Player() { pacmen = new ArrayList<>(); }
}
*/

type Player struct {
	Index                   int
	Score                   int
	Pellets                 int
	Pacmen                  []*Pacman
	Deactivated             bool
	DeactivationReasonValue string
	TimedOut                bool
	InputLines              []string
	Outputs                 []string
	OutputError             error
	ExecuteFunc             func() error
}

func NewPlayer(index int) *Player {
	return &Player{Index: index}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Player.java:20-35

public void addPacman(Pacman pacman)   { pacmen.add(pacman); }
public List<Pacman> getPacmen()        { return pacmen; }
public Stream<Pacman> getAlivePacmen() { return pacmen.stream().filter(pac -> !pac.isDead()); }
public Stream<Pacman> getDeadPacmen()  { return pacmen.stream().filter(pac -> pac.isDead()); }
*/

func (p *Player) AddPacman(pac *Pacman) {
	p.Pacmen = append(p.Pacmen, pac)
}

// AlivePacmen returns pacmen that are not dead, in insertion order.
func (p *Player) AlivePacmen() []*Pacman {
	out := make([]*Pacman, 0, len(p.Pacmen))
	for _, pac := range p.Pacmen {
		if !pac.Dead {
			out = append(out, pac)
		}
	}
	return out
}

// DeadPacmen returns dead pacmen in insertion order.
func (p *Player) DeadPacmen() []*Pacman {
	out := make([]*Pacman, 0)
	for _, pac := range p.Pacmen {
		if pac.Dead {
			out = append(out, pac)
		}
	}
	return out
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Player.java:42-44

public void turnReset() { pacmen.forEach(a -> a.turnReset()); }
*/

// TurnReset forwards to each owned pacman.
func (p *Player) TurnReset() {
	for _, pac := range p.Pacmen {
		pac.TurnReset()
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Player.java:37-40,46-52

@Override public int getExpectedOutputLines() { return 1; }
public boolean isTimedOut()                   { return timeout; }
public void setTimedOut(boolean timeout)      { this.timeout = timeout; }
*/

// arena.Player interface implementation.

func (p *Player) GetIndex() int  { return p.Index }
func (p *Player) GetScore() int  { return p.Score }
func (p *Player) SetScore(s int) { p.Score = s }
func (p *Player) IsDeactivated() bool {
	return p.Deactivated
}
func (p *Player) Deactivate(reason string) {
	p.Deactivated = true
	p.DeactivationReasonValue = reason
}
func (p *Player) DeactivationReason() string { return p.DeactivationReasonValue }
func (p *Player) IsTimedOut() bool           { return p.TimedOut }
func (p *Player) SetTimedOut(v bool)         { p.TimedOut = v }
func (p *Player) GetExpectedOutputLines() int {
	return 1
}
func (p *Player) SendInputLine(line string) {
	p.InputLines = append(p.InputLines, line)
}
func (p *Player) ConsumeInputLines() []string {
	lines := append([]string(nil), p.InputLines...)
	p.InputLines = p.InputLines[:0]
	return lines
}
func (p *Player) GetOutputs() []string { return p.Outputs }
func (p *Player) SetOutputs(lines []string) {
	p.Outputs = lines
	p.OutputError = nil
}
func (p *Player) GetOutputError() error { return p.OutputError }
func (p *Player) SetExecuteFunc(fn func() error) {
	p.ExecuteFunc = fn
}
func (p *Player) Execute() error {
	if p.ExecuteFunc == nil {
		return nil
	}
	err := p.ExecuteFunc()
	p.OutputError = err
	return err
}
