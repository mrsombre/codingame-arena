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
// side of the map, index 1 = right). Bots[i] is the bot basename that played
// on side i; Scores[i] is that side's final score; Winner is 0, 1, or -1 for
// draw. Random side-swap is intentionally not recorded — the bot→side mapping
// here is ground truth for downstream trace consumers (e.g. training).
type TraceMatch struct {
	MatchID int         `json:"match_id"`
	Seed    int64       `json:"seed,string"`
	Winner  int         `json:"winner"`
	Scores  [2]int      `json:"scores"`
	Bots    [2]string   `json:"bots"`
	TTFO    [2]float64  `json:"ttfo_ms"`
	AOT     [2]float64  `json:"aot_ms"`
	Turns   []TraceTurn `json:"turns"`
}

// TraceTurn captures one turn of game state for replay/debug.
type TraceTurn struct {
	Turn      int             `json:"turn"`
	GameInput traceTurnInput  `json:"game_input"`
	P0Output  string          `json:"p0_output,omitempty"`
	P1Output  string          `json:"p1_output,omitempty"`
	Events    []TurnEvent     `json:"events,omitempty"`
	GameState json.RawMessage `json:"game_state,omitempty"`
}

type traceTurnInput struct {
	P0 []string `json:"p0,omitempty"`
	P1 []string `json:"p1,omitempty"`
}

// TraceWriter writes per-match JSON trace files to a directory.
type TraceWriter struct {
	mu  sync.Mutex
	dir string
}

// NewTraceWriter creates a TraceWriter that writes to the given directory.
// Returns nil if dir is empty.
func NewTraceWriter(dir string) *TraceWriter {
	if dir == "" {
		return nil
	}
	return &TraceWriter{dir: dir}
}

// WriteMatch writes a single match trace as a JSON file: <dir>/<match_id>.json
func (w *TraceWriter) WriteMatch(match TraceMatch) error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("create trace directory: %w", err)
	}

	path := filepath.Join(w.dir, fmt.Sprintf("%d.json", match.MatchID))
	data, err := json.MarshalIndent(match, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trace: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write trace file: %w", err)
	}
	return nil
}
