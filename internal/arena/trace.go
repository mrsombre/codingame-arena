package arena

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TraceMatch is the per-match trace file structure.
//
// Everything is reported from the in-match side perspective (index 0 = left
// side of the map, index 1 = right). Players[i] is the bot basename that
// played on side i; Scores[i] is that side's final score; Winner is 0, 1, or
// -1 for draw. Random side-swap is intentionally not recorded — the bot→side
// mapping here is ground truth for downstream trace consumers (e.g. training).
type TraceMatch struct {
	TraceID int64        `json:"trace_id,omitempty"`
	MatchID int          `json:"match_id"`
	Seed    int64        `json:"seed,string"`
	Winner  int          `json:"winner"`
	Scores  [2]int       `json:"scores"`
	Players [2]string    `json:"players"`
	Timing  *TraceTiming `json:"timing,omitempty"`
	Turns   []TraceTurn  `json:"turns"`
}

// TraceTiming aggregates per-side response timings in milliseconds.
//
// FirstResponse is the turn-0 latency (typically dominated by bot startup and
// not representative of steady-state). ResponseAverage and ResponseMedian
// summarize the remaining turns and intentionally exclude turn 0.
type TraceTiming struct {
	FirstResponse   [2]float64 `json:"first_response"`
	ResponseAverage [2]float64 `json:"response_average"`
	ResponseMedian  [2]float64 `json:"response_median"`
}

// TraceTurn captures one turn of game state for replay/debug.
type TraceTurn struct {
	Turn      int              `json:"turn"`
	GameInput traceTurnInput   `json:"game_input"`
	P0Output  string           `json:"p0_output,omitempty"`
	P1Output  string           `json:"p1_output,omitempty"`
	Timing    *TraceTurnTiming `json:"timing,omitempty"`
	Events    []TurnEvent      `json:"events,omitempty"`
	GameState json.RawMessage  `json:"game_state,omitempty"`
}

// TraceTurnTiming carries per-side response time for one turn in milliseconds.
// Zero entries mean the side did not execute (deactivated or skipped).
type TraceTurnTiming struct {
	Response [2]float64 `json:"response"`
}

type traceTurnInput struct {
	P0 []string `json:"p0,omitempty"`
	P1 []string `json:"p1,omitempty"`
}

// TraceWriter writes per-match JSON trace files to a directory. All matches in
// a batch share traceID (typically the batch start timestamp); each file is
// keyed by traceID + per-match MatchID so multiple batches can coexist.
type TraceWriter struct {
	mu      sync.Mutex
	dir     string
	traceID int64
}

// NewTraceWriter creates a TraceWriter that writes to the given directory and
// stamps every match with traceID. Returns nil if dir is empty.
func NewTraceWriter(dir string, traceID int64) *TraceWriter {
	if dir == "" {
		return nil
	}
	return &TraceWriter{dir: dir, traceID: traceID}
}

// WriteMatch writes a single match trace as a JSON file:
// <dir>/trace-<trace_id>-<match_id>.json
func (w *TraceWriter) WriteMatch(match TraceMatch) error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("create trace directory: %w", err)
	}

	match.TraceID = w.traceID
	path := filepath.Join(w.dir, fmt.Sprintf("trace-%d-%d.json", w.traceID, match.MatchID))
	data, err := json.MarshalIndent(match, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trace: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write trace file: %w", err)
	}
	return nil
}
