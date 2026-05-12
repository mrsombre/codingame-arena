// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java
package engine

// Player is the engine-side handle for one agent. Implements arena.Player.
// Game-specific fields (inventory, owned units, etc.) will be added when the
// engine is ported; this file currently provides only the boilerplate that
// satisfies the arena.Player interface.
type Player struct {
	index              int
	score              int
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

// GetExpectedOutputLines: troll bots emit a single semicolon-joined line per
// turn, matching the upstream protocol.
func (p *Player) GetExpectedOutputLines() int { return 1 }

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
