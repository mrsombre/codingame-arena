package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func stateWithDay(t *testing.T, day int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(TraceTurnState{Day: &day})
	require.NoError(t, err)
	return raw
}

func stateWithDayAndTrees(t *testing.T, day int, trees [][][4]int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(TraceTurnState{Day: &day, Trees: trees})
	require.NoError(t, err)
	return raw
}

func TestAnalyzeSpring2021FirstCutDayRecordsDayOfFirstCompletePerSide(t *testing.T) {
	// Two trace turns can share a day in PhaseTurnModel (gathering then
	// actions). The metric must report game day, not trace turn — both
	// sides cutting in turn 8 of day 3 should both come back as 3. The
	// fixture exercises CUTS, SEEDS, GROWS, and SUN_GATHERED so a single
	// trace covers all action-volume specs.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			// Day 2: side 0 grows once, side 1 grows then seeds. Action
			// counts must accumulate per type without double-counting.
			{Turn: 5, State: stateWithDay(t, 2), Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceGrow, GrowData{Cell: 0, Cost: 3})},
				{
					arena.MakeTurnTrace(TraceGrow, GrowData{Cell: 9, Cost: 3}),
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 9, Target: 12, Cost: 0}),
				},
			}},
			// Day 3, gathering frame: side 0 harvests size-3 tree, size-2 tree;
			// side 1 harvests size-3 tree but a size-1 tree is shadowed (Sun=0).
			// SUN_GATHERED must sum across all GATHER events but skip Sun=0
			// rows so the totals reflect actual income.
			{Turn: 6, State: stateWithDay(t, 3), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 0, Sun: TREE_TALL}),
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 1, Sun: TREE_MEDIUM}),
				},
				{
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 9, Sun: TREE_TALL}),
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 22, Sun: 0}),
				},
			}},
			// Day 3, actions frame: side 1 cuts first.
			{Turn: 8, State: stateWithDay(t, 3), Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceWait, WaitData{})},
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 5, Points: 20, Cost: LIFECYCLE_END_COST})},
			}},
			// Day 5: side 0 cuts; side 1 cuts a second tree. FIRST_CUT_DAY
			// for side 1 must stay locked at 3, but CUTS keeps incrementing.
			{Turn: 12, State: stateWithDay(t, 5), Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 3, Points: 18, Cost: LIFECYCLE_END_COST})},
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 7, Points: 22, Cost: LIFECYCLE_END_COST})},
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, [2]int{5, 3}, stats[spring2021MetricFirstCutDay])
	assert.Equal(t, [2]int{1, 2}, stats[spring2021MetricCuts])
	// side 0 grew once; side 1 grew once and seeded once.
	assert.Equal(t, [2]int{1, 1}, stats[spring2021MetricGrows])
	assert.Equal(t, [2]int{0, 1}, stats[spring2021MetricSeeds])
	// side 0: TREE_TALL (3) + TREE_MEDIUM (2) = 5; side 1: TREE_TALL (3) only,
	// the shadowed Sun=0 GATHER must NOT add a sample.
	assert.Equal(t, [2]int{TREE_TALL + TREE_MEDIUM, TREE_TALL}, stats[spring2021MetricSunGathered])
}

func TestAnalyzeSpring2021FirstCutDayReportsZeroForSidesThatNeverComplete(t *testing.T) {
	// Side 0 never completes — value stays 0 so the arena per_match_value
	// pipeline drops it as a missing sample (rather than treating day 0 as
	// a real first-cut record, which the game rules make unreachable).
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, State: stateWithDay(t, 1), Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceGrow, GrowData{Cell: 0, Cost: 3})},
				nil,
			}},
			{Turn: 7, State: stateWithDay(t, 4), Traces: [2][]arena.TurnTrace{
				nil,
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 5, Points: 20, Cost: LIFECYCLE_END_COST})},
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, [2]int{0, 4}, stats[spring2021MetricFirstCutDay])
}

func TestAnalyzeSpring2021FirstCutDaySkipsTurnsMissingDayInState(t *testing.T) {
	// A turn without State (engine frame predating the decorator, or any
	// future trace shape that omits Day) must not crash — the analyzer
	// just walks past it. The same COMPLETE on a later, well-stamped turn
	// is the value that wins.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			// No State — analyzer must skip without panicking even though
			// a COMPLETE is present.
			{Turn: 4, Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 0, Points: 18, Cost: LIFECYCLE_END_COST})},
				nil,
			}},
			{Turn: 9, State: stateWithDay(t, 6), Traces: [2][]arena.TurnTrace{
				{arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 1, Points: 18, Cost: LIFECYCLE_END_COST})},
				nil,
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, [2]int{6, 0}, stats[spring2021MetricFirstCutDay])
}

func TestAnalyzeSpring2021ShadowLostSumsOwnTreeSizesForShadowedGathersOnly(t *testing.T) {
	// Side 0 owns a TALL gathering and a SMALL that got spookied.
	// Side 1 owns a TALL gathering and a SEED (size 0): the seed emits
	// Sun=0 too but rules grant no sun to seeds in any case, so its row
	// must NOT count as a shadow loss.
	trees := [][][4]int{
		{
			{0, 3, TREE_TALL, 0},
			{7, 1, TREE_SMALL, 0},
		},
		{
			{19, 0, TREE_TALL, 0},
			{22, 0, TREE_SEED, 0},
		},
	}
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDayAndTrees(t, 3, trees), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 0, Sun: TREE_TALL}),
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 7, Sun: 0}),
				},
				{
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 19, Sun: TREE_TALL}),
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 22, Sun: 0}),
				},
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	// side 0 forfeited TREE_SMALL (1) sun to a spooky shadow.
	// side 1's Sun=0 came from a seed, which doesn't count.
	assert.Equal(t, [2]int{TREE_SMALL, 0}, stats[spring2021MetricShadowLost])
	assert.Equal(t, [2]int{TREE_TALL, TREE_TALL}, stats[spring2021MetricSunGathered])
}

func TestAnalyzeSpring2021PointsPerCutComputesWithinMatchMedian(t *testing.T) {
	// Side 0 cuts three trees scoring {12, 14, 20} → median 14.
	// Side 1 cuts two trees scoring {8, 16} → median 12 (half-up rounded
	// from 12.0; this exact pair has no rounding so the rounding rule is
	// exercised separately below).
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDay(t, 5), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 0, Points: 12, Cost: LIFECYCLE_END_COST}),
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 1, Points: 14, Cost: LIFECYCLE_END_COST}),
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 2, Points: 20, Cost: LIFECYCLE_END_COST}),
				},
				{
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 9, Points: 8, Cost: LIFECYCLE_END_COST}),
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 10, Points: 16, Cost: LIFECYCLE_END_COST}),
				},
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, [2]int{14, 12}, stats[spring2021MetricPointsPerCut])
}

func TestAnalyzeSpring2021PointsPerCutRoundsHalfUpOnEvenCounts(t *testing.T) {
	// Two cuts {18, 21} have median 19.5 → rounds half-up to 20. The
	// alternative truncation behavior would land on 19 and silently bias
	// every even-count match downward, so this rounding rule is pinned.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDay(t, 5), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 0, Points: 18, Cost: LIFECYCLE_END_COST}),
					arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: 1, Points: 21, Cost: LIFECYCLE_END_COST}),
				},
				nil,
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, 20, stats[spring2021MetricPointsPerCut][0])
}

func TestAnalyzeSpring2021SeedRichnessLooksUpRichnessFromStateTrees(t *testing.T) {
	// Set up Trees in two turns so the cell→richness map covers every
	// SEED Target without the seed itself needing to be in the same
	// turn's Trees. Cell 5 = LUSH (3), cell 12 = OK (2), cell 18 = POOR (1).
	turn1Trees := [][][4]int{
		{{0, RICHNESS_LUSH, TREE_TALL, 0}, {5, RICHNESS_LUSH, TREE_SEED, 0}},
		{{9, RICHNESS_OK, TREE_TALL, 0}, {12, RICHNESS_OK, TREE_SEED, 0}},
	}
	turn2Trees := [][][4]int{
		{{18, RICHNESS_POOR, TREE_SEED, 0}},
		nil,
	}
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDayAndTrees(t, 3, turn1Trees), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 0, Target: 5, Cost: 0}),
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 0, Target: 18, Cost: 1}),
				},
				{
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 9, Target: 12, Cost: 0}),
				},
			}},
			{Turn: 9, State: stateWithDayAndTrees(t, 4, turn2Trees), Traces: [2][]arena.TurnTrace{
				nil,
				nil,
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	// side 0 seeded {LUSH=3, POOR=1} → median = (1+3+1)/2 = 2 (half-up).
	// side 1 seeded {OK=2} → median = 2.
	assert.Equal(t, [2]int{2, 2}, stats[spring2021MetricSeedRichness])
}

func TestAnalyzeSpring2021SeedRichnessSkipsCellsNeverSeenInTrees(t *testing.T) {
	// SEED.Target = 99 is a cell that no Trees turn ever shows (e.g. a
	// seed conflict where neither tree gets placed AND nobody re-seeds).
	// The richness can't be resolved, so the sample drops out of the
	// per-side collector — side 0's median is computed over the one
	// resolvable cell only.
	trees := [][][4]int{
		{{0, RICHNESS_LUSH, TREE_TALL, 0}, {5, RICHNESS_LUSH, TREE_SEED, 0}},
		nil,
	}
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDayAndTrees(t, 3, trees), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 0, Target: 5, Cost: 0}),
					arena.MakeTurnTrace(TraceSeed, SeedData{Source: 0, Target: 99, Cost: 0}),
				},
				nil,
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	// Only the LUSH seed counts; cell 99 is dropped.
	assert.Equal(t, RICHNESS_LUSH, stats[spring2021MetricSeedRichness][0])
}

func TestAnalyzeSpring2021ShadowLostSilentlyZeroWhenStateLacksTrees(t *testing.T) {
	// State has Day but no Trees: SHADOW_LOST can't be computed without the
	// tree list, so the analyzer must fall back to 0 rather than crash.
	// SUN_GATHERED still works (it doesn't need state).
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 6, State: stateWithDay(t, 3), Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 0, Sun: TREE_TALL}),
					arena.MakeTurnTrace(TraceGather, GatherData{Cell: 7, Sun: 0}),
				},
				nil,
			}},
		},
	}

	stats := analyzeSpring2021TraceMetrics(trace)

	assert.Equal(t, [2]int{0, 0}, stats[spring2021MetricShadowLost])
	assert.Equal(t, [2]int{TREE_TALL, 0}, stats[spring2021MetricSunGathered])
}

func TestSpring2021TraceMetricSpecsListExpectedMetrics(t *testing.T) {
	specs := NewFactory().(arena.TraceMetricAnalyzer).TraceMetricSpecs()

	require.Len(t, specs, 8)
	keys := make([]string, len(specs))
	for i, spec := range specs {
		keys[i] = spec.Key
		assert.NotEmpty(t, spec.Description, "spec %s must carry a description so the legend renders", spec.Key)
	}
	assert.Equal(t, []string{
		spring2021MetricFirstCutDay,
		spring2021MetricCuts,
		spring2021MetricSeeds,
		spring2021MetricGrows,
		spring2021MetricSunGathered,
		spring2021MetricShadowLost,
		spring2021MetricPointsPerCut,
		spring2021MetricSeedRichness,
	}, keys)

	specByKey := make(map[string]arena.TraceMetricSpec, len(specs))
	for _, spec := range specs {
		specByKey[spec.Key] = spec
	}
	assert.Equal(t, arena.TraceMetricPerMatchValue, specByKey[spring2021MetricFirstCutDay].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specByKey[spring2021MetricCuts].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specByKey[spring2021MetricSeeds].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specByKey[spring2021MetricGrows].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specByKey[spring2021MetricSunGathered].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchCount, specByKey[spring2021MetricShadowLost].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchValue, specByKey[spring2021MetricPointsPerCut].Kind)
	assert.Equal(t, arena.TraceMetricPerMatchValue, specByKey[spring2021MetricSeedRichness].Kind)
	// Positive (HigherIsBetter) specs are excluded from WORST. SHADOW_LOST
	// is the only hazard — its peak names the replay where blue's
	// positioning collapsed worst.
	assert.True(t, specByKey[spring2021MetricCuts].HigherIsBetter)
	assert.True(t, specByKey[spring2021MetricSeeds].HigherIsBetter)
	assert.True(t, specByKey[spring2021MetricGrows].HigherIsBetter)
	assert.True(t, specByKey[spring2021MetricSunGathered].HigherIsBetter)
	assert.True(t, specByKey[spring2021MetricPointsPerCut].HigherIsBetter)
	assert.True(t, specByKey[spring2021MetricSeedRichness].HigherIsBetter)
	assert.False(t, specByKey[spring2021MetricShadowLost].HigherIsBetter)
}
