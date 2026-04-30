package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeSpringTraceMetricsAttributesAndCollapsesPerTurnMetrics(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 0, Coord: [2]int{1, 1}, Cost: 1}),
					arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 1, Coord: [2]int{2, 1}, Cost: 10}),
					arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
					arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
					arena.MakeTurnTrace(TraceKilled, KilledMeta{Pac: 2, Coord: [2]int{3, 1}, Killer: 1}),
					arena.MakeTurnTrace(TraceSwitch, SwitchMeta{Pac: 3, Type: "ROCK"}),
				},
			},
			{
				Traces: []arena.TurnTrace{
					arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
					arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 3}),
					arena.MakeTurnTrace(TraceCollideEnemy, PacMeta{Pac: 0}),
					arena.MakeTurnTrace(TraceCollideEnemy, PacMeta{Pac: 2}),
				},
			},
		},
	}

	stats := analyzeSpringTraceMetrics(trace)

	assert.Equal(t, [2]int{0, 1}, stats[TraceSwitch])
	assert.Equal(t, [2]int{1, 0}, stats[TraceKilled])
	assert.Equal(t, [2]int{0, 1}, stats[springMetricEatSuper])
	assert.Equal(t, [2]int{0, 2}, stats[TraceCollideSelf])
	assert.Equal(t, [2]int{1, 0}, stats[TraceCollideEnemy])
	assert.Equal(t, [2]int{1, 1}, stats[springMetricNoEatTurn])
}

func TestSpringTraceMetricSpecsListExpectedMetrics(t *testing.T) {
	specs := (&Factory{}).TraceMetricSpecs()

	require.Len(t, specs, 6)
	assert.Equal(t,
		[]string{TraceSwitch, TraceKilled, springMetricEatSuper, TraceCollideSelf, TraceCollideEnemy, springMetricNoEatTurn},
		metricSpecKeys(specs),
	)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specs[0].Kind)
	assert.Equal(t, arena.TraceMetricPerTurnRate, specs[3].Kind)
}

func metricSpecKeys(metrics []arena.TraceMetricSpec) []string {
	keys := make([]string, len(metrics))
	for i, m := range metrics {
		keys[i] = m.Key
	}
	return keys
}
