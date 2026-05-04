package arena

import (
	"github.com/spf13/viper"
)

// Referee drives the game through a standard protocol.
// Match calls these methods in order during the game loop.
type Referee interface {
	Init(players []Player)
	GlobalInfoFor(player Player) []string
	FrameInfoFor(player Player) []string
	ParsePlayerOutputs(players []Player)
	PerformGameUpdate(turn int)
	ResetGameTurnData()
	Ended() bool
	EndGame()
	OnEnd()
	ShouldSkipPlayerTurn(player Player) bool
	ActivePlayers(players []Player) int
}

// Player represents one agent in the match.
// Engine owns the concrete type; match interacts through this interface.
type Player interface {
	GetIndex() int
	GetScore() int
	SetScore(int)
	IsDeactivated() bool
	Deactivate(reason string)
	DeactivationReason() string
	IsTimedOut() bool
	SetTimedOut(bool)
	GetExpectedOutputLines() int
	SendInputLine(string)
	ConsumeInputLines() []string
	GetOutputs() []string
	SetOutputs([]string)
	GetOutputError() error
	SetExecuteFunc(func() error)
	Execute() error
}

// GameFactory creates game instances for each match.
type GameFactory interface {
	Name() string
	PuzzleID() int
	// PuzzleTitle returns the human-readable CodinGame puzzle title
	// (e.g. "SnakeByte - Winter Challenge 2026"). convert uses it to
	// recover replays where the API returned puzzleId=0 but did include
	// a puzzleTitle entry.
	PuzzleTitle() string
	// LeaderboardSlug returns the puzzle pretty-id used in the CodinGame
	// leaderboard URL (e.g. "winter-challenge-2026-snakebyte"), so the
	// replay command can resolve a player's last battles without the
	// caller passing the URL on each invocation.
	LeaderboardSlug() string
	NewGame(seed int64, options *viper.Viper) (Referee, []Player)
	MaxTurns() int
}

// MetricsProvider produces game-specific metrics for a completed match.
// Referee should implement this.
type MetricsProvider interface {
	Metrics() []Metric
}

// TurnTraceProvider produces structured game traces per turn.
// Optional — if Referee also implements this, match captures traces.
type TurnTraceProvider interface {
	TurnTraces(turn int, players []Player) []TurnTrace
}

// RawScoresProvider returns per-player raw scores before any end-of-game
// tiebreaker adjustments run. Used by the trace writer so match traces record
// the engine's intrinsic scoring state rather than the adjusted value the
// referee reports via Player.GetScore after OnEnd.
// Optional — if Referee also implements this, match captures raw scores.
type RawScoresProvider interface {
	RawScores() [2]int
}

// LeagueResolver returns the league level a factory will run with for the
// given options (applying its game-specific default when "league" is unset).
// Optional — if a GameFactory implements this, match stamps the resolved
// value onto each trace as "league".
type LeagueResolver interface {
	ResolveLeague(options *viper.Viper) int
}

// EndReasonProvider returns a categorized reason for why the match ended.
// turn is the final loop turn; deactivationTurns[i] is the turn player i
// was deactivated (or -1). Optional — if Referee implements this, match
// stamps the value onto the trace as "end_reason".
type EndReasonProvider interface {
	EndReason(turn int, players []Player, deactivationTurns [2]int) string
}

// GameOverFrameReporter signals that the engine has detected game-over and
// is about to run its post-end "game over frame" (Java spring2020's
// gameOverFrame branch — one extra gameTurn that calls performGameOver and
// endGame). The runner skips player polling and command parsing on that
// iteration: the outcome is decided, and re-polling exhausted replay bots
// would interpret an empty stdout as a Timeout and erroneously deactivate
// the surviving side. Engines that end the match on the same turn game-over
// is detected (winter2026) don't need to implement this.
type GameOverFrameReporter interface {
	InGameOverFrame() bool
}

// PostEndFrameEmitter is the static factory-level companion to
// GameOverFrameReporter: GameFactories whose engines always emit a separate
// post-end trace turn for their game-over frame implement this so replay
// verification can pre-compute the expected trace turn count without
// instantiating a referee. Spring 2020 implements this; Winter 2026 does
// not.
type PostEndFrameEmitter interface {
	EmitsPostEndFrame() bool
}

// ReplayPhaseFrameEmitter signals that the engine emits standalone trace
// turns for non-decision phases (e.g. Spring 2021's GATHERING and SUN_MOVE
// frames) which appear in the CodinGame replay as empty-stdout frames.
// Replay verification needs this hint because the default decision-only
// counting model treats empty-stdout frames as sub-turn flushes that don't
// add their own turn — correct for spring2020/winter2026 but undercounts
// spring2021's per-round phase trio. Spring 2021 implements this.
type ReplayPhaseFrameEmitter interface {
	EmitsReplayPhaseFrames() bool
}
