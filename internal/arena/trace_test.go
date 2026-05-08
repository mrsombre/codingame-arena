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
		MatchID:     3,
		Seed:        12345,
		CreatedAt:   "2026-04-29T18:34:46Z",
		Scores:      [2]TraceScore{15, 12},
		FinalScores: [2]TraceScore{17, 12},
		Ranks:       [2]int{0, 1},
		Timing: &TraceTiming{
			FirstResponse:   [2]float64{820, 910},
			ResponseAverage: [2]float64{12, 14},
			ResponseMedian:  [2]float64{10, 13},
		},
		Turns: []TraceTurn{
			{
				Turn:      0,
				GameInput: []string{"5 3 2", "apple 1 2"},
				Output:    [2]string{"UP 0 RIGHT 1", "DOWN 0 LEFT 1"},
				Timing:    &TraceTurnTiming{Response: [2]float64{820, 910}},
				Traces: [2][]TurnTrace{
					{MakeTurnTrace("eat", map[string]any{"bot": "bot0", "score": 14.5})},
					nil,
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
	assert.Equal(t, int64(12345), got.Seed)
	assert.Equal(t, "2026-04-29T18:34:46Z", got.CreatedAt)
	assert.Equal(t, [2]TraceScore{15, 12}, got.Scores)
	assert.Equal(t, [2]TraceScore{17, 12}, got.FinalScores)
	assert.Equal(t, [2]int{0, 1}, got.Ranks)
	require.NotNil(t, got.Timing)
	assert.Equal(t, [2]float64{820, 910}, got.Timing.FirstResponse)
	assert.Equal(t, [2]float64{12, 14}, got.Timing.ResponseAverage)
	assert.Equal(t, [2]float64{10, 13}, got.Timing.ResponseMedian)
	require.Len(t, got.Turns, 1)
	assert.Equal(t, [2]string{"UP 0 RIGHT 1", "DOWN 0 LEFT 1"}, got.Turns[0].Output)
	require.NotNil(t, got.Turns[0].Timing)
	assert.Equal(t, [2]float64{820, 910}, got.Turns[0].Timing.Response)
	require.Len(t, got.Turns[0].Traces[0], 1)
	assert.Empty(t, got.Turns[0].Traces[1])
	assert.Equal(t, "eat", got.Turns[0].Traces[0][0].Type)
	decoded, err := DecodeData[map[string]any](got.Turns[0].Traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, "bot0", decoded["bot"])
	assert.InDelta(t, 14.5, decoded["score"], 0.0001)
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
}

// Setup carries the raw global-info lines blue's bot received on stdin —
// game-specific format, captured verbatim. Verify the field round-trips
// through the trace file.
func TestTraceWriterPreservesSetup(t *testing.T) {
	dir := t.TempDir()
	writer := NewTraceWriter(dir, 1)

	match := TraceMatch{
		MatchID: 7,
		Setup:   []string{"37", "0 3 1 2 3 4 5 6", "1 0 7 8 2 0 6 18"},
	}
	require.NoError(t, writer.WriteMatch(match))

	data, err := os.ReadFile(filepath.Join(dir, "trace-1-7.json"))
	require.NoError(t, err)

	var got TraceMatch
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, match.Setup, got.Setup)
}

func TestTraceWriterNilIsNoop(t *testing.T) {
	var writer *TraceWriter
	assert.NoError(t, writer.WriteMatch(TraceMatch{}))
}

func TestTraceMatchBlueSide(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		blue string
		ps   [2]string
		want int
	}{
		{name: "blue is left", blue: "bot-cpp", ps: [2]string{"bot-cpp", "bot-py"}, want: 0},
		{name: "blue is right (post-swap)", blue: "bot-cpp", ps: [2]string{"bot-py", "bot-cpp"}, want: 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := TraceMatch{Blue: tc.blue, Players: tc.ps}
			assert.Equal(t, tc.want, m.BlueSide())
		})
	}
}
