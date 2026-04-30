package engine

import (
	"fmt"
	"io"
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// AnalyzeTraces implements arena.TraceAnalyzer for Winter 2026 traces. The
// report focuses on end-reason distribution, with extra attention to whether
// the blue (our) side caused fault terminations (timeouts / invalid commands).
func (f *factory) AnalyzeTraces(input arena.TraceAnalysisInput) (arena.TraceAnalysis, error) {
	report := winterAnalyzeReport{
		TraceDir:        input.TraceDir,
		Files:           len(input.Files),
		EndReasonCounts: make(map[string]int),
		BlueFaults:      make(map[string]int),
		OpponentFaults:  make(map[string]int),
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
	}

	return report, nil
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

	return r.writeBlueFaultSummary(w)
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
