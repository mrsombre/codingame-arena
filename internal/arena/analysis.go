package arena

import "io"

// TraceFile is one decoded arena trace file selected for analysis.
type TraceFile struct {
	Name  string
	Trace TraceMatch
}

// TraceAnalysisInput is the generic input passed from the CLI to the shared
// trace analyzer.
type TraceAnalysisInput struct {
	TraceDir   string
	Files      []TraceFile
	PuzzleName string
}

// TraceAnalysis is an arena-rendered analysis report.
type TraceAnalysis interface {
	Write(io.Writer) error
}

// TraceMetricKind tells arena how to normalize a game-owned trace metric.
type TraceMetricKind string

const (
	// TraceMetricPerMatchCount is averaged as raw occurrences per match.
	TraceMetricPerMatchCount TraceMetricKind = "per_match_count"
	// TraceMetricPerTurnRate is averaged as a per-match percentage of turns.
	TraceMetricPerTurnRate TraceMetricKind = "per_turn_rate"
)

// TraceMetricSpec declares one game-owned metric that arena can aggregate.
// Arena owns ordering, comparison, and presentation; games own metric meaning.
//
// Description is a one-line human explanation rendered in a `Metrics:` legend
// before the comparison sections, so a reader unfamiliar with the game can
// interpret labels like DEAD or HIT_ENEMY without reading source. Optional —
// specs without a description are omitted from the legend, and a game whose
// specs all lack one renders no legend block.
//
// HigherIsBetter flips the polarity for the WORST section: hazard metrics
// (DEAD_*, HIT_*, FALL counts) leave it false so the WORST list reports the
// match with the highest blue value, while score-style metrics (EAT_T20)
// set it true and are excluded from WORST entirely (their "worst" would be
// a min-eat match, which deserves a different framing than a hazard peak).
type TraceMetricSpec struct {
	Key            string
	Label          string
	Kind           TraceMetricKind
	ShowZero       bool
	Description    string
	HigherIsBetter bool
}

// TraceMetricStats holds per-side metric values for one match.
// Values are indexed by in-match side: [0] = left/player 0, [1] = right/player 1.
type TraceMetricStats map[string][2]int

// TraceMetricAnalyzer is an optional game factory capability for trace metric
// extraction. Implementations may inspect TraceMatch, including opaque
// turns[].traces, but arena never interprets game trace labels or payloads.
type TraceMetricAnalyzer interface {
	TraceMetricSpecs() []TraceMetricSpec
	AnalyzeTraceMetrics(TraceMatch) (TraceMetricStats, error)
}
