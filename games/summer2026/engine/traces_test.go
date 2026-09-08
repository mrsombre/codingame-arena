package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// traceTypes lists the event discriminators of one player's turn bucket, in
// emission order.
func traceTypes(traces []arena.TurnTrace) []string {
	types := make([]string, 0, len(traces))
	for _, t := range traces {
		types = append(types, t.Type)
	}
	return types
}

// traceData decodes the single event of the given type in a turn bucket.
func traceData[T any](t *testing.T, traces []arena.TurnTrace, typ string) T {
	t.Helper()
	var found []T
	for _, tr := range traces {
		if tr.Type != typ {
			continue
		}
		data, err := arena.DecodeData[T](tr)
		require.NoError(t, err)
		found = append(found, data)
	}
	require.Lenf(t, found, 1, "expected exactly one %s event", typ)
	return found[0]
}

// A two-turn scenario driven end to end: player 0 lays a three-cell line
// joining the two towns and is paid for both directions of the connection;
// player 1 answers with a disruption that tips the middle region over its
// instability threshold, inking player 0's rails out and ending the match
// because no route between the towns remains.
func TestTracesRecordATurnOfPlayAndTheTurnThatEndsIt(t *testing.T) {
	game, p0, p1 := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Regions: []string{"abbba"},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0, Desires: []int{0}},
		},
	})

	// ——— turn 1: three rails, both directions of the connection pay out ———

	runTurn(game, "PLACE_TRACK 1 0;PLACE_TRACK 2 0;PLACE_TRACK 3 0", "MESSAGE hello")
	traces := game.TurnTraces()

	assert.Equal(t,
		[]string{TraceTrack, TraceTrack, TraceTrack, TraceScore, TraceScore, TraceTurnSummary},
		traceTypes(traces[0]),
	)
	assert.Equal(t, []string{TraceMessage, TraceTurnSummary}, traceTypes(traces[1]))

	first, err := arena.DecodeData[TrackData](traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, TrackData{Cell: [2]int{1, 0}, Cost: 1, Terrain: "PLAINS"}, first)

	score, err := arena.DecodeData[ScoreData](traces[0][3])
	require.NoError(t, err)
	assert.Equal(t, ScoreData{From: 0, To: 1, Points: 3, PathLength: 5, Detour: 1}, score)

	assert.Equal(t, TurnSummaryData{
		Turn:              1,
		PaintAvailable:    PASSIVE_INCOME,
		PaintSpent:        3,
		PaintLeft:         0,
		TracksPlaced:      3,
		TracksAutoplaced:  0,
		DisruptAvailable:  BLOT_POINTS_PER_TURN,
		DisruptSpent:      0,
		ConnectionsActive: 2,
		PointsEarned:      6,
		Score:             6,
	}, traceData[TurnSummaryData](t, traces[0], TraceTurnSummary))

	assert.Equal(t, "hello", traceData[MessageData](t, traces[1], TraceMessage).Text)
	assert.Equal(t, TurnSummaryData{
		Turn:             1,
		PaintAvailable:   PASSIVE_INCOME,
		PaintLeft:        PASSIVE_INCOME,
		DisruptAvailable: BLOT_POINTS_PER_TURN,
	}, traceData[TurnSummaryData](t, traces[1], TraceTurnSummary))

	assert.Equal(t, 6, p0.GetScore())
	assert.Equal(t, 0, p1.GetScore())

	// ——— turn 2: the region inks out, the match ends ———

	// One short of the threshold, so player 1's single blot this turn tips it.
	game.Grid.Zones[1].Instability = INSTABILITY_THRESHOLD_BASE - 1

	runTurn(game, "PLACE_TRACK 1 0", "DISRUPT 1")
	traces = game.TurnTraces()

	assert.Equal(t, []string{TraceFailed, TraceInk, TraceTurnSummary}, traceTypes(traces[0]))
	assert.Equal(t, []string{TraceDisrupt, TraceInk, TraceTurnSummary}, traceTypes(traces[1]))

	assert.Equal(t,
		"Cannot place tracks on existing tracks at (1, 0)",
		traceData[FailedData](t, traces[0], TraceFailed).Reason,
	)
	assert.Equal(t,
		DisruptData{Zone: 1, Instability: INSTABILITY_THRESHOLD_BASE},
		traceData[DisruptData](t, traces[1], TraceDisrupt),
	)

	// The ink event is cross-owner, so both buckets carry the same payload.
	ink := InkData{Zone: 1, Credited: []int{1}, TracksLost: [3]int{3, 0, 0}}
	assert.Equal(t, ink, traceData[InkData](t, traces[0], TraceInk))
	assert.Equal(t, ink, traceData[InkData](t, traces[1], TraceInk))

	// Rails are gone, so nobody is paid this turn and player 0's score stands.
	assert.Equal(t, TurnSummaryData{
		Turn:             2,
		PaintAvailable:   PASSIVE_INCOME,
		PaintLeft:        PASSIVE_INCOME,
		DisruptAvailable: BLOT_POINTS_PER_TURN,
		Score:            6,
	}, traceData[TurnSummaryData](t, traces[0], TraceTurnSummary))
	assert.Equal(t, TurnSummaryData{
		Turn:             2,
		PaintAvailable:   PASSIVE_INCOME,
		PaintLeft:        PASSIVE_INCOME,
		DisruptAvailable: BLOT_POINTS_PER_TURN,
		DisruptSpent:     1,
	}, traceData[TurnSummaryData](t, traces[1], TraceTurnSummary))

	assert.True(t, game.Ended(), "no route between the towns survives the inking")
}

// Autoplace expands into one placement event per rail it lays, and the
// expansion itself is recorded so an analyzer can tell an autoplaced rail
// from a hand-placed one.
func TestTracesRecordAutoplaceExpansion(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"....."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0, Desires: []int{1}},
			{ID: 1, X: 4, Y: 0, Desires: []int{0}},
		},
	})

	runTurn(game, "AUTOPLACE 1 0 3 0", "WAIT")
	traces := game.TurnTraces()

	assert.Equal(t,
		[]string{TraceAutoplace, TraceTrack, TraceTrack, TraceTrack, TraceScore, TraceScore, TraceTurnSummary},
		traceTypes(traces[0]),
	)
	assert.Equal(t,
		AutoplaceData{From: [2]int{1, 0}, To: [2]int{3, 0}, Placements: 3},
		traceData[AutoplaceData](t, traces[0], TraceAutoplace),
	)

	auto, err := arena.DecodeData[TrackData](traces[0][1])
	require.NoError(t, err)
	assert.True(t, auto.Autoplaced)

	summary := traceData[TurnSummaryData](t, traces[0], TraceTurnSummary)
	assert.Equal(t, 3, summary.TracksPlaced)
	assert.Equal(t, 3, summary.TracksAutoplaced)
}

// A cell both players claim on the same turn resolves to neutral track. The
// event belongs to neither side, so it is mirrored into both buckets.
func TestTracesRecordContestedPlacement(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0},
			{ID: 1, X: 2, Y: 0},
		},
	})

	runTurn(game, "PLACE_TRACK 1 0", "PLACE_TRACK 1 0")
	traces := game.TurnTraces()

	contested := ContestedData{Cell: [2]int{1, 0}}
	assert.Equal(t, contested, traceData[ContestedData](t, traces[0], TraceContested))
	assert.Equal(t, contested, traceData[ContestedData](t, traces[1], TraceContested))
	assert.Equal(t, TRACK_NEUTRAL, game.Grid.GetXY(1, 0).Track)
}

// The buffer is per turn: a turn's events never leak into the next one.
func TestTurnTracesAreClearedEachTurn(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0},
			{ID: 1, X: 2, Y: 0},
		},
	})

	runTurn(game, "PLACE_TRACK 1 0", "WAIT")
	require.Equal(t, []string{TraceTrack, TraceTurnSummary}, traceTypes(game.TurnTraces()[0]))

	runTurn(game, "WAIT", "WAIT")
	assert.Equal(t, []string{TraceTurnSummary}, traceTypes(game.TurnTraces()[0]))
}

// The referee hands the arena runner a copy, so later turns cannot mutate a
// trace the runner has already taken.
func TestRefereeTurnTracesAreIndependentCopies(t *testing.T) {
	game, _, _ := loadScenario(t, scenario{
		Terrain: []string{"..."},
		Towns: []townSpec{
			{ID: 0, X: 0, Y: 0},
			{ID: 1, X: 2, Y: 0},
		},
	})
	referee := NewReferee(game)

	runTurn(game, "PLACE_TRACK 1 0", "WAIT")
	taken := referee.TurnTraces(1, nil)
	require.Len(t, taken[0], 2)

	runTurn(game, "WAIT", "WAIT")
	assert.Len(t, taken[0], 2)
}
