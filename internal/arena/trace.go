package arena

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// TraceScore is the per-side score as written into the trace. It marshals as a
// JSON float with at least one fractional digit (e.g. 127.0) so the file shape
// matches CodinGame's replay encoding.
type TraceScore float64

// MarshalJSON renders the value as a JSON number that always carries a decimal
// point ("127" -> "127.0", "22.7713" stays unchanged).
func (s TraceScore) MarshalJSON() ([]byte, error) {
	str := strconv.FormatFloat(float64(s), 'f', -1, 64)
	if !strings.ContainsRune(str, '.') {
		str += ".0"
	}
	return []byte(str), nil
}

// TraceMatch is the per-match trace file structure.
//
// Everything is reported from the in-match side perspective (index 0 = left
// side of the map, index 1 = right). Players[i] is the bot basename that
// played on side i; Scores[i] is that side's final score; Ranks encodes the
// winner CodinGame-style (0 = first place, [0,0] = draw). Random side-swap is
// intentionally not recorded — the bot→side mapping here is ground truth for
// downstream trace consumers (e.g. training).
type TraceMatch struct {
	TraceID  int64         `json:"trace_id,omitempty"`
	MatchID  int           `json:"match_id"`
	Type     string        `json:"type,omitempty"`
	File     string        `json:"file,omitempty"`
	GameID   string        `json:"gameId,omitempty"`
	PuzzleID int           `json:"puzzleId,omitempty"`
	Seed     int64         `json:"seed,string"`
	Scores   [2]TraceScore `json:"scores"`
	Ranks    [2]int        `json:"ranks"`
	Players  [2]string     `json:"players"`
	Timing   *TraceTiming  `json:"timing,omitempty"`
	Turns    []TraceTurn   `json:"turns"`
}

// RanksFromWinner returns the CodinGame-style ranks array for a 2-player match
// from a winner index. 0 = first place; tied players share the better rank, so
// a draw becomes [0,0] (standard competition ranking).
//
// In our sample of 160 downloaded replays no draws appeared, so the [0,0]
// convention is inferred — adjust if a real tied replay disagrees.
func RanksFromWinner(winner int) [2]int {
	switch winner {
	case 0:
		return [2]int{0, 1}
	case 1:
		return [2]int{1, 0}
	default:
		return [2]int{0, 0}
	}
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

const (
	TraceTypeTrace  = "trace"
	TraceTypeReplay = "replay"
)

var replayTraceFilePattern = regexp.MustCompile(`^replay-\d+\.json$`)

// NewTraceWriter creates a TraceWriter that writes to the given directory and
// stamps every match with traceID. Returns nil if dir is empty.
func NewTraceWriter(dir string, traceID int64) *TraceWriter {
	if dir == "" {
		return nil
	}
	return &TraceWriter{dir: dir, traceID: traceID}
}

// WriteMatch writes a single match trace as a JSON file:
// <dir>/trace-<trace_id>-<match_id>.json or <dir>/replay-<trace_id>.json
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
	if match.Type == "" {
		match.Type = TraceTypeTrace
	}
	match.File = TraceFileName(match.Type, w.traceID, match.MatchID)
	path := filepath.Join(w.dir, match.File)
	data, err := json.MarshalIndent(match, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trace: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write trace file: %w", err)
	}
	return nil
}

// TraceFileName returns the on-disk JSON filename for a trace artifact.
func TraceFileName(kind string, traceID int64, matchID int) string {
	if kind == TraceTypeReplay {
		return fmt.Sprintf("replay-%d.json", traceID)
	}
	return fmt.Sprintf("trace-%d-%d.json", traceID, matchID)
}

// TraceTypeFromFileName infers whether the stored artifact is a normal trace
// or a replay-derived trace from its JSON filename.
func TraceTypeFromFileName(name string) string {
	if replayTraceFilePattern.MatchString(name) {
		return TraceTypeReplay
	}
	return TraceTypeTrace
}
