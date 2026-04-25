package arena

// CodinGameReplay is the shape of a CodinGame match replay JSON as returned by
// /services/gameResult/findInformationById, and as stored on disk after the
// `arena replay` downloader strips the viewer/per-frame view payloads.
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
