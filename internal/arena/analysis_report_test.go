package arena

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMetricDied  = "DIED"
	testMetricNoEat = "NO_EAT"
)

type testMetricAnalyzer struct {
	specs []TraceMetricSpec
	stats map[string]TraceMetricStats
}

func (a testMetricAnalyzer) TraceMetricSpecs() []TraceMetricSpec {
	return a.specs
}

func (a testMetricAnalyzer) AnalyzeTraceMetrics(trace TraceMatch) (TraceMetricStats, error) {
	return a.stats[trace.Type], nil
}

func runTestAnalysis(t *testing.T, files []TraceFile, analyzer TraceMetricAnalyzer) string {
	t.Helper()
	report, err := AnalyzeTraceFiles(
		TraceAnalysisInput{TraceDir: "traces", GameID: "test", Files: files},
		analyzer,
	)
	require.NoError(t, err)
	var out bytes.Buffer
	require.NoError(t, report.Write(&out))
	return out.String()
}

func testTurns(n int) []TraceTurn {
	turns := make([]TraceTurn, n)
	for i := range turns {
		turns[i].Turn = i
	}
	return turns
}

func TestAnalysisReportSummarizesGenericMatchStats(t *testing.T) {
	files := []TraceFile{
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason: EndReasonScore,
			Scores:    [2]TraceScore{10, 5}, Ranks: [2]int{0, 1},
			Timing: &TraceTiming{FirstResponse: [2]float64{100, 200}, ResponseAverage: [2]float64{10, 20}},
			Turns:  testTurns(2),
		}},
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason:   EndReasonTimeout,
			Deactivated: [2]bool{true, false},
			Scores:      [2]TraceScore{0, 7}, Ranks: [2]int{1, 0},
			Turns: testTurns(1),
		}},
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"rival", "us"},
			EndReason: EndReasonEliminated,
			Scores:    [2]TraceScore{0, 8}, Ranks: [2]int{1, 0},
			Turns: testTurns(3),
		}},
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason: EndReasonTurnsOut,
			Scores:    [2]TraceScore{4, 4}, Ranks: [2]int{0, 0},
		}},
	}

	text := runTestAnalysis(t, files, nil)

	assert.Contains(t, text, "test — 4 traces — ./traces")
	assert.Contains(t, text, "OUTCOME")
	assert.Contains(t, text, "Decided   75.0%   Draws   25.0%")
	assert.Contains(t, text, "Blue     W  50.0%   L  25.0%   D  25.0%")
	assert.NotContains(t, text, "Side wins")
	assert.Contains(t, text, "MATCH")
	assert.Contains(t, text, "Turns    avg 1.5   min 0   max 3")
	assert.Contains(t, text, "Scores   blue 5.5   red 4.0   margin 6.7")
	assert.Contains(t, text, "Timing   first  blue 100ms / red 200ms")
	assert.Contains(t, text, "turn   blue 10ms / red 20ms")
}

func TestAnalysisReportEndReasonsAttributeBlueFault(t *testing.T) {
	files := []TraceFile{
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason:   EndReasonTimeout,
			Deactivated: [2]bool{true, false},
			Scores:      [2]TraceScore{0, 5}, Ranks: [2]int{1, 0},
		}},
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason:   EndReasonInvalid,
			Deactivated: [2]bool{false, true},
			Scores:      [2]TraceScore{6, 0}, Ranks: [2]int{0, 1},
		}},
		{Trace: TraceMatch{
			Blue: "us", Players: [2]string{"us", "rival"},
			EndReason: EndReasonScore,
			Scores:    [2]TraceScore{10, 4}, Ranks: [2]int{0, 1},
		}},
	}

	text := runTestAnalysis(t, files, nil)

	assert.Contains(t, text, "END REASONS")
	assert.Contains(t, text, "SCORE")
	assert.Contains(t, text, "TIMEOUT         33.3%  (blue 33.3%)")
	assert.Contains(t, text, "INVALID         33.3%  (blue 0.0%)")
	assert.Contains(t, text, "TIMEOUT_START    0.0%")
}

func TestAnalysisReportAggregatesMetricKinds(t *testing.T) {
	files := []TraceFile{
		{Trace: TraceMatch{
			Type: "A",
			Blue: "us", Players: [2]string{"us", "rival"},
			Scores: [2]TraceScore{10, 4}, Ranks: [2]int{0, 1},
			Turns: testTurns(4),
		}},
		{Trace: TraceMatch{
			Type: "B",
			Blue: "us", Players: [2]string{"us", "rival"},
			Scores: [2]TraceScore{3, 8}, Ranks: [2]int{1, 0},
			Turns: testTurns(2),
		}},
	}
	analyzer := testMetricAnalyzer{
		specs: []TraceMetricSpec{
			{Key: testMetricDied, Label: "DIED", Kind: TraceMetricPerMatchCount},
			{Key: testMetricNoEat, Label: "NO_EAT", Kind: TraceMetricPerTurnRate},
		},
		stats: map[string]TraceMetricStats{
			"A": {
				testMetricDied:  [2]int{1, 3},
				testMetricNoEat: [2]int{1, 2},
			},
			"B": {
				testMetricDied:  [2]int{4, 2},
				testMetricNoEat: [2]int{2, 0},
			},
		},
	}

	text := runTestAnalysis(t, files, analyzer)

	assert.Contains(t, text, "METRICS — winner vs loser")
	assert.Contains(t, text, "DIED       winner  1.50/match   loser  3.50/match   (loser 2.33x winner)")
	assert.Contains(t, text, "NO_EAT     winner  12.5%   loser  75.0%   (loser 6.00x winner)")
	assert.Contains(t, text, "METRICS — blue vs red")
	assert.Contains(t, text, "DIED       blue  2.50/match   red  2.50/match   (equal)")
	assert.Contains(t, text, "NO_EAT     blue  62.5%   red  25.0%   (blue 2.50x red)")
}

func TestAnalysisReportRejectsPerTurnMetricAboveTurnCount(t *testing.T) {
	files := []TraceFile{{Name: "trace.json", Trace: TraceMatch{
		Type:  "bad",
		Blue:  "us", Players: [2]string{"us", "rival"},
		Turns: testTurns(2),
	}}}
	analyzer := testMetricAnalyzer{
		specs: []TraceMetricSpec{{Key: testMetricNoEat, Kind: TraceMetricPerTurnRate}},
		stats: map[string]TraceMetricStats{
			"bad": {testMetricNoEat: [2]int{3, 0}},
		},
	}

	_, err := AnalyzeTraceFiles(TraceAnalysisInput{TraceDir: "traces", GameID: "test", Files: files}, analyzer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count 3 exceeds turns 2")
}

func TestAnalysisReportSkipsPerTurnMetricsForZeroTurnMatches(t *testing.T) {
	files := []TraceFile{{Trace: TraceMatch{
		Type:   "zero",
		Blue:   "us", Players: [2]string{"us", "rival"},
		Scores: [2]TraceScore{1, 0}, Ranks: [2]int{0, 1},
	}}}
	analyzer := testMetricAnalyzer{
		specs: []TraceMetricSpec{{Key: testMetricNoEat, Label: "NO_EAT", Kind: TraceMetricPerTurnRate}},
		stats: map[string]TraceMetricStats{
			"zero": {testMetricNoEat: [2]int{0, 0}},
		},
	}

	text := runTestAnalysis(t, files, analyzer)

	assert.Contains(t, text, "METRICS — winner vs loser")
	assert.Contains(t, text, "  none")
}
