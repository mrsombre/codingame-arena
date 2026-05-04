package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

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

	// p0 grows, p1 has no action set so defaults to WAIT.
	types := traceTypes(g.traces)
	assert.Contains(t, types, TraceGrow)
	assert.Contains(t, types, TraceWait)
}

func TestTraceSeedEmitsOnSuccessfulAction(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 5
	g.placeTree(p0, 19, TREE_TALL)

	runActionTurn(g, func() {
		p0.SetAction(NewSeedAction(19, 7))
	})

	assert.Contains(t, traceTypes(g.traces), TraceSeed)
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

	assert.Contains(t, traceTypes(g.traces), TraceComplete)
	assert.Greater(t, p0.GetScore(), startScore, "completion should award points")
}

func TestTraceWaitEmitsWhenPlayerWaits(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]

	runActionTurn(g, func() {
		p0.SetAction(NewWaitAction())
	})

	assert.Contains(t, traceTypes(g.traces), TraceWait)
	assert.True(t, p0.IsWaiting())
}

func TestRefereeTurnTracesReturnsCopy(t *testing.T) {
	g := newScenario(4)
	r := NewReferee(g)
	g.traces = append(g.traces, arena.MakeTurnTrace(TraceWait, PlayerMeta{Player: 0}))

	out := r.TurnTraces(0, []arena.Player{g.Players[0], g.Players[1]})
	require.Len(t, out, 1)
	assert.Equal(t, TraceWait, out[0].Type)

	// Mutating the engine slice must not bleed into the returned copy.
	g.traces = g.traces[:0]
	assert.Len(t, out, 1, "TurnTraces returns an independent copy")
}

// PerformGameUpdate resets g.traces[:0] at the start; verify the full
// public API behaves as the runner expects (drained per turn, not stale).
func TestPerformGameUpdateResetsTraces(t *testing.T) {
	g := newScenario(4)
	g.traces = append(g.traces, arena.MakeTurnTrace(TraceWait, PlayerMeta{Player: 0}))

	g.PerformGameUpdate(0)

	for _, e := range g.traces {
		assert.NotEqual(t, TraceWait, e.Type, "stale traces must be cleared at the top of PerformGameUpdate")
	}
}

// findTraceMeta locates the first trace of the given type and decodes its
// meta into T. Fails the test if no matching trace is present.
func findTraceMeta[T any](t *testing.T, traces []arena.TurnTrace, typ string) T {
	t.Helper()
	for _, tr := range traces {
		if tr.Type != typ {
			continue
		}
		v, err := arena.DecodeMeta[T](tr)
		require.NoError(t, err)
		return v
	}
	require.Failf(t, "trace not found", "expected trace type %q in %v", typ, traceTypes(traces))
	var zero T
	return zero
}

// Phase markers fire even when no per-player action does — guarantees a
// non-empty trace turn for every CG phase frame.

func TestTraceGatherPhaseEmittedOnGatheringFrame(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameGathering
	g.NextFrameType = FrameActions
	g.Round = 3

	g.PerformGameUpdate(0)

	meta := findTraceMeta[GatherPhaseMeta](t, g.traces, TraceGatherPhase)
	assert.Equal(t, 3, meta.Round)
}

func TestTraceSunMoveEmittedWithNewDirection(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameSunMove
	g.NextFrameType = FrameGathering
	g.Round = 0
	g.Sun.SetOrientation(0)

	g.PerformGameUpdate(0)

	meta := findTraceMeta[SunMoveMeta](t, g.traces, TraceSunMove)
	assert.Equal(t, 0, meta.Round, "round that just ended")
	assert.Equal(t, 1, meta.Direction, "sun.move() advances 0 -> 1")
}

// On the final SUN_MOVE the engine skips Sun.Move() and the game ends; the
// trace must signal "no further direction" rather than report the stale value.
func TestTraceSunMoveDirectionMinusOneOnFinalRound(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameSunMove
	g.NextFrameType = FrameGathering
	g.Round = g.MAX_ROUNDS - 1
	priorOrientation := g.Sun.Orientation

	g.PerformGameUpdate(0)

	meta := findTraceMeta[SunMoveMeta](t, g.traces, TraceSunMove)
	assert.Equal(t, g.MAX_ROUNDS-1, meta.Round)
	assert.Equal(t, -1, meta.Direction)
	assert.Equal(t, priorOrientation, g.Sun.Orientation, "sun must not move past the final round")
}
