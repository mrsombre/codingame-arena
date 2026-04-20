package arena

import "encoding/json"

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
	NewGame(seed int64, options map[string]string) (Referee, []Player)
	MaxTurns() int
}

// MetricsProvider produces game-specific metrics for a completed match.
// Referee should implement this.
type MetricsProvider interface {
	Metrics() []Metric
}

// TraceProvider produces an opaque per-turn game state snapshot.
// Optional — if Referee also implements this, match captures snapshots.
type TraceProvider interface {
	SnapshotTurn(turn int, players []Player) json.RawMessage
}

// TurnEventProvider produces structured game events per turn.
// Optional — if Referee also implements this, match captures events.
type TurnEventProvider interface {
	TurnEvents(turn int, players []Player) []TurnEvent
}
