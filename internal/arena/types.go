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
type MatchResult struct {
	ID                int
	Seed              int64
	Turns             int
	Scores            [2]int
	Winner            int // 0, 1, or -1 for draw
	LossReasons       [2]LossReason
	BadCommands       []BadCommandInfo
	TimeToFirstAnswer [2]time.Duration
	TimeToTurnP99     [2]time.Duration
	TimeToTurnMax     [2]time.Duration
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
