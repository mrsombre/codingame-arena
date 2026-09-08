package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func summer2026Summary(d TurnSummaryData) arena.TurnTrace {
	return arena.MakeTurnTrace(TraceTurnSummary, d)
}

func TestAnalyzeSummer2026SumsTurnSummaries(t *testing.T) {
	// The TURN summary is the per-turn ledger, so every economy metric is a
	// straight sum of its field across the match. PAINT_WASTED sums paintLeft
	// because paint does not carry over — what is left is gone.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{summer2026Summary(TurnSummaryData{
					Turn: 1, PaintAvailable: 10, PaintSpent: 8, PaintLeft: 2,
					TracksPlaced: 3, TracksAutoplaced: 1,
					DisruptAvailable: 1, DisruptSpent: 1,
					ConnectionsActive: 2, PointsEarned: 5, Score: 5,
				})},
				{summer2026Summary(TurnSummaryData{
					Turn: 1, PaintAvailable: 10, PaintSpent: 10, PaintLeft: 0,
					TracksPlaced: 4, TracksAutoplaced: 4,
					DisruptAvailable: 1, DisruptSpent: 0,
					ConnectionsActive: 1, PointsEarned: 3, Score: 3,
				})},
			}},
			{Turn: 2, Traces: [2][]arena.TurnTrace{
				{summer2026Summary(TurnSummaryData{
					Turn: 2, PaintAvailable: 12, PaintSpent: 12, PaintLeft: 0,
					TracksPlaced: 5, TracksAutoplaced: 0,
					DisruptAvailable: 1, DisruptSpent: 1,
					ConnectionsActive: 3, PointsEarned: 7, Score: 12,
				})},
				{summer2026Summary(TurnSummaryData{
					Turn: 2, PaintAvailable: 12, PaintSpent: 6, PaintLeft: 6,
					TracksPlaced: 2, TracksAutoplaced: 0,
					DisruptAvailable: 1, DisruptSpent: 1,
					ConnectionsActive: 1, PointsEarned: 2, Score: 5,
				})},
			}},
		},
	}

	stats := analyzeSummer2026TraceMetrics(trace)

	assert.Equal(t, [2]int{20, 16}, stats[summer2026MetricPaintSpent])
	assert.Equal(t, [2]int{2, 6}, stats[summer2026MetricPaintWasted])
	assert.Equal(t, [2]int{8, 6}, stats[summer2026MetricTracks])
	assert.Equal(t, [2]int{1, 4}, stats[summer2026MetricTracksAuto])
	assert.Equal(t, [2]int{2, 1}, stats[summer2026MetricDisrupts])
	assert.Equal(t, [2]int{5, 2}, stats[summer2026MetricConnections])
	assert.Equal(t, [2]int{12, 5}, stats[summer2026MetricPoints])
	// PAINT_FULL counts turns where the whole budget was spent: side 0 on
	// turn 2, side 1 on turn 1.
	assert.Equal(t, [2]int{1, 1}, stats[summer2026MetricPaintFull])
}

func TestAnalyzeSummer2026PaintFullSkipsBudgetlessTurns(t *testing.T) {
	// A turn granting no paint at all trivially ends with nothing left, but
	// that is not the bot using its budget — it must not inflate the rate.
	// The turn still lands in arena's denominator, so it counts against the
	// share rather than dropping out of it.
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{summer2026Summary(TurnSummaryData{Turn: 1, PaintAvailable: 0, PaintSpent: 0, PaintLeft: 0})},
				{summer2026Summary(TurnSummaryData{Turn: 1, PaintAvailable: 5, PaintSpent: 5, PaintLeft: 0})},
			}},
		},
	}

	stats := analyzeSummer2026TraceMetrics(trace)

	assert.Equal(t, [2]int{0, 1}, stats[summer2026MetricPaintFull])
}

func TestAnalyzeSummer2026AttributesMirroredEvents(t *testing.T) {
	// CONTESTED and INK are mirrored into both buckets, so counting them per
	// bucket would double them. CONTESTED costs both sides equally; INK is
	// credited only to the players named in the payload, and TRACKS_LOST is
	// read off the per-side slots regardless of credit.
	ink := arena.MakeTurnTrace(TraceInk, InkData{Zone: 4, Credited: []int{1}, TracksLost: [3]int{3, 2, 1}})
	contested := arena.MakeTurnTrace(TraceContested, ContestedData{Cell: [2]int{2, 3}})
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{contested, ink},
				{contested, ink},
			}},
		},
	}

	stats := analyzeSummer2026TraceMetrics(trace)

	assert.Equal(t, [2]int{1, 1}, stats[summer2026MetricContested])
	assert.Equal(t, [2]int{0, 1}, stats[summer2026MetricInked])
	assert.Equal(t, [2]int{3, 2}, stats[summer2026MetricTracksLost])
}

func TestAnalyzeSummer2026UncreditedInkStillCountsLostTracks(t *testing.T) {
	// A region that tips over on nobody's blot counts for neither side's
	// INKED, but the rails it swallows are still lost.
	ink := arena.MakeTurnTrace(TraceInk, InkData{Zone: 2, TracksLost: [3]int{1, 4, 0}})
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{{ink}, {ink}}},
		},
	}

	stats := analyzeSummer2026TraceMetrics(trace)

	assert.Equal(t, [2]int{0, 0}, stats[summer2026MetricInked])
	assert.Equal(t, [2]int{1, 4}, stats[summer2026MetricTracksLost])
}

func TestAnalyzeSummer2026CountsFailedCommands(t *testing.T) {
	trace := arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{
					arena.MakeTurnTrace(TraceFailed, FailedData{Reason: "Cannot place tracks on existing tracks at (2,3)"}),
					arena.MakeTurnTrace(TraceFailed, FailedData{Reason: "Not enough paint"}),
					arena.MakeTurnTrace(TraceMessage, MessageData{Text: "hi"}),
					arena.MakeTurnTrace(TraceTrack, TrackData{Cell: [2]int{1, 1}, Cost: 2, Terrain: "PLAINS"}),
					arena.MakeTurnTrace(TraceAutoplace, AutoplaceData{From: [2]int{0, 0}, To: [2]int{4, 4}, Placements: 6}),
					arena.MakeTurnTrace(TraceScore, ScoreData{From: 0, To: 1, Points: 3, PathLength: 7, Detour: 2}),
					arena.MakeTurnTrace(TraceDisrupt, DisruptData{Zone: 1, Instability: 2}),
				},
				{arena.MakeTurnTrace(TraceFailed, FailedData{Reason: "Zone 1 is already inked"})},
			}},
		},
	}

	stats := analyzeSummer2026TraceMetrics(trace)

	assert.Equal(t, [2]int{2, 1}, stats[summer2026MetricFailed])
	// Non-summary economy events must not be double counted alongside TURN.
	assert.Equal(t, [2]int{0, 0}, stats[summer2026MetricTracks])
	assert.Equal(t, [2]int{0, 0}, stats[summer2026MetricPoints])
	assert.Equal(t, [2]int{0, 0}, stats[summer2026MetricDisrupts])
}

func TestSummer2026TraceMetricSpecsMatchStats(t *testing.T) {
	// validateTraceMetricStats rejects a stat key without a spec, so the two
	// lists must stay in lockstep.
	f := &factory{}
	specs := f.TraceMetricSpecs()
	stats := analyzeSummer2026TraceMetrics(arena.TraceMatch{})

	specKeys := make(map[string]bool, len(specs))
	for _, spec := range specs {
		assert.NotEmpty(t, spec.Description, "metric %s needs a legend description", spec.Key)
		assert.False(t, specKeys[spec.Key], "duplicate spec %s", spec.Key)
		specKeys[spec.Key] = true
	}
	for key := range stats {
		assert.True(t, specKeys[key], "stat key %s missing from TraceMetricSpecs", key)
	}
	assert.Equal(t, len(specs), len(stats))
}

func TestSummer2026AnalyzerIsRegisteredOnFactory(t *testing.T) {
	analyzer, ok := NewFactory().(arena.TraceMetricAnalyzer)
	assert.True(t, ok, "summer2026 factory must implement arena.TraceMetricAnalyzer")

	stats, err := analyzer.AnalyzeTraceMetrics(arena.TraceMatch{
		Turns: []arena.TraceTurn{
			{Turn: 1, Traces: [2][]arena.TurnTrace{
				{summer2026Summary(TurnSummaryData{Turn: 1, PaintAvailable: 4, PaintSpent: 4, PointsEarned: 2})},
				nil,
			}},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, [2]int{4, 0}, stats[summer2026MetricPaintSpent])
	assert.Equal(t, [2]int{2, 0}, stats[summer2026MetricPoints])
}

// summer2026TurnPair builds one trace turn from a TURN summary per side.
func summer2026TurnPair(n int, left, right TurnSummaryData) arena.TraceTurn {
	left.Turn, right.Turn = n, n
	return arena.TraceTurn{Turn: n, Traces: [2][]arena.TurnTrace{
		{summer2026Summary(left)},
		{summer2026Summary(right)},
	}}
}

// summer2026ReportLine returns the report line whose first field is label,
// so an assertion can name the row without hardcoding column padding.
func summer2026ReportLine(t *testing.T, report, label string) string {
	t.Helper()
	for _, line := range strings.Split(report, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == label {
			return line
		}
	}
	t.Fatalf("report has no %q line:\n%s", label, report)
	return ""
}

// summer2026FixedTraceSet is three matches with hand-set per-turn ledgers:
// blue wins one from side 0, loses one from side 1, and times out on side 0.
// The side swap in the middle match is what makes the blue/red split a real
// assertion rather than a restatement of side 0 vs side 1.
func summer2026FixedTraceSet() []arena.TraceFile {
	strong := TurnSummaryData{
		PaintAvailable: 10, PaintSpent: 10, PaintLeft: 0,
		TracksPlaced: 2, ConnectionsActive: 1, PointsEarned: 5,
	}
	weak := TurnSummaryData{
		PaintAvailable: 10, PaintSpent: 5, PaintLeft: 5,
		TracksPlaced: 1, ConnectionsActive: 0, PointsEarned: 0,
	}

	return []arena.TraceFile{
		{Name: "trace-1.json", Trace: arena.TraceMatch{
			Blue: "blue", Players: [2]string{"blue", "red"},
			EndReason: arena.EndReasonScore,
			Scores:    [2]arena.TraceScore{10, 5}, Ranks: [2]int{0, 1},
			Turns: []arena.TraceTurn{
				summer2026TurnPair(1, strong, weak),
				summer2026TurnPair(2, strong, weak),
			},
		}},
		{Name: "trace-2.json", Trace: arena.TraceMatch{
			Blue: "blue", Players: [2]string{"red", "blue"},
			EndReason: arena.EndReasonTurnsOut,
			Scores:    [2]arena.TraceScore{8, 3}, Ranks: [2]int{0, 1},
			Turns: []arena.TraceTurn{
				summer2026TurnPair(1, strong, weak),
				summer2026TurnPair(2, strong, weak),
			},
		}},
		{Name: "trace-3.json", Trace: arena.TraceMatch{
			Blue: "blue", Players: [2]string{"blue", "red"},
			EndReason: arena.EndReasonTimeout, Disqualified: [2]bool{true, false},
			Scores: [2]arena.TraceScore{1, 9}, Ranks: [2]int{1, 0},
			Turns: []arena.TraceTurn{
				summer2026TurnPair(1,
					TurnSummaryData{PaintAvailable: 10, PaintSpent: 0, PaintLeft: 10},
					TurnSummaryData{
						PaintAvailable: 10, PaintSpent: 10, PaintLeft: 0,
						TracksPlaced: 3, ConnectionsActive: 2, PointsEarned: 4,
					}),
			},
		}},
	}
}

func TestSummer2026AnalyzeReportsBatchAggregates(t *testing.T) {
	report, err := arena.AnalyzeTraceFiles(arena.TraceAnalysisInput{
		TraceDir:   "traces",
		PuzzleName: "summer2026",
		Files:      summer2026FixedTraceSet(),
	}, NewFactory().(arena.TraceMetricAnalyzer))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, report.Write(&buf))
	out := buf.String()

	// The report opens with winner-vs-loser rows, so scope the assertions to
	// the blue-vs-red block before looking a metric up by name.
	blueVsRed := out[strings.Index(out, "METRICS — blue vs red"):]

	// Blue's per-match paint spend is 20/10/0 and red's is 10/20/10, so the
	// row must read 10.00 against 13.33. Following the raw side columns
	// instead of Blue would swap match 2's contributions.
	spent := summer2026ReportLine(t, blueVsRed, summer2026MetricPaintSpent)
	assert.Contains(t, spent, "blue    10.00/m")
	assert.Contains(t, spent, "red    13.33/m")

	wasted := summer2026ReportLine(t, blueVsRed, summer2026MetricPaintWasted)
	assert.Contains(t, wasted, "blue     6.67/m")
	assert.Contains(t, wasted, "red     3.33/m")

	// PAINT_FULL is a per-turn rate: blue empties its budget on both turns of
	// match 1 and on none of the other three turns.
	full := summer2026ReportLine(t, blueVsRed, summer2026MetricPaintFull)
	assert.Contains(t, full, "blue      33.3%")
	assert.Contains(t, full, "red      66.7%")

	tracks := summer2026ReportLine(t, blueVsRed, summer2026MetricTracks)
	assert.Contains(t, tracks, "blue     2.00/m")
	assert.Contains(t, tracks, "red     3.00/m")

	conns := summer2026ReportLine(t, blueVsRed, summer2026MetricConnections)
	assert.Contains(t, conns, "blue     0.67/m")
	assert.Contains(t, conns, "red     1.33/m")

	points := summer2026ReportLine(t, blueVsRed, summer2026MetricPoints)
	assert.Contains(t, points, "blue     3.33/m")
	assert.Contains(t, points, "red     4.67/m")

	// Every metric carries a legend entry, and the WORST section points at the
	// match to open for blue's worst paint waste — match 2 and match 3 both
	// waste 10, and the first seen wins.
	assert.Contains(t, out, "Metrics:")
	assert.Contains(t, out, "- "+summer2026MetricPaintWasted+": Paint left unspent per match")
	worst := out[strings.Index(out, "WORST"):]
	assert.Contains(t, summer2026ReportLine(t, worst, summer2026MetricPaintWasted), "  10   2")
}

func TestSummer2026PaintFullCountsBudgetlessTurnsInTheDenominator(t *testing.T) {
	// Arena divides a per-turn rate by every turn in the match, so blue
	// emptying its budget on one of two turns reads 50%, not the 100% a
	// metric that dropped the budgetless turn entirely would report.
	report, err := arena.AnalyzeTraceFiles(arena.TraceAnalysisInput{
		TraceDir:   "traces",
		PuzzleName: "summer2026",
		Files: []arena.TraceFile{{Name: "trace-1.json", Trace: arena.TraceMatch{
			Blue: "blue", Players: [2]string{"blue", "red"},
			EndReason: arena.EndReasonScore,
			Scores:    [2]arena.TraceScore{5, 1}, Ranks: [2]int{0, 1},
			Turns: []arena.TraceTurn{
				summer2026TurnPair(1,
					TurnSummaryData{PaintAvailable: 0, PaintSpent: 0, PaintLeft: 0},
					TurnSummaryData{PaintAvailable: 5, PaintSpent: 1, PaintLeft: 4}),
				summer2026TurnPair(2,
					TurnSummaryData{PaintAvailable: 5, PaintSpent: 5, PaintLeft: 0},
					TurnSummaryData{PaintAvailable: 5, PaintSpent: 1, PaintLeft: 4}),
			},
		}}},
	}, NewFactory().(arena.TraceMetricAnalyzer))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, report.Write(&buf))
	out := buf.String()

	blueVsRed := out[strings.Index(out, "METRICS — blue vs red"):]
	full := summer2026ReportLine(t, blueVsRed, summer2026MetricPaintFull)
	assert.Contains(t, full, "blue      50.0%")
	assert.Contains(t, full, "red       0.0%")
}

func TestSummer2026AnalyzeReportsEndReasonDistribution(t *testing.T) {
	report, err := arena.AnalyzeTraceFiles(arena.TraceAnalysisInput{
		TraceDir:   "traces",
		PuzzleName: "summer2026",
		Files:      summer2026FixedTraceSet(),
	}, NewFactory().(arena.TraceMetricAnalyzer))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, report.Write(&buf))
	out := buf.String()

	assert.Contains(t, out, "END REASONS")
	assert.Contains(t, summer2026ReportLine(t, out, arena.EndReasonScore), "33.3%")
	assert.Contains(t, summer2026ReportLine(t, out, arena.EndReasonTurnsOut), "33.3%")
	// The timeout is blue's own (side 0 was deactivated and blue played side
	// 0), and TIMEOUT lists the matches so the author can open the trace.
	timeout := summer2026ReportLine(t, out, arena.EndReasonTimeout)
	assert.Contains(t, timeout, "33.3%")
	assert.Contains(t, timeout, "(blue 100.0%)")
	assert.Contains(t, timeout, "3")
}
