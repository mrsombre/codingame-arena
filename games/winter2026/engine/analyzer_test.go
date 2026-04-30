package engine

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeTracesReportsEndReasonsAndBlueFaults(t *testing.T) {
	analyzer, ok := NewFactory().(arena.TraceAnalyzer)
	require.True(t, ok)

	files := []arena.TraceFile{
		// Blue (p0=our_bot) wins on score.
		{
			Name: "trace-1.json",
			Trace: arena.TraceMatch{
				GameID:    "winter2026",
				Blue:      "our_bot",
				EndReason: arena.EndReasonScore,
				Scores:    [2]arena.TraceScore{12, 7},
				Ranks:     [2]int{0, 1},
				Players:   [2]string{"our_bot", "rival"},
			},
		},
		// Blue times out — our fault.
		{
			Name: "trace-2.json",
			Trace: arena.TraceMatch{
				GameID:      "winter2026",
				Blue:        "our_bot",
				EndReason:   arena.EndReasonTimeout,
				Deactivated: [2]bool{true, false},
				Scores:      [2]arena.TraceScore{0, 5},
				Ranks:       [2]int{1, 0},
				Players:     [2]string{"our_bot", "rival"},
			},
		},
		// Opponent disqualified.
		{
			Name: "trace-3.json",
			Trace: arena.TraceMatch{
				GameID:      "winter2026",
				Blue:        "our_bot",
				EndReason:   arena.EndReasonInvalid,
				Deactivated: [2]bool{false, true},
				Scores:      [2]arena.TraceScore{6, 0},
				Ranks:       [2]int{0, 1},
				Players:     [2]string{"our_bot", "rival"},
			},
		},
		// Match capped on turns (no fault).
		{
			Name: "trace-4.json",
			Trace: arena.TraceMatch{
				GameID:    "winter2026",
				Blue:      "our_bot",
				EndReason: arena.EndReasonTurnsOut,
				Scores:    [2]arena.TraceScore{4, 4},
				Ranks:     [2]int{0, 0},
				Players:   [2]string{"our_bot", "rival"},
			},
		},
		// Blue plays on side 1 due to swap; blue gets eliminated.
		{
			Name: "trace-5.json",
			Trace: arena.TraceMatch{
				GameID:    "winter2026",
				Blue:      "our_bot",
				EndReason: arena.EndReasonEliminated,
				Scores:    [2]arena.TraceScore{8, 0},
				Ranks:     [2]int{0, 1},
				Players:   [2]string{"rival", "our_bot"},
			},
		},
	}

	report, err := analyzer.AnalyzeTraces(arena.TraceAnalysisInput{
		TraceDir: "traces",
		Files:    files,
	})
	require.NoError(t, err)

	var out bytes.Buffer
	require.NoError(t, report.Write(&out))
	text := out.String()

	assert.Contains(t, text, "Winter 2026 trace analysis: 5 trace files")
	assert.Contains(t, text, "Decided matches: 4  draws: 1")
	// Blue stats: 5 matches, 2 wins (trace-1, trace-3), 2 losses (trace-2, trace-5), 1 draw.
	assert.Contains(t, text, "Blue side: matches=5  wins=2 (40.0%)  losses=2  draws=1")

	assert.Contains(t, text, "End reasons")
	assert.Contains(t, text, "TIMEOUT")
	assert.Contains(t, text, "INVALID")
	assert.Contains(t, text, "SCORE")
	assert.Contains(t, text, "TURNS_OUT")
	assert.Contains(t, text, "ELIMINATED")
	// Fault rows include blue/opponent breakdown; non-fault rows do not.
	assert.Contains(t, text, "(blue: 1, opponent: 0)") // TIMEOUT row
	assert.Contains(t, text, "(blue: 0, opponent: 1)") // INVALID row

	// Blue-fault summary: 1 timeout, 0 disqualifications, 1 of 5 matches (20%).
	assert.Contains(t, text, "Blue-side faults: 1 / 5 matches (20.0%) — timeouts=1 disqualifications=0")
	assert.Contains(t, text, "Of 2 fault matches: blue=1 (50.0%)  opponent=1 (50.0%)  unknown=0")
}

func TestAnalyzeTracesWithoutBlueOmitsBlueSummary(t *testing.T) {
	analyzer, ok := NewFactory().(arena.TraceAnalyzer)
	require.True(t, ok)

	report, err := analyzer.AnalyzeTraces(arena.TraceAnalysisInput{
		TraceDir: "traces",
		Files: []arena.TraceFile{
			{
				Name: "trace-1.json",
				Trace: arena.TraceMatch{
					GameID:    "winter2026",
					EndReason: arena.EndReasonScore,
					Scores:    [2]arena.TraceScore{5, 3},
					Ranks:     [2]int{0, 1},
				},
			},
		},
	})
	require.NoError(t, err)

	var out bytes.Buffer
	require.NoError(t, report.Write(&out))
	text := out.String()

	assert.NotContains(t, text, "Blue side: matches=")
	assert.Contains(t, text, "Blue not identified in traces")
}

func TestAnalyzeTracesEmptyInput(t *testing.T) {
	analyzer, ok := NewFactory().(arena.TraceAnalyzer)
	require.True(t, ok)

	report, err := analyzer.AnalyzeTraces(arena.TraceAnalysisInput{TraceDir: "traces"})
	require.NoError(t, err)

	var out bytes.Buffer
	require.NoError(t, report.Write(&out))
	text := out.String()
	assert.Contains(t, text, "Winter 2026 trace analysis: 0 trace files")
	assert.Contains(t, text, "none recorded")
}
