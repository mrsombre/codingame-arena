package arena

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraceWriterWritesMatchFile(t *testing.T) {
	dir := t.TempDir()
	writer := NewTraceWriter(dir)

	match := TraceMatch{
		MatchID: 3,
		Seed:    12345,
		Winner:  0,
		Scores:  [2]int{15, 12},
		Turns: []TraceTurn{
			{
				Turn: 0,
				GameInput: traceTurnInput{
					P0: []string{"5 3 2", "apple 1 2"},
					P1: []string{"5 3 2", "apple 1 2"},
				},
				P0Output: "UP 0 RIGHT 1",
				P1Output: "DOWN 0 LEFT 1",
				Events: []TurnEvent{
					{Label: "eat", Payload: "bot0:14.5"},
				},
			},
		},
	}

	require.NoError(t, writer.WriteMatch(match))

	path := filepath.Join(dir, "3.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got TraceMatch
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, int64(12345), got.Seed)
	assert.Equal(t, 0, got.Winner)
	require.Len(t, got.Turns, 1)
	assert.Equal(t, "UP 0 RIGHT 1", got.Turns[0].P0Output)
	assert.Equal(t, "DOWN 0 LEFT 1", got.Turns[0].P1Output)
	require.Len(t, got.Turns[0].Events, 1)
	assert.Equal(t, "eat", got.Turns[0].Events[0].Label)
}

func TestTraceWriterNilIsNoop(t *testing.T) {
	var writer *TraceWriter
	assert.NoError(t, writer.WriteMatch(TraceMatch{}))
}
