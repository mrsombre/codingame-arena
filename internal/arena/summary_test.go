package arena

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeMatchesAggregatesMetrics(t *testing.T) {
	results := []MatchResult{
		{
			Metrics: []Metric{
				{Label: "score_p0", Value: 17},
				{Label: "ttfo_p0", Value: 111.11},
			},
		},
		{
			Metrics: []Metric{
				{Label: "score_p0", Value: 19},
				{Label: "ttfo_p0", Value: 111.05},
			},
		},
	}

	summary := SummarizeMatches(results)

	score := summary.Get("score_p0")
	require.NotNil(t, score)
	require.NotNil(t, score.Total)
	assert.Equal(t, 36.0, *score.Total)
	assert.Equal(t, 18.0, score.Avg)

	timing := summary.Get("ttfo_p0")
	require.NotNil(t, timing)
	assert.Nil(t, timing.Total)
	assert.Equal(t, 111.08, timing.Avg)
}

func TestFindWorstLossesReturnsTopN(t *testing.T) {
	results := []MatchResult{
		{Winner: 0, Metrics: []Metric{{Label: "wins_p0", Value: 1}, {Label: "loses_p0", Value: 0}, {Label: "score_p0", Value: 20}, {Label: "score_p1", Value: 10}}},
		{Winner: 1, Metrics: []Metric{{Label: "wins_p0", Value: 0}, {Label: "loses_p0", Value: 1}, {Label: "score_p0", Value: 5}, {Label: "score_p1", Value: 25}}},
		{Winner: 1, Metrics: []Metric{{Label: "wins_p0", Value: 0}, {Label: "loses_p0", Value: 1}, {Label: "score_p0", Value: 10}, {Label: "score_p1", Value: 15}}},
	}
	indices := FindWorstLosses(results, 1)
	require.Len(t, indices, 1)
	assert.Equal(t, 1, indices[0])
}

func TestWriteShortSummary(t *testing.T) {
	summary := MatchSummary{
		Simulations: 10,
		Metrics: []MetricSummary{
			{Label: "wins_p0", Avg: 0.7},
			{Label: "loses_p0", Avg: 0.2},
			{Label: "draws", Avg: 0.1},
			{Label: "score_p0", Avg: 15.3},
			{Label: "score_p1", Avg: 12.1},
			{Label: "ttfo_p0", Avg: 820},
			{Label: "aot_p0", Avg: 12},
			{Label: "ttfo_p1", Avg: 900},
			{Label: "aot_p1", Avg: 14},
		},
	}
	var buf bytes.Buffer
	err := WriteShortSummary(&buf, summary, 1234*time.Millisecond)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Summary: 10 matches played (1.234s)")
	assert.Contains(t, buf.String(), "Stats: wins=70%")
	assert.Contains(t, buf.String(), "losses=20%")
	assert.Contains(t, buf.String(), "Timing: avg_first_response=820msx900ms")
	assert.Contains(t, buf.String(), "avg_turn_response=12msx14ms")
}
