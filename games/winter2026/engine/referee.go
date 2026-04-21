// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

type Referee struct {
	game           *Game
	commandManager *CommandManager
}

func NewReferee(game *Game) *Referee {
	return &Referee{
		game:           game,
		commandManager: NewCommandManager(&game.summary),
	}
}

func (r *Referee) Init(players []arena.Player) {
	typed := make([]*Player, 0, len(players))
	for _, player := range players {
		typed = append(typed, player.(*Player))
	}
	r.game.Init(typed)
}

func (r *Referee) GlobalInfoFor(player arena.Player) []string {
	return serializeGlobalInfoFor(player.(*Player), r.game)
}

func (r *Referee) FrameInfoFor(player arena.Player) []string {
	return serializeFrameInfoFor(player.(*Player), r.game)
}

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	for _, player := range players {
		p := player.(*Player)
		if p.IsDeactivated() || r.game.ShouldSkipPlayerTurn(p) {
			continue
		}
		r.commandManager.ParseCommands(p, p.GetOutputs())
	}
}

func (r *Referee) PerformGameUpdate(turn int) {
	r.game.PerformGameUpdate(turn)
}

func (r *Referee) ResetGameTurnData() {
	r.game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.game.ended
}

func (r *Referee) EndGame() {
	r.game.EndGame()
}

func (r *Referee) OnEnd() {
	r.game.OnEnd()
}

func (r *Referee) ShouldSkipPlayerTurn(player arena.Player) bool {
	return r.game.ShouldSkipPlayerTurn(player.(*Player))
}

func (r *Referee) ActivePlayers(players []arena.Player) int {
	active := 0
	for _, player := range players {
		if !player.IsDeactivated() {
			active++
		}
	}
	return active
}

func (r *Referee) TurnEvents(_ int, _ []arena.Player) []arena.TurnEvent {
	if len(r.game.events) == 0 {
		return nil
	}
	out := make([]arena.TurnEvent, len(r.game.events))
	copy(out, r.game.events)
	return out
}

// RawScores returns each player's score as the sum of alive bird segments.
// This is the intrinsic game metric before OnEnd applies its tie-breaking
// adjustment (subtracting losses when raw scores tie) that can produce
// negative values in Player.GetScore.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.game.players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		for _, b := range p.birds {
			if b.Alive {
				scores[idx] += len(b.Body)
			}
		}
	}
	return scores
}

func (r *Referee) Metrics() []arena.Metric {
	return []arena.Metric{
		{Label: "apples_remaining", Value: float64(len(r.game.grid.Apples))},
		{Label: "losses_p0", Value: float64(r.game.losses[0])},
		{Label: "losses_p1", Value: float64(r.game.losses[1])},
	}
}
