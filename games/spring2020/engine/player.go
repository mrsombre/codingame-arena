// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Player.java
package engine

type Player struct {
	index              int
	score              int
	Pellets            int
	Pacmen             []*Pacman
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

// TurnReset forwards to each owned pacman.
func (p *Player) TurnReset() {
	for _, pac := range p.Pacmen {
		pac.TurnReset()
	}
}

// arena.Player interface implementation.

func (p *Player) GetIndex() int  { return p.index }
func (p *Player) GetScore() int  { return p.score }
func (p *Player) SetScore(s int) { p.score = s }
func (p *Player) IsDeactivated() bool {
	return p.deactivated
}
func (p *Player) Deactivate(reason string) {
	p.deactivated = true
	p.deactivationReason = reason
}
func (p *Player) DeactivationReason() string { return p.deactivationReason }
func (p *Player) IsTimedOut() bool           { return p.timedOut }
func (p *Player) SetTimedOut(v bool)         { p.timedOut = v }
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
func (p *Player) GetOutputs() []string { return p.outputs }
func (p *Player) SetOutputs(lines []string) {
	p.outputs = lines
	p.outputError = nil
}
func (p *Player) GetOutputError() error { return p.outputError }
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
