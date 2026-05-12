// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java
//
// The upstream "referee" responsibilities live inside engine.Board.tick and
// the per-task classes under engine/task/. This file wires the arena.Referee
// contract to the in-progress port — most lifecycle methods are stubs that
// panic until the port lands.
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

type Referee struct {
	Game *Game
}

func NewReferee(game *Game) *Referee {
	return &Referee{Game: game}
}

func (r *Referee) Init(players []arena.Player) {
	typed := make([]*Player, 0, len(players))
	for _, player := range players {
		typed = append(typed, player.(*Player))
	}
	r.Game.Init(typed)
}

func (r *Referee) GlobalInfoFor(player arena.Player) []string {
	panic("spring2026: Referee.GlobalInfoFor not yet implemented")
}

func (r *Referee) FrameInfoFor(player arena.Player) []string {
	panic("spring2026: Referee.FrameInfoFor not yet implemented")
}

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	panic("spring2026: Referee.ParsePlayerOutputs not yet implemented")
}

func (r *Referee) PerformGameUpdate(turn int) {
	r.Game.PerformGameUpdate(turn)
}

func (r *Referee) ResetGameTurnData() {
	r.Game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.Game.ended
}

func (r *Referee) EndGame() {
	r.Game.EndGame()
}

func (r *Referee) OnEnd() {
	r.Game.OnEnd()
}

func (r *Referee) ShouldSkipPlayerTurn(player arena.Player) bool {
	return r.Game.ShouldSkipPlayerTurn(player.(*Player))
}

func (r *Referee) ActivePlayers(players []arena.Player) int {
	count := 0
	for _, player := range players {
		if !player.IsDeactivated() {
			count++
		}
	}
	return count
}
