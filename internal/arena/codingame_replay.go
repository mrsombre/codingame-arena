package arena

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// CodinGameReplay is the shape of a CodinGame match replay JSON as returned by
// /services/gameResult/findInformationById, and as stored on disk after the
// `arena replay` downloader strips the no-op top-level "viewer" payload (see
// StripReplayViewer).
//
// Top-level fields are shared across every CodinGame game. The F parameter is
// the per-game frame shape: games that don't need custom per-turn fields can
// use the default CodinGameReplayFrame; games that do can declare their own
// struct and pass it as F.
type CodinGameReplay[F any] struct {
	PuzzleID      int                      `json:"puzzleId"`
	PuzzleTitle   []string                 `json:"puzzleTitle"`
	QuestionTitle string                   `json:"questionTitle"`
	GameResult    CodinGameReplayResult[F] `json:"gameResult"`
}

// CodinGameReplayResult is the gameResult sub-object. Frames is the only
// game-parameterized field.
type CodinGameReplayResult[F any] struct {
	GameID       int64                  `json:"gameId"`
	RefereeInput string                 `json:"refereeInput"`
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

var replayLeaguePattern = regexp.MustCompile(`(?i)level(\d)`)

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
			moves.P0 = append(moves.P0, f.Stdout)
		case 1:
			moves.P1 = append(moves.P1, f.Stdout)
		}
	}
	return moves
}

// ReplayTurnCount returns the number of recorded decision turns in the replay.
func ReplayTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	moves := ReplayMovesFromFrames(replay)
	if len(moves.P0) > len(moves.P1) {
		return len(moves.P0)
	}
	return len(moves.P1)
}

// ReplayTraceTurnCount returns the number of engine frames represented by the
// replay. CodinGame stores simultaneous player outputs as one frame per player.
// Mid-replay empty-stdout frames are Spring 2020 speed sub-turns, which the
// engine folds into the preceding main turn and emits no trace turn for. The
// trailing empty-stdout frame is the game-over marker; the engine emits one
// extra trace turn for it (mirroring Java's post-game-over gameTurn frame).
func ReplayTraceTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int {
	turns := 0
	seenOutput := map[int]bool{}

	flushOutputTurn := func() {
		if len(seenOutput) == 0 {
			return
		}
		turns++
		clear(seenOutput)
	}

	for _, frame := range replay.GameResult.Frames {
		if frame.AgentID < 0 {
			continue
		}

		if strings.TrimSpace(frame.Stdout) == "" {
			flushOutputTurn()
			continue
		}

		if seenOutput[frame.AgentID] {
			flushOutputTurn()
		}
		seenOutput[frame.AgentID] = true
		if len(seenOutput) == 2 {
			flushOutputTurn()
		}
	}
	flushOutputTurn()

	if hasTrailingEngineFrame(replay) {
		turns++
	}

	return turns
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
	names := [2]string{"p0", "p1"}
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

// StripReplayViewer removes the top-level "viewer" field and the per-frame
// "view" payloads from a raw CodinGame replay JSON body, then pretty-prints
// the result. Those fields carry serialized viewer state that nothing in this
// codebase reads, and they are typically the majority of the file size. If the
// body is not a JSON object, the original bytes are returned unchanged.
func StripReplayViewer(body []byte) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return body, nil
	}

	delete(top, "viewer")

	if err := stripReplayFrameViews(top); err != nil {
		return nil, err
	}

	var err error
	body, err = json.Marshal(top)
	if err != nil {
		return nil, err
	}

	return prettyJSON(body)
}

func stripReplayFrameViews(top map[string]json.RawMessage) error {
	var gameResult map[string]json.RawMessage
	if err := json.Unmarshal(top["gameResult"], &gameResult); err != nil {
		return nil
	}

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
		if _, ok := frame["view"]; !ok {
			continue
		}
		delete(frame, "view")

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
	if err != nil {
		return err
	}
	top["gameResult"], err = json.Marshal(gameResult)
	return err
}

func prettyJSON(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
