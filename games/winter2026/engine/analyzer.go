package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Winter 2026.
//
// Every metric is per-match because the actionable cost in this game is
// segments lost over the match, not the rate of incidents. DEAD/HIT_* events
// each represent a discrete segment loss (1 per HIT, 3 per DEAD beheading);
// FALL accumulates the bird's body length at fall time since a fall kills
// regardless of length.
func (f *factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{Key: TraceDead, Label: TraceDead, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceHitEnemy, Label: TraceHitEnemy, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceHitWall, Label: TraceHitWall, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceHitSelf, Label: TraceHitSelf, Kind: arena.TraceMetricPerMatchCount},
		{Key: TraceFall, Label: TraceFall, Kind: arena.TraceMetricPerMatchCount},
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
		TraceDead:     {},
		TraceHitEnemy: {},
		TraceHitWall:  {},
		TraceHitSelf:  {},
		TraceFall:     {},
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
			case TraceDead, TraceHitEnemy, TraceHitWall, TraceHitSelf:
				addWinterTraceMetricCount(stats, ev.Type, side, 1)
			case TraceFall:
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
