package arena

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Replay source tags written into the saved replay's top-level "source" field.
const (
	ReplaySourceGet         = "get"
	ReplaySourceLeaderboard = "leaderboard"
)

// CodinGameReplay is the shape of a CodinGame match replay JSON as returned by
// /services/gameResult/findInformationById, and as stored on disk after the
// `arena replay` downloader normalizes it (see PrepareReplay).
//
// Top-level fields are shared across every CodinGame game. The F parameter is
// the per-game frame shape: games that don't need custom per-turn fields can
// use the default CodinGameReplayFrame; games that do can declare their own
// struct and pass it as F.
//
// Blue, League, Source, FetchedAt, and Leaderboard are arena-only annotations
// (not part of the upstream CodinGame body): Blue is the username of the
// player we are playing for, League is the league level the replay belongs
// to, Source records which `arena replay` subcommand produced the file
// ("get" or "leaderboard"), FetchedAt is the RFC 3339 download timestamp,
// and Leaderboard carries the player's rank/division at fetch time
// (populated only when downloaded via `replay leaderboard`).
type CodinGameReplay[F any] struct {
	PuzzleID      int                      `json:"puzzleId"`
	QuestionTitle string                   `json:"questionTitle"`
	Blue          string                   `json:"blue,omitempty"`
	League        int                      `json:"league,omitempty"`
	Source        string                   `json:"source,omitempty"`
	FetchedAt     string                   `json:"fetched_at,omitempty"`
	// Seed is the referee seed promoted to the top level by PrepareReplay so
	// callers don't need to re-parse refereeInput. Encoded as a JSON string
	// because seeds routinely exceed JS Number precision.
	Seed        int64                    `json:"seed,string,omitempty"`
	Leaderboard *ReplayLeaderboardInfo   `json:"leaderboard,omitempty"`
	GameResult  CodinGameReplayResult[F] `json:"gameResult"`
}

// ReplayLeaderboardInfo captures the player's leaderboard standing at the
// moment the replay was downloaded. Division mirrors the API's
// `league.divisionIndex`. Score is the elo-like value the API returns; 0 if
// unknown.
type ReplayLeaderboardInfo struct {
	Rank     int     `json:"rank"`
	Division int     `json:"division"`
	Score    float64 `json:"score,omitempty"`
}

// CodinGameReplayResult is the gameResult sub-object. Frames is the only
// game-parameterized field. RefereeInput is preserved on parsing so older
// saved replays (written before the top-level Seed field existed) can still
// be re-read; PrepareReplay strips it from new files once the seed is
// promoted to the top level.
type CodinGameReplayResult[F any] struct {
	GameID       int64                  `json:"gameId"`
	RefereeInput string                 `json:"refereeInput,omitempty"`
	Scores       []float64              `json:"scores"`
	Ranks        []int                  `json:"ranks"`
	Agents       []CodinGameReplayAgent `json:"agents"`
	Frames       []F                    `json:"frames"`
}

// CodinGameReplayAgent is one entry in gameResult.agents.
type CodinGameReplayAgent struct {
	Index      int                 `json:"index"`
	AgentID    int64               `json:"agentId"`
	Score      float64             `json:"score"`
	Valid      bool                `json:"valid"`
	CodinGamer CodinGameReplayUser `json:"codingamer"`
}

// CodinGameReplayUser is gameResult.agents[i].codingamer.
type CodinGameReplayUser struct {
	UserID int64  `json:"userId"`
	Pseudo string `json:"pseudo"`
	Avatar int64  `json:"avatar"`
}

// CodinGameReplayFrame is the default frame shape: the fields every CodinGame
// referee emits. Fits every game checked so far (Spring 2020, Winter 2026).
// Games that need extra per-turn fields can define their own frame struct and
// instantiate CodinGameReplay with it.
type CodinGameReplayFrame struct {
	AgentID int    `json:"agentId"`
	Stdout  string `json:"stdout"`
	Summary string `json:"summary"`
}

// ResolveReplaySeed returns the replay's seed, preferring the top-level Seed
// field (set by PrepareReplay since the seed-promotion change) and falling
// back to parsing the legacy gameResult.refereeInput for older saved files.
func ResolveReplaySeed[F any](replay CodinGameReplay[F]) (int64, bool) {
	if replay.Seed != 0 {
		return replay.Seed, true
	}
	return ParseReplaySeed(replay.GameResult.RefereeInput)
}

// ParseReplaySeed extracts the referee seed from the replay's refereeInput.
func ParseReplaySeed(refereeInput string) (int64, bool) {
	for _, tok := range strings.Fields(refereeInput) {
		if strings.HasPrefix(tok, "seed=") {
			n, err := strconv.ParseInt(strings.TrimPrefix(tok, "seed="), 10, 64)
			if err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

var replayLeaguePattern = regexp.MustCompile(`(?i)level\s*(\d)`)

// ParseReplayLeague extracts the CodinGame league level from the replay title.
func ParseReplayLeague(questionTitle string) int {
	m := replayLeaguePattern.FindStringSubmatch(questionTitle)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

// ReplayMovesFromFrames converts replay frames into per-turn bot outputs.
func ReplayMovesFromFrames(replay CodinGameReplay[CodinGameReplayFrame]) ReplayMoves {
	var moves ReplayMoves
	for _, f := range replay.GameResult.Frames {
		if strings.TrimSpace(f.Stdout) == "" {
			continue
		}
		switch f.AgentID {
		case 0:
			moves.Left = append(moves.Left, f.Stdout)
		case 1:
			moves.Right = append(moves.Right, f.Stdout)
		}
	}
	return moves
}

// ReplayTurnCount returns the number of recorded decision turns in the replay.
func ReplayTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	moves := ReplayMovesFromFrames(replay)
	if len(moves.Left) > len(moves.Right) {
		return len(moves.Left)
	}
	return len(moves.Right)
}

// FlatTurnModel is the default TurnModel: 1 engine iteration = 1 player-
// decision turn = 1 trace turn. The engine ends the match on the same turn
// game-over is detected (no separate post-end frame). A trailing empty
// stdout that pairs with the other side's stdout is the deactivation/
// timeout close of the in-progress turn — already counted by the regular
// flush — so it must not be double-counted. Used by Winter 2026.
type FlatTurnModel struct{}

// PostEndTurnModel handles engines that emit a separate post-end trace
// turn for their game-over frame (Java spring2020's gameOverFrame branch).
// Mid-replay empty stdouts are sub-turn markers (e.g. speed) that the
// engine folds into the preceding main turn. The replay's trailing empty
// stdout always represents the gameOverFrame and adds +1 unconditionally,
// even when it appears to pair with another agent's stdout (in that case
// the pair is a polling-timeout close of the final main turn, with the
// gameOverFrame following separately). Used by Spring 2020.
type PostEndTurnModel struct{}

// PhaseTurnModel handles engines that emit standalone trace turns for
// non-decision phases (Spring 2021's GATHERING and SUN_MOVE). Every
// empty-stdout frame in the replay is its own phase trace turn AND
// flushes any pending single-stdout pair. The trailing engine frame is
// the final phase, already counted, so no post-end add. Used by Spring 2021.
type PhaseTurnModel struct{}

// MainTurnCount counts player-decision turns in a replay: stdout-pair
// flushes only, with empty-stdout phase frames and trailing engine
// markers excluded. The same logic applies to all three current games,
// so each TurnModel's MainTurnCount delegates here.
func mainTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	turns := 0
	seen := map[int]bool{}
	flush := func() {
		if len(seen) == 0 {
			return
		}
		turns++
		clear(seen)
	}
	for _, frame := range replay.GameResult.Frames {
		if frame.AgentID < 0 {
			continue
		}
		if strings.TrimSpace(frame.Stdout) == "" {
			flush()
			continue
		}
		if seen[frame.AgentID] {
			flush()
		}
		seen[frame.AgentID] = true
		if len(seen) == 2 {
			flush()
		}
	}
	flush()
	return turns
}

func (FlatTurnModel) ExpectedTraceTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	turns, trailingEmptyClosedTurn := pairFlushCount(replay)
	if hasTrailingEngineFrame(replay) && !trailingEmptyClosedTurn {
		turns++
	}
	return turns
}

func (FlatTurnModel) MainTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	return mainTurnCount(replay)
}

func (PostEndTurnModel) ExpectedTraceTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	turns, _ := pairFlushCount(replay)
	if hasTrailingEngineFrame(replay) {
		turns++
	}
	return turns
}

func (PostEndTurnModel) MainTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	return mainTurnCount(replay)
}

func (PhaseTurnModel) ExpectedTraceTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	turns := 0
	seen := map[int]bool{}
	flush := func() {
		if len(seen) == 0 {
			return
		}
		turns++
		clear(seen)
	}
	for _, frame := range replay.GameResult.Frames {
		if frame.AgentID < 0 {
			continue
		}
		if strings.TrimSpace(frame.Stdout) == "" {
			flush()
			turns++
			continue
		}
		if seen[frame.AgentID] {
			flush()
		}
		seen[frame.AgentID] = true
		if len(seen) == 2 {
			flush()
		}
	}
	flush()
	return turns
}

func (PhaseTurnModel) MainTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	return mainTurnCount(replay)
}

// pairFlushCount runs the shared flat counting loop used by FlatTurnModel
// and PostEndTurnModel. It returns the pair-flush turn count and whether
// the very last replay frame was an empty stdout that flushed a pending
// pair (matters for FlatTurnModel: a deactivation pair close must not be
// double-counted as a post-end frame).
func pairFlushCount(replay CodinGameReplay[CodinGameReplayFrame]) (turns int, trailingEmptyClosedTurn bool) {
	seen := map[int]bool{}
	flush := func() bool {
		if len(seen) == 0 {
			return false
		}
		turns++
		clear(seen)
		return true
	}
	frames := replay.GameResult.Frames
	for i, frame := range frames {
		if frame.AgentID < 0 {
			continue
		}
		if strings.TrimSpace(frame.Stdout) == "" {
			flushed := flush()
			if i == len(frames)-1 {
				trailingEmptyClosedTurn = flushed
			}
			continue
		}
		if seen[frame.AgentID] {
			flush()
		}
		seen[frame.AgentID] = true
		if len(seen) == 2 {
			flush()
		}
	}
	flush()
	return turns, trailingEmptyClosedTurn
}

// ReplayOutcome captures the L0 verification surface of a CodinGame replay:
// what's visible from gameResult alone, without re-simulating the engine.
// Winner is 0, 1, or -1 for draw; Scores carries -1 for any side CG marked
// disqualified (its gameResult.scores[i] is -1); Deactivated mirrors the
// score check (deactivated[i] = scores[i] == -1).
type ReplayOutcome struct {
	Winner      int
	Scores      [2]int
	Deactivated [2]bool
}

// ExtractReplayOutcome derives the replay's outcome from gameResult.scores
// (with -1 indicating disqualification) and gameResult.ranks (winner). No
// summary parsing — every field is read directly from CG's structured
// match result. ok=false when scores or ranks are malformed.
func ExtractReplayOutcome(replay CodinGameReplay[CodinGameReplayFrame]) (ReplayOutcome, bool) {
	if len(replay.GameResult.Scores) < 2 {
		return ReplayOutcome{}, false
	}
	out := ReplayOutcome{}
	for i := 0; i < 2; i++ {
		out.Scores[i] = int(replay.GameResult.Scores[i])
		out.Deactivated[i] = out.Scores[i] == -1
	}
	ranks, ok := RanksFromCGRanks(replay.GameResult.Ranks)
	if !ok {
		return ReplayOutcome{}, false
	}
	switch {
	case ranks[0] == ranks[1]:
		out.Winner = -1
	case ranks[0] < ranks[1]:
		out.Winner = 0
	default:
		out.Winner = 1
	}
	return out, true
}

// hasTrailingEngineFrame reports whether the replay ends with the game-over
// engine marker frame CodinGame appends after every match. The marker has no
// stdout; its agentId is either -1 (engine-attributed, e.g. when the first
// player to be polled has been deactivated) or the agentId of the player
// whose slot would have been polled next.
func hasTrailingEngineFrame(replay CodinGameReplay[CodinGameReplayFrame]) bool {
	frames := replay.GameResult.Frames
	if len(frames) == 0 {
		return false
	}
	return strings.TrimSpace(frames[len(frames)-1].Stdout) == ""
}

// ReplayPlayerNames extracts player display names from replay agent metadata.
func ReplayPlayerNames(replay CodinGameReplay[CodinGameReplayFrame]) [2]string {
	names := [2]string{"left", "right"}
	for _, a := range replay.GameResult.Agents {
		switch a.Index {
		case 0:
			if pseudo := strings.TrimSpace(a.CodinGamer.Pseudo); pseudo != "" {
				names[0] = pseudo
			}
		case 1:
			if pseudo := strings.TrimSpace(a.CodinGamer.Pseudo); pseudo != "" {
				names[1] = pseudo
			}
		}
	}
	return names
}

// ReplayAnnotations carries the arena-only metadata written into a saved
// replay JSON by `arena replay`. Zero values are omitted from the output.
type ReplayAnnotations struct {
	// Blue is the username of the player we are playing for.
	Blue string
	// League is the league level the match was played at (parsed from the
	// CodinGame question title).
	League int
	// Source identifies which subcommand produced the file (ReplaySourceGet
	// or ReplaySourceLeaderboard).
	Source string
	// FetchedAt is the time the replay was downloaded; serialized as RFC 3339.
	FetchedAt time.Time
	// Leaderboard captures the player's rank/division at fetch time. Set only
	// when the replay was discovered via `replay leaderboard`.
	Leaderboard *ReplayLeaderboardInfo
	// PuzzleID is the canonical CodinGame puzzleId for the game we're
	// downloading replays for. Layered into the saved replay even when the
	// API returned 0, which CG occasionally does and which would otherwise
	// make convert reject the file.
	PuzzleID int
	// PuzzleTitle is the canonical title for the puzzle (e.g.
	// "SnakeByte - Winter Challenge 2026"). Written as the same two-element
	// array CG's API uses so on-disk replays have a uniform shape.
	PuzzleTitle string
}

// Top-level keys that exist only to drive the CodinGame web viewer. Stripped
// from saved replays so the on-disk shape mirrors what convert/analyze read.
var replayStripTopLevel = []string{"viewer", "shareable"}

// gameResult sub-keys with the same "viewer-only" property.
var replayStripGameResult = []string{"metadata", "tooltips"}

// Per-frame keys we drop. "view" is the serialized viewer state and is
// usually the majority of the file size; "gameInformation" / "keyframe" are
// viewer hints unused by convert/analyze.
var replayStripFrame = []string{"view", "gameInformation", "keyframe"}

// PrepareReplay normalizes a raw CodinGame replay JSON body for local
// storage: removes viewer-only payloads, layers in the arena annotations,
// and pretty-prints the result. If the body is not a JSON object, the
// original bytes are returned unchanged.
func PrepareReplay(body []byte, ann ReplayAnnotations) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return body, nil
	}

	for _, key := range replayStripTopLevel {
		delete(top, key)
	}

	if err := promoteReplaySeed(top); err != nil {
		return nil, err
	}

	if err := stripReplayInner(top); err != nil {
		return nil, err
	}

	if err := setIfNotEmpty(top, "blue", ann.Blue); err != nil {
		return nil, err
	}
	if err := setIfNotZero(top, "league", ann.League); err != nil {
		return nil, err
	}
	if err := setIfNotEmpty(top, "source", ann.Source); err != nil {
		return nil, err
	}
	if !ann.FetchedAt.IsZero() {
		if err := setRaw(top, "fetched_at", ann.FetchedAt.UTC().Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}
	if ann.Leaderboard != nil {
		if err := setRaw(top, "leaderboard", ann.Leaderboard); err != nil {
			return nil, err
		}
	}
	// CG's API occasionally returns puzzleId=0 and omits puzzleTitle for some
	// games (observed for game IDs 882653023, 882783026, 882785040 in the
	// winter2026 leaderboard pull). Override with the canonical values from
	// the factory so convert doesn't have to recover after the fact.
	if ann.PuzzleID != 0 {
		if err := setRaw(top, "puzzleId", ann.PuzzleID); err != nil {
			return nil, err
		}
	}
	if ann.PuzzleTitle != "" {
		if err := setRaw(top, "puzzleTitle", ann.PuzzleTitle); err != nil {
			return nil, err
		}
	}

	body, err := json.Marshal(top)
	if err != nil {
		return nil, err
	}
	return prettyJSON(body)
}

// promoteReplaySeed reads gameResult.refereeInput, lifts the seed to the
// top-level "seed" field as a JSON string, then drops refereeInput. Silent
// no-op if gameResult is absent or no seed token is found.
func promoteReplaySeed(top map[string]json.RawMessage) error {
	var gameResult map[string]json.RawMessage
	if err := json.Unmarshal(top["gameResult"], &gameResult); err != nil {
		return nil
	}
	raw, ok := gameResult["refereeInput"]
	if !ok {
		return nil
	}
	var refereeInput string
	if err := json.Unmarshal(raw, &refereeInput); err != nil {
		return nil
	}
	seed, ok := ParseReplaySeed(refereeInput)
	if !ok {
		return nil
	}

	if err := setRaw(top, "seed", strconv.FormatInt(seed, 10)); err != nil {
		return err
	}
	delete(gameResult, "refereeInput")

	var err error
	top["gameResult"], err = json.Marshal(gameResult)
	return err
}

// stripReplayInner removes viewer-only fields nested under gameResult and
// inside each frame. Silent no-op if gameResult or frames are absent.
func stripReplayInner(top map[string]json.RawMessage) error {
	var gameResult map[string]json.RawMessage
	if err := json.Unmarshal(top["gameResult"], &gameResult); err != nil {
		return nil
	}
	for _, key := range replayStripGameResult {
		delete(gameResult, key)
	}

	if err := stripReplayFrames(gameResult); err != nil {
		return err
	}

	var err error
	top["gameResult"], err = json.Marshal(gameResult)
	return err
}

func stripReplayFrames(gameResult map[string]json.RawMessage) error {
	var frames []json.RawMessage
	if err := json.Unmarshal(gameResult["frames"], &frames); err != nil {
		return nil
	}

	changed := false
	for i, frameBody := range frames {
		var frame map[string]json.RawMessage
		if err := json.Unmarshal(frameBody, &frame); err != nil {
			continue
		}
		dropped := false
		for _, key := range replayStripFrame {
			if _, ok := frame[key]; ok {
				delete(frame, key)
				dropped = true
			}
		}
		if !dropped {
			continue
		}
		var err error
		frames[i], err = json.Marshal(frame)
		if err != nil {
			return err
		}
		changed = true
	}
	if !changed {
		return nil
	}

	var err error
	gameResult["frames"], err = json.Marshal(frames)
	return err
}

func setIfNotEmpty(top map[string]json.RawMessage, key, value string) error {
	if value == "" {
		return nil
	}
	return setRaw(top, key, value)
}

func setIfNotZero(top map[string]json.RawMessage, key string, value int) error {
	if value == 0 {
		return nil
	}
	return setRaw(top, key, value)
}

func setRaw(top map[string]json.RawMessage, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	top[key] = raw
	return nil
}

func prettyJSON(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
