package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// winterEarlyTurnCutoff bounds the EAT_T20 milestone metric: only EAT events
// from turns strictly before this index count toward the early-game grow rate.
// 20 turns is roughly the opening phase of a winter2026 match before bots
// settle into long-range routing — matches typically run 100–200 turns.
const winterEarlyTurnCutoff = 20

const (
	winterMetricEatByTurn20 = "EAT_T20"
	winterMetricDeadWall    = "DEAD_WALL"
	winterMetricDeadEnemy   = "DEAD_ENEMY"
	winterMetricDeadSelf    = "DEAD_SELF"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Winter 2026.
//
// Every metric is per-match because the actionable cost in this game is
// segments lost over the match, not the rate of incidents. DEAD_* and HIT_*
// events each represent a discrete segment loss (1 per HIT, 3 per DEAD
// beheading) split by hazard so wall/enemy/self bugs surface separately.
// FALL accumulates the bird's body length at fall time since a fall kills
// regardless of length and is not a beheading — it stays separate from the
// DEAD_* breakdown. EAT_T20 is a milestone count of power sources eaten in
// the opening phase, isolating early-game grow rate from the rest of the
// match.
func (f *factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{Key: winterMetricDeadSelf, Label: winterMetricDeadSelf, Kind: arena.TraceMetricPerMatchCount,
			Description: "Snake deaths per match from crashing into own or ally body"},
		{Key: winterMetricDeadWall, Label: winterMetricDeadWall, Kind: arena.TraceMetricPerMatchCount,
			Description: "Snake deaths per match from crashing into a wall"},
		{Key: winterMetricDeadEnemy, Label: winterMetricDeadEnemy, Kind: arena.TraceMetricPerMatchCount,
			Description: "Snake deaths per match from crashing into an enemy snake"},
		{Key: TraceDeadFall, Label: TraceDeadFall, Kind: arena.TraceMetricPerMatchCount,
			Description: "Segments lost per match from snakes falling off the grid (variable per death, unlike DEAD_* beheadings which always lose 3)"},
		{Key: TraceHitSelf, Label: TraceHitSelf, Kind: arena.TraceMetricPerMatchCount,
			Description: "Segments lost per match from crashing into own or ally body"},
		{Key: TraceHitWall, Label: TraceHitWall, Kind: arena.TraceMetricPerMatchCount,
			Description: "Segments lost per match from crashing into a wall"},
		{Key: TraceHitEnemy, Label: TraceHitEnemy, Kind: arena.TraceMetricPerMatchCount,
			Description: "Segments lost per match from crashing into an enemy snake"},
		{Key: winterMetricEatByTurn20, Label: winterMetricEatByTurn20, Kind: arena.TraceMetricPerMatchCount,
			Description:    "Power sources eaten in the first 20 turns (early-game grow rate)",
			HigherIsBetter: true},
	}
}

// AnalyzeTraceMetrics interprets Winter 2026's typed trace metas and returns
// side-attributed match-total counts. HIT_*/DEAD increment by 1 per event;
// FALL increments by the segment count carried in BirdSegmentsMeta.
func (f *factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeWinterTraceMetrics(trace), nil
}

func analyzeWinterTraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		winterMetricDeadSelf:    {},
		winterMetricDeadWall:    {},
		winterMetricDeadEnemy:   {},
		TraceDeadFall:           {},
		TraceHitSelf:            {},
		TraceHitWall:            {},
		TraceHitEnemy:           {},
		winterMetricEatByTurn20: {},
	}

	birdSide := winterBirdSideMap(trace)
	if len(birdSide) == 0 {
		return stats
	}

	for _, turn := range trace.Turns {
		for _, ev := range turn.Traces {
			birdID, ok := decodeWinterSubject(ev)
			if !ok {
				continue
			}
			side, ok := birdSide[birdID]
			if !ok {
				continue
			}

			switch ev.Type {
			case TraceHitSelf, TraceHitWall, TraceHitEnemy:
				addWinterTraceMetricCount(stats, ev.Type, side, 1)
			case TraceDead:
				meta, err := arena.DecodeMeta[BirdDeathMeta](ev)
				if err != nil {
					continue
				}
				key, ok := winterDeadKeyForCause(meta.Cause)
				if !ok {
					continue
				}
				addWinterTraceMetricCount(stats, key, side, 1)
			case TraceEat:
				if turn.Turn < winterEarlyTurnCutoff {
					addWinterTraceMetricCount(stats, winterMetricEatByTurn20, side, 1)
				}
			case TraceDeadFall:
				meta, err := arena.DecodeMeta[BirdSegmentsMeta](ev)
				if err != nil {
					continue
				}
				addWinterTraceMetricCount(stats, ev.Type, side, meta.Segments)
			}
		}
	}

	return stats
}

// winterDeadKeyForCause maps a BirdDeathMeta.Cause string to the matching
// metric key. Unknown causes are dropped (returns ok=false) so a future
// engine extension that adds a new cause silently no-ops in old analyzers
// rather than crashing or miscategorizing.
func winterDeadKeyForCause(cause string) (string, bool) {
	switch cause {
	case DeathCauseWall:
		return winterMetricDeadWall, true
	case DeathCauseEnemy:
		return winterMetricDeadEnemy, true
	case DeathCauseSelf:
		return winterMetricDeadSelf, true
	default:
		return "", false
	}
}

// decodeWinterSubject extracts the subject bird ID from any Winter 2026 trace.
// Every winter meta carries a top-level "bird" field, so a single shape covers
// all event types.
func decodeWinterSubject(t arena.TurnTrace) (int, bool) {
	meta, err := arena.DecodeMeta[BirdMeta](t)
	if err != nil {
		return 0, false
	}
	return meta.Bird, true
}

// winterBirdSideMap reads the per-turn command outputs to learn which side
// owns each bird ID. A move command starts with `<birdID> <DIR>`; only the
// bird's owner can issue commands for it, so the leading int identifies
// ownership unambiguously. MARK commands (which lead with `MARK`) are skipped.
func winterBirdSideMap(trace arena.TraceMatch) map[int]int {
	birdSide := make(map[int]int)
	for _, turn := range trace.Turns {
		for side := 0; side < 2; side++ {
			for _, cmd := range winterSplitCommands(turn.Output[side]) {
				birdID, ok := parseLeadingBirdID(cmd)
				if !ok {
					continue
				}
				birdSide[birdID] = side
			}
		}
	}
	return birdSide
}

func winterSplitCommands(output string) []string {
	if output == "" {
		return nil
	}
	replacer := strings.NewReplacer(";", "\n", "\r", "\n")
	lines := strings.Split(replacer.Replace(output), "\n")
	commands := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		commands = append(commands, line)
	}
	return commands
}

// parseLeadingBirdID extracts the first whitespace-delimited integer from a
// command line. Used only by winterBirdSideMap; trace metas are decoded via
// DecodeMeta and do not need string parsing.
func parseLeadingBirdID(cmd string) (int, bool) {
	head, _, _ := strings.Cut(cmd, " ")
	n, err := strconv.Atoi(head)
	if err != nil {
		return 0, false
	}
	return n, true
}

func addWinterTraceMetricCount(stats arena.TraceMetricStats, key string, side, delta int) {
	values := stats[key]
	values[side] += delta
	stats[key] = values
}

var _ arena.TraceMetricAnalyzer = (*factory)(nil)
