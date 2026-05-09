package arena

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteDebugStderrEmitsHeaderAndLines(t *testing.T) {
	var buf bytes.Buffer
	writeDebugStderr(&buf, 7, 1, []string{"hello", "world"})
	assert.Equal(t, "--- turn 7 right stderr ---\nhello\nworld\n", buf.String())
}

func TestWriteDebugStderrLeftSide(t *testing.T) {
	var buf bytes.Buffer
	writeDebugStderr(&buf, 0, 0, []string{"x"})
	assert.Equal(t, "--- turn 0 left stderr ---\nx\n", buf.String())
}

func TestWriteDebugStderrEmptyIsNoOp(t *testing.T) {
	var buf bytes.Buffer
	writeDebugStderr(&buf, 7, 0, nil)
	assert.Empty(t, buf.String())
}

type mockPlayer struct {
	index               int
	score               int
	deactivated         bool
	deactivationReason  string
	timedOut            bool
	expectedOutputLines int
	inputLines          []string
	outputs             []string
	outputError         error
	executeFn           func() error
}

func (p *mockPlayer) GetIndex() int              { return p.index }
func (p *mockPlayer) GetScore() int              { return p.score }
func (p *mockPlayer) SetScore(s int)             { p.score = s }
func (p *mockPlayer) IsDeactivated() bool        { return p.deactivated }
func (p *mockPlayer) Deactivate(reason string)   { p.deactivated = true; p.deactivationReason = reason }
func (p *mockPlayer) DeactivationReason() string { return p.deactivationReason }
func (p *mockPlayer) IsTimedOut() bool           { return p.timedOut }
func (p *mockPlayer) SetTimedOut(v bool)         { p.timedOut = v }
func (p *mockPlayer) GetExpectedOutputLines() int {
	if p.expectedOutputLines == 0 {
		return 1
	}
	return p.expectedOutputLines
}
func (p *mockPlayer) SendInputLine(line string)      { p.inputLines = append(p.inputLines, line) }
func (p *mockPlayer) ConsumeInputLines() []string    { l := p.inputLines; p.inputLines = nil; return l }
func (p *mockPlayer) GetOutputs() []string           { return p.outputs }
func (p *mockPlayer) SetOutputs(o []string)          { p.outputs = o }
func (p *mockPlayer) GetOutputError() error          { return p.outputError }
func (p *mockPlayer) SetExecuteFunc(fn func() error) { p.executeFn = fn }
func (p *mockPlayer) Execute() error {
	if p.executeFn != nil {
		return p.executeFn()
	}
	return nil
}

func TestLossReasonForTimeout(t *testing.T) {
	p := &mockPlayer{timedOut: true}
	assert.Equal(t, LossReasonTimeout, lossReasonFor(p, 1, 0))
}

func TestLossReasonForBadCommand(t *testing.T) {
	p := &mockPlayer{deactivated: true}
	assert.Equal(t, LossReasonBadCommand, lossReasonFor(p, 1, 0))
}

func TestLossReasonForScore(t *testing.T) {
	p := &mockPlayer{}
	assert.Equal(t, LossReasonScore, lossReasonFor(p, 1, 0))
}

func TestLossReasonForNone(t *testing.T) {
	p := &mockPlayer{}
	assert.Equal(t, LossReasonNone, lossReasonFor(p, 0, 0))
}

func TestLossReasonForDraw(t *testing.T) {
	p := &mockPlayer{}
	assert.Equal(t, LossReasonNone, lossReasonFor(p, -1, 0))
}

func TestSwapMatchSidesSwapsScoresAndWinner(t *testing.T) {
	r := MatchResult{
		Scores:            [2]int{10, 20},
		Winner:            0,
		LossReasons:       [2]LossReason{LossReasonNone, LossReasonScore},
		TimeToFirstOutput: [2]time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
		AverageOutputTime: [2]time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
		Metrics: []Metric{
			{Label: "wins_blue", Value: 1},
			{Label: "wins_red", Value: 0},
			{Label: "loses_blue", Value: 0},
			{Label: "loses_red", Value: 1},
			{Label: "score_blue", Value: 10},
			{Label: "score_red", Value: 20},
		},
	}

	swapped := swapMatchSides(r)

	assert.Equal(t, [2]int{20, 10}, swapped.Scores)
	assert.Equal(t, 1, swapped.Winner)
	assert.Equal(t, LossReasonScore, swapped.LossReasons[0])
	assert.Equal(t, LossReasonNone, swapped.LossReasons[1])
	assert.Equal(t, [2]float64{200, 100}, swapped.TTFO())
	assert.Equal(t, [2]float64{20, 10}, swapped.AOT())

	metricMap := make(map[string]float64)
	for _, m := range swapped.Metrics {
		metricMap[m.Label] = m.Value
	}
	assert.Equal(t, 0.0, metricMap["wins_blue"])
	assert.Equal(t, 1.0, metricMap["wins_red"])
	assert.Equal(t, 1.0, metricMap["loses_blue"])
	assert.Equal(t, 0.0, metricMap["loses_red"])
	assert.Equal(t, 20.0, metricMap["score_blue"])
	assert.Equal(t, 10.0, metricMap["score_red"])
}

func TestSwapMatchSidesBadCommands(t *testing.T) {
	r := MatchResult{
		Winner: -1,
		BadCommands: []BadCommandInfo{
			{Player: 0, Turn: 5},
		},
	}
	swapped := swapMatchSides(r)
	assert.Equal(t, 1, swapped.BadCommands[0].Player)
}
