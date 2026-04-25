// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"strconv"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// maxTurns gives the arena some headroom over Java's 200 main-turn limit so
// speed sub-turns (which Java inserts on top of 200) still fit within a single
// arena loop.
const maxTurns = 400

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "spring2020" }

func (f *factory) MaxTurns() int { return maxTurns }

func (f *factory) NewGame(seed int64, options map[string]string) (arena.Referee, []arena.Player) {
	leagueLevel := 4
	if raw := options["league"]; raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			leagueLevel = value
		}
	}
	game := NewGame(seed, leagueLevel)
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}
