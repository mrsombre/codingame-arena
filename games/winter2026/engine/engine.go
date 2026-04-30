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

func (f *factory) PuzzleID() int { return 13771 }

func (f *factory) PuzzleTitle() string { return "SnakeByte - Winter Challenge 2026" }

func (f *factory) MaxTurns() int { return 200 }

func (f *factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	game := NewGame(seed, f.ResolveLeague(options))
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}

// ResolveLeague returns the league level the factory will run with for the
// given options, falling back to the Winter 2026 default of 4 when the
// "league" option is unset or unparseable.
func (f *factory) ResolveLeague(options *viper.Viper) int {
	if options != nil {
		if raw := options.GetString("league"); raw != "" {
			if value, err := strconv.Atoi(raw); err == nil {
				return value
			}
		}
	}
	return 4
}
