package arena

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"time"
)

// MetricSummary is an aggregated metric across multiple matches.
type MetricSummary struct {
	Label string   `json:"label"`
	Avg   float64  `json:"avg"`
	Total *float64 `json:"total,omitempty"`
}

// MatchSummary aggregates metrics across a batch of matches.
type MatchSummary struct {
	Simulations int             `json:"simulations"`
	Metrics     []MetricSummary `json:"metrics"`
}

func (s *MatchSummary) Get(label string) *MetricSummary {
	for i := range s.Metrics {
		if s.Metrics[i].Label == label {
			return &s.Metrics[i]
		}
	}
	return nil
}

// SummarizeMatches aggregates metrics from all match results.
func SummarizeMatches(results []MatchResult) MatchSummary {
	if len(results) == 0 {
		return MatchSummary{}
	}

	firstMetrics := results[0].Metrics
	totals := make([]float64, len(firstMetrics))
	for _, result := range results {
		for i, metric := range result.Metrics {
			totals[i] += metric.Value
		}
	}

	timingPrefix := "time_to_"
	metrics := make([]MetricSummary, len(firstMetrics))
	for i, metric := range firstMetrics {
		ms := MetricSummary{
			Label: metric.Label,
			Avg:   round2(totals[i] / float64(len(results))),
		}
		if len(metric.Label) < len(timingPrefix) || metric.Label[:len(timingPrefix)] != timingPrefix {
			t := round2(totals[i])
			ms.Total = &t
		}
		metrics[i] = ms
	}

	return MatchSummary{
		Simulations: len(results),
		Metrics:     metrics,
	}
}

// FindWorstLosses returns indices of the worst P0 losses sorted by margin.
func FindWorstLosses(results []MatchResult, limit int) []int {
	type lossEntry struct {
		idx    int
		margin float64
	}

	var losses []lossEntry
	for idx, result := range results {
		var wonByP1 bool
		var scoreP0, scoreP1 float64
		for _, metric := range result.Metrics {
			switch metric.Label {
			case "loses_p0":
				wonByP1 = metric.Value == 1
			case "score_p0":
				scoreP0 = metric.Value
			case "score_p1":
				scoreP1 = metric.Value
			}
		}
		if wonByP1 {
			losses = append(losses, lossEntry{idx: idx, margin: scoreP0 - scoreP1})
		}
	}

	sort.Slice(losses, func(i, j int) bool {
		return losses[i].margin < losses[j].margin
	})

	if limit > len(losses) {
		limit = len(losses)
	}
	indices := make([]int, limit)
	for i := 0; i < limit; i++ {
		indices[i] = losses[i].idx
	}
	return indices
}

// RenderMatch serializes a MatchResult to JSON for verbose output.
// Scores and Winner prefer raw alive-segment sums when the engine provides
// them, so the viewer's single-match status never shows the negative values
// that referees like winter-2026 emit after tie-break adjustments.
func (r MatchResult) RenderMatch() string {
	scores := r.Scores
	winner := r.Winner
	if r.HaveRawScores {
		scores = r.RawScores
		switch {
		case scores[0] > scores[1]:
			winner = 0
		case scores[1] > scores[0]:
			winner = 1
		default:
			winner = -1
		}
	}
	payload := struct {
		ID           int        `json:"id"`
		Seed         int64      `json:"seed,string"`
		Turns        int        `json:"turns"`
		Winner       int        `json:"winner"`
		LossReasonP0 LossReason `json:"loss_reason_p0"`
		LossReasonP1 LossReason `json:"loss_reason_p1"`
		ScoreP0      int        `json:"score_p0"`
		ScoreP1      int        `json:"score_p1"`
		Metrics      []Metric   `json:"metrics,omitempty"`
		Swapped      bool       `json:"swapped,omitempty"`
	}{
		ID:           r.ID,
		Seed:         r.Seed,
		Turns:        r.Turns,
		Winner:       winner,
		LossReasonP0: r.LossReasons[0],
		LossReasonP1: r.LossReasons[1],
		ScoreP0:      scores[0],
		ScoreP1:      scores[1],
		Metrics:      r.Metrics,
		Swapped:      r.Swapped,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func WriteShortSummary(w io.Writer, s MatchSummary) error {
	get := func(label string) float64 {
		if m := s.Get(label); m != nil {
			return m.Avg
		}
		return 0
	}
	_, err := fmt.Fprintf(w, "W=%.0f%% L=%.0f%% D=%.0f%% score=%.1fv%.1f turns=%.1f\n",
		get("wins_p0")*100,
		get("loses_p0")*100,
		get("draws")*100,
		get("score_p0"),
		get("score_p1"),
		get("turns"),
	)
	return err
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func durationMillis(value time.Duration) float64 {
	return round2(float64(value) / float64(time.Millisecond))
}
