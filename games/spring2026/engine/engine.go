// Package engine
package engine

import (
	"strconv"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/games/spring2026"
	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
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

func (f *factory) LeaderboardSlug() string { return "spring-challenge-2026-troll-farm" }

// IsChallengeLeaderboard tells the replay command this game's leaderboard
// lives under the CodinGame "challenge" (community contest) API rather than
// the standard puzzle API. Troll Farm is hosted at /contests/, not
// /multiplayer/, so the puzzleLeaderboardId resolution step doesn't apply.
func (f *factory) IsChallengeLeaderboard() bool { return true }

// MaxTurns is the absolute cap (league 3+). Leagues 1-2 stop at 100 turns
// via the engine's own end-of-game logic.
func (f *factory) MaxTurns() int { return 300 }

func (f *factory) TurnModel() arena.TurnModel { return arena.FlatTurnModel{} }

func (f *factory) NewGame(seed int64, options *viper.Viper) (arena.Referee, []arena.Player) {
	league := f.ResolveLeague(options)
	p0 := NewPlayer(0)
	p1 := NewPlayer(1)
	// SHA1PRNG matches the CG SDK's MultiplayerGameManager.getRandom(); plain
	// java.util.Random would diverge on the very first map-height roll. See
	// engine_board_seed_test.go for byte-for-byte parity coverage.
	board := CreateMap([]*Player{p0, p1}, sha1prng.New(seed), league)
	board.Seed = seed
	return NewReferee(board), []arena.Player{p0, p1}
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
