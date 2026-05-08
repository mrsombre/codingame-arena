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

// giveSun emits one GATHER event per tree in TreeOrder, regardless of whether
// the tree harvested. Spooky-shadowed trees and seeds (size 0) emit Sun=0;
// non-shadowed trees emit Sun=size. This lets analyzers see shadow outcomes
// without recomputing them from sunDirection + trees.
func TestTraceGatherEmitsZeroForShadowedAndSeedTrees(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	g.Sun.SetOrientation(0)
	g.placeTree(p0, 0, TREE_TALL)   // casts size-3 shadow on 1, 7, 19
	g.placeTree(p0, 4, TREE_SMALL)  // not shadowed; shadows 0 with size 1 (1 < 3, not spooky)
	g.placeTree(p0, 7, TREE_SMALL)  // spooky-shadowed by cell 0 (3 >= 1)
	g.placeTree(p0, 19, TREE_TALL)  // spooky-shadowed by cell 0 (3 >= 3)
	g.placeTree(p0, 22, TREE_SEED)  // size 0 — never harvests
	g.calculateShadows()

	g.giveSun()

	want := []GatherData{
		{Cell: 0, Sun: TREE_TALL},   // shadowed by size 1 → not spooky → harvests
		{Cell: 4, Sun: TREE_SMALL},  // unshadowed → harvests
		{Cell: 7, Sun: 0},           // spooky → 0
		{Cell: 19, Sun: 0},          // spooky → 0
		{Cell: 22, Sun: 0},          // seed → 0
	}
	require.Len(t, g.traces[0], len(want), "one GATHER per tree in TreeOrder")
	for i, ev := range g.traces[0] {
		assert.Equal(t, TraceGather, ev.Type, "event %d type", i)
		data, err := arena.DecodeData[GatherData](ev)
		require.NoError(t, err)
		assert.Equal(t, want[i], data, "event %d", i)
	}
	assert.Empty(t, g.traces[1], "owner-bucketed: p1 has no trees")
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

// Action events report the sun cost actually debited so analyzers don't have
// to recompute it from sun deltas. Costs follow the rules: GROW = base + count
// of player-owned trees at the target size; SEED = number of size-0 trees the
// player owns; COMPLETE = LIFECYCLE_END_COST (4).
func TestTraceActionsCarryCost(t *testing.T) {
	t.Run("grow base cost only", func(t *testing.T) {
		g := newScenario(4)
		p0 := g.Players[0]
		p0.Sun = 10
		g.placeTree(p0, 0, TREE_SMALL) // grow target → size 2; no other size-2 trees

		runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

		data, err := arena.DecodeData[GrowData](g.traces[0][0])
		require.NoError(t, err)
		assert.Equal(t, 0, data.Cell)
		assert.Equal(t, 3, data.Cost, "TREE_BASE_COST[2] = 3, no same-size trees")
	})

	t.Run("grow cost includes same-size penalty", func(t *testing.T) {
		g := newScenario(4)
		p0 := g.Players[0]
		p0.Sun = 10
		g.placeTree(p0, 0, TREE_SMALL)   // target → size 2 (action subject)
		g.placeTree(p0, 6, TREE_MEDIUM)  // already size 2 (own, +1)
		g.placeTree(p0, 12, TREE_MEDIUM) // already size 2 (own, +1)
		// Opponent same-size trees do NOT count.
		g.placeTree(g.Players[1], 18, TREE_MEDIUM)

		runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

		data, err := arena.DecodeData[GrowData](g.traces[0][0])
		require.NoError(t, err)
		assert.Equal(t, 3+2, data.Cost, "base 3 + 2 own size-2 trees")
	})

	t.Run("seed cost equals owned seed count", func(t *testing.T) {
		g := newScenario(4)
		p0 := g.Players[0]
		p0.Sun = 10
		g.placeTree(p0, 19, TREE_TALL) // source
		// Two seeds the player already owns ⇒ next seed costs 2.
		g.placeTree(p0, 12, TREE_SEED)
		g.placeTree(p0, 13, TREE_SEED)
		// Opponent seeds don't count.
		g.placeTree(g.Players[1], 14, TREE_SEED)

		runActionTurn(g, func() { p0.SetAction(NewSeedAction(19, 7)) })

		data, err := arena.DecodeData[SeedData](g.traces[0][0])
		require.NoError(t, err)
		assert.Equal(t, 19, data.Source)
		assert.Equal(t, 7, data.Target)
		assert.Equal(t, 2, data.Cost, "two own seeds at action time")
	})

	t.Run("complete cost is LIFECYCLE_END_COST", func(t *testing.T) {
		g := newScenario(4)
		p0 := g.Players[0]
		p0.Sun = LIFECYCLE_END_COST
		g.placeTree(p0, 19, TREE_TALL)

		runActionTurn(g, func() { p0.SetAction(NewCompleteAction(19)) })

		var complete arena.TurnTrace
		for _, ev := range g.traces[0] {
			if ev.Type == TraceComplete {
				complete = ev
			}
		}
		require.Equal(t, TraceComplete, complete.Type, "expected COMPLETE event")
		data, err := arena.DecodeData[CompleteData](complete)
		require.NoError(t, err)
		assert.Equal(t, LIFECYCLE_END_COST, data.Cost)
	})
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

// CommandManager parses "WAIT GL HF" as action=WAIT plus message="GL HF" and
// stamps it on the player. The trailing chat text rides on the action event
// itself as `data.debug` (no separate DEBUG event); empty messages omit the
// field entirely so the WAIT trace stays bare-`{type:"WAIT"}` shape.
func TestTraceWaitCarriesInlineDebugMessage(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p1 := g.Players[1]

	runActionTurn(g, func() {
		p0.SetAction(NewWaitAction())
		p0.SetMessage("GL HF")
		p1.SetAction(NewWaitAction())
	})

	require.Equal(t, []string{TraceWait}, traceTypes(g.traces[0]), "no separate DEBUG event")
	require.Equal(t, []string{TraceWait}, traceTypes(g.traces[1]))

	// p0 carries the message inline.
	data, err := arena.DecodeData[WaitData](g.traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, "GL HF", data.Debug)

	// p1 had no message — emit a bare TurnTrace with no `data` field at all.
	assert.Empty(t, g.traces[1][0].Data, "WAIT without message must not carry a data payload")
}

// On a `COMPLETE <cell> <message>` output, the message rides on the COMPLETE
// event's `debug` field rather than via a trailing DEBUG event. This was the
// case that previously broke ordering (DEBUG arrived before COMPLETE because
// COMPLETE was emitted late in removeDyingTrees). The redesign removes the
// ordering question entirely.
func TestTraceCompleteCarriesInlineDebugMessage(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = LIFECYCLE_END_COST
	g.placeTree(p0, 19, TREE_TALL)

	runActionTurn(g, func() {
		p0.SetAction(NewCompleteAction(19))
		p0.SetMessage("complete tree")
	})

	require.Equal(t, []string{TraceComplete}, traceTypes(g.traces[0]),
		"COMPLETE must be the only event in p0's slot — debug rides inline")
	data, err := arena.DecodeData[CompleteData](g.traces[0][0])
	require.NoError(t, err)
	assert.Equal(t, "complete tree", data.Debug)
}

// On-disk JSON keys for the per-turn state payload must be camelCase to match
// the cross-game trace envelope (commit 46a18fd2 standardized this). The Go
// field names are camelCase already, but the json tags are easy to drift —
// pin them here so a snake_case regression fails this test.
func TestDecorateTraceTurnUsesCamelCaseJSONKeys(t *testing.T) {
	g := newScenario(4)
	g.CurrentFrameType = FrameActions
	g.DayActionIndex = 1
	cell := 7
	g.seedConflictCell = &cell

	raw := g.DecorateTraceTurn(0, nil)

	var keys map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &keys))
	for _, k := range []string{"day", "phase", "dayActionIndex", "sunDirection", "nutrients", "sun", "trees", "seedConflictCell"} {
		_, ok := keys[k]
		assert.Truef(t, ok, "expected camelCase key %q in state payload, got %v", k, keysOf(keys))
	}
	for _, k := range []string{"day_action_index", "sun_direction", "seed_conflict_cell"} {
		_, ok := keys[k]
		assert.Falsef(t, ok, "snake_case key %q must not appear in state payload", k)
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
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
	g.Nutrients = 17
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
	require.NotNil(t, state.Nutrients)
	assert.Equal(t, 17, *state.Nutrients)
	assert.Equal(t, []int{8, 6}, state.Sun)
	require.NotNil(t, state.DayActionIndex)
	assert.Equal(t, 2, *state.DayActionIndex)
	assert.Equal(t, [][][4]int{
		{{1, RICHNESS_LUSH, TREE_SEED, 0}, {24, RICHNESS_POOR, TREE_SMALL, 0}},
		{{7, RICHNESS_OK, TREE_MEDIUM, 0}, {19, RICHNESS_POOR, TREE_TALL, 0}},
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
// frame and the orientation reflects on the next trace turn's sunDirection.
// (The SUN_MOVE event itself is no longer emitted; phase=sun + next-turn
// sunDirection are the consumer signals.)
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
