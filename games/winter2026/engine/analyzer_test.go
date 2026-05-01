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
				Turn:   0,
				Output: [2]string{"0 UP;1 UP", "2 UP;3 UP"},
				Traces: []arena.TurnTrace{
					// Two HIT_WALL on the same side in one turn must both count
					// — each event is one lost segment.
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 0, Coord: [2]int{5, 5}}),
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 1, Coord: [2]int{4, 5}}),
					arena.MakeTurnTrace(TraceHitEnemy, BirdCoordMeta{Bird: 2, Coord: [2]int{1, 1}}),
					// Turn 0 < cutoff: this EAT counts toward EAT_T20.
					arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: 0, Coord: [2]int{2, 2}}),
					arena.MakeTurnTrace(TraceDead, BirdDeathMeta{Bird: 3, Cause: DeathCauseEnemy}),
					arena.MakeTurnTrace(TraceDeadFall, BirdSegmentsMeta{Bird: 2, Segments: 5}),
				},
			},
			{
				Turn:   1,
				Output: [2]string{"0 UP;1 UP", "2 UP;3 UP"},
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: 2, Coord: [2]int{0, 0}}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	// Bird 3 died with cause ENEMY → DEAD_ENEMY for side 1; DEAD_WALL stays 0.
	assert.Equal(t, [2]int{0, 0}, stats[winterMetricDeadWall])
	assert.Equal(t, [2]int{0, 1}, stats[winterMetricDeadEnemy])
	assert.Equal(t, [2]int{0, 0}, stats[winterMetricDeadSelf])
	assert.Equal(t, [2]int{0, 1}, stats[TraceHitEnemy])
	assert.Equal(t, [2]int{2, 1}, stats[TraceHitWall])
	assert.Equal(t, [2]int{0, 5}, stats[TraceDeadFall])
	assert.Equal(t, [2]int{1, 0}, stats[winterMetricEatByTurn20])
}

func TestAnalyzeWinterTraceMetricsAttributesEachDeathCause(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Output: [2]string{"0 UP;1 UP", "2 UP;3 UP"},
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceDead, BirdDeathMeta{Bird: 0, Cause: DeathCauseWall}),
					arena.MakeTurnTrace(TraceDead, BirdDeathMeta{Bird: 1, Cause: DeathCauseSelf}),
					arena.MakeTurnTrace(TraceDead, BirdDeathMeta{Bird: 2, Cause: DeathCauseEnemy}),
					arena.MakeTurnTrace(TraceDead, BirdDeathMeta{Bird: 3, Cause: DeathCauseSelf}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{1, 0}, stats[winterMetricDeadWall])
	assert.Equal(t, [2]int{0, 1}, stats[winterMetricDeadEnemy])
	assert.Equal(t, [2]int{1, 1}, stats[winterMetricDeadSelf])
}

func TestAnalyzeWinterTraceMetricsDropsDeadWithUnknownCause(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Output: [2]string{"0 UP", "2 UP"},
				Traces: []arena.TurnTrace{
					// Empty cause (e.g. legacy BirdMeta-shaped DEAD trace from
					// before the cause field existed) must be dropped, not
					// miscategorized into a side bucket.
					arena.MakeTurnTrace(TraceDead, BirdMeta{Bird: 0}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{}, stats[winterMetricDeadWall])
	assert.Equal(t, [2]int{}, stats[winterMetricDeadEnemy])
	assert.Equal(t, [2]int{}, stats[winterMetricDeadSelf])
}

func TestAnalyzeWinterTraceMetricsExcludesLateEatsFromEatT20(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				// Turn 19 is the last turn that counts (cutoff is < 20).
				Turn:   19,
				Output: [2]string{"0 UP", "2 UP"},
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: 0, Coord: [2]int{1, 1}}),
					arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: 2, Coord: [2]int{2, 2}}),
				},
			},
			{
				// Turn 20 is past the cutoff and must NOT count.
				Turn:   20,
				Output: [2]string{"0 UP", "2 UP"},
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: 0, Coord: [2]int{3, 3}}),
				},
			},
		},
	}

	stats := analyzeWinterTraceMetrics(trace)

	assert.Equal(t, [2]int{1, 1}, stats[winterMetricEatByTurn20])
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

	require.Len(t, specs, 8)
	assert.Equal(t,
		[]string{
			winterMetricDeadSelf, winterMetricDeadWall, winterMetricDeadEnemy, TraceDeadFall,
			TraceHitSelf, TraceHitWall, TraceHitEnemy,
			winterMetricEatByTurn20,
		},
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
