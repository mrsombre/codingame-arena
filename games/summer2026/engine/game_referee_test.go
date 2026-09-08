package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// arenaPlayers lifts the engine players into the arena interface the optional
// capability hooks are called with.
func arenaPlayers(game *Game) []arena.Player {
	players := make([]arena.Player, len(game.Players))
	for i, p := range game.Players {
		players[i] = p
	}
	return players
}

// ——— RawScoresProvider ————————————————————————————————————————————————————

func TestRawScoresReportLivePointsBeforeOnEnd(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{Terrain: []string{"."}})
	referee := NewReferee(game)

	p0.SetScore(7)
	p1.SetScore(3)

	assert.Equal(t, [2]int{7, 3}, referee.RawScores())
}

func TestRawScoresKeepIntrinsicPointsOfDeactivatedPlayer(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{Terrain: []string{"."}})
	referee := NewReferee(game)

	p0.SetScore(7)
	p1.SetScore(3)
	p1.Deactivate("Invalid action.")
	referee.OnEnd()

	require.Equal(t, -1, p1.GetScore(), "OnEnd overwrites a deactivated score")
	assert.Equal(t, [2]int{7, 3}, referee.RawScores())
}

func TestRawScoresKeepIntrinsicPointsThroughTutorialEnd(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{Terrain: []string{"."}, League: 1})
	referee := NewReferee(game)

	p0.SetScore(5)
	p1.SetScore(2)
	referee.OnEnd()

	assert.Equal(t, [2]int{5, 2}, referee.RawScores())
}

// ——— EndReasonProvider ————————————————————————————————————————————————————

func TestEndReasonMapsEveryEndPath(t *testing.T) {
	tests := []struct {
		name              string
		deactivate        int
		reason            string
		timedOut          bool
		deactivationTurns [2]int
		firstOutputTurns  [2]int
		want              string
	}{
		{
			name:              "timeout on the first prompt",
			deactivate:        0,
			reason:            "Timeout!",
			deactivationTurns: [2]int{1, -1},
			firstOutputTurns:  [2]int{1, 1},
			want:              arena.EndReasonTimeoutStart,
		},
		{
			name:              "timeout later in the match",
			deactivate:        1,
			reason:            "Timeout!",
			deactivationTurns: [2]int{-1, 12},
			firstOutputTurns:  [2]int{1, 1},
			want:              arena.EndReasonTimeout,
		},
		{
			name:              "arena hard timeout, whose message is not Java's",
			deactivate:        1,
			reason:            "external player timed out after 50ms on turn 12 (bot)",
			timedOut:          true,
			deactivationTurns: [2]int{-1, 12},
			firstOutputTurns:  [2]int{1, 1},
			want:              arena.EndReasonTimeout,
		},
		{
			name:              "unparseable output",
			deactivate:        0,
			reason:            "Invalid action. Expected AUTOPLACE x1 y1 x2 y2",
			deactivationTurns: [2]int{4, -1},
			firstOutputTurns:  [2]int{1, 1},
			want:              arena.EndReasonInvalid,
		},
		{
			name:              "no deactivation",
			deactivate:        -1,
			deactivationTurns: [2]int{-1, -1},
			firstOutputTurns:  [2]int{1, 1},
			want:              arena.EndReasonScore,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game, _, _ := loadScenario(t, scenario{Terrain: []string{"."}})
			referee := NewReferee(game)
			if tc.deactivate >= 0 {
				game.Players[tc.deactivate].Deactivate(tc.reason)
				game.Players[tc.deactivate].SetTimedOut(tc.timedOut)
			}

			got := referee.EndReason(MAX_TURNS, arenaPlayers(game), tc.deactivationTurns, tc.firstOutputTurns)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Both normal finishes are SCORE: reaching the turn cap and running out of
// reachable connections are equally ordinary endings here.
func TestEndReasonIsScoreForBothNormalFinishes(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..", ".."},
		Towns:   []townSpec{{ID: 0, X: 0, Y: 0, Desires: []int{1}}, {ID: 1, X: 1, Y: 1}},
	})
	referee := NewReferee(game)
	players := arenaPlayers(game)

	assert.Equal(t, arena.EndReasonScore, referee.EndReason(MAX_TURNS, players, [2]int{-1, -1}, [2]int{1, 1}))
	assert.Equal(t, arena.EndReasonScore, referee.EndReason(7, players, [2]int{-1, -1}, [2]int{1, 1}))
}

// ——— MetricsProvider ——————————————————————————————————————————————————————

func TestMetricsExposeTheJavaMetadataCounters(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{Terrain: []string{"."}})
	referee := NewReferee(game)

	game.PlacedTracks = [2]int{11, 21}
	game.TracksPlacedOnPlains = [2]int{12, 22}
	game.TracksPlacedOnRiver = [2]int{13, 23}
	game.TracksPlacedOnMountains = [2]int{14, 24}
	game.ZonesInked = [2]int{15, 25}
	game.OwnTracksInkedOut = [2]int{16, 26}
	game.EnemyTracksInkedOut = [2]int{17, 27}
	game.ExtraTilesInConnection = [2]int{18, 28}
	game.AutobuildCalled = [2]int{19, 29}
	game.SideQuestPoints = [2]int{1, 0}
	game.TrackOwnershipPercentagePerActiveConnection = [2]float32{1.5, 0}
	game.TrackOwnershipPercentagePerActiveConnectionTotal = [2]int{2, 0}

	got := map[string]float64{}
	for _, metric := range referee.Metrics() {
		got[metric.Label] = metric.Value
	}

	want := map[string]float64{
		"tracksPlaced_0":            11,
		"tracksPlacedOnPlains_0":    12,
		"tracksPlacedOnRiver_0":     13,
		"tracksPlacedOnMountains_0": 14,
		"zonesInked_0":              15,
		"ownTracksInkedOut_0":       16,
		"enemyTracksInkedOut_0":     17,
		"extraTilesInConnection_0":  18,
		"autobuildCalled_0":         19,
		"sideQuestPoints_0":         1,
		"averageTrackOwnershipPercentagePerActiveConnection_0": 0.75,

		"tracksPlaced_1":            21,
		"tracksPlacedOnPlains_1":    22,
		"tracksPlacedOnRiver_1":     23,
		"tracksPlacedOnMountains_1": 24,
		"zonesInked_1":              25,
		"ownTracksInkedOut_1":       26,
		"enemyTracksInkedOut_1":     27,
		"extraTilesInConnection_1":  28,
		"autobuildCalled_1":         29,
		"sideQuestPoints_1":         0,
		// No scoring event, so the average is 0 rather than a division by zero.
		"averageTrackOwnershipPercentagePerActiveConnection_1": 0,
	}

	assert.Equal(t, want, got)
	assert.Len(t, referee.Metrics(), 22, "eleven counters per player")
}

// ——— capability wiring ————————————————————————————————————————————————————

func TestRefereeImplementsOptionalArenaCapabilities(t *testing.T) {
	var referee any = NewReferee(NewGame(nil, DEFAULT_LEAGUE))

	assert.Implements(t, (*arena.RawScoresProvider)(nil), referee)
	assert.Implements(t, (*arena.EndReasonProvider)(nil), referee)
	assert.Implements(t, (*arena.MetricsProvider)(nil), referee)
}
