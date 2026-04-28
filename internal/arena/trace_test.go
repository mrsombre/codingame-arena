package arena

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraceWriterWritesMatchFile(t *testing.T) {
	dir := t.TempDir()
	const traceID int64 = 1717000000
	writer := NewTraceWriter(dir, traceID)

	match := TraceMatch{
		MatchID: 3,
		Seed:    12345,
		Scores:  [2]TraceScore{15, 12},
		Ranks:   [2]int{0, 1},
		Timing: &TraceTiming{
			FirstResponse:   [2]float64{820, 910},
			ResponseAverage: [2]float64{12, 14},
			ResponseMedian:  [2]float64{10, 13},
		},
		Turns: []TraceTurn{
			{
				Turn: 0,
				GameInput: traceTurnInput{
					P0: []string{"5 3 2", "apple 1 2"},
					P1: []string{"5 3 2", "apple 1 2"},
				},
				P0Output: "UP 0 RIGHT 1",
				P1Output: "DOWN 0 LEFT 1",
				Timing:   &TraceTurnTiming{Response: [2]float64{820, 910}},
				Events: []TurnEvent{
					{Label: "eat", Payload: "bot0:14.5"},
				},
			},
		},
	}

	require.NoError(t, writer.WriteMatch(match))

	path := filepath.Join(dir, fmt.Sprintf("trace-%d-3.json", traceID))
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got TraceMatch
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, traceID, got.TraceID)
	assert.Equal(t, TraceTypeTrace, got.Type)
	assert.Equal(t, fmt.Sprintf("trace-%d-3.json", traceID), got.File)
	assert.Equal(t, int64(12345), got.Seed)
	assert.Equal(t, [2]int{0, 1}, got.Ranks)
	require.NotNil(t, got.Timing)
	assert.Equal(t, [2]float64{820, 910}, got.Timing.FirstResponse)
	assert.Equal(t, [2]float64{12, 14}, got.Timing.ResponseAverage)
	assert.Equal(t, [2]float64{10, 13}, got.Timing.ResponseMedian)
	require.Len(t, got.Turns, 1)
	assert.Equal(t, "UP 0 RIGHT 1", got.Turns[0].P0Output)
	assert.Equal(t, "DOWN 0 LEFT 1", got.Turns[0].P1Output)
	require.NotNil(t, got.Turns[0].Timing)
	assert.Equal(t, [2]float64{820, 910}, got.Turns[0].Timing.Response)
	require.Len(t, got.Turns[0].Events, 1)
	assert.Equal(t, "eat", got.Turns[0].Events[0].Label)
}

func TestTraceWriterWritesReplayFile(t *testing.T) {
	dir := t.TempDir()
	const traceID int64 = 875143752
	writer := NewTraceWriter(dir, traceID)

	match := TraceMatch{
		MatchID: 0,
		Type:    TraceTypeReplay,
		Seed:    12345,
		Scores:  [2]TraceScore{15, 12},
		Ranks:   [2]int{0, 1},
	}

	require.NoError(t, writer.WriteMatch(match))

	path := filepath.Join(dir, fmt.Sprintf("replay-%d.json", traceID))
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got TraceMatch
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, traceID, got.TraceID)
	assert.Equal(t, TraceTypeReplay, got.Type)
	assert.Equal(t, fmt.Sprintf("replay-%d.json", traceID), got.File)
}

func TestTraceWriterNilIsNoop(t *testing.T) {
	var writer *TraceWriter
	assert.NoError(t, writer.WriteMatch(TraceMatch{}))
}
