// Package engine
package engine

import (
	"github.com/mrsombre/codingame-arena/games/summer2026"
	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

type factory struct{}

func NewFactory() arena.GameFactory {
	return &factory{}
}

func (f *factory) Name() string { return "summer2026" }

func (f *factory) Rules() string { return summer2026.Rules }

func (f *factory) Trace() string { return summer2026.Trace }

// PuzzleID is 0: the contest is still running and CodinGame has not
// published puzzle metadata for Back Track King yet.
func (f *factory) PuzzleID() int { return 0 }

func (f *factory) PuzzleTitle() string { return "Back Track King - Summer Challenge 2026" }

func (f *factory) LeaderboardSlug() string { return "summer-challenge-2026-back-track-king" }

// IsChallengeLeaderboard: Back Track King is hosted at /contests/, so its
// leaderboard lives under the CodinGame challenge API rather than the puzzle
// API.
func (f *factory) IsChallengeLeaderboard() bool { return true }

// MaxTurns is Game.MAX_TURNS. The SDK's setMaxTurns(400) is a safety ceiling,
// not a rule of the game.
func (f *factory) MaxTurns() int { return MAX_TURNS }

func (f *factory) TurnModel() arena.TurnModel { return arena.FlatTurnModel{} }

func (f *factory) NewGame(seed int64, options arena.GameOptions) (arena.Referee, []arena.Player) {
	p0 := NewPlayer(0)
	p1 := NewPlayer(1)
	// SHA1PRNG matches the SDK's MultiplayerGameManager.getRandom().
	game := NewGame(sha1prng.New(seed), f.ResolveLeague(options))
	// Online replays with POIs publish sideQuestPoints even when it is zero;
	// the later rules omit those keys and generate no POIs.
	_, sideQuest0 := options.ReplayMetadata["sideQuestPoints_0"]
	_, sideQuest1 := options.ReplayMetadata["sideQuestPoints_1"]
	game.EnableSideQuest = sideQuest0 || sideQuest1
	return NewReferee(game), []arena.Player{p0, p1}
}

// ResolveLeague returns the league level the factory will run with for the
// given options, falling back to the full game when "league" is unset or
// unparseable. Leagues 1-2 are tutorials whose win condition comes from
// TutorialManager; 3-5 all run the same rules.
func (f *factory) ResolveLeague(options arena.GameOptions) int {
	if options.League > 0 {
		return options.League
	}
	return DEFAULT_LEAGUE
}
