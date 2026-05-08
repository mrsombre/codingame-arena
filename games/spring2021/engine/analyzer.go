package engine

import (
	"encoding/json"
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

const (
	spring2021MetricFirstCutDay  = "FIRST_CUT_DAY"
	spring2021MetricCuts         = "CUTS"
	spring2021MetricSeeds        = "SEEDS"
	spring2021MetricGrows        = "GROWS"
	spring2021MetricSunGathered  = "SUN_GATHERED"
	spring2021MetricShadowLost   = "SHADOW_LOST"
	spring2021MetricPointsPerCut = "POINTS_PER_CUT"
	spring2021MetricSeedRichness = "SEED_RICHNESS"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Spring 2021.
//
// FIRST_CUT_DAY is a per-match scalar — the game day on which the side
// emitted its first COMPLETE event — aggregated as a median across matches.
// Day (not trace turn) so the metric is invariant to the engine's PhaseTurnModel
// expansion (each day fans out into gathering / per-action / sun-move trace
// turns, and that fan-out drifts as bots take different action counts; day is
// the stable axis players think in). Median (not mean) so a side that never
// completed (no sample for that side) doesn't drag the central day toward
// MAX_ROUNDS. Day-0 cut is unreachable (trees start SMALL and COMPLETE
// requires TALL — at least one grow chain must run), so the per_match_value
// 0-as-missing sentinel is safe.
//
// CUTS counts COMPLETE events per match. Higher is better — every COMPLETE
// is a scoring action, so this is the volume side of the scoring loop
// (FIRST_CUT_DAY captures *when*, CUTS captures *how many*).
//
// SEEDS / GROWS count the input-action volume per match: SEEDS plant new
// trees (income capacity), GROWS upgrade them toward TALL (the size required
// to COMPLETE). Higher is better — together they trace investment volume,
// since SUN_GATHERED can only rise if there are trees on the board, and
// CUTS can only rise if those trees reach TALL.
//
// SUN_GATHERED sums the GATHER event Sun field per match. Sun is the input
// resource for every grow/seed/complete action, so this is the income side
// of the same loop — a high cut count with low sun-gathered means the bot
// converted efficiently; the inverse means it left sun on the table.
//
// SHADOW_LOST sums the size of own trees whose GATHER returned Sun=0 because
// a same-or-larger tree spookied them. Seeds (size 0) are excluded — they
// emit Sun=0 as well, but rules grant no sun to seeds regardless of shadow,
// so counting them would inflate the metric without indicating a positioning
// loss. The size is the cost: a shadowed TALL tree forfeited 3 sun, a
// shadowed SMALL forfeited 1.
//
// POINTS_PER_CUT is the within-match median of CompleteData.Points across
// the side's COMPLETE events. Median (not mean) so a single high-yield cut
// doesn't drag the per-match value upward — it reports the *typical* cut.
// The cross-match aggregation is a second median (per_match_value pipeline),
// so the rendered value answers "what's the typical cut quality the bot
// makes in a typical match?" — orthogonal to CUTS (volume).
//
// SEED_RICHNESS is the within-match median richness (1=POOR, 2=OK, 3=LUSH)
// of the cells the side seeded onto. Resolved by scanning every turn's
// state.Trees once to build a cell→richness map, then looking up each
// SEED.Target in the map. Cells that never carry a tree in the trace
// (rare — seed conflicts on cells nobody re-seeds successfully) are
// silently dropped. Higher is better — a richness-3 tree pays +4 nutrient
// bonus on every COMPLETE, so seed placement compounds across the whole
// scoring loop.
func (f *Factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{
			Key:         spring2021MetricFirstCutDay,
			Label:       spring2021MetricFirstCutDay,
			Kind:        arena.TraceMetricPerMatchValue,
			Description: "Game day of the side's first COMPLETE action (median across matches; sides that never completed are skipped)",
		},
		{
			Key:            spring2021MetricCuts,
			Label:          spring2021MetricCuts,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Trees completed per match (count of COMPLETE events; averaged across matches)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2021MetricSeeds,
			Label:          spring2021MetricSeeds,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Seeds planted per match (count of SEED actions; averaged across matches)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2021MetricGrows,
			Label:          spring2021MetricGrows,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Grow actions per match (count of GROW actions; averaged across matches)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2021MetricSunGathered,
			Label:          spring2021MetricSunGathered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Sun gathered per match (sum of GATHER event sun field; averaged across matches)",
			HigherIsBetter: true,
		},
		{
			Key:         spring2021MetricShadowLost,
			Label:       spring2021MetricShadowLost,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Sun lost to spooky shadows per match (sum of own-tree sizes whose GATHER returned Sun=0; seeds excluded)",
		},
		{
			Key:            spring2021MetricPointsPerCut,
			Label:          spring2021MetricPointsPerCut,
			Kind:           arena.TraceMetricPerMatchValue,
			Description:    "Median points per COMPLETE within a match (median across matches); 0 = no cuts (zero-yield cuts are also dropped due to the per_match_value sentinel)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2021MetricSeedRichness,
			Label:          spring2021MetricSeedRichness,
			Kind:           arena.TraceMetricPerMatchValue,
			Description:    "Median richness (1=POOR, 2=OK, 3=LUSH) of cells the side seeded within a match (median across matches)",
			HigherIsBetter: true,
		},
	}
}

// AnalyzeTraceMetrics walks the trace once, recording the game day of the
// first COMPLETE event observed per side. The day is read from
// TraceTurnState.Day in the per-turn State payload (the engine writes it via
// DecorateTraceTurn). Sides that never complete report 0 — the arena
// per_match_value pipeline drops 0 as "no sample" rather than treating it as
// a real day-0 cut (which the rules preclude).
func (f *Factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeSpring2021TraceMetrics(trace), nil
}

func analyzeSpring2021TraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		spring2021MetricFirstCutDay:  {},
		spring2021MetricCuts:         {},
		spring2021MetricSeeds:        {},
		spring2021MetricGrows:        {},
		spring2021MetricSunGathered:  {},
		spring2021MetricShadowLost:   {},
		spring2021MetricPointsPerCut: {},
		spring2021MetricSeedRichness: {},
	}

	cellRichness := spring2021CellRichnessMap(trace)
	var pointsPerSide [2][]int
	var richnessPerSide [2][]int

	var firstCutFound [2]bool
	for _, turn := range trace.Turns {
		state, stateOK := decodeSpring2021TraceState(turn.State)
		dayOK := stateOK && state.Day != nil
		for side := 0; side < 2; side++ {
			for _, ev := range turn.Traces[side] {
				switch ev.Type {
				case TraceComplete:
					addSpring2021TraceMetricCount(stats, spring2021MetricCuts, side, 1)
					if data, err := arena.DecodeData[CompleteData](ev); err == nil {
						pointsPerSide[side] = append(pointsPerSide[side], data.Points)
					}
					if !firstCutFound[side] && dayOK {
						values := stats[spring2021MetricFirstCutDay]
						values[side] = *state.Day
						stats[spring2021MetricFirstCutDay] = values
						firstCutFound[side] = true
					}
				case TraceSeed:
					addSpring2021TraceMetricCount(stats, spring2021MetricSeeds, side, 1)
					if data, err := arena.DecodeData[SeedData](ev); err == nil {
						if r, ok := cellRichness[data.Target]; ok && r > 0 {
							richnessPerSide[side] = append(richnessPerSide[side], r)
						}
					}
				case TraceGrow:
					addSpring2021TraceMetricCount(stats, spring2021MetricGrows, side, 1)
				case TraceGather:
					data, err := arena.DecodeData[GatherData](ev)
					if err != nil {
						continue
					}
					if data.Sun > 0 {
						addSpring2021TraceMetricCount(stats, spring2021MetricSunGathered, side, data.Sun)
						continue
					}
					// Sun=0: spooky shadow, or seed. Need the tree's size to
					// distinguish — seeds (size 0) emit Sun=0 by rule and
					// must NOT count as a shadow loss. Without State.Trees
					// we can't tell, so silently skip the row.
					if !stateOK {
						continue
					}
					if size, ok := spring2021TraceTreeSize(state, side, data.Cell); ok && size > 0 {
						addSpring2021TraceMetricCount(stats, spring2021MetricShadowLost, side, size)
					}
				}
			}
		}
	}

	// Per-match medians for the value-kind metrics. Stored as int per the
	// [2]int contract; medianIntHalfUp rounds half-integers up so a {18, 21}
	// pair lands on 20 rather than truncating to 19. Sides with an empty
	// sample slice keep value 0, which the per_match_value pipeline drops
	// as "no sample" — there's a known cost: a side whose true median
	// POINTS_PER_CUT is 0 (cuts at nutrients=0 on POOR cells) is also
	// dropped, but that's a rare degenerate case in practice.
	for side := 0; side < 2; side++ {
		if len(pointsPerSide[side]) > 0 {
			values := stats[spring2021MetricPointsPerCut]
			values[side] = medianIntHalfUp(pointsPerSide[side])
			stats[spring2021MetricPointsPerCut] = values
		}
		if len(richnessPerSide[side]) > 0 {
			values := stats[spring2021MetricSeedRichness]
			values[side] = medianIntHalfUp(richnessPerSide[side])
			stats[spring2021MetricSeedRichness] = values
		}
	}

	return stats
}

// spring2021CellRichnessMap returns a cell→richness lookup built by scanning
// every turn's State.Trees. Richness is fixed per cell at board generation,
// so any turn that observes a tree on cell C tells us C's richness for the
// whole match. Cells nobody ever has a tree on stay missing — analyzers
// fall back to "no sample" rather than guess. Cheap to rebuild per trace
// (one map per trace, not per turn).
func spring2021CellRichnessMap(trace arena.TraceMatch) map[int]int {
	out := make(map[int]int)
	for _, turn := range trace.Turns {
		state, ok := decodeSpring2021TraceState(turn.State)
		if !ok {
			continue
		}
		for _, sideTrees := range state.Trees {
			for _, tuple := range sideTrees {
				cell, richness := tuple[0], tuple[1]
				if richness > 0 {
					out[cell] = richness
				}
			}
		}
	}
	return out
}

// medianIntHalfUp returns the integer median of values, rounding half-up on
// even-count inputs (so {2, 3} → 3 rather than truncating to 2). Returns 0
// for an empty slice — callers treat that as "no sample" via the
// per_match_value 0-as-missing sentinel.
func medianIntHalfUp(values []int) int {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int(nil), values...)
	sort.Ints(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2] + 1) / 2
}

func addSpring2021TraceMetricCount(stats arena.TraceMetricStats, key string, side, delta int) {
	values := stats[key]
	values[side] += delta
	stats[key] = values
}

// decodeSpring2021TraceState parses the per-turn State payload back into the
// engine's typed struct so the analyzer can read Day and Trees. Returns
// ok=false on missing or malformed payloads — callers fall back to skipping
// the metric for that turn rather than crashing the whole report.
func decodeSpring2021TraceState(raw json.RawMessage) (TraceTurnState, bool) {
	if len(raw) == 0 {
		return TraceTurnState{}, false
	}
	var state TraceTurnState
	if err := json.Unmarshal(raw, &state); err != nil {
		return TraceTurnState{}, false
	}
	return state, true
}

// spring2021TraceTreeSize linearly scans the side's tree list for the cell
// and returns the tree's size (tuple index 2 in the on-disk shape). Linear
// scan is fine: spring2021 caps at ~24 trees per side per turn and the
// analyzer is offline, so a map per turn would cost more setup than it saves.
func spring2021TraceTreeSize(state TraceTurnState, side, cell int) (int, bool) {
	if side < 0 || side >= len(state.Trees) {
		return 0, false
	}
	for _, tuple := range state.Trees[side] {
		if tuple[0] == cell {
			return tuple[2], true
		}
	}
	return 0, false
}

var _ arena.TraceMetricAnalyzer = (*Factory)(nil)
