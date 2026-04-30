package arena

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const TraceAnalysisNoSide = -1

type TraceAnalysisEndReasonSpec struct {
	Key      string
	Label    string
	ShowZero bool
	ShowBlue bool
}

func StandardTraceEndReasons() []TraceAnalysisEndReasonSpec {
	return []TraceAnalysisEndReasonSpec{
		{Key: EndReasonTurnsOut, Label: EndReasonTurnsOut},
		{Key: EndReasonScore, Label: EndReasonScore},
		{Key: EndReasonScoreEarly, Label: EndReasonScoreEarly},
		{Key: EndReasonEliminated, Label: EndReasonEliminated, ShowBlue: true},
		{Key: EndReasonTimeoutStart, Label: EndReasonTimeoutStart, ShowZero: true, ShowBlue: true},
		{Key: EndReasonTimeout, Label: EndReasonTimeout, ShowZero: true, ShowBlue: true},
		{Key: EndReasonInvalid, Label: EndReasonInvalid, ShowZero: true, ShowBlue: true},
	}
}

func AnalyzeTraceFiles(input TraceAnalysisInput, metricAnalyzer TraceMetricAnalyzer) (TraceAnalysis, error) {
	specs, specByKey, err := normalizedTraceMetricSpecs(metricAnalyzer)
	if err != nil {
		return nil, err
	}

	report := &traceAnalysisReport{
		traceDir:            input.TraceDir,
		gameID:              input.GameID,
		files:               len(input.Files),
		endReasonSpecs:      StandardTraceEndReasons(),
		endReasonCounts:     make(map[string]int),
		endReasonBlueCounts: make(map[string]int),
		metricSpecs:         specs,
		winnerA:             make(map[string]traceAnalysisMetricAggregate),
		winnerB:             make(map[string]traceAnalysisMetricAggregate),
		blueA:               make(map[string]traceAnalysisMetricAggregate),
		blueB:               make(map[string]traceAnalysisMetricAggregate),
	}

	for _, file := range input.Files {
		var stats TraceMetricStats
		if metricAnalyzer != nil {
			stats, err = metricAnalyzer.AnalyzeTraceMetrics(file.Trace)
			if err != nil {
				return nil, fmt.Errorf("analyze %s: %w", file.Name, err)
			}
			if err := validateTraceMetricStats(file.Name, stats, specByKey, len(file.Trace.Turns)); err != nil {
				return nil, err
			}
		}
		report.add(file.Trace, stats)
	}

	return report, nil
}

func normalizedTraceMetricSpecs(metricAnalyzer TraceMetricAnalyzer) ([]TraceMetricSpec, map[string]TraceMetricSpec, error) {
	if metricAnalyzer == nil {
		return nil, nil, nil
	}

	specs := append([]TraceMetricSpec(nil), metricAnalyzer.TraceMetricSpecs()...)
	byKey := make(map[string]TraceMetricSpec, len(specs))
	for i := range specs {
		spec := &specs[i]
		if spec.Key == "" {
			return nil, nil, fmt.Errorf("trace metric spec %d has empty key", i)
		}
		if spec.Label == "" {
			spec.Label = spec.Key
		}
		switch spec.Kind {
		case TraceMetricPerMatchCount, TraceMetricPerTurnRate:
		default:
			return nil, nil, fmt.Errorf("trace metric %q has unsupported kind %q", spec.Key, spec.Kind)
		}
		if _, exists := byKey[spec.Key]; exists {
			return nil, nil, fmt.Errorf("duplicate trace metric spec %q", spec.Key)
		}
		byKey[spec.Key] = *spec
	}
	return specs, byKey, nil
}

func validateTraceMetricStats(file string, stats TraceMetricStats, specs map[string]TraceMetricSpec, turns int) error {
	for key, values := range stats {
		spec, ok := specs[key]
		if !ok {
			return fmt.Errorf("analyze %s: trace metric %q was returned without a spec", file, key)
		}
		for side, value := range values {
			if value < 0 {
				return fmt.Errorf("analyze %s: trace metric %q side %d is negative", file, key, side)
			}
			if spec.Kind == TraceMetricPerTurnRate && value > turns {
				return fmt.Errorf("analyze %s: trace metric %q side %d count %d exceeds turns %d", file, key, side, value, turns)
			}
		}
	}
	return nil
}

// TraceWinner returns the winner side for a 2-player trace, or -1 for a draw.
func TraceWinner(trace TraceMatch) int {
	if trace.Ranks[0] == trace.Ranks[1] {
		switch {
		case trace.Scores[0] > trace.Scores[1]:
			return 0
		case trace.Scores[1] > trace.Scores[0]:
			return 1
		default:
			return TraceAnalysisNoSide
		}
	}
	if trace.Ranks[0] < trace.Ranks[1] {
		return 0
	}
	return 1
}

// TraceEndReasonSide returns the side that a side-specific end reason applies
// to, or -1 when the reason is not side-specific or cannot be attributed.
func TraceEndReasonSide(trace TraceMatch, winner int) int {
	switch {
	case trace.Deactivated[0] && !trace.Deactivated[1]:
		return 0
	case trace.Deactivated[1] && !trace.Deactivated[0]:
		return 1
	case trace.Deactivated[0] && trace.Deactivated[1]:
		return TraceAnalysisNoSide
	}

	switch trace.EndReason {
	case EndReasonTimeoutStart, EndReasonTimeout, EndReasonInvalid, EndReasonEliminated:
		if winner == 0 {
			return 1
		}
		if winner == 1 {
			return 0
		}
	}
	return TraceAnalysisNoSide
}

type traceAnalysisReport struct {
	traceDir string
	gameID   string
	files    int

	decided  int
	draws    int
	sideWins [2]int

	blueMatches int
	blueWins    int
	blueLosses  int

	turnSum  int
	turnMin  int
	turnMax  int
	turnSeen bool

	scoreSum       [2]float64
	decidedMargin  float64
	timingMatches  int
	firstResponse  [2]float64
	turnResponse   [2]float64
	endReasonSpecs []TraceAnalysisEndReasonSpec

	endReasonCounts     map[string]int
	endReasonBlueCounts map[string]int

	metricSpecs []TraceMetricSpec
	winnerA     map[string]traceAnalysisMetricAggregate
	winnerB     map[string]traceAnalysisMetricAggregate
	blueA       map[string]traceAnalysisMetricAggregate
	blueB       map[string]traceAnalysisMetricAggregate
}

type traceAnalysisMetricAggregate struct {
	Sum     float64
	Samples int
}

func (r *traceAnalysisReport) add(trace TraceMatch, stats TraceMetricStats) {
	winner := TraceWinner(trace)
	if winner == 0 || winner == 1 {
		r.decided++
		r.sideWins[winner]++
		r.decidedMargin += math.Abs(float64(trace.Scores[0] - trace.Scores[1]))
	} else {
		r.draws++
	}

	blueSide := trace.BlueSide()
	if blueSide == 0 || blueSide == 1 {
		r.blueMatches++
		switch winner {
		case blueSide:
			r.blueWins++
		case 1 - blueSide:
			r.blueLosses++
		}
	}

	turns := len(trace.Turns)
	r.turnSum += turns
	if !r.turnSeen || turns < r.turnMin {
		r.turnMin = turns
		r.turnSeen = true
	}
	if turns > r.turnMax {
		r.turnMax = turns
	}

	r.scoreSum[0] += float64(trace.Scores[0])
	r.scoreSum[1] += float64(trace.Scores[1])

	if trace.Timing != nil {
		r.timingMatches++
		for side := 0; side < 2; side++ {
			r.firstResponse[side] += trace.Timing.FirstResponse[side]
			r.turnResponse[side] += trace.Timing.ResponseAverage[side]
		}
	}

	if trace.EndReason != "" {
		r.endReasonCounts[trace.EndReason]++
		if blueSide >= 0 && TraceEndReasonSide(trace, winner) == blueSide {
			r.endReasonBlueCounts[trace.EndReason]++
		}
	}

	for _, spec := range r.metricSpecs {
		values := stats[spec.Key]
		if winner == 0 || winner == 1 {
			loser := 1 - winner
			addTraceMetricSample(r.winnerA, spec, values[winner], turns)
			addTraceMetricSample(r.winnerB, spec, values[loser], turns)
		}
		if blueSide == 0 || blueSide == 1 {
			addTraceMetricSample(r.blueA, spec, values[blueSide], turns)
			addTraceMetricSample(r.blueB, spec, values[1-blueSide], turns)
		}
	}
}

func addTraceMetricSample(dst map[string]traceAnalysisMetricAggregate, spec TraceMetricSpec, value, turns int) {
	sample, ok := traceMetricSample(spec.Kind, value, turns)
	if !ok {
		return
	}
	current := dst[spec.Key]
	current.Sum += sample
	current.Samples++
	dst[spec.Key] = current
}

func traceMetricSample(kind TraceMetricKind, value, turns int) (float64, bool) {
	switch kind {
	case TraceMetricPerMatchCount:
		return float64(value), true
	case TraceMetricPerTurnRate:
		if turns == 0 {
			return 0, false
		}
		return float64(value) / float64(turns) * 100, true
	default:
		return 0, false
	}
}

func (r *traceAnalysisReport) Write(w io.Writer) error {
	title := r.gameID
	if title == "" {
		title = "Trace"
	}
	if _, err := fmt.Fprintf(w, "%s analysis: %d trace files analyzed [%s]\n",
		title, r.files, analysisTraceDirLabel(r.traceDir)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Decided matches: %.1f%% / Draws: %.1f%%\n",
		percent(r.decided, r.files), percent(r.draws, r.files)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Side wins: p0 %.1f%% / p1 %.1f%%\n",
		percent(r.sideWins[0], r.files), percent(r.sideWins[1], r.files)); err != nil {
		return err
	}
	if r.blueMatches > 0 {
		if _, err := fmt.Fprintf(w, "Blue: Wins: %.1f%% Losses: %.1f%% Draws: %.1f%%\n",
			percent(r.blueWins, r.blueMatches),
			percent(r.blueLosses, r.blueMatches),
			percent(r.blueMatches-r.blueWins-r.blueLosses, r.blueMatches)); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(w, "Blue: not identified in traces"); err != nil {
		return err
	}

	if err := r.writeGenericStats(w); err != nil {
		return err
	}
	if err := r.writeEndReasons(w); err != nil {
		return err
	}
	if len(r.metricSpecs) > 0 {
		if err := r.writeMetricComparison(w, "Winner vs loser metrics", r.winnerA, r.winnerB, "winner", "loser"); err != nil {
			return err
		}
		if err := r.writeMetricComparison(w, "Blue vs enemy metrics", r.blueA, r.blueB, "blue", "enemy"); err != nil {
			return err
		}
	}
	return nil
}

func (r *traceAnalysisReport) writeGenericStats(w io.Writer) error {
	if r.files == 0 {
		if _, err := fmt.Fprintln(w, "Turns: none"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Scores: none"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "Turns: avg %.1f min %d max %d\n",
		float64(r.turnSum)/float64(r.files), r.turnMin, r.turnMax); err != nil {
		return err
	}

	line := fmt.Sprintf("Scores: avg p0 %.1f p1 %.1f",
		r.scoreSum[0]/float64(r.files), r.scoreSum[1]/float64(r.files))
	if r.decided > 0 {
		line += fmt.Sprintf(" margin %.1f", r.decidedMargin/float64(r.decided))
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}

	if r.timingMatches > 0 {
		if _, err := fmt.Fprintf(w, "Timing: first_response %.0fmsx%.0fms avg_turn_response %.0fmsx%.0fms\n",
			r.firstResponse[0]/float64(r.timingMatches),
			r.firstResponse[1]/float64(r.timingMatches),
			r.turnResponse[0]/float64(r.timingMatches),
			r.turnResponse[1]/float64(r.timingMatches)); err != nil {
			return err
		}
	}
	return nil
}

func (r *traceAnalysisReport) writeEndReasons(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "\nEnd reasons:"); err != nil {
		return err
	}
	if len(r.endReasonCounts) == 0 {
		_, err := fmt.Fprintln(w, "  none recorded")
		return err
	}

	rows := r.endReasonRows()
	for _, row := range rows {
		if row.count == 0 && !row.spec.ShowZero {
			continue
		}
		line := fmt.Sprintf("  %-14s %5.1f%%", row.label(), percent(row.count, r.files))
		if row.count > 0 && row.spec.ShowBlue && r.blueMatches > 0 {
			line += fmt.Sprintf(" (blue: %.1f%%)", percent(row.blueCount, r.blueMatches))
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

type traceAnalysisEndReasonRow struct {
	key       string
	count     int
	blueCount int
	spec      TraceAnalysisEndReasonSpec
	order     int
}

func (r *traceAnalysisReport) endReasonRows() []traceAnalysisEndReasonRow {
	specs := make(map[string]TraceAnalysisEndReasonSpec, len(r.endReasonSpecs))
	orders := make(map[string]int, len(r.endReasonSpecs))
	for i, spec := range r.endReasonSpecs {
		specs[spec.Key] = spec
		orders[spec.Key] = i
	}

	seen := make(map[string]struct{}, len(r.endReasonCounts)+len(specs))
	rows := make([]traceAnalysisEndReasonRow, 0, len(r.endReasonCounts)+len(specs))
	add := func(key string) {
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		spec := specs[key]
		if spec.Key == "" {
			spec = TraceAnalysisEndReasonSpec{Key: key, Label: key}
		}
		order, ok := orders[key]
		if !ok {
			order = len(orders) + len(rows)
		}
		rows = append(rows, traceAnalysisEndReasonRow{
			key:       key,
			count:     r.endReasonCounts[key],
			blueCount: r.endReasonBlueCounts[key],
			spec:      spec,
			order:     order,
		})
	}

	for _, spec := range r.endReasonSpecs {
		add(spec.Key)
	}
	for key := range r.endReasonCounts {
		add(key)
	}

	sortTraceAnalysisEndReasons(rows)
	return rows
}

func sortTraceAnalysisEndReasons(rows []traceAnalysisEndReasonRow) {
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		j := i - 1
		for ; j >= 0 && traceAnalysisEndReasonLess(row, rows[j]); j-- {
			rows[j+1] = rows[j]
		}
		rows[j+1] = row
	}
}

func traceAnalysisEndReasonLess(a, b traceAnalysisEndReasonRow) bool {
	if a.count != b.count {
		return a.count > b.count
	}
	if a.order != b.order {
		return a.order < b.order
	}
	return a.label() < b.label()
}

func (r traceAnalysisEndReasonRow) label() string {
	if r.spec.Label != "" {
		return r.spec.Label
	}
	return r.key
}

func (r *traceAnalysisReport) writeMetricComparison(w io.Writer, title string,
	a, b map[string]traceAnalysisMetricAggregate, aLabel, bLabel string,
) error {
	if _, err := fmt.Fprintf(w, "\n%s:\n", title); err != nil {
		return err
	}

	wrote := false
	for _, spec := range r.metricSpecs {
		av := averageTraceMetric(a[spec.Key])
		bv := averageTraceMetric(b[spec.Key])
		if !av.ok && !bv.ok {
			continue
		}
		if !spec.ShowZero && av.value == 0 && bv.value == 0 {
			continue
		}
		wrote = true
		if _, err := fmt.Fprintf(w, "  %-14s %s %s  %s %s  (%s)\n",
			spec.Label,
			aLabel, formatTraceMetricValue(spec.Kind, av.value),
			bLabel, formatTraceMetricValue(spec.Kind, bv.value),
			traceAnalysisValueExplanation(av.value, bv.value, aLabel, bLabel)); err != nil {
			return err
		}
	}
	if !wrote {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	return nil
}

type traceAnalysisMetricAverage struct {
	value float64
	ok    bool
}

func averageTraceMetric(agg traceAnalysisMetricAggregate) traceAnalysisMetricAverage {
	if agg.Samples == 0 {
		return traceAnalysisMetricAverage{}
	}
	return traceAnalysisMetricAverage{value: agg.Sum / float64(agg.Samples), ok: true}
}

func formatTraceMetricValue(kind TraceMetricKind, value float64) string {
	switch kind {
	case TraceMetricPerMatchCount:
		return fmt.Sprintf("%5.2f/match", value)
	case TraceMetricPerTurnRate:
		return fmt.Sprintf("%5.1f%%", value)
	default:
		return fmt.Sprintf("%5.1f", value)
	}
}

func percent(num, denom int) float64 {
	if denom == 0 {
		return 0
	}
	return float64(num) / float64(denom) * 100
}

func analysisTraceDirLabel(dir string) string {
	if dir == "" {
		return "."
	}
	label := filepath.Clean(dir)
	if cwd, err := os.Getwd(); err == nil {
		if abs, err := filepath.Abs(label); err == nil {
			if rel, err := filepath.Rel(cwd, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				label = rel
			}
		}
	}
	if label == "." || filepath.IsAbs(label) || strings.HasPrefix(label, ".") {
		return label
	}
	return "." + string(filepath.Separator) + label
}

func traceAnalysisValueExplanation(a, b float64, aLabel, bLabel string) string {
	a = math.Round(a*10) / 10
	b = math.Round(b*10) / 10
	const epsilon = 0.0000001
	switch {
	case math.Abs(a-b) < epsilon:
		return "same value"
	case a == 0:
		return fmt.Sprintf("%s never; %s did", aLabel, bLabel)
	case b == 0:
		return fmt.Sprintf("%s did; %s never", aLabel, bLabel)
	case a < b:
		return fmt.Sprintf("%s only %.0f%% as often as %s; %s %.1fx %s",
			aLabel, a/b*100, bLabel, bLabel, b/a, aLabel)
	default:
		return fmt.Sprintf("%s %.1fx %s; %s only %.0f%% as often as %s",
			aLabel, a/b, bLabel, bLabel, b/a*100, aLabel)
	}
}
