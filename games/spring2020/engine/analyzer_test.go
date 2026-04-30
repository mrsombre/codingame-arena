package engine

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeTracesComparesWinnersAgainstLosers(t *testing.T) {
	analyzer, ok := NewFactory().(arena.TraceAnalyzer)
	require.True(t, ok)

	traceSummaryA := arena.TraceSummary{
		{
			TraceEat:    [][]int{{1, 2, 3}},
			TraceKilled: [][]int{{5}},
		},
		{
			TraceEat:    [][]int{{1}},
			TraceKilled: [][]int{{2, 3}},
		},
	}
	traceSummaryB := arena.TraceSummary{
		{
			TraceEat:    [][]int{{1}},
			TraceKilled: [][]int{{2, 3}},
		},
		{
			TraceEat:    [][]int{{1, 2, 3, 4}},
			TraceKilled: [][]int{{}},
		},
	}

	report, err := analyzer.AnalyzeTraces(arena.TraceAnalysisInput{
		TraceDir: "traces",
		Files: []arena.TraceFile{
			{
				Name: "trace-1-0.json",
				Trace: arena.TraceMatch{
					GameID:       "spring2020",
					Scores:       [2]arena.TraceScore{10, 4},
					Ranks:        [2]int{0, 1},
					TraceSummary: &traceSummaryA,
					Turns: []arena.TraceTurn{
						{Turn: 0, Output: [2]string{"SPEED 0 | MOVE 1 1 1", "MOVE 0 2 2"}},
						{Turn: 1, Output: [2]string{"MOVE 0 1 1", "SPEED 0"}},
					},
				},
			},
			{
				Name: "replay-2.json",
				Trace: arena.TraceMatch{
					GameID:       "spring2020",
					Scores:       [2]arena.TraceScore{5, 8},
					Ranks:        [2]int{1, 0},
					TraceSummary: &traceSummaryB,
					Turns: []arena.TraceTurn{
						{Turn: 0, Output: [2]string{"MOVE 0 1 1", "MOVE 0 2 2 | SPEED 1"}},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	var out bytes.Buffer
	require.NoError(t, report.Write(&out))
	text := out.String()
	assert.Contains(t, text, "Spring 2020 trace analysis: 2 trace files")
	assert.Contains(t, text, "Decided matches: 2  draws: 0")
	assert.Contains(t, text, "Side wins: p0=1 p1=1")
	assert.Contains(t, text, "Winner command rates")
	assert.Contains(t, text, "Winner pac events")
	assert.Contains(t, text, "EAT")
	assert.Contains(t, text, "KILLED")
	assert.Contains(t, text, "winner only 25% as often as loser")
}
