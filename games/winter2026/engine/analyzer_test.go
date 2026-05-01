package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeWinterTraceMetricsCountsEverySegmentLossEvent(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Output: [2]string{"0 UP;1 UP", "2 UP;3 UP"},
				Traces: []arena.TurnTrace{
					// Two HIT_WALL on the same side in one turn must both count
					// — each event is one lost segment.
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 0, Coord: [2]int{5, 5}}),
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 1, Coord: [2]int{4, 5}}),
					arena.MakeTurnTrace(TraceHitEnemy, BirdCoordMeta{Bird: 2, Coord: [2]int{1, 1}}),
					arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: 0, Coord: [2]int{2, 2}}),
					arena.MakeTurnTrace(TraceDead, BirdMeta{Bird: 3}),
					arena.MakeTurnTrace(TraceFall, BirdSegmentsMeta{Bird: 2, Segments: 5}),
				},
			},
			{
				Output: [2]string{"0 UP;1 UP", "2 UP;3 UP"},
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 2, Coord: [2]int{0, 0}}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{0, 1}, stats[TraceDead])
	assert.Equal(t, [2]int{0, 1}, stats[TraceHitEnemy])
	assert.Equal(t, [2]int{2, 1}, stats[TraceHitWall])
	assert.Equal(t, [2]int{0, 5}, stats[TraceFall])
}

func TestAnalyzeWinterTraceMetricsWithoutBirdMappingReturnsZeroes(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Traces: []arena.TurnTrace{
				arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 0, Coord: [2]int{5, 5}}),
			}},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{}, stats[TraceHitWall])
}

func TestAnalyzeWinterTraceMetricsSkipsUnknownBirdsAndMalformedMeta(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Output: [2]string{"0 UP", "2 UP"},
				Traces: []arena.TurnTrace{
					{Type: TraceHitWall, Meta: json.RawMessage(`{"bird": "MARK"}`)},
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 99, Coord: [2]int{7, 7}}),
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 0, Coord: [2]int{1, 1}}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{1, 0}, stats[TraceHitWall])
}

func TestWinterTraceMetricSpecsListExpectedMetrics(t *testing.T) {
	specs := NewFactory().(arena.TraceMetricAnalyzer).TraceMetricSpecs()

	require.Len(t, specs, 5)
	assert.Equal(t,
		[]string{TraceDead, TraceHitEnemy, TraceHitWall, TraceHitSelf, TraceFall},
		metricSpecKeys(specs),
	)
	for _, spec := range specs {
		assert.Equal(t, arena.TraceMetricPerMatchCount, spec.Kind, spec.Key)
	}
}

func metricSpecKeys(metrics []arena.TraceMetricSpec) []string {
	keys := make([]string, len(metrics))
	for i, m := range metrics {
		keys[i] = m.Key
	}
	return keys
}
