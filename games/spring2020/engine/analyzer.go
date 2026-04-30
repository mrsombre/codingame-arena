package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

const (
	springMetricNoEatTurn = "NO_EAT_TURN"
	springMetricEatSuper  = "EAT_SUPER"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Spring 2020.
func (f *Factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{Key: TraceSwitch, Label: TraceSwitch, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceKilled, Label: TraceKilled, Kind: arena.TraceMetricPerMatchCount},
		{Key: springMetricEatSuper, Label: springMetricEatSuper, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceCollideSelf, Label: TraceCollideSelf, Kind: arena.TraceMetricPerTurnRate},
		{Key: TraceCollideEnemy, Label: TraceCollideEnemy, Kind: arena.TraceMetricPerTurnRate},
		{Key: springMetricNoEatTurn, Label: "NO_EAT", Kind: arena.TraceMetricPerTurnRate},
	}
}

// AnalyzeTraceMetrics interprets Spring 2020's typed trace metas and returns
// side-attributed metric counts. Per-turn metrics collapse duplicate same-side
// occurrences within a turn. Super-pellet eats are derived from EAT.cost > 1.
func (f *Factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeSpringTraceMetrics(trace), nil
}

func analyzeSpringTraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		TraceSwitch:           {},
		TraceKilled:           {},
		springMetricEatSuper:  {},
		TraceCollideSelf:      {},
		TraceCollideEnemy:     {},
		springMetricNoEatTurn: {},
	}

	for _, turn := range trace.Turns {
		eatBySide := [2]bool{}
		turnMetrics := map[string][2]bool{
			TraceCollideSelf:  {},
			TraceCollideEnemy: {},
		}

		for _, ev := range turn.Traces {
			pacID, ok := decodeSpringSubject(ev)
			if !ok {
				continue
			}
			side, ok := springPacSide(pacID)
			if !ok {
				continue
			}

			switch ev.Type {
			case TraceSwitch, TraceKilled:
				addTraceMetricCount(stats, ev.Type, side, 1)
			case TraceEat:
				eatBySide[side] = true
				meta, err := arena.DecodeMeta[EatMeta](ev)
				if err == nil && meta.Cost > 1 {
					addTraceMetricCount(stats, springMetricEatSuper, side, 1)
				}
			case TraceCollideSelf, TraceCollideEnemy:
				sides := turnMetrics[ev.Type]
				sides[side] = true
				turnMetrics[ev.Type] = sides
			}
		}

		for key, sides := range turnMetrics {
			for side, happened := range sides {
				if happened {
					addTraceMetricCount(stats, key, side, 1)
				}
			}
		}
		for side, ate := range eatBySide {
			if !ate {
				addTraceMetricCount(stats, springMetricNoEatTurn, side, 1)
			}
		}
	}

	return stats
}

// decodeSpringSubject extracts the subject pac ID from any Spring 2020 trace.
// Every spring meta carries a top-level "pac" field, so a single shape covers
// all event types.
func decodeSpringSubject(t arena.TurnTrace) (int, bool) {
	meta, err := arena.DecodeMeta[PacMeta](t)
	if err != nil {
		return 0, false
	}
	return meta.Pac, true
}

func springPacSide(pacID int) (int, bool) {
	if pacID < 0 {
		return 0, false
	}
	return pacID % 2, true
}

func addTraceMetricCount(stats arena.TraceMetricStats, key string, side, delta int) {
	values := stats[key]
	values[side] += delta
	stats[key] = values
}

var _ arena.TraceMetricAnalyzer = (*Factory)(nil)
