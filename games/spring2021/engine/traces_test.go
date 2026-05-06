package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// decodeState calls DecorateTraceTurn and decodes the returned bytes into
// TraceTurnState. Tests use this to assert on the typed payload.
func decodeState(t *testing.T, g *Game) TraceTurnState {
	t.Helper()
	raw := g.DecorateTraceTurn(0, nil)
	require.NotEmpty(t, raw, "DecorateTraceTurn returned empty payload")
	var state TraceTurnState
	require.NoError(t, json.Unmarshal(raw, &state))
	return state
}

func traceTypes(traces []arena.TurnTrace) []string {
	out := make([]string, len(traces))
	for i, t := range traces {
		out[i] = t.Type
	}
	return out
}

func TestTraceGrowEmitsOnSuccessfulAction(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 5
	g.placeTree(p0, 0, TREE_SMALL)

	runActionTurn(g, func() {
		p0.SetAction(NewGrowAction(0))
	})

	// p0 grows in slot 0; p1 has no action set so defaults to WAIT in slot 1.
	assert.Contains(t, traceTypes(g.traces[0]), TraceGrow)
	assert.Contains(t, traceTypes(g.traces[1]), TraceWait)
}

func TestTraceSeedEmitsOnSuccessfulAction(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 5
	g.placeTree(p0, 19, TREE_TALL)

	runActionTurn(g, func() {
		p0.SetAction(NewSeedAction(19, 7))
	})

	assert.Contains(t, traceTypes(g.traces[0]), TraceSeed)
}

func TestTraceCompleteEmitsWithPoints(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = LIFECYCLE_END_COST
	cell := g.Board.Cells[19]
	g.placeTree(p0, cell.GetIndex(), TREE_TALL)
	startScore := p0.GetScore()

	runActionTurn(g, func() {
		p0.SetAction(NewCompleteAction(cell.GetIndex()))
	})

	assert.Contains(t, traceTypes(g.traces[0]), TraceComplete)
	assert.Greater(t, p0.GetScore(), startScore, "completion should award points")
}

func TestTraceWaitEmitsWhenPlayerWaits(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]

	runActionTurn(g, func() {
		p0.SetAction(NewWaitAction())
	})

	assert.Contains(t, traceTypes(g.traces[0]), TraceWait)
	assert.True(t, p0.IsWaiting())
}

func TestDecorateTraceTurnAddsDecisionTraces(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p1 := g.Players[1]
	p0.Sun = 8
	p1.Sun = 6
	p0.SetScore(3)
	p1.SetScore(5)
	g.Round = 4
	g.Sun.SetOrientation(g.Round)
	g.DayActionIndex = 2
	g.CurrentFrameType = FrameActions
	g.NextFrameType = FrameActions
	g.placeTree(p0, 1, TREE_SEED)
	g.placeTree(p0, 24, TREE_SMALL)
	g.placeTree(p1, 7, TREE_MEDIUM)
	g.placeTree(p1, 19, TREE_TALL)
	p0.SetAction(NewGrowAction(24))
	p1.SetAction(NewSeedAction(7, 8))

	state := decodeState(t, g)

	require.NotNil(t, state.Day)
	assert.Equal(t, 4, *state.Day)
	assert.Equal(t, "actions", state.Phase)
	require.NotNil(t, state.SunDirection)
	assert.Equal(t, 4, *state.SunDirection)
	assert.Equal(t, []int{8, 6}, state.Sun)
	require.NotNil(t, state.DayActionIndex)
	assert.Equal(t, 2, *state.DayActionIndex)
	assert.Equal(t, [][][3]int{
		{{1, TREE_SEED, RICHNESS_LUSH}, {24, TREE_SMALL, RICHNESS_POOR}},
		{{7, TREE_MEDIUM, RICHNESS_OK}, {19, TREE_TALL, RICHNESS_POOR}},
	}, state.Trees)
}

func TestRefereeTurnTracesReturnsCopy(t *testing.T) {
	g := newScenario(4)
	r := NewReferee(g)
	g.traces[0] = append(g.traces[0], arena.TurnTrace{Type: TraceWait})

	out := r.TurnTraces(0, []arena.Player{g.Players[0], g.Players[1]})
	require.Len(t, out[0], 1)
	assert.Empty(t, out[1])
	assert.Equal(t, TraceWait, out[0][0].Type)

	// Mutating the engine slice must not bleed into the returned copy.
	g.traces[0] = g.traces[0][:0]
	assert.Len(t, out[0], 1, "TurnTraces returns an independent copy")
}

// PerformGameUpdate resets g.traces at the start; verify the full
// public API behaves as the runner expects (drained per turn, not stale).
func TestPerformGameUpdateResetsTraces(t *testing.T) {
	g := newScenario(4)
	g.traces[0] = append(g.traces[0], arena.TurnTrace{Type: TraceWait})

	g.PerformGameUpdate(0)

	for _, slot := range g.traces {
		for _, e := range slot {
			assert.NotEqual(t, TraceWait, e.Type, "stale traces must be cleared at the top of PerformGameUpdate")
		}
	}
}

// DecorateTraceTurn stamps the current frame type onto the State payload
// so consumers can identify gather/action/sun frames without inspecting
// traces.
func TestDecorateTraceTurnStampsPhase(t *testing.T) {
	g := newScenario(4)

	cases := []struct {
		frame FrameType
		want  string
	}{
		{FrameGathering, "gathering"},
		{FrameActions, "actions"},
		{FrameSunMove, "sun"},
	}
	for _, tc := range cases {
		g.CurrentFrameType = tc.frame
		state := decodeState(t, g)
		assert.Equal(t, tc.want, state.Phase, "frame %v", tc.frame)
	}
}

// TestSunMoveAdvancesOrientation verifies the sun rotates after a SUN_MOVE
// frame and the orientation reflects on the next trace turn's sun_direction.
// (The SUN_MOVE event itself is no longer emitted; phase=sun + next-turn
// sun_direction are the consumer signals.)
func TestSunMoveAdvancesOrientation(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameSunMove
	g.NextFrameType = FrameGathering
	g.Round = 0
	g.Sun.SetOrientation(0)

	g.PerformGameUpdate(0)

	assert.Equal(t, 1, g.Sun.Orientation, "sun.move() advances 0 -> 1")
	assert.Equal(t, [2][]arena.TurnTrace{}, g.traces, "sun-move frame emits no per-player events")
}

// On the final SUN_MOVE the engine skips Sun.Move() so the orientation
// stays put; phase=sun is still stamped at decorate time.
func TestSunMoveSkippedOnFinalRound(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameSunMove
	g.NextFrameType = FrameGathering
	g.Round = g.MAX_ROUNDS - 1
	priorOrientation := g.Sun.Orientation

	g.PerformGameUpdate(0)

	assert.Equal(t, priorOrientation, g.Sun.Orientation, "sun must not move past the final round")
}

// SEED_CONFLICT info is preserved in the State payload rather than as a
// per-player event.
func TestSeedConflictPromotedToTurnRoot(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameActions
	g.NextFrameType = FrameActions
	p0 := g.Players[0]
	p1 := g.Players[1]
	p0.Sun = 10
	p1.Sun = 10
	g.placeTree(p0, 0, TREE_TALL)
	g.placeTree(p1, 18, TREE_TALL)
	target := 5
	p0.SetAction(NewSeedAction(0, target))
	p1.SetAction(NewSeedAction(18, target))

	g.PerformGameUpdate(0)
	require.NotNil(t, g.seedConflictCell)
	assert.Equal(t, target, *g.seedConflictCell)

	state := decodeState(t, g)
	require.NotNil(t, state.SeedConflictCell)
	assert.Equal(t, target, *state.SeedConflictCell)
}
