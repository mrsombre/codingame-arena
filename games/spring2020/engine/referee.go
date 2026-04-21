// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Referee drives the Game through the arena runner lifecycle. It tracks the
// Spring 2020 "speed sub-turn" mechanic locally so the arena can treat every
// iteration of its main loop uniformly.
type Referee struct {
	game           *Game
	commandManager *CommandManager
	speedTurn      bool
	gameOverFrame  bool
}

func NewReferee(game *Game) *Referee {
	r := &Referee{game: game}
	r.commandManager = NewCommandManager(&game.summary, game)
	return r
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
	if r.speedTurn {
		return
	}
	for _, player := range players {
		p := player.(*Player)
		if p.IsDeactivated() {
			continue
		}
		r.commandManager.ParseCommands(p, p.GetOutputs())
	}
}

func (r *Referee) PerformGameUpdate(turn int) {
	if r.gameOverFrame {
		r.game.ResetGameTurnData()
		r.game.PerformGameUpdate()
		r.game.PerformGameOver()
		r.game.EndGame()
		return
	}

	if r.speedTurn {
		r.game.PerformGameSpeedUpdate()
	} else {
		r.game.PerformGameUpdate()
	}

	if r.game.IsGameOver() {
		r.gameOverFrame = true
	}
	r.speedTurn = r.game.IsSpeedTurn()
}

func (r *Referee) ResetGameTurnData() {
	if r.speedTurn {
		return
	}
	r.game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.game.Ended()
}

func (r *Referee) EndGame() {
	r.game.EndGame()
}

func (r *Referee) OnEnd() {
	r.game.OnEnd()
}

func (r *Referee) ShouldSkipPlayerTurn(player arena.Player) bool {
	return r.speedTurn
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

// RawScores returns per-player pellet counts — the raw in-match score.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.game.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		scores[idx] = p.Pellets
	}
	return scores
}

// Metrics emits Spring 2020-specific per-match metrics.
func (r *Referee) Metrics() []arena.Metric {
	remainingPellets := 0
	remainingCherries := 0
	for _, cell := range r.game.Grid.Cells() {
		if cell.HasPellet {
			remainingPellets++
		}
		if cell.HasCherry {
			remainingCherries++
		}
	}
	alive0, alive1 := 0, 0
	for _, pac := range r.game.Pacmen {
		if pac.Owner.GetIndex() == 0 {
			alive0++
		} else {
			alive1++
		}
	}
	return []arena.Metric{
		{Label: "pellets_remaining", Value: float64(remainingPellets)},
		{Label: "cherries_remaining", Value: float64(remainingCherries)},
		{Label: "pacs_alive_p0", Value: float64(alive0)},
		{Label: "pacs_alive_p1", Value: float64(alive1)},
	}
}
