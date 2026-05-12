package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

const (
	spring2026MetricPlumDelivered   = "PLUM_DELIVERED"
	spring2026MetricLemonDelivered  = "LEMON_DELIVERED"
	spring2026MetricAppleDelivered  = "APPLE_DELIVERED"
	spring2026MetricBananaDelivered = "BANANA_DELIVERED"
	spring2026MetricIronDelivered   = "IRON_DELIVERED"
	spring2026MetricWoodDelivered   = "WOOD_DELIVERED"
	spring2026MetricGoblinsTrained  = "GOBLINS_TRAINED"
	spring2026MetricChops           = "CHOPS"
	spring2026MetricHarvests        = "HARVESTS"
	spring2026MetricPlants          = "PLANTS"
	spring2026MetricFailedActions   = "FAILED_ACTIONS"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Spring 2026.
//
// The six *_DELIVERED metrics sum the items the side handed over to its shack
// via DROP events. DROP is the only event that adds to the shack inventory
// (PICK and CHOP/MINE only fill the troll's own carry), so summing DropData.Items
// over the match equals the side's gross resource intake. Score is
// PLUM + LEMON + APPLE + BANANA + 4·WOOD, so splitting by item ordinal keeps
// the score-weighting visible: WOOD pays 4× and is typically the dominant
// source in leagues 3+; IRON is not scored but feeds future TRAIN/CHOP work.
// Higher is better for all six (more resources delivered).
//
// GOBLINS_TRAINED counts TRAIN events — every successful TRAIN spawns one new
// troll, so this is the side's investment-volume axis (mirrors the role
// CUTS plays for spring2021).
//
// CHOPS, HARVESTS, PLANTS count the input-action volumes for the three
// resource-gathering loops (wood / fruit / seeding). These are work counts,
// not score counts — the score side is the *_DELIVERED group. Together they
// trace where the bot spent its turns. Higher is better as a proxy for
// activity, though CHOPS or PLANTS dominance is a strategic signal rather
// than a correctness signal.
//
// FAILED_ACTIONS counts non-critical input errors per match (one FAILED
// trace per raw error — PopErrors collapses runs of the same code on the
// summary tape but FAILED traces preserve the actual count). Lower is
// better (no HigherIsBetter flag): a high count usually flags pathfinding
// or coordination bugs (target blocked, opponent contradicts, troll
// already used this turn). Critical errors deactivate the player and
// surface via endReason instead, so this metric tracks playable-but-wasted
// commands.
func (f *factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{
			Key:            spring2026MetricPlumDelivered,
			Label:          spring2026MetricPlumDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Plums dropped at the shack per match (sum of DROP.items[PLUM]; 1 point each in score)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricLemonDelivered,
			Label:          spring2026MetricLemonDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Lemons dropped at the shack per match (sum of DROP.items[LEMON]; 1 point each in score)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricAppleDelivered,
			Label:          spring2026MetricAppleDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Apples dropped at the shack per match (sum of DROP.items[APPLE]; 1 point each in score)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricBananaDelivered,
			Label:          spring2026MetricBananaDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Bananas dropped at the shack per match (sum of DROP.items[BANANA]; 1 point each in score)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricIronDelivered,
			Label:          spring2026MetricIronDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Iron dropped at the shack per match (sum of DROP.items[IRON]; not scored, feeds future investment)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricWoodDelivered,
			Label:          spring2026MetricWoodDelivered,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Wood dropped at the shack per match (sum of DROP.items[WOOD]; 4 points each in score)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricGoblinsTrained,
			Label:          spring2026MetricGoblinsTrained,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Trolls trained per match (count of TRAIN events; each spawns one new troll on the shack)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricChops,
			Label:          spring2026MetricChops,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Chop actions per match (count of CHOP events; wood-loop work volume)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricHarvests,
			Label:          spring2026MetricHarvests,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Harvest actions per match (count of HARVEST events; fruit-loop work volume)",
			HigherIsBetter: true,
		},
		{
			Key:            spring2026MetricPlants,
			Label:          spring2026MetricPlants,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Plant actions per match (count of PLANT events; seeding work volume)",
			HigherIsBetter: true,
		},
		{
			Key:         spring2026MetricFailedActions,
			Label:       spring2026MetricFailedActions,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Rejected commands per match (count of FAILED events; non-critical input errors only — lower is better)",
		},
	}
}

// AnalyzeTraceMetrics walks the trace once and accumulates per-side counts.
// The DROP event is the source of truth for "delivered to shack" because it
// is the only path through which items reach Player.Inv: PICK only moves
// items from shack to troll; CHOP / MINE / HARVEST only fill the troll's
// own carry. Summing DropData.Items across all DROPs reproduces the gross
// intake of each item type. TRAIN / CHOP / HARVEST / PLANT counts come from
// straight event counts — every event is one applied task by construction
// (rejected commands emit no trace).
func (f *factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeSpring2026TraceMetrics(trace), nil
}

func analyzeSpring2026TraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		spring2026MetricPlumDelivered:   {},
		spring2026MetricLemonDelivered:  {},
		spring2026MetricAppleDelivered:  {},
		spring2026MetricBananaDelivered: {},
		spring2026MetricIronDelivered:   {},
		spring2026MetricWoodDelivered:   {},
		spring2026MetricGoblinsTrained:  {},
		spring2026MetricChops:           {},
		spring2026MetricHarvests:        {},
		spring2026MetricPlants:          {},
		spring2026MetricFailedActions:   {},
	}

	deliveredKeys := [ItemsCount]string{
		spring2026MetricPlumDelivered,
		spring2026MetricLemonDelivered,
		spring2026MetricAppleDelivered,
		spring2026MetricBananaDelivered,
		spring2026MetricIronDelivered,
		spring2026MetricWoodDelivered,
	}

	for _, turn := range trace.Turns {
		for side := 0; side < 2; side++ {
			for _, ev := range turn.Traces[side] {
				switch ev.Type {
				case TraceDrop:
					data, err := arena.DecodeData[DropData](ev)
					if err != nil {
						continue
					}
					for i := 0; i < ItemsCount; i++ {
						if data.Items[i] != 0 {
							addSpring2026TraceMetricCount(stats, deliveredKeys[i], side, data.Items[i])
						}
					}
				case TraceTrain:
					addSpring2026TraceMetricCount(stats, spring2026MetricGoblinsTrained, side, 1)
				case TraceChop:
					addSpring2026TraceMetricCount(stats, spring2026MetricChops, side, 1)
				case TraceHarvest:
					addSpring2026TraceMetricCount(stats, spring2026MetricHarvests, side, 1)
				case TracePlant:
					addSpring2026TraceMetricCount(stats, spring2026MetricPlants, side, 1)
				case TraceFailed:
					addSpring2026TraceMetricCount(stats, spring2026MetricFailedActions, side, 1)
				}
			}
		}
	}

	return stats
}

func addSpring2026TraceMetricCount(stats arena.TraceMetricStats, key string, side, delta int) {
	values := stats[key]
	values[side] += delta
	stats[key] = values
}

var _ arena.TraceMetricAnalyzer = (*factory)(nil)
