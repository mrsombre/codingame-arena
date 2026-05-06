package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeSpringTraceMetricsAttributesAndCollapsesPerTurnMetrics(t *testing.T) {
	// COLLIDE_ENEMY events are mirrored into both player slots by the
	// engine, so the fixture replicates that pattern.
	collideEnemyP0 := arena.MakeTurnTrace(TraceCollideEnemy, PacMeta{Pac: 0})
	collideEnemyP2 := arena.MakeTurnTrace(TraceCollideEnemy, PacMeta{Pac: 2})

	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{
				Turn: 0,
				Traces: [2][]arena.TurnTrace{
					{
						arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 0, Coord: [2]int{1, 1}, Cost: 1}),
						arena.MakeTurnTrace(TraceKilled, KilledMeta{Pac: 2, Coord: [2]int{3, 1}, Killer: 1}),
						arena.MakeTurnTrace(TraceSpeed, PacMeta{Pac: 0}),
					},
					{
						arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 1, Coord: [2]int{2, 1}, Cost: 10}),
						arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
						arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
						arena.MakeTurnTrace(TraceSwitch, SwitchMeta{Pac: 3, Type: "ROCK"}),
					},
				},
			},
			{
				Turn: 1,
				Traces: [2][]arena.TurnTrace{
					{
						collideEnemyP0,
						collideEnemyP2,
						arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 0, Coord: [2]int{4, 1}, Cost: 1}),
					},
					{
						arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 1}),
						arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: 3}),
						collideEnemyP0,
						collideEnemyP2,
					},
				},
			},
			{
				Turn: 25,
				Traces: [2][]arena.TurnTrace{
					{
						arena.MakeTurnTrace(TraceEat, EatMeta{Pac: 0, Coord: [2]int{5, 1}, Cost: 1}),
					},
					nil,
				},
			},
		},
	}

	stats := analyzeSpringTraceMetrics(trace)

	assert.Equal(t, [2]int{0, 1}, stats[TraceSwitch])
	assert.Equal(t, [2]int{1, 0}, stats[TraceKilled])
	assert.Equal(t, [2]int{0, 1}, stats[springMetricKills])
	assert.Equal(t, [2]int{0, 1}, stats[springMetricEatSuper])
	assert.Equal(t, [2]int{3, 1}, stats[springMetricEatT50])
	assert.Equal(t, [2]int{0, 2}, stats[TraceCollideSelf])
	// COLLIDE_ENEMY appears in both slots when mirrored, so per-turn rate
	// counts both sides as having "experienced" the cross-team collision.
	assert.Equal(t, [2]int{1, 1}, stats[TraceCollideEnemy])
	assert.Equal(t, [2]int{0, 2}, stats[springMetricNoEatTurn])
}

func TestSpringTraceMetricSpecsListExpectedMetrics(t *testing.T) {
	specs := (&Factory{}).TraceMetricSpecs()

	require.Len(t, specs, 8)
	assert.Equal(t,
		[]string{
			TraceKilled, springMetricKills,
			TraceSwitch,
			springMetricEatSuper, springMetricEatT50,
			TraceCollideSelf, TraceCollideEnemy, springMetricNoEatTurn,
		},
		metricSpecKeys(specs),
	)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specs[0].Kind)
	assert.Equal(t, arena.TraceMetricPerTurnRate, specs[5].Kind)

	for _, s := range specs {
		assert.NotEmpty(t, s.Description, "metric %q missing Description", s.Key)
	}
	for _, key := range []string{springMetricKills, springMetricEatSuper, springMetricEatT50} {
		var spec arena.TraceMetricSpec
		for _, s := range specs {
			if s.Key == key {
				spec = s
				break
			}
		}
		assert.True(t, spec.HigherIsBetter, "metric %q should be HigherIsBetter", key)
	}
}

func metricSpecKeys(metrics []arena.TraceMetricSpec) []string {
	keys := make([]string, len(metrics))
	for i, m := range metrics {
		keys[i] = m.Key
	}
	return keys
}
