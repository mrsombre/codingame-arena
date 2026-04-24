package arena

import "time"

// Metric is a labeled numeric value used for per-match and aggregate statistics.
type Metric struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// TurnEvent is a structured game event produced by the engine per turn.
type TurnEvent struct {
	Label   string `json:"label"`
	Payload string `json:"payload"`
}

// LossReason describes why a player lost a match.
type LossReason string

const (
	LossReasonNone       LossReason = ""
	LossReasonScore      LossReason = "score"
	LossReasonTimeout    LossReason = "timeout"
	LossReasonBadCommand LossReason = "bad_command"
)

// MatchResult holds the outcome of a single match simulation.
//
// Scores are the referee's reported final values (may include tie-break
// adjustments and therefore be negative). RawScores is populated when the
// Referee implements RawScoresProvider and reflects the intrinsic game state
// (e.g. sum of alive bird segments). Both arrays follow the user-selected bot
// perspective — Scores[0]/RawScores[0] belong to the bot the user chose as P0
// even when sides were randomly swapped during the match.
type MatchResult struct {
	ID                int
	Seed              int64
	Turns             int
	Scores            [2]int
	RawScores         [2]int
	HaveRawScores     bool
	Winner            int // 0, 1, or -1 for draw
	LossReasons       [2]LossReason
	BadCommands       []BadCommandInfo
	TimeToFirstOutput [2]time.Duration
	AverageOutputTime [2]time.Duration
	Swapped           bool
	Metrics           []Metric
}

// SimulationID returns the match ID for sorting.
func (r MatchResult) SimulationID() int { return r.ID }

// BadCommandInfo records a player sending an invalid command.
type BadCommandInfo struct {
	Seed    int64  `json:"seed"`
	Player  int    `json:"player"`
	Turn    int    `json:"turn"`
	Command string `json:"command"`
	Reason  string `json:"reason"`
}

// BatchOptions controls parallel batch execution.
type BatchOptions struct {
	Simulations   int
	Parallel      int
	Seed          int64
	SeedIncrement *int64
	OutputMatches bool
}
