// Package engine
package engine

import (
	"github.com/mrsombre/codingame-arena/games/winter2026"
	"github.com/mrsombre/codingame-arena/internal/arena"
)

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "winter2026" }

func (f *factory) Rules() string { return winter2026.Rules }

func (f *factory) Trace() string { return winter2026.Trace }

func (f *factory) PuzzleID() int { return 13771 }

func (f *factory) PuzzleTitle() string { return "SnakeByte - Winter Challenge 2026" }

func (f *factory) LeaderboardSlug() string { return "winter-challenge-2026-snakebyte" }

func (f *factory) MaxTurns() int { return 200 }

func (f *factory) TurnModel() arena.TurnModel { return arena.FlatTurnModel{} }

func (f *factory) NewGame(seed int64, options arena.GameOptions) (arena.Referee, []arena.Player) {
	game := NewGame(seed, f.ResolveLeague(options))
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}

// ResolveLeague returns the league level the factory will run with for the
// given options, falling back to the Winter 2026 default of 4 when the
// "league" option is unset or unparseable.
func (f *factory) ResolveLeague(options arena.GameOptions) int {
	if options.League > 0 {
		return options.League
	}
	return 4
}
