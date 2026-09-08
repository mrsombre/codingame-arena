package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// parse runs one output line through CommandManager against a throwaway game.
// The grid is irrelevant here — parsing never touches it — but Init still
// generates one, so the game needs a real RNG.
func parse(t *testing.T, line string) (*Game, *Player) {
	t.Helper()
	game := NewGame(sha1prng.New(0), DEFAULT_LEAGUE)
	player := NewPlayer(0)
	game.Init([]*Player{player, NewPlayer(1)})
	NewCommandManager(game).ParseCommands(player, []string{line})
	return game, player
}

func TestParseAcceptsBothPlaceTrackSpellings(t *testing.T) {
	for _, line := range []string{"PLACE_TRACK 3 4", "PLACE_TRACKS 3 4"} {
		t.Run(line, func(t *testing.T) {
			_, player := parse(t, line)

			require.False(t, player.IsDeactivated())
			require.Len(t, player.Intents, 1)
			assert.Equal(t, ACTION_PLACE_TRACK, player.Intents[0].Type)
			assert.Equal(t, Coord{3, 4}, player.Intents[0].Coord)
		})
	}
}

func TestParseIsCaseInsensitive(t *testing.T) {
	cases := map[string]ActionType{
		"place_track 1 2":   ACTION_PLACE_TRACK,
		"Place_Tracks 1 2":  ACTION_PLACE_TRACK,
		"autoplace 0 0 1 1": ACTION_AUTOPLACE,
		"disrupt 2":         ACTION_DISRUPT,
		"DiSrUpT 1 2":       ACTION_DISRUPT_ALT,
		"wait":              ACTION_WAIT,
	}
	for line, want := range cases {
		t.Run(line, func(t *testing.T) {
			_, player := parse(t, line)

			require.False(t, player.IsDeactivated())
			require.Len(t, player.Intents, 1)
			assert.Equal(t, want, player.Intents[0].Type)
		})
	}
}

func TestParseSplitsOnSemicolonsAndTrimsEachCommand(t *testing.T) {
	_, player := parse(t, " PLACE_TRACK 1 1 ; WAIT ;PLACE_TRACK 2 2")

	require.False(t, player.IsDeactivated())
	require.Len(t, player.Intents, 3)
	assert.Equal(t, Coord{1, 1}, player.Intents[0].Coord)
	assert.Equal(t, ACTION_WAIT, player.Intents[1].Type)
	assert.Equal(t, Coord{2, 2}, player.Intents[2].Coord)
}

// MESSAGE is consumed by the player rather than queued, and its pattern
// excludes ';' so it cannot swallow the commands that follow it.
func TestParseMessageIsStoredOnThePlayerAndNotQueuedAsAnIntent(t *testing.T) {
	_, player := parse(t, "MESSAGE hello there;PLACE_TRACK 0 0")

	require.False(t, player.IsDeactivated())
	assert.Equal(t, "hello there", player.GetMessage())
	require.Len(t, player.Intents, 1)
	assert.Equal(t, ACTION_PLACE_TRACK, player.Intents[0].Type)
}

func TestParseDisqualifiesWithTheExpectedSyntaxForTheClosestCommand(t *testing.T) {
	cases := map[string]string{
		"PLACE_TRACK 1":      "Invalid Input: Expected PLACE_TRACK x y but got 'PLACE_TRACK 1'",
		"PLACE_TRACK -1 2":   "Invalid Input: Expected PLACE_TRACK x y but got 'PLACE_TRACK -1 2'",
		"PLACE_TRACKS 1 2 3": "Invalid Input: Expected PLACE_TRACK x y but got 'PLACE_TRACKS 1 2 3'",
		"AUTOPLACE 1 2":      "Invalid Input: Expected AUTOPLACE x1 y1 x2 y2 but got 'AUTOPLACE 1 2'",
		"DISRUPT":            "Invalid Input: Expected DISRUPT zoneId but got 'DISRUPT'",
		"MESSAGE":            "Invalid Input: Expected MESSAGE text but got 'MESSAGE'",
		"BUILD 1 2":          "Invalid Input: Expected AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT but got 'BUILD 1 2'",
		"place_track 1":      "Invalid Input: Expected AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT but got 'place_track 1'",
	}
	for line, want := range cases {
		t.Run(line, func(t *testing.T) {
			game, player := parse(t, line)

			assert.True(t, player.IsDeactivated())
			assert.Equal(t, -1, player.GetScore())
			assert.Equal(t, want, player.DeactivationReason())
			assert.Equal(t, []string{
				want,
				"¤RED¤Player 0: disqualified!§RED§",
			}, game.Summary)
		})
	}
}

// GetExpected matches on a case-sensitive prefix even though the patterns
// themselves are not, so a lowercase near-miss falls through to the generic
// syntax line. The case above pins that; this names why.
func TestGetExpectedIsCaseSensitiveUnlikeThePatterns(t *testing.T) {
	assert.Equal(t, "PLACE_TRACK x y", GetExpected("PLACE_TRACKS 1"))
	assert.Equal(t,
		"AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT",
		GetExpected("place_tracks 1"),
	)
}

// String.split drops trailing empty fields, so a trailing separator is not an
// empty command and does not disqualify.
func TestParseIgnoresATrailingSemicolon(t *testing.T) {
	_, player := parse(t, "WAIT;")

	assert.False(t, player.IsDeactivated())
	assert.Len(t, player.Intents, 1)
}

// Parsing stops at the first bad command, so intents queued before it are
// irrelevant — the player is out either way.
func TestParseStopsAtTheFirstUnparseableCommand(t *testing.T) {
	_, player := parse(t, "WAIT;NONSENSE;WAIT")

	assert.True(t, player.IsDeactivated())
	assert.Equal(t,
		"Invalid Input: Expected AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT but got 'NONSENSE'",
		player.DeactivationReason(),
	)
}
