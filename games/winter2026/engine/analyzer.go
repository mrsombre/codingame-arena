package engine

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AnalyzeTraces implements arena.TraceAnalyzer for Winter 2026 traces. The
// report focuses on end-reason distribution (with attention to blue-side
// faults) and aggregates per-side trace event counts (HIT_WALL, HIT_ENEMY,
// HIT_ITSELF, EAT, FALL, DEAD) to compare winner vs loser play.
func (f *factory) AnalyzeTraces(input arena.TraceAnalysisInput) (arena.TraceAnalysis, error) {
	report := winterAnalyzeReport{
		TraceDir:        input.TraceDir,
		Files:           len(input.Files),
		EndReasonCounts: make(map[string]int),
		BlueFaults:      make(map[string]int),
		OpponentFaults:  make(map[string]int),
		Winner:          newWinterAnalyzeSideAggregate(),
		Loser:           newWinterAnalyzeSideAggregate(),
		Blue:            newWinterAnalyzeSideAggregate(),
		Opponent:        newWinterAnalyzeSideAggregate(),
	}

	for _, file := range input.Files {
		trace := file.Trace
		winner := winterTraceWinner(trace)
		switch winner {
		case 0:
			report.SideWins[0]++
			report.Decided++
		case 1:
			report.SideWins[1]++
			report.Decided++
		default:
			report.Draws++
		}

		blueSide := trace.BlueSide()
		if blueSide >= 0 {
			report.BlueMatches++
			if winner == blueSide {
				report.BlueWins++
			} else if winner >= 0 {
				report.BlueLosses++
			}
		}

		if trace.EndReason != "" {
			report.EndReasonCounts[trace.EndReason]++
			if winterIsFaultEndReason(trace.EndReason) {
				report.FaultMatches++
				faultSide := winterFaultSide(trace, winner)
				switch {
				case blueSide < 0 || faultSide < 0:
					// Unknown attribution — counted in the global tally only.
					report.UnknownFaultMatches++
				case faultSide == blueSide:
					report.BlueFaults[trace.EndReason]++
					report.BlueFaultMatches++
				default:
					report.OpponentFaults[trace.EndReason]++
				}
			}
		}

		stats, ok := summarizeWinterTraceEvents(trace)
		if !ok {
			continue
		}
		if winner >= 0 {
			loser := 1 - winner
			report.EventMatches++
			report.Winner.add(stats[winner])
			report.Loser.add(stats[loser])
		}
		if blueSide >= 0 {
			report.BlueEventMatches++
			report.Blue.add(stats[blueSide])
			report.Opponent.add(stats[1-blueSide])
		}
	}

	return report, nil
}

// summarizeWinterTraceEvents derives a per-side event tally for one match by
// scanning the per-turn traces. Each trace payload starts with the subject
// bird ID; we map bird→side by reading the same turn's command outputs (each
// side commands only its own birds), so this works on any winter2026 trace
// without needing a precomputed TraceSummary. Returns ok=false when the trace
// has no per-turn traces or no commands to derive ownership.
func summarizeWinterTraceEvents(trace arena.TraceMatch) ([2]winterAnalyzeSideMatchStats, bool) {
	stats := [2]winterAnalyzeSideMatchStats{
		newWinterAnalyzeSideMatchStats(),
		newWinterAnalyzeSideMatchStats(),
	}

	birdSide := winterBirdSideMap(trace)
	if len(birdSide) == 0 {
		return stats, false
	}

	hasEvents := false
	for _, turn := range trace.Turns {
		for _, ev := range turn.Traces {
			birdID, ok := parseLeadingBirdID(ev.Payload)
			if !ok {
				continue
			}
			side, ok := birdSide[birdID]
			if !ok {
				continue
			}
			stats[side].EventCounts[ev.Label]++
			hasEvents = true
		}
	}
	return stats, hasEvents
}

// winterBirdSideMap reads the per-turn command outputs to learn which side
// owns each bird ID. A move command starts with `<birdID> <DIR>`; only the
// bird's owner can issue commands for it, so the leading int identifies
// ownership unambiguously. MARK commands (which lead with `MARK`) are
// skipped.
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

type winterAnalyzeSideAggregate struct {
	Matches     int
	EventCounts map[string]int
}

type winterAnalyzeSideMatchStats struct {
	EventCounts map[string]int
}

func newWinterAnalyzeSideAggregate() winterAnalyzeSideAggregate {
	return winterAnalyzeSideAggregate{EventCounts: make(map[string]int)}
}

func newWinterAnalyzeSideMatchStats() winterAnalyzeSideMatchStats {
	return winterAnalyzeSideMatchStats{EventCounts: make(map[string]int)}
}

func (a *winterAnalyzeSideAggregate) add(stats winterAnalyzeSideMatchStats) {
	a.Matches++
	for label, n := range stats.EventCounts {
		a.EventCounts[label] += n
	}
}

type winterAnalyzeReport struct {
	TraceDir string
	Files    int
	Decided  int
	Draws    int
	SideWins [2]int

	BlueMatches int
	BlueWins    int
	BlueLosses  int

	EndReasonCounts     map[string]int
	FaultMatches        int
	BlueFaults          map[string]int
	BlueFaultMatches    int
	OpponentFaults      map[string]int
	UnknownFaultMatches int

	EventMatches     int
	Winner           winterAnalyzeSideAggregate
	Loser            winterAnalyzeSideAggregate
	BlueEventMatches int
	Blue             winterAnalyzeSideAggregate
	Opponent         winterAnalyzeSideAggregate
}

// winterFaultSide picks which side caused a fault end_reason. Prefers the
// explicit Deactivated marker, falling back to the loser side for older
// traces that predate the marker (the deactivated side always loses).
// Returns -1 when attribution is ambiguous (both deactivated, or no winner).
func winterFaultSide(trace arena.TraceMatch, winner int) int {
	switch {
	case trace.Deactivated[0] && !trace.Deactivated[1]:
		return 0
	case trace.Deactivated[1] && !trace.Deactivated[0]:
		return 1
	case trace.Deactivated[0] && trace.Deactivated[1]:
		return -1
	}
	if winner == 0 {
		return 1
	}
	if winner == 1 {
		return 0
	}
	return -1
}

func winterTraceWinner(trace arena.TraceMatch) int {
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

func winterIsFaultEndReason(reason string) bool {
	switch reason {
	case arena.EndReasonTimeoutStart, arena.EndReasonTimeout, arena.EndReasonInvalid:
		return true
	default:
		return false
	}
}

func (r winterAnalyzeReport) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "Winter 2026 trace analysis: %d trace files from %s\n", r.Files, r.TraceDir); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Decided matches: %d  draws: %d\n", r.Decided, r.Draws); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Side wins: p0=%d p1=%d\n", r.SideWins[0], r.SideWins[1]); err != nil {
		return err
	}
	if r.BlueMatches > 0 {
		blueWinRate := percent(r.BlueWins, r.BlueMatches)
		if _, err := fmt.Fprintf(w, "Blue side: matches=%d  wins=%d (%.1f%%)  losses=%d  draws=%d\n",
			r.BlueMatches, r.BlueWins, blueWinRate, r.BlueLosses, r.BlueMatches-r.BlueWins-r.BlueLosses); err != nil {
			return err
		}
	}

	if err := r.writeEndReasons(w); err != nil {
		return err
	}

	if err := r.writeBlueFaultSummary(w); err != nil {
		return err
	}

	if err := r.writeEventStats(w, "Winner vs loser events (avg per decided match)",
		r.Winner, r.Loser, r.EventMatches, "winner", "loser"); err != nil {
		return err
	}

	return r.writeEventStats(w, "Blue vs opponent events (avg per match with summary)",
		r.Blue, r.Opponent, r.BlueEventMatches, "blue", "opp")
}

func (r winterAnalyzeReport) writeEventStats(w io.Writer, title string,
	a, b winterAnalyzeSideAggregate, matches int, aLabel, bLabel string) error {
	if _, err := fmt.Fprintf(w, "\n%s:\n", title); err != nil {
		return err
	}
	if matches == 0 {
		_, err := fmt.Fprintln(w, "  no traces with summary")
		return err
	}

	rows := winterEventDiffRows(a.EventCounts, b.EventCounts, matches)
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "  %-12s %s %.2f  %s %.2f  diff %+0.2f (%s)\n",
			row.Label, aLabel, row.A, bLabel, row.B, row.Diff,
			winterEventExplanation(row, aLabel, bLabel)); err != nil {
			return err
		}
	}
	return nil
}

type winterEventDiffRow struct {
	Label string
	A     float64
	B     float64
	Diff  float64
}

func winterEventDiffRows(a, b map[string]int, matches int) []winterEventDiffRow {
	if matches == 0 {
		return nil
	}
	labels := make(map[string]struct{}, len(a)+len(b))
	for label := range a {
		if label == TraceEat {
			continue
		}
		labels[label] = struct{}{}
	}
	for label := range b {
		if label == TraceEat {
			continue
		}
		labels[label] = struct{}{}
	}
	rows := make([]winterEventDiffRow, 0, len(labels))
	for label := range labels {
		av := float64(a[label]) / float64(matches)
		bv := float64(b[label]) / float64(matches)
		rows = append(rows, winterEventDiffRow{
			Label: label,
			A:     av,
			B:     bv,
			Diff:  av - bv,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		ai := math.Abs(rows[i].Diff)
		aj := math.Abs(rows[j].Diff)
		if ai == aj {
			return rows[i].Label < rows[j].Label
		}
		return ai > aj
	})
	return rows
}

func winterEventExplanation(row winterEventDiffRow, aLabel, bLabel string) string {
	switch {
	case row.A == 0 && row.B == 0:
		return "same rate"
	case row.A == 0:
		return fmt.Sprintf("%s never; %s did", aLabel, bLabel)
	case row.B == 0:
		return fmt.Sprintf("%s did; %s never", aLabel, bLabel)
	case row.A < row.B:
		return fmt.Sprintf("%s only %.0f%% as often as %s; %s %.1fx %s",
			aLabel, row.A/row.B*100, bLabel, bLabel, row.B/row.A, aLabel)
	case row.A > row.B:
		return fmt.Sprintf("%s %.1fx %s; %s only %.0f%% as often as %s",
			aLabel, row.A/row.B, bLabel, bLabel, row.B/row.A*100, aLabel)
	default:
		return "same rate"
	}
}

func (r winterAnalyzeReport) writeEndReasons(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "\nEnd reasons (% of trace files):"); err != nil {
		return err
	}
	if len(r.EndReasonCounts) == 0 {
		_, err := fmt.Fprintln(w, "  none recorded")
		return err
	}

	rows := make([]winterEndReasonRow, 0, len(r.EndReasonCounts))
	for reason, count := range r.EndReasonCounts {
		rows = append(rows, winterEndReasonRow{
			Reason:        reason,
			Count:         count,
			BlueCount:     r.BlueFaults[reason],
			OpponentCount: r.OpponentFaults[reason],
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].Reason < rows[j].Reason
		}
		return rows[i].Count > rows[j].Count
	})
	for _, row := range rows {
		share := percent(row.Count, r.Files)
		line := fmt.Sprintf("  %-14s %4d  %5.1f%%", row.Reason, row.Count, share)
		if winterIsFaultEndReason(row.Reason) {
			line += fmt.Sprintf("  (blue: %d, opponent: %d)", row.BlueCount, row.OpponentCount)
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func (r winterAnalyzeReport) writeBlueFaultSummary(w io.Writer) error {
	if r.BlueMatches == 0 {
		_, err := fmt.Fprintln(w, "\nBlue not identified in traces (no Blue field set); fault attribution unavailable.")
		return err
	}
	timeouts := r.BlueFaults[arena.EndReasonTimeout] + r.BlueFaults[arena.EndReasonTimeoutStart]
	invalids := r.BlueFaults[arena.EndReasonInvalid]
	total := timeouts + invalids
	if _, err := fmt.Fprintf(w, "\nBlue-side faults: %d / %d matches (%.1f%%) — timeouts=%d disqualifications=%d\n",
		total, r.BlueMatches, percent(total, r.BlueMatches), timeouts, invalids); err != nil {
		return err
	}
	if r.FaultMatches > 0 {
		opponentTotal := r.FaultMatches - r.BlueFaultMatches - r.UnknownFaultMatches
		if _, err := fmt.Fprintf(w, "Of %d fault matches: blue=%d (%.1f%%)  opponent=%d (%.1f%%)  unknown=%d\n",
			r.FaultMatches,
			r.BlueFaultMatches, percent(r.BlueFaultMatches, r.FaultMatches),
			opponentTotal, percent(opponentTotal, r.FaultMatches),
			r.UnknownFaultMatches); err != nil {
			return err
		}
	}
	return nil
}

type winterEndReasonRow struct {
	Reason        string
	Count         int
	BlueCount     int
	OpponentCount int
}

func percent(num, denom int) float64 {
	if denom == 0 {
		return 0
	}
	return float64(num) / float64(denom) * 100
}
