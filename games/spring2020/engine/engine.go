// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// MaxTurns gives the arena some headroom over Java's 200 main-turn limit so
// speed sub-turns (which Java inserts on top of 200) still fit within a single
// arena loop.
const MaxTurns = 400

type Factory struct{}

func NewFactory() arena.GameFactory {
	return &Factory{}
}

func (f *Factory) Name() string { return "spring2020" }

func (f *Factory) PuzzleID() int { return 592 }

func (f *Factory) MaxTurns() int { return MaxTurns }

func (f *Factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
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
