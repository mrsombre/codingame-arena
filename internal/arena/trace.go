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
// JSON float with at least one fractional digit (e.g., 127.0), so the file shape
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
// played on side i; Scores[i] is that side's raw pre-OnEnd score and
// FinalScores[i] is the post-OnEnd value. For self-play traces both fields
// hold engine truth. For replay-converted traces the convention only differs
// when the match is disqualified: Scores keeps the engine's count and
// FinalScores is overwritten with CodinGame's gameResult.scores so the
// trace's outcome matches the official replay; the Disqualified[i] flag
// marks which side(s) CodinGame deactivated. Ranks encodes the winner
// CodinGame-style (0 = first place, [0,0] = draw). Random side-swap is
// intentionally not recorded — the bot→side mapping here is ground truth
// for downstream trace consumers (e.g., training).
//
// Blue is the bot/agent name analyze treats as "us". It is required on every
// trace: self-play sets it to the basename of --blue (which always equals one
// of Players); convert sets it to the saved replay's blue field (which is
// the username supplied to `replay get` / `replay leaderboard`). Analyzers
// locate our side by scanning Players[i] == Blue (so seed-driven side swaps
// still resolve correctly).
type TraceMatch struct {
	TraceID  int64  `json:"trace_id,omitempty"`
	MatchID  int    `json:"match_id"`
	Type     string `json:"type,omitempty"`
	GameID   string `json:"gameId,omitempty"`
	PuzzleID int    `json:"puzzleId,omitempty"`
	Seed     int64  `json:"seed,string"`
	Blue     string `json:"blue,omitempty"`
	League   int    `json:"league,omitempty"`
	// CreatedAt is the RFC 3339 timestamp, the trace was produced. For
	// self-play traces it's stamped at match completion; for replay traces
	// it's copied from the source replay's fetched_at so analyze can sort
	// converted replays chronologically without re-reading the JSON.
	CreatedAt string `json:"created_at,omitempty"`
	// EndReason categorizes how the match terminated. Game-specific; see the
	// EndReason* constants for shared values. Empty when the referee doesn't
	// implement EndReasonProvider.
	EndReason string `json:"end_reason,omitempty"`
	// Disqualified[i] is true when side i was deactivated (timeout / bad
	// command) during the match. Used by analyzers to attribute fault end
	// reasons to a specific side; the verifier uses it to short-circuit
	// score and turn-count checks for replays where CodinGame's record is
	// known to diverge from a clean simulation.
	Disqualified [2]bool `json:"disqualified,omitzero"`
	// Scores carry the raw pre-OnEnd value reported by RawScoresProvider
	// (intrinsic in-game count, e.g., spring2021 tree segments before the sun
	// bonus). Always engine truth; disqualified sides keep their actual
	// accumulated value here.
	Scores [2]TraceScore `json:"scores"`
	// FinalScores carries the post-OnEnd value (with bonuses and tiebreakers
	// applied). For self-play traces this is engine truth. For replay-
	// converted traces the trace inherits CodinGame's gameResult.scores
	// when the match was disqualified; otherwise it equals the engine's
	// post-OnEnd value (which matches CG for clean matches).
	FinalScores [2]TraceScore `json:"final_scores"`
	Ranks       [2]int        `json:"ranks"`
	Players     [2]string     `json:"players"`
	Timing      *TraceTiming  `json:"timing,omitempty"`
	Turns       []TraceTurn   `json:"turns"`
	// MainTurns is the count of player-decision trace turns. Excludes
	// non-decision phase turns (Spring 2021 GATHERING/SUN_MOVE) and
	// post-end frames (Spring 2020 gameOverFrame). Populated going forward
	// only; older trace files load with 0 ("unknown").
	MainTurns int `json:"main_turns,omitempty"`
}

// Shared EndReason values. Games may use these or add their own.
const (
	EndReasonTimeoutStart = "TIMEOUT_START"
	EndReasonTimeout      = "TIMEOUT"
	EndReasonInvalid      = "INVALID"
	EndReasonEliminated   = "ELIMINATED"
	EndReasonScore        = "SCORE"
	EndReasonScoreEarly   = "SCORE_EARLY"
	EndReasonTurnsOut     = "TURNS_OUT"
)

// BlueSide returns the index (0 or 1) of the side identified by Blue in
// Players. Blue is an invariant on loaded traces — convert refuses to write
// a trace without it and analyze refuses to load one — so callers can treat
// the result as 0/1. When both Players entries equal Blue (e.g., self-play
// of identical binary names), the lower index wins. Returns -1 only on
// unloaded zero-value TraceMatch instances.
func (t TraceMatch) BlueSide() int {
	if t.Blue == "" {
		return -1
	}
	for i, p := range t.Players {
		if p == t.Blue {
			return i
		}
	}
	return -1
}

// RanksFromWinner returns the CodinGame-style ranks array for a 2-player match
// from a winner index. 0 = first place; tied players share the better rank, so
// a draw becomes [0,0] (standard competition ranking).
//
// In our sample of 160 downloaded replays, no draws appeared, so the [0,0]
// conventions are inferred — adjust if a real-tied replay disagrees.
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

// RanksFromCGRanks normalizes CodinGame's gameResult.ranks (any indexing
// scheme; lower = better; ties allowed) into the [2]int form TraceMatch.Ranks
// uses (0 = first place, equal entries = draw). Replay traces use this so the
// trace's ranks reflect CG's ground-truth ordering — including any post-OnEnd
// tiebreaker the engine cannot derive from raw scores alone — rather than
// disagreeing with reality whenever rawScores happened to be tied. Returns
// ok=false when the input is malformed (fewer than 2 entries).
func RanksFromCGRanks(ranks []int) ([2]int, bool) {
	if len(ranks) < 2 {
		return [2]int{}, false
	}
	switch {
	case ranks[0] < ranks[1]:
		return [2]int{0, 1}, true
	case ranks[1] < ranks[0]:
		return [2]int{1, 0}, true
	default:
		return [2]int{0, 0}, true
	}
}

// TraceWinnerFromScores derives the winner shown in a trace. Deactivation
// outranks raw scores: a deactivated side cannot win, while an active side
// beats a deactivated opponent regardless of raw counts. When neither side is
// deactivated (or both are), the higher score wins; equal scores → -1 (draw).
//
// Why deactivation takes precedence: a TIMEOUT_START/TIMEOUT/INVALID_INPUT
// player never finishes the game, but their pre-existing pieces (e.g., winter
// 2026 birds) keep contributing to the raw alive-segment count. Without this
// override the trace would call a 12-vs-12 raw tie a draw even though CG
// scores it as -1 vs. 12.
func TraceWinnerFromScores(scores [2]int, deactivated [2]bool) int {
	switch {
	case deactivated[0] && !deactivated[1]:
		return 1
	case deactivated[1] && !deactivated[0]:
		return 0
	case scores[0] > scores[1]:
		return 0
	case scores[1] > scores[0]:
		return 1
	default:
		return -1
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

// TraceTurn captures one turn of the game state for replay/debug.
//
// GameInput is the stdin lines the engine fed the blue side this turn (the
// user's bot — see TraceMatch.Blue). For symmetric-input games it equals
// what either side received; for fog-of-war games it is blue's perspective
// only. Absent on turns where blue did not execute (deactivated, skipped,
// or game-over frame).
//
// Output[i] is the raw stdout the side-i bot emitted this turn (empty when
// the side was deactivated or skipped). Indexed [left, right] in match-side
// space; the bot→side mapping is in TraceMatch.Players.
//
// State is an opaque per-turn payload owned by the game (mirrors the
// TurnTrace.Data pattern at the per-turn level). Arena never inspects it;
// games marshal a typed struct describing whatever board/scoring/phase
// info downstream consumers need.
type TraceTurn struct {
	Turn      int              `json:"turn"`
	GameInput []string         `json:"game_input,omitempty"`
	Output    [2]string        `json:"output,omitzero"`
	// IsOutputTurn[i] records whether side i was prompted for output this
	// turn (i.e., the runner asked the bot for a command). Independent of
	// game-specific phase rules: false on engine-only frames (Spring 2021
	// GATHERING/SUN_MOVE, post-end frames) and on turns where the side was
	// already deactivated or the engine flagged it skipped (e.g., spring2021
	// IsWaiting). Lets analyzers and the runner identify the first turn a
	// bot was actually asked to act, regardless of the game's frame model.
	IsOutputTurn [2]bool          `json:"is_output_turn,omitzero"`
	Timing       *TraceTurnTiming `json:"timing,omitempty"`
	// Score carries the per-player raw score going into this turn, sampled
	// from RawScoresProvider before PerformGameUpdate. Zero values when the
	// referee doesn't implement RawScoresProvider.
	Score [2]int `json:"score"`
	// Traces partitions per-turn structured events by player: Traces[0] is
	// everything player 0 owned this turn, Traces[1] is everything player 1
	// owned. Cross-owner events (e.g. spring2020 COLLIDE_ENEMY, winter2026
	// HIT_ENEMY) are mirrored into both slots.
	Traces [2][]TurnTrace  `json:"traces,omitzero"`
	State  json.RawMessage `json:"state,omitempty"`
}

// TraceTurnTiming carries per-side response time for one turn in milliseconds.
// Zero entries mean the side did not execute (deactivated or skipped).
type TraceTurnTiming struct {
	Response [2]float64 `json:"response"`
}

// TraceSink receives a completed TraceMatch from the match runner. Implemented
// by *TraceWriter (file-backed) and by ad-hoc capture sinks (e.g., the HTTP
// single-match handler that returns the trace inline).
type TraceSink interface {
	WriteMatch(TraceMatch) error
}

// TraceWriter writes per-match JSON trace files to a directory. All matches in
// a batch share traceID (typically the batch start timestamp); each file is
// keyed by traceID + per-match MatchID so multiple batches can coexist.
type TraceWriter struct {
	mu        sync.Mutex
	dir       string
	traceID   int64
	fixedName string
}

const (
	TraceTypeTrace  = "trace"
	TraceTypeReplay = "replay"
)

var replayTraceFilePattern = regexp.MustCompile(`^replay-\d+\.json$`)

// NewTraceWriter creates a TraceWriter that writes to the given directory and
// stamps every match with traceID. Returns nil if the dir is empty.
func NewTraceWriter(dir string, traceID int64) *TraceWriter {
	if dir == "" {
		return nil
	}
	return &TraceWriter{dir: dir, traceID: traceID}
}

// NewFixedTraceWriter creates a TraceWriter that overwrites a single JSON file.
func NewFixedTraceWriter(path string, traceID int64) *TraceWriter {
	if path == "" {
		return nil
	}
	return &TraceWriter{dir: filepath.Dir(path), traceID: traceID, fixedName: filepath.Base(path)}
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
	name := TraceFileName(match.Type, w.traceID, match.MatchID)
	if w.fixedName != "" {
		name = w.fixedName
	}
	path := filepath.Join(w.dir, name)
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
