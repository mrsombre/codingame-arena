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

// TraceSummaryProvider returns the per-match aggregate of trace events.
// Optional — if Referee also implements this, match writes the summary into
// the trace JSON root under "trace_summary".
type TraceSummaryProvider interface {
	TraceSummary() TraceSummary
}

// RawScoresProvider returns per-player raw scores before any end-of-game
// tiebreaker adjustments run. Used by the trace writer so match traces record
// the intrinsic game state (e.g. sum of alive bird segments) rather than the
// adjusted value the referee reports via Player.GetScore after OnEnd.
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
