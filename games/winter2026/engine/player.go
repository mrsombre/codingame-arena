// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Player.java
package engine

import "github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"

type Player struct {
	index              int
	score              int
	points             int
	birds              []*Bird
	marks              []grid.Coord
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

func (p *Player) Init() {
	p.birds = make([]*Bird, 0)
	p.marks = make([]grid.Coord, 0)
}

func (p *Player) Reset() {
	for _, bird := range p.birds {
		bird.Direction = grid.DirUnset
		bird.HasMove = false
		bird.Message = ""
		bird.HasMsg = false
	}
	p.marks = p.marks[:0]
}

func (p *Player) AddScore(points int) {
	p.points += points
	p.SetScore(p.points)
}

func (p *Player) BirdByID(id int) *Bird {
	for _, bird := range p.birds {
		if bird.ID == id {
			return bird
		}
	}
	return nil
}

func (p *Player) AddMark(coord grid.Coord) bool {
	if len(p.marks) < 4 {
		p.marks = append(p.marks, coord)
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
