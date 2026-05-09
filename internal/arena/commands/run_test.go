package commands

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestBuildRunnerOutputAggregatesRunMetadata(t *testing.T) {
	opts := RunOptions{
		BatchOptions: arena.BatchOptions{
			Simulations:   3,
			Parallel:      2,
			Seed:          100,
			SeedIncrement: 7,
			OutputMatches: true,
		},
		BlueBotBin: "./bin/blue",
		RedBotBin:  "./bin/red",
		MaxTurns:   123,
		Trace:      true,
		TraceDir:   "./tmp/traces",
		NoSwap:     true,
	}
	results := []arena.MatchResult{
		matchResult(0, 100, false, 30, 20),
		matchResult(1, -107, true, 4, 18),
		{
			ID:      2,
			Seed:    114,
			Swapped: false,
			Metrics: resultMetrics(0, 1, 10, 40),
			BadCommands: []arena.BadCommandInfo{
				{Seed: 114, Player: 0, Turn: 3, Command: "WAIT", Reason: "invalid action"},
			},
		},
	}

	out := buildRunnerOutput(opts, results)

	assert.Equal(t, "./bin/blue", out.BlueBotBin)
	assert.Equal(t, "./bin/red", out.RedBotBin)
	assert.Equal(t, 3, out.Runner.Simulations)
	assert.Equal(t, 2, out.Runner.Parallel)
	assert.Equal(t, int64(100), out.Runner.Seed)
	assert.Equal(t, int64(7), out.Runner.SeedIncrement)
	assert.Equal(t, "./tmp/traces", out.Runner.TraceDir)
	assert.Equal(t, 123, out.Runner.MaxTurns)
	assert.True(t, out.Runner.NoSwap)
	assert.Equal(t, 2, out.Runner.BlueLeft)
	assert.Equal(t, 1, out.Runner.BlueRight)
	assert.Len(t, out.BadCommands, 1)
	assert.Len(t, out.WorstLosses, 2)
	assert.Len(t, out.Matches, 3)
	assert.Equal(t, 3, out.Summary.Simulations)
}

func TestBuildRunnerOutputOmitsOptionalFieldsWhenDisabled(t *testing.T) {
	opts := RunOptions{
		BatchOptions: arena.BatchOptions{Simulations: 1, Parallel: 1, SeedIncrement: 1},
		BlueBotBin:   "./bin/blue",
		RedBotBin:    "./bin/red",
		MaxTurns:     200,
	}

	out := buildRunnerOutput(opts, []arena.MatchResult{matchResult(0, 1, false, 20, 10)})

	assert.Empty(t, out.Runner.TraceDir)
	assert.Empty(t, out.BadCommands)
	assert.Empty(t, out.WorstLosses)
	assert.Empty(t, out.Matches)
}

func TestWriteRunOutputWritesVerboseJSON(t *testing.T) {
	opts := RunOptions{
		BatchOptions: arena.BatchOptions{
			Simulations:   1,
			Parallel:      1,
			SeedIncrement: 1,
		},
		BlueBotBin: "./bin/blue",
		RedBotBin:  "./bin/red",
		MaxTurns:   200,
		Verbose:    true,
	}
	var stdout bytes.Buffer

	err := writeRunOutput(&stdout, opts, []arena.MatchResult{matchResult(0, 1, false, 20, 10)}, time.Second, nil)

	require.NoError(t, err)
	var out runnerOutput
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	assert.Equal(t, 1, out.Runner.Simulations)
	assert.Equal(t, 1, out.Summary.Simulations)
}

func TestWriteRunOutputDebugWritesCapturedTraceJSON(t *testing.T) {
	opts := RunOptions{
		BatchOptions: arena.BatchOptions{Simulations: 1, Parallel: 1, SeedIncrement: 1},
		BlueBotBin:   "./bin/blue",
		RedBotBin:    "./bin/red",
		MaxTurns:     200,
		Debug:        true,
	}
	debugSink := &debugTraceCapture{traceID: 42}
	debugSink.match = arena.TraceMatch{
		PuzzleName: "winter2026",
		Players:    [2]string{"blue", "red"},
		Seed:       12345,
		MatchID:    0,
	}
	var stdout bytes.Buffer

	err := writeRunOutput(&stdout, opts, []arena.MatchResult{matchResult(0, 12345, false, 20, 10)}, time.Second, debugSink)

	require.NoError(t, err)
	var trace arena.TraceMatch
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &trace))
	assert.Equal(t, "winter2026", trace.PuzzleName)
	assert.Equal(t, int64(12345), trace.Seed)
}

func TestDebugTraceCaptureStampsTraceIDAndDefaultsType(t *testing.T) {
	sink := &debugTraceCapture{traceID: 7}
	require.NoError(t, sink.WriteMatch(arena.TraceMatch{Seed: 99}))
	assert.Equal(t, int64(7), sink.match.TraceID)
	assert.Equal(t, arena.TraceTypeTrace, sink.match.Type)
}

func TestWriteRunOutputWritesShortSummaryByDefault(t *testing.T) {
	opts := RunOptions{
		BatchOptions: arena.BatchOptions{
			Simulations:   1,
			Parallel:      1,
			SeedIncrement: 1,
		},
		BlueBotBin: "./bin/blue",
		RedBotBin:  "./bin/red",
		MaxTurns:   200,
	}
	var stdout bytes.Buffer

	err := writeRunOutput(&stdout, opts, []arena.MatchResult{matchResult(0, 1, false, 20, 10)}, time.Second, nil)

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Summary: 1 matches played")
	assert.Contains(t, stdout.String(), "Stats: wins=100%")
}

func matchResult(id int, seed int64, swapped bool, blueScore int, redScore int) arena.MatchResult {
	winner := 0
	blueWins, blueLoses := 1.0, 0.0
	if redScore > blueScore {
		winner = 1
		blueWins, blueLoses = 0, 1
	}

	return arena.MatchResult{
		ID:      id,
		Seed:    seed,
		Winner:  winner,
		Scores:  [2]int{blueScore, redScore},
		Swapped: swapped,
		Metrics: resultMetrics(blueWins, blueLoses, float64(blueScore), float64(redScore)),
	}
}

func resultMetrics(blueWins, blueLoses, blueScore, redScore float64) []arena.Metric {
	draws := 0.0
	if blueWins == 0 && blueLoses == 0 {
		draws = 1
	}
	return []arena.Metric{
		{Label: "wins_blue", Value: blueWins},
		{Label: "loses_blue", Value: blueLoses},
		{Label: "draws", Value: draws},
		{Label: "score_blue", Value: blueScore},
		{Label: "score_red", Value: redScore},
		{Label: "turns", Value: 10},
		{Label: "ttfo_blue", Value: 1},
		{Label: "ttfo_red", Value: 2},
		{Label: "aot_blue", Value: 3},
		{Label: "aot_red", Value: 4},
	}
}
