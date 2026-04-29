package engine

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AnalyzeTraces implements arena.TraceAnalyzer for Spring 2020 traces.
func (f *Factory) AnalyzeTraces(input arena.TraceAnalysisInput) (arena.TraceAnalysis, error) {
	report := springAnalyzeReport{
		TraceDir: input.TraceDir,
		Files:    len(input.Files),
		Winner:   newSpringAnalyzeSideAggregate(),
		Loser:    newSpringAnalyzeSideAggregate(),
	}

	for _, file := range input.Files {
		trace := file.Trace
		winner := springTraceWinner(trace)
		if winner < 0 {
			report.Draws++
			continue
		}

		loser := 1 - winner
		sideStats := analyzeSpringTraceSides(trace)

		report.Decided++
		report.SideWins[winner]++
		report.TurnSum += len(trace.Turns)
		report.MarginSum += sideStats[winner].Score - sideStats[loser].Score
		report.Winner.add(sideStats[winner])
		report.Loser.add(sideStats[loser])
	}

	return report, nil
}

type springAnalyzeReport struct {
	TraceDir string
	Files    int
	Decided  int
	Draws    int
	SideWins [2]int

	TurnSum   int
	MarginSum float64
	Winner    springAnalyzeSideAggregate
	Loser     springAnalyzeSideAggregate
}

type springAnalyzeSideAggregate struct {
	Matches       int
	ScoreSum      float64
	CommandCounts map[string]int
	EventCounts   map[string]int
}

type springAnalyzeSideMatchStats struct {
	Score         float64
	CommandCounts map[string]int
	EventCounts   map[string]int
}

func springTraceWinner(trace arena.TraceMatch) int {
	if trace.Ranks[0] == trace.Ranks[1] {
		switch {
		case trace.Scores[0] > trace.Scores[1]:
			return 0
		case trace.Scores[1] > trace.Scores[0]:
			return 1
		default:
			return -1
		}
	}
	if trace.Ranks[0] < trace.Ranks[1] {
		return 0
	}
	return 1
}

func analyzeSpringTraceSides(trace arena.TraceMatch) [2]springAnalyzeSideMatchStats {
	stats := [2]springAnalyzeSideMatchStats{
		newSpringAnalyzeSideMatchStats(float64(trace.Scores[0])),
		newSpringAnalyzeSideMatchStats(float64(trace.Scores[1])),
	}

	for _, turn := range trace.Turns {
		stats[0].addOutput(turn.P0Output)
		stats[1].addOutput(turn.P1Output)
	}

	if trace.TraceSummary != nil {
		summary := *trace.TraceSummary
		for side := 0; side < 2; side++ {
			for label, units := range summary[side] {
				for _, turns := range units {
					stats[side].EventCounts[label] += len(turns)
				}
			}
		}
	}

	return stats
}

func newSpringAnalyzeSideAggregate() springAnalyzeSideAggregate {
	return springAnalyzeSideAggregate{
		CommandCounts: make(map[string]int),
		EventCounts:   make(map[string]int),
	}
}

func (a *springAnalyzeSideAggregate) add(stats springAnalyzeSideMatchStats) {
	a.Matches++
	a.ScoreSum += stats.Score
	addSpringAnalyzeCounts(a.CommandCounts, stats.CommandCounts)
	addSpringAnalyzeCounts(a.EventCounts, stats.EventCounts)
}

func newSpringAnalyzeSideMatchStats(score float64) springAnalyzeSideMatchStats {
	return springAnalyzeSideMatchStats{
		Score:         score,
		CommandCounts: make(map[string]int),
		EventCounts:   make(map[string]int),
	}
}

func (s *springAnalyzeSideMatchStats) addOutput(output string) {
	for _, command := range springOutputCommands(output) {
		s.CommandCounts[command]++
	}
}

func springOutputCommands(output string) []string {
	replacer := strings.NewReplacer("|", "\n", ";", "\n", "\r", "\n")
	lines := strings.Split(replacer.Replace(output), "\n")

	commands := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		commands = append(commands, strings.ToUpper(fields[0]))
	}
	return commands
}

func addSpringAnalyzeCounts(dst, src map[string]int) {
	for label, n := range src {
		dst[label] += n
	}
}

func (r springAnalyzeReport) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "Spring 2020 trace analysis: %d trace files from %s\n", r.Files, r.TraceDir); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Decided matches: %d  draws: %d\n", r.Decided, r.Draws); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Side wins: p0=%d p1=%d\n", r.SideWins[0], r.SideWins[1]); err != nil {
		return err
	}
	if r.Decided == 0 {
		_, err := fmt.Fprintln(w, "No decided matches to compare winner strengths.")
		return err
	}

	winnerScore := r.Winner.ScoreSum / float64(r.Decided)
	loserScore := r.Loser.ScoreSum / float64(r.Decided)
	avgTurns := float64(r.TurnSum) / float64(r.Decided)
	avgMargin := r.MarginSum / float64(r.Decided)
	if _, err := fmt.Fprintf(w, "Winner score: avg %.1f vs loser %.1f (margin +%.1f), avg turns %.1f\n",
		winnerScore, loserScore, avgMargin, avgTurns); err != nil {
		return err
	}

	if err := writeSpringRateRows(w, "Winner command rates (% of decisions per side)", springRateDiffRows(
		r.Winner.CommandCounts,
		r.Loser.CommandCounts,
	)); err != nil {
		return err
	}
	if err := writeSpringDiffRows(w, "Winner pac events (avg per decided match)", springCountDiffRows(
		r.Winner.EventCounts,
		r.Loser.EventCounts,
		r.Decided,
	)); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "Command rates normalize for surviving pacs; events are absolute counts. Diff = winner - loser.")
	return err
}

type springAnalyzeDiffRow struct {
	Label  string
	Winner float64
	Loser  float64
	Diff   float64
}

func springCountDiffRows(winner, loser map[string]int, matches int) []springAnalyzeDiffRow {
	if matches == 0 {
		return nil
	}

	labels := make(map[string]struct{}, len(winner)+len(loser))
	for label := range winner {
		labels[label] = struct{}{}
	}
	for label := range loser {
		labels[label] = struct{}{}
	}

	rows := make([]springAnalyzeDiffRow, 0, len(labels))
	for label := range labels {
		w := float64(winner[label]) / float64(matches)
		l := float64(loser[label]) / float64(matches)
		rows = append(rows, springAnalyzeDiffRow{
			Label:  label,
			Winner: w,
			Loser:  l,
			Diff:   w - l,
		})
	}
	sortSpringDiffRows(rows)
	return rows
}

func sortSpringDiffRows(rows []springAnalyzeDiffRow) {
	sort.Slice(rows, func(i, j int) bool {
		ai := math.Abs(rows[i].Diff)
		aj := math.Abs(rows[j].Diff)
		if ai == aj {
			return rows[i].Label < rows[j].Label
		}
		return ai > aj
	})
}

func writeSpringDiffRows(w io.Writer, title string, rows []springAnalyzeDiffRow) error {
	if _, err := fmt.Fprintf(w, "\n%s:\n", title); err != nil {
		return err
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "  %-12s winner %.2f  loser %.2f  diff %+0.2f (%s)\n",
			row.Label, row.Winner, row.Loser, row.Diff, springCountExplanation(row)); err != nil {
			return err
		}
	}
	return nil
}

// springRateDiffRows expresses each command label as a percentage of that
// side's total commands. Each turn, every alive pac emits exactly one command,
// so this normalizes for asymmetric pac counts after deaths — the rate
// reflects decision frequency, not how many pacs were left to decide.
func springRateDiffRows(winner, loser map[string]int) []springAnalyzeDiffRow {
	winnerTotal := 0
	for _, n := range winner {
		winnerTotal += n
	}
	loserTotal := 0
	for _, n := range loser {
		loserTotal += n
	}
	if winnerTotal == 0 && loserTotal == 0 {
		return nil
	}

	labels := make(map[string]struct{}, len(winner)+len(loser))
	for label := range winner {
		labels[label] = struct{}{}
	}
	for label := range loser {
		labels[label] = struct{}{}
	}

	rows := make([]springAnalyzeDiffRow, 0, len(labels))
	for label := range labels {
		var w, l float64
		if winnerTotal > 0 {
			w = float64(winner[label]) / float64(winnerTotal) * 100
		}
		if loserTotal > 0 {
			l = float64(loser[label]) / float64(loserTotal) * 100
		}
		rows = append(rows, springAnalyzeDiffRow{
			Label:  label,
			Winner: w,
			Loser:  l,
			Diff:   w - l,
		})
	}
	sortSpringDiffRows(rows)
	return rows
}

func writeSpringRateRows(w io.Writer, title string, rows []springAnalyzeDiffRow) error {
	if _, err := fmt.Fprintf(w, "\n%s:\n", title); err != nil {
		return err
	}
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "  %-12s winner %5.2f%%  loser %5.2f%%  diff %+0.2fpp\n",
			row.Label, row.Winner, row.Loser, row.Diff); err != nil {
			return err
		}
	}
	return nil
}

func springCountExplanation(row springAnalyzeDiffRow) string {
	switch {
	case row.Winner == 0 && row.Loser == 0:
		return "same rate"
	case row.Winner == 0:
		return "winner never did this; loser did"
	case row.Loser == 0:
		return "winner did this; loser never did"
	case row.Winner < row.Loser:
		return fmt.Sprintf("winner only %.0f%% as often as loser; loser %.1fx winner",
			row.Winner/row.Loser*100, row.Loser/row.Winner)
	case row.Winner > row.Loser:
		return fmt.Sprintf("winner %.1fx loser; loser only %.0f%% as often as winner",
			row.Winner/row.Loser, row.Loser/row.Winner*100)
	default:
		return "same rate"
	}
}

