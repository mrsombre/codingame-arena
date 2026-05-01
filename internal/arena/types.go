package arena

import (
	"encoding/json"
	"time"
)

// Metric is a labeled numeric value used for per-match and aggregate statistics.
type Metric struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// TurnTrace is opaque game-owned data produced by the engine per turn.
// Arena stores the type discriminator and an opaque JSON-encoded meta object
// for game viewers and metric analyzers, but never interprets their meaning.
type TurnTrace struct {
	Type string          `json:"type"`
	Meta json.RawMessage `json:"meta,omitempty"`
}

// MakeTurnTrace marshals a typed game-owned meta into a TurnTrace. Games use
// this to emit traces with structured payloads. The marshal must succeed:
// trace metas are typed structs, so a panic indicates a programming error.
func MakeTurnTrace[T any](typ string, meta T) TurnTrace {
	raw, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}
	return TurnTrace{Type: typ, Meta: raw}
}

// DecodeMeta unmarshals a TurnTrace's opaque meta into a typed game struct.
// Returns the zero value plus an error on malformed input.
func DecodeMeta[T any](t TurnTrace) (T, error) {
	var v T
	if len(t.Meta) == 0 {
		return v, nil
	}
	err := json.Unmarshal(t.Meta, &v)
	return v, err
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
// before end-of-game adjustments. Both arrays follow the user-selected bot
// perspective: index 0 is the blue/our bot selected by --p0, index 1 is the
// red/their bot selected by --p1, even when sides were randomly swapped during
// the match.
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
	SeedIncrement int64
	OutputMatches bool
}
