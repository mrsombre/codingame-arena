package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

const winterMetricNoEatTurn = "NO_EAT_TURN"

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Winter 2026.
func (f *factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{Key: TraceDead, Label: TraceDead, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceHitEnemy, Label: TraceHitEnemy, Kind: arena.TraceMetricPerTurnRate},
		{Key: TraceHitWall, Label: TraceHitWall, Kind: arena.TraceMetricPerTurnRate},
		{Key: TraceHitSelf, Label: TraceHitSelf, Kind: arena.TraceMetricPerTurnRate},
		{Key: TraceFall, Label: TraceFall, Kind: arena.TraceMetricPerTurnRate},
		{Key: winterMetricNoEatTurn, Label: "NO_EAT", Kind: arena.TraceMetricPerTurnRate},
	}
}

// AnalyzeTraceMetrics interprets Winter 2026's typed trace metas and returns
// side-attributed metric counts. Per-turn metrics collapse duplicate same-side
// occurrences within a turn.
func (f *factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeWinterTraceMetrics(trace), nil
}

func analyzeWinterTraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		TraceDead:             {},
		TraceHitEnemy:         {},
		TraceHitWall:          {},
		TraceHitSelf:          {},
		TraceFall:             {},
		winterMetricNoEatTurn: {},
	}

	birdSide := winterBirdSideMap(trace)
	if len(birdSide) == 0 {
		return stats
	}

	for _, turn := range trace.Turns {
		eatBySide := [2]bool{}
		turnMetrics := map[string][2]bool{
			TraceHitEnemy: {},
			TraceHitWall:  {},
			TraceHitSelf:  {},
			TraceFall:     {},
		}

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
			case TraceDead:
				addWinterTraceMetricCount(stats, ev.Type, side, 1)
			case TraceEat:
				eatBySide[side] = true
			case TraceHitEnemy, TraceHitWall, TraceHitSelf, TraceFall:
				sides := turnMetrics[ev.Type]
				sides[side] = true
				turnMetrics[ev.Type] = sides
			}
		}

		for key, sides := range turnMetrics {
			for side, happened := range sides {
				if happened {
					addWinterTraceMetricCount(stats, key, side, 1)
				}
			}
		}
		for side, ate := range eatBySide {
			if !ate {
				addWinterTraceMetricCount(stats, winterMetricNoEatTurn, side, 1)
			}
		}
	}

	return stats
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
