// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/games/spring2021"
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// MaxTurns is a generous cap. The Java engine ends a match in at most
// MAX_ROUNDS * (GATHERING + ACTIONS + SUN_MOVE) turns plus a few action
// sub-turns per round; 24 rounds * (number of action-stops + 2 frame turns)
// stays well under 1000.
const MaxTurns = 1000

type Factory struct{}

func NewFactory() arena.GameFactory {
	return &Factory{}
}

func (f *Factory) Name() string { return "spring2021" }

func (f *Factory) Rules() string { return spring2021.Rules }

func (f *Factory) Trace() string { return spring2021.Trace }

func (f *Factory) PuzzleID() int { return 730 }

// TurnModel selects PhaseTurnModel: Spring 2021's engine emits standalone
// trace turns for GATHERING and SUN_MOVE phases (alongside ACTIONS), so
// every empty-stdout frame in the CG replay is counted as its own turn.
func (f *Factory) TurnModel() arena.TurnModel { return arena.PhaseTurnModel{} }

func (f *Factory) PuzzleTitle() string { return "Spring Challenge 2021 - Photosynthesis" }

func (f *Factory) LeaderboardSlug() string { return "spring-challenge-2021" }

func (f *Factory) MaxTurns() int { return MaxTurns }

func (f *Factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	game := NewGame(seed, f.ResolveLeague(options))
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	return NewReferee(game), players
}

// ResolveLeague returns the league level the factory will run with, defaulting
// to Bronze+ (the highest league branch) when "league" is unset/unparseable.
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
