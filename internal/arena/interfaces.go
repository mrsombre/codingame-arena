package arena

import (
	"encoding/json"

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
	// (e.g. "<Game> - <Season> <Year>"). convert uses it to recover
	// replays where the API returned puzzleId=0 but did include a
	// puzzleTitle entry.
	PuzzleTitle() string
	// LeaderboardSlug returns the puzzle pretty-id used in the CodinGame
	// leaderboard URL (e.g. "<season>-<year>-<game>"), so the replay
	// command can resolve a player's last battles without the caller
	// passing the URL on each invocation.
	LeaderboardSlug() string
	NewGame(seed int64, options *viper.Viper) (Referee, []Player)
	MaxTurns() int
}

// MetricsProvider produces game-specific metrics for a completed match.
// Referee should implement this.
type MetricsProvider interface {
	Metrics() []Metric
}

// TurnTraceProvider produces structured game traces per turn, partitioned by
// player: index 0 is everything player 0 owned this turn, index 1 player 1.
// Cross-owner events should be mirrored into both slots.
// Optional — if Referee also implements this, match captures traces.
type TurnTraceProvider interface {
	TurnTraces(turn int, players []Player) [2][]TurnTrace
}

// TraceTurnDecorator returns a game-owned opaque per-turn payload that the
// arena copies into TraceTurn.State as-is. The arena never inspects it;
// downstream consumers (game viewers, analyzers) decode the bytes back
// into a game-specific struct. Mirrors the TurnTrace.Data pattern at the
// per-turn level.
//
// Match calls it after command parsing and before PerformGameUpdate, so
// the payload reflects the state the players saw when choosing actions.
// Returning nil/empty bytes leaves State unset on that turn.
type TraceTurnDecorator interface {
	DecorateTraceTurn(turn int, players []Player) json.RawMessage
}

// TraceGlobalInfoProducer returns the global-info lines to record on the
// trace's setup field. Optional — when implemented, the runner uses it
// instead of falling back to Referee.GlobalInfoFor(players[0]).
//
// Use this when the standard per-side serializer leaks side-specific state
// (e.g., fog-of-war filters, player-index headers) that an analyzer wants
// to skip. Implementations should produce a side-agnostic / god-mode view
// describing the full game state.
type TraceGlobalInfoProducer interface {
	TraceGlobalInfo() []string
}

// TraceFrameInfoProducer returns the per-turn frame-info lines to record
// on each TraceTurn.GameInput. Optional — when implemented, the runner
// uses it instead of falling back to Referee.FrameInfoFor(players[0]).
//
// Use this for fog-of-war games (e.g., Spring 2020) so the trace records
// every entity's state every turn rather than blue's filtered view.
// Implementations should produce a side-agnostic / god-mode view from a
// canonical perspective (typically side 0 with all visibility filters
// disabled).
type TraceFrameInfoProducer interface {
	TraceFrameInfo() []string
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
// turn is the final loop turn; deactivationTurns[i] is the turn player i was
// deactivated (or -1); firstOutputTurns[i] is the turn player i was first
// prompted for output (or -1 if never prompted — only possible when the bot
// was deactivated before any output turn or the match never reached one).
// Implementations should use deactivationTurns[i] == firstOutputTurns[i] to
// distinguish TIMEOUT_START (timed out on first prompt) from TIMEOUT (timed
// out later) — this is independent of the game's frame model, so games with
// non-decision leading frames (e.g., Spring 2021 GATHERING) classify
// correctly. Optional — if Referee implements this, match stamps the value
// onto the trace as "endReason".
type EndReasonProvider interface {
	EndReason(turn int, players []Player, deactivationTurns, firstOutputTurns [2]int) string
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

// TurnModel describes how a game's trace turns relate to the frames in a
// CodinGame replay. Replay verification asks the model two things:
//
//   - ExpectedTraceTurnCount: how many trace turns the engine should emit
//     for this replay (matches len(trace.Turns)).
//   - MainTurnCount: how many of those are player-decision turns (matches
//     trace.MainTurns). Phase frames (spring2021 GATHERING/SUN_MOVE) and
//     post-end frames (spring2020 gameOverFrame) are excluded.
//
// Concrete implementations are FlatTurnModel (winter2026), PostEndTurnModel
// (spring2020), and PhaseTurnModel (spring2021).
type TurnModel interface {
	ExpectedTraceTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int
	MainTurnCount(replay CodinGameReplay[CodinGameReplayFrame]) int
}

// TurnModeler is the factory-level hook that names which TurnModel a game
// uses. Factories without this hook fall back to FlatTurnModel.
type TurnModeler interface {
	TurnModel() TurnModel
}

// RulesProvider returns the bundled rules.md contents for a game so the
// `arena game <game> rules` subcommand can print them without a sidecar
// filesystem path. Optional — factories without this hook surface a "no
// rules bundled" error from the rules subcommand.
type RulesProvider interface {
	Rules() string
}

// TraceProvider returns the bundled trace.md contents for a game so the
// `arena game <game> trace` subcommand can print the per-game trace-format
// reference without a sidecar filesystem path. Optional — factories without
// this hook surface a "no trace docs bundled" error from the trace
// subcommand.
type TraceProvider interface {
	Trace() string
}
