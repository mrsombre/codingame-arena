package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestAnalyzeSpring2026SumsDropItemsPerOrdinal(t *testing.T) {
	// A single DROP can carry mixed items; the analyzer must split the
	// vector by ordinal so PLUM/LEMON/APPLE/BANANA/IRON/WOOD land in their
	// own metric. Side 0 drops [2,0,1,0,0,3], side 1 drops [0,1,0,0,4,0]
	// plus a second drop [0,0,0,2,0,1] — totals are summed across drops.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceDrop, DropData{Unit: 0, Items: [ItemsCount]int{2, 0, 1, 0, 0, 3}})},
				{arena.MakeTurnTrace(TraceDrop, DropData{Unit: 1, Items: [ItemsCount]int{0, 1, 0, 0, 4, 0}})},
			}},
			{Turn: 2, Traces: [2][]arena.TurnTrace{
				nil,
				{arena.MakeTurnTrace(TraceDrop, DropData{Unit: 1, Items: [ItemsCount]int{0, 0, 0, 2, 0, 1}})},
			}},
		},
	}

	stats := analyzeSpring2026TraceMetrics(trace)

	assert.Equal(t, [2]int{2, 0}, stats[spring2026MetricPlumDelivered])
	assert.Equal(t, [2]int{0, 1}, stats[spring2026MetricLemonDelivered])
	assert.Equal(t, [2]int{1, 0}, stats[spring2026MetricAppleDelivered])
	assert.Equal(t, [2]int{0, 2}, stats[spring2026MetricBananaDelivered])
	assert.Equal(t, [2]int{0, 4}, stats[spring2026MetricIronDelivered])
	assert.Equal(t, [2]int{3, 1}, stats[spring2026MetricWoodDelivered])
}

func TestAnalyzeSpring2026CountsActionEvents(t *testing.T) {
	// TRAIN, CHOP, HARVEST, PLANT, FAILED are 1 event = 1 count by construction
	// (every emitted trace is one applied task or one raw rejection). MOVE /
	// PICK / WAIT / MSG must NOT influence these counts.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceTrain, TrainData{Unit: 2, Talents: [4]int{1, 1, 1, 0}}),
					arena.MakeTurnTrace(TraceChop, ChopData{Unit: 0, Cell: [2]int{5, 3}, Damage: 1, Wood: 0}),
					arena.MakeTurnTrace(TraceHarvest, HarvestData{Unit: 0, Cell: [2]int{4, 2}, Type: "PLUM", Amount: 1}),
					arena.MakeTurnTrace(TracePlant, PlantData{Unit: 0, Cell: [2]int{3, 3}, Type: "LEMON"}),
					arena.MakeTurnTrace(TraceMove, MoveData{Unit: 0, To: [2]int{4, 3}}),
					arena.MakeTurnTrace(TracePick, PickData{Unit: 0, Type: "PLUM"}),
					arena.TurnTrace{Type: TraceWait},
					arena.MakeTurnTrace(TraceMessage, MessageData{Text: "hi"}),
					arena.MakeTurnTrace(TraceFailed, FailedData{Code: ErrMoveBlocked, Reason: "blocked"}),
				},
				{
					arena.MakeTurnTrace(TraceChop, ChopData{Unit: 1, Cell: [2]int{10, 5}, Damage: 1, Wood: 0}),
					arena.MakeTurnTrace(TraceChop, ChopData{Unit: 1, Cell: [2]int{10, 5}, Damage: 1, Wood: 2, Killed: true}),
					arena.MakeTurnTrace(TraceFailed, FailedData{Code: ErrAlreadyUsed, Reason: "Troll 1 already used"}),
					arena.MakeTurnTrace(TraceFailed, FailedData{Code: ErrOpponentBlocking, Reason: "contradicted"}),
				},
			}},
			{Turn: 2, Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceTrain, TrainData{Unit: 3, Talents: [4]int{1, 2, 0, 1}})},
				nil,
			}},
		},
	}

	stats := analyzeSpring2026TraceMetrics(trace)

	assert.Equal(t, [2]int{2, 0}, stats[spring2026MetricGoblinsTrained])
	assert.Equal(t, [2]int{1, 2}, stats[spring2026MetricChops])
	assert.Equal(t, [2]int{1, 0}, stats[spring2026MetricHarvests])
	assert.Equal(t, [2]int{1, 0}, stats[spring2026MetricPlants])
	assert.Equal(t, [2]int{1, 2}, stats[spring2026MetricFailedActions])
}

func TestAnalyzeSpring2026SilentSideStaysZero(t *testing.T) {
	// A side that emits no events must report 0 across every metric. 0 is a
	// real value for PerMatchCount metrics (not a "no sample" sentinel), so
	// the report will compare it normally against the other side.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceDrop, DropData{Unit: 0, Items: [ItemsCount]int{1, 0, 0, 0, 0, 0}})},
				nil,
			}},
		},
	}

	stats := analyzeSpring2026TraceMetrics(trace)

	assert.Equal(t, [2]int{1, 0}, stats[spring2026MetricPlumDelivered])
	for _, key := range []string{
		spring2026MetricLemonDelivered,
		spring2026MetricAppleDelivered,
		spring2026MetricBananaDelivered,
		spring2026MetricIronDelivered,
		spring2026MetricWoodDelivered,
		spring2026MetricGoblinsTrained,
		spring2026MetricChops,
		spring2026MetricHarvests,
		spring2026MetricPlants,
		spring2026MetricFailedActions,
	} {
		assert.Equal(t, [2]int{0, 0}, stats[key], "metric %s should be zero for both sides", key)
	}
}

func TestSpring2026TraceMetricSpecsAllRegistered(t *testing.T) {
	// Every key returned by AnalyzeTraceMetrics must have a matching spec —
	// validateTraceMetricStats rejects unknown keys at the arena layer, so
	// the test guards against a future drift between the spec list and the
	// stats map.
	f := &factory{}
	specs := f.TraceMetricSpecs()
	stats := analyzeSpring2026TraceMetrics(arena.TraceMatch{})

	specKeys := make(map[string]bool, len(specs))
	for _, spec := range specs {
		specKeys[spec.Key] = true
		assert.Equal(t, arena.TraceMetricPerMatchCount, spec.Kind, "metric %s should be per_match_count", spec.Key)
		// FAILED_ACTIONS is lower-is-better (hazard metric); everything else is HigherIsBetter.
		if spec.Key == spring2026MetricFailedActions {
			assert.False(t, spec.HigherIsBetter, "FAILED_ACTIONS should NOT be HigherIsBetter")
		} else {
			assert.True(t, spec.HigherIsBetter, "metric %s should be HigherIsBetter", spec.Key)
		}
	}
	for key := range stats {
		assert.True(t, specKeys[key], "stat key %s missing from TraceMetricSpecs", key)
	}
	assert.Equal(t, len(specs), len(stats))
}
