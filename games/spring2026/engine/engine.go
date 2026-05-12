// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/games/spring2026"
	"github.com/mrsombre/codingame-arena/internal/arena"
)

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "spring2026" }

func (f *factory) Rules() string { return spring2026.Rules }

// PuzzleID is 0: Troll Farm is a community contest, not under the
// CodinGame/ GitHub org, and the CodinGame puzzle metadata is not yet
// reconciled with the arena replay machinery.
func (f *factory) PuzzleID() int { return 0 }

func (f *factory) PuzzleTitle() string { return "Troll Farm - Spring Challenge 2026" }

func (f *factory) LeaderboardSlug() string { return "spring-challenge-2026-troller-farm" }

// MaxTurns is the absolute cap (league 3+). Leagues 1-2 stop at 100 turns
// via the engine's own end-of-game logic.
func (f *factory) MaxTurns() int { return 300 }

func (f *factory) TurnModel() arena.TurnModel { return arena.FlatTurnModel{} }

func (f *factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	game := NewGame(seed, f.ResolveLeague(options))
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}

// ResolveLeague returns the league level the factory will run with for the
// given options, falling back to the Spring 2026 default of 4 when the
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
