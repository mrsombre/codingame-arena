package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func findTrace(traces []arena.TurnTrace, label string) (arena.TurnTrace, bool) {
	for _, tr := range traces {
		if tr.Label == label {
			return tr, true
		}
	}
	return arena.TurnTrace{}, false
}

func countTraces(traces []arena.TurnTrace, label string) int {
	n := 0
	for _, tr := range traces {
		if tr.Label == label {
			n++
		}
	}
	return n
}

func TestTraceEatPelletValueOne(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})

	tr, ok := findTrace(g.traces, TraceEat)
	require.True(t, ok, "EAT trace emitted")
	assert.Equal(t, "0 2,1 1", tr.Payload)
}

func TestTraceEatSuperPelletValueTen(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	g.Grid.Get(Coord{X: 2, Y: 1}).HasCherry = true
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 3, Y: 1})
	})

	tr, ok := findTrace(g.traces, TraceEat)
	require.True(t, ok, "EAT trace emitted")
	assert.Equal(t, "0 2,1 10", tr.Payload)
}

func TestTraceKilledOnRPSCombat(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	attacker := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	victim := spawn(g, 1, 0, TypeScissors, Coord{X: 3, Y: 1})

	runTurn(g, func() {
		attacker.Intent = NewMoveAction(Coord{X: 3, Y: 1})
		victim.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	})

	require.True(t, victim.Dead)
	tr, ok := findTrace(g.traces, TraceKilled)
	require.True(t, ok, "KILLED trace emitted")
	// Payload format: "<deadId> <x>,<y> <killerId>". Position is the victim's
	// position at the moment of the kill, whatever the movement resolution
	// left it at this turn.
	assert.Equal(t, traceKilledPayload(victim.ID, victim.Position, attacker.ID), tr.Payload)
}

func TestTraceSpeedAbility(t *testing.T) {
	g := newScenario(4, []string{"#######", "#     #", "#######"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewSpeedAction()
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	})

	tr, ok := findTrace(g.traces, TraceSpeed)
	require.True(t, ok, "SPEED trace emitted")
	assert.Equal(t, "0", tr.Payload)
}

func TestTraceSwitchAbility(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypePaper)
		pac.AbilityToUse = AbilitySetPaper
		pac.HasAbilityToUse = true
	})

	tr, ok := findTrace(g.traces, TraceSwitch)
	require.True(t, ok, "SWITCH trace emitted")
	assert.Equal(t, "0 PAPER", tr.Payload)
}

func TestTraceNoEmissionWhenAbilityBlockedByCooldown(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	// First switch fires and sets cooldown.
	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypePaper)
		pac.AbilityToUse = AbilitySetPaper
		pac.HasAbilityToUse = true
	})
	require.Equal(t, 1, countTraces(g.traces, TraceSwitch))

	// Second switch on the next turn is blocked by remaining cooldown.
	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypeScissors)
		pac.AbilityToUse = AbilitySetScissors
		pac.HasAbilityToUse = true
	})
	assert.Equal(t, 0, countTraces(g.traces, TraceSwitch), "blocked switch emits nothing")
}

func TestTracesClearedPerTurn(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})
	require.NotEmpty(t, g.traces, "first turn produced traces")

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})
	// One EAT for the new cell stepped onto, no leftover from the first turn.
	require.Len(t, g.traces, 1)
	assert.Equal(t, TraceEat, g.traces[0].Label)
	assert.Equal(t, "0 3,1 1", g.traces[0].Payload)
}

func TestTraceSpeedSubTurnAccumulatesEats(t *testing.T) {
	g := newScenario(4, []string{"#######", "#.....#", "#######"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	// Second player keeps the game alive so the speed sub-step is not skipped.
	spawn(g, 1, 0, TypeRock, Coord{X: 5, Y: 1})

	// Turn 1: activate SPEED.
	runTurn(g, func() {
		pac.Intent = NewSpeedAction()
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	})

	// Turn 2: MOVE — main step + speed sub-step both resolve in one turn.
	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 4, Y: 1})
	})

	assert.Equal(t, 2, countTraces(g.traces, TraceEat),
		"two pellets eaten in one main turn (main step + speed sub-step)")
}

func TestRefereeTurnTracesReturnsCopy(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	r := NewReferee(g)

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})
	require.NotEmpty(t, g.traces)

	asArena := []arena.Player{g.Players[0], g.Players[1]}
	out := r.TurnTraces(0, asArena)
	require.NotEmpty(t, out)
	out[0].Label = "MUTATED"
	assert.NotEqual(t, "MUTATED", g.traces[0].Label, "TurnTraces returned a copy")
}

func TestRefereeTurnTracesNilWhenEmpty(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	r := NewReferee(g)

	asArena := []arena.Player{g.Players[0], g.Players[1]}
	assert.Nil(t, r.TurnTraces(0, asArena), "no traces yet → nil")
}

func TestTraceSummaryAggregatesEatsPerPacAcrossTurns(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	p0 := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	p1 := spawn(g, 1, 0, TypeRock, Coord{X: 5, Y: 1})
	r := NewReferee(g)

	// Turn 0: p0 steps to (2,1) and eats; p1 stays on (5,1) and eats.
	r.ResetGameTurnData()
	p0.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	p1.Intent = NoAction
	r.PerformGameUpdate(0)

	// Turn 1: p0 steps to (3,1) and eats; p1's cell already empty.
	r.ResetGameTurnData()
	p0.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	p1.Intent = NoAction
	r.PerformGameUpdate(1)

	summary := r.TraceSummary()
	assert.Equal(t, [][]int{{0, 1}}, summary[0][TraceEat])
	assert.Equal(t, [][]int{{0}}, summary[1][TraceEat])
}

func TestTraceSummaryRecordsKilledUnderVictim(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	attacker := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	victim := spawn(g, 1, 0, TypeScissors, Coord{X: 3, Y: 1})
	r := NewReferee(g)

	r.ResetGameTurnData()
	attacker.Intent = NewMoveAction(Coord{X: 3, Y: 1})
	victim.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	r.PerformGameUpdate(7)

	summary := r.TraceSummary()
	// KILLED uses the dead pac as the subject → bucketed under p1, pac 0.
	assert.Equal(t, [][]int{{7}}, summary[1][TraceKilled])
	assert.Empty(t, summary[0][TraceKilled], "killer side has no KILLED entry")
}

func TestTraceSummaryIsEmptyWhenNoEvents(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	p0 := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	p1 := spawn(g, 1, 0, TypeRock, Coord{X: 3, Y: 1})
	r := NewReferee(g)

	r.ResetGameTurnData()
	p0.Intent = NoAction
	p1.Intent = NoAction
	r.PerformGameUpdate(0)

	assert.True(t, r.TraceSummary().IsEmpty(), "no traces → summary is empty")
}
