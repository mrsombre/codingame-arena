// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Board.java
package engine

// Game owns the per-match simulation state. This is the Go counterpart of
// the upstream engine.Board class. The port is in progress — the engine is
// currently a structural stub and does not run matches.
type Game struct {
	Seed   int64
	League int

	// ended is set to true when the engine detects a terminal condition
	// (max turns, 10-turn stall, or stuck-leader win). The Referee polls
	// it after every PerformGameUpdate.
	ended bool
}

func NewGame(seed int64, league int) *Game {
	return &Game{Seed: seed, League: league}
}

func (g *Game) Init(players []*Player) {
	panic("spring2026: Game.Init not yet implemented")
}

func (g *Game) ResetGameTurnData() {
	panic("spring2026: Game.ResetGameTurnData not yet implemented")
}

func (g *Game) PerformGameUpdate(turn int) {
	panic("spring2026: Game.PerformGameUpdate not yet implemented")
}

func (g *Game) ShouldSkipPlayerTurn(player *Player) bool {
	return false
}

func (g *Game) EndGame() {
	panic("spring2026: Game.EndGame not yet implemented")
}

func (g *Game) OnEnd() {
	panic("spring2026: Game.OnEnd not yet implemented")
}
