package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

const (
	summer2026MetricPaintSpent  = "PAINT_SPENT"
	summer2026MetricPaintWasted = "PAINT_WASTED"
	summer2026MetricPaintFull   = "PAINT_FULL"
	summer2026MetricTracks      = "TRACKS"
	summer2026MetricTracksAuto  = "TRACKS_AUTO"
	summer2026MetricTracksLost  = "TRACKS_LOST"
	summer2026MetricConnections = "CONNECTIONS"
	summer2026MetricPoints      = "POINTS"
	summer2026MetricDisrupts    = "DISRUPTS"
	summer2026MetricInked       = "INKED"
	summer2026MetricContested   = "CONTESTED"
	summer2026MetricFailed      = "FAILED"
)

// TraceMetricSpecs implements arena.TraceMetricAnalyzer for Summer 2026.
//
// Every economy metric is a per-match sum of the TURN summary's per-turn
// ledger, so a value read against the report's average turn count gives the
// per-turn rate. Matches run at most MAX_TURNS turns and both sides get the
// same income schedule, so the sums are directly comparable between sides.
//
// The paint group is the utilisation axis. Paint is granted fresh each turn
// and does not carry over, so PAINT_SPENT and PAINT_WASTED partition the
// budget: what the bot converted into rails against what it threw away.
// PAINT_FULL is the same question at turn granularity — the share of turns
// the bot emptied its budget — which separates a bot that wastes a little
// every turn from one that stalls completely on a few. Arena divides a
// per-turn rate by every turn in the match, so a turn that granted no paint
// at all counts against the rate rather than being dropped from it: the
// metric reads "turns that ended with the budget emptied", not "turns with a
// budget that ended emptied".
//
// TRACKS counts every rail paid for, TRACKS_AUTO the subset the AUTOPLACE
// planner chose rather than the bot naming cells itself; the gap between them
// is how much routing the bot delegates. TRACKS_LOST is rails swallowed by an
// inked region, the only way a placed rail is ever taken back, and is a
// hazard: it is paint already spent and now unpaid.
//
// CONNECTIONS is connection-payouts summed over turns (a connection standing
// for ten turns counts ten), which is the throughput POINTS is earned from.
// DISRUPTS counts honoured blots — the offensive spend — and INKED the
// regions that actually tipped over on this side's blot, so the two together
// show whether the disruption budget converted.
//
// CONTESTED is cells both players bought on the same turn: nobody is refunded
// and the cell scores for neither, so it is a mutual waste and reads equal for
// both sides by construction. FAILED counts commands the engine rejected
// without disqualifying — a high count usually flags budget or legality bugs.
func (f *factory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{
		{
			Key: summer2026MetricPaintSpent, Label: summer2026MetricPaintSpent,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Paint spent on rails per match (sum of TURN.paintSpent)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricPaintWasted, Label: summer2026MetricPaintWasted,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Paint left unspent per match (sum of TURN.paintLeft; paint does not carry over — lower is better)",
		},
		{
			Key: summer2026MetricPaintFull, Label: summer2026MetricPaintFull,
			Kind:           arena.TraceMetricPerTurnRate,
			Description:    "Share of match turns that ended with the whole paint budget spent (a turn granting no paint counts against the share)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricTracks, Label: summer2026MetricTracks,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Rails placed per match (sum of TURN.tracksPlaced; includes autoplaced rails)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricTracksAuto, Label: summer2026MetricTracksAuto,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Rails placed by the AUTOPLACE planner per match (subset of TRACKS)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricTracksLost, Label: summer2026MetricTracksLost,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Rails swallowed by an inked region per match (sum of INK.tracksLost for this side — lower is better)",
		},
		{
			Key: summer2026MetricConnections, Label: summer2026MetricConnections,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Connection payouts per match (sum of TURN.connectionsActive; one standing connection counts once per turn)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricPoints, Label: summer2026MetricPoints,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Points earned per match (sum of TURN.pointsEarned)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricDisrupts, Label: summer2026MetricDisrupts,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Disruptions honoured per match (sum of TURN.disruptSpent)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricInked, Label: summer2026MetricInked,
			Kind:           arena.TraceMetricPerMatchCount,
			Description:    "Regions inked out on this side's blot per match (INK events crediting this side)",
			HigherIsBetter: true,
		},
		{
			Key: summer2026MetricContested, Label: summer2026MetricContested,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Cells both players bought on the same turn per match (neither is refunded, the cell goes neutral — lower is better)",
		},
		{
			Key: summer2026MetricFailed, Label: summer2026MetricFailed,
			Kind:        arena.TraceMetricPerMatchCount,
			Description: "Commands rejected without disqualification per match (count of FAILED events — lower is better)",
		},
	}
}

// AnalyzeTraceMetrics walks the trace once and accumulates per-side match
// totals. The TURN summary is the source of truth for the economy metrics
// rather than the individual TRACK / SCORE / DISRUPT events, so a deactivated
// player — who still gets a summary but emits no events — is counted with the
// same zeroes the engine saw.
func (f *factory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	return analyzeSummer2026TraceMetrics(trace), nil
}

func analyzeSummer2026TraceMetrics(trace arena.TraceMatch) arena.TraceMetricStats {
	stats := arena.TraceMetricStats{
		summer2026MetricPaintSpent:  {},
		summer2026MetricPaintWasted: {},
		summer2026MetricPaintFull:   {},
		summer2026MetricTracks:      {},
		summer2026MetricTracksAuto:  {},
		summer2026MetricTracksLost:  {},
		summer2026MetricConnections: {},
		summer2026MetricPoints:      {},
		summer2026MetricDisrupts:    {},
		summer2026MetricInked:       {},
		summer2026MetricContested:   {},
		summer2026MetricFailed:      {},
	}

	for _, turn := range trace.Turns {
		for side := 0; side < 2; side++ {
			for _, ev := range turn.Traces[side] {
				switch ev.Type {
				case TraceTurnSummary:
					data, err := arena.DecodeData[TurnSummaryData](ev)
					if err != nil {
						continue
					}
					addSummer2026TraceMetricCount(stats, summer2026MetricPaintSpent, side, data.PaintSpent)
					addSummer2026TraceMetricCount(stats, summer2026MetricPaintWasted, side, data.PaintLeft)
					addSummer2026TraceMetricCount(stats, summer2026MetricTracks, side, data.TracksPlaced)
					addSummer2026TraceMetricCount(stats, summer2026MetricTracksAuto, side, data.TracksAutoplaced)
					addSummer2026TraceMetricCount(stats, summer2026MetricConnections, side, data.ConnectionsActive)
					addSummer2026TraceMetricCount(stats, summer2026MetricPoints, side, data.PointsEarned)
					addSummer2026TraceMetricCount(stats, summer2026MetricDisrupts, side, data.DisruptSpent)
					if data.PaintAvailable > 0 && data.PaintLeft == 0 {
						addSummer2026TraceMetricCount(stats, summer2026MetricPaintFull, side, 1)
					}
				case TraceFailed:
					addSummer2026TraceMetricCount(stats, summer2026MetricFailed, side, 1)
				case TraceContested:
					// Mirrored into both buckets: read it once and charge both
					// sides, since both paid for the cell and neither kept it.
					if side != 0 {
						continue
					}
					addSummer2026TraceMetricCount(stats, summer2026MetricContested, 0, 1)
					addSummer2026TraceMetricCount(stats, summer2026MetricContested, 1, 1)
				case TraceInk:
					// Also mirrored, so attribution comes from the payload
					// rather than the bucket it was found in.
					if side != 0 {
						continue
					}
					data, err := arena.DecodeData[InkData](ev)
					if err != nil {
						continue
					}
					for _, credited := range data.Credited {
						if credited < 0 || credited > 1 {
							continue
						}
						addSummer2026TraceMetricCount(stats, summer2026MetricInked, credited, 1)
					}
					for owner := 0; owner < 2; owner++ {
						addSummer2026TraceMetricCount(stats, summer2026MetricTracksLost, owner, data.TracksLost[owner])
					}
				}
			}
		}
	}

	return stats
}

func addSummer2026TraceMetricCount(stats arena.TraceMetricStats, key string, side, delta int) {
	values := stats[key]
	values[side] += delta
	stats[key] = values
}

var _ arena.TraceMetricAnalyzer = (*factory)(nil)
