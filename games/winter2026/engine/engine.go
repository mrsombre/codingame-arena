// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"strconv"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "winter2026" }

func (f *factory) MaxTurns() int { return 200 }

func (f *factory) NewGame(seed int64, options map[string]string) (arena.Referee, []arena.Player) {
	leagueLevel := 3
	if raw := options["league"]; raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			leagueLevel = value
		}
	}
	game := NewGame(seed, leagueLevel)
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}
