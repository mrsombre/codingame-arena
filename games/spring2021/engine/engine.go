// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

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

func (f *Factory) PuzzleID() int { return 730 }

func (f *Factory) EmitsReplayPhaseFrames() bool { return true }

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
