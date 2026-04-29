// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// MaxTurns matches Java's 200 main-turn cap. Speed sub-steps are folded into
// a single PerformGameUpdate call (see Game.PerformGameUpdate), so an arena
// turn maps 1:1 to a Java main turn — no extra headroom needed.
const MaxTurns = 200

type Factory struct{}

func NewFactory() arena.GameFactory {
	return &Factory{}
}

func (f *Factory) Name() string { return "spring2020" }

func (f *Factory) PuzzleID() int { return 592 }

func (f *Factory) MaxTurns() int { return MaxTurns }

func (f *Factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	game := NewGame(seed, f.ResolveLeague(options))
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}

// ResolveLeague returns the league level the factory will run with for the
// given options, falling back to the Spring 2020 default of 4 when the
// "league" option is unset or unparseable.
func (f *Factory) ResolveLeague(options *viper.Viper) int {
	if options != nil {
		if raw := options.GetString("league"); raw != "" {
			if value, err := strconv.Atoi(raw); err == nil {
				return value
			}
		}
	}
	return 4
}
