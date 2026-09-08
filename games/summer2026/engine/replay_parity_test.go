package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Fixtures: https://www.codingame.com/replay/901942746 (POIs) and
// https://www.codingame.com/replay/902038382 (no POIs). Setup comes from the
// raw replays' global graphics; arena's saved replays omit that field.
func TestOnlineReplayParity(t *testing.T) {
	for _, id := range []string{"901942746", "902038382"} {
		t.Run(id, func(t *testing.T) {
			checkOnlineReplayParity(t, id)
		})
	}
}

func checkOnlineReplayParity(t *testing.T, id string) {
	t.Helper()
	data, err := os.ReadFile("testdata/replay-" + id + ".json")
	require.NoError(t, err)
	var reference struct {
		Seed     int64 `json:"seed,string"`
		League   int
		Metadata map[string]any
		Setup    []string
		Scores   [2]int
		Ranks    [2]int
		Turns    []struct {
			Output [2]string
			Failed []string
		}
	}
	require.NoError(t, json.Unmarshal(data, &reference))
	var moves arena.ReplayMoves
	for _, turn := range reference.Turns {
		moves.Left = append(moves.Left, turn.Output[0])
		moves.Right = append(moves.Right, turn.Output[1])
	}
	replay := arena.CodinGameReplay[arena.CodinGameReplayFrame]{
		GameResult: arena.CodinGameReplayResult[arena.CodinGameReplayFrame]{Metadata: reference.Metadata},
	}
	options := arena.ReplayGameOptions(replay, reference.League)
	trace, scores := arena.RunReplay(NewFactory(), reference.Seed, options, moves, [2]string{"blue", "red"}, 0)
	assert.Equal(t, reference.Setup, trace.Setup)
	assert.Equal(t, reference.Scores, scores)
	assert.Equal(t, reference.Ranks, trace.Ranks)
	require.Len(t, trace.Turns, len(reference.Turns))
	assert.Equal(t, len(reference.Turns), trace.MainTurns)
	for i, turn := range trace.Turns {
		var failed []string
		for player, events := range turn.Traces {
			for _, event := range events {
				if event.Type == TraceFailed {
					failure, err := arena.DecodeData[FailedData](event)
					require.NoError(t, err)
					failed = append(failed, fmt.Sprintf("$%d %s", player, failure.Reason))
				}
			}
		}
		assert.ElementsMatch(t, reference.Turns[i].Failed, failed, "turn %d", i)
	}
}
