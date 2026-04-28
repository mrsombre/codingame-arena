// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "winter2026" }

func (f *factory) MaxTurns() int { return 200 }

func (f *factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	leagueLevel := 4
	if raw := options.GetString("league"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			leagueLevel = value
		}
	}
	game := NewGame(seed, leagueLevel)
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}
