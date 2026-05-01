package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

// springEarlyTurnCutoff bounds the EAT_T50 milestone metric: only EAT events
// from turns strictly before this index count toward the early-game
// collection rate. 50 turns covers most of the opening phase before the
// pellet field thins out — matches typically run 60–200 turns, so 20 turns
// was too short to separate fast vs. slow openers.
const springEarlyTurnCutoff = 50

const (
	springMetricNoEatTurn = "NO_EAT_TURN"
	springMetricEatSuper  = "EAT_SUPER"
	springMetricEatT50    = "EAT_T50"
	springMetricKills     = "KILLS"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Spring 2020.
//
// Combat metrics (KILLED, KILLS) and ability use (SWITCH) are per-match
// counts because the actionable cost is the total over a match, not how
// often a side did something each turn. Eating metrics (EAT_SUPER, EAT_T50)
// are per-match counts of pellets collected; EAT_T50 isolates the early
// grab. Movement-quality metrics (HIT_SELF/HIT_ENEMY display labels for
// COLLIDE_SELF/COLLIDE_ENEMY trace keys, NO_EAT_TURN) are per-turn rates
// because the bug they surface is "fraction of turns wasted" rather than
// absolute count.
func (f *Factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{Key: TraceKilled, Label: TraceKilled, Kind: arena.TraceMetricPerMatchCount,
			Description: "Pacs killed (lost) per match in RPS collisions"},
		{Key: springMetricKills, Label: springMetricKills, Kind: arena.TraceMetricPerMatchCount,
			Description:    "Pacs killed (scored) per match in RPS collisions",
			HigherIsBetter: true},
		{Key: TraceSwitch, Label: TraceSwitch, Kind: arena.TraceMetricPerMatchCount,
			Description: "SWITCH ability activations per match (type change; shares 10-turn cooldown with SPEED)"},
		{Key: springMetricEatSuper, Label: springMetricEatSuper, Kind: arena.TraceMetricPerMatchCount,
			Description:    "Super-pellets eaten per match (worth 10 points each; usually 4 spawn per map)",
			HigherIsBetter: true},
		{Key: springMetricEatT50, Label: springMetricEatT50, Kind: arena.TraceMetricPerMatchCount,
			Description:    "Pellets and super-pellets eaten in the first 50 turns (early-game collection rate)",
			HigherIsBetter: true},
		{Key: TraceCollideSelf, Label: "HIT_SELF", Kind: arena.TraceMetricPerTurnRate,
			Description: "Turns where own pacs blocked each other (same type / same team collisions cancel both moves)"},
		{Key: TraceCollideEnemy, Label: "HIT_ENEMY", Kind: arena.TraceMetricPerTurnRate,
			Description: "Turns where enemy pacs blocked own pacs (same-type enemy collision cancels move)"},
		{Key: springMetricNoEatTurn, Label: "NO_EAT", Kind: arena.TraceMetricPerTurnRate,
			Description: "Turns where no pac on this side ate anything (idle / wasted turn rate)"},
	}
}

// AnalyzeTraceMetrics interprets Spring 2020's typed trace metas and returns
// side-attributed metric counts. Per-turn metrics collapse duplicate same-side
// occurrences within a turn. Super-pellet eats are derived from EAT.cost > 1.
// KILLS is derived from KilledMeta.Killer (the killer's side gets a kill, the
// victim's side gets a KILLED).
func (f *Factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeSpringTraceMetrics(trace), nil
}

func analyzeSpringTraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		TraceKilled:           {},
		springMetricKills:     {},
		TraceSwitch:           {},
		springMetricEatSuper:  {},
		springMetricEatT50:    {},
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
			case TraceSwitch:
				addTraceMetricCount(stats, ev.Type, side, 1)
			case TraceKilled:
				addTraceMetricCount(stats, TraceKilled, side, 1)
				meta, err := arena.DecodeMeta[KilledMeta](ev)
				if err == nil {
					if killerSide, ok := springPacSide(meta.Killer); ok {
						addTraceMetricCount(stats, springMetricKills, killerSide, 1)
					}
				}
			case TraceEat:
				eatBySide[side] = true
				if turn.Turn < springEarlyTurnCutoff {
					addTraceMetricCount(stats, springMetricEatT50, side, 1)
				}
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
