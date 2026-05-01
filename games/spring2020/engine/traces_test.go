package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func findTrace(traces []arena.TurnTrace, typ string) (arena.TurnTrace, bool) {
	for _, tr := range traces {
		if tr.Type == typ {
			return tr, true
		}
	}
	return arena.TurnTrace{}, false
}

func countTraces(traces []arena.TurnTrace, typ string) int {
	n := 0
	for _, tr := range traces {
		if tr.Type == typ {
			n++
		}
	}
	return n
}

func decodeMeta[T any](t *testing.T, tr arena.TurnTrace) T {
	t.Helper()
	v, err := arena.DecodeMeta[T](tr)
	require.NoError(t, err)
	return v
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
	assert.Equal(t, EatMeta{Pac: 0, Coord: [2]int{2, 1}, Cost: 1}, decodeMeta[EatMeta](t, tr))
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
	assert.Equal(t, EatMeta{Pac: 0, Coord: [2]int{2, 1}, Cost: 10}, decodeMeta[EatMeta](t, tr),
		"super pellets are identified by Cost > 1")
	assert.Equal(t, 1, countTraces(g.traces, TraceEat), "super pellet emits a single EAT trace")
}

func TestTraceCollideEnemyOnSameTypeBodyBlock(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	a := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	b := spawn(g, 1, 0, TypeRock, Coord{X: 3, Y: 1})

	runTurn(g, func() {
		a.Intent = NewMoveAction(Coord{X: 3, Y: 1})
		b.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	})

	// Both pacs are body-blocked; each emits its own COLLIDE_ENEMY.
	assert.Equal(t, 2, countTraces(g.traces, TraceCollideEnemy))
	assert.Equal(t, 0, countTraces(g.traces, TraceCollideSelf))
}

func TestTraceCollideSelfOnFriendlyBodyBlock(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	a := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	b := spawn(g, 0, 1, TypeRock, Coord{X: 3, Y: 1})

	runTurn(g, func() {
		a.Intent = NewMoveAction(Coord{X: 3, Y: 1})
		b.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	})

	assert.Equal(t, 2, countTraces(g.traces, TraceCollideSelf))
	assert.Equal(t, 0, countTraces(g.traces, TraceCollideEnemy))
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
	assert.Equal(t, KilledMeta{
		Pac:    victim.ID,
		Coord:  coordPair(victim.Position),
		Killer: attacker.ID,
	}, decodeMeta[KilledMeta](t, tr))
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
	assert.Equal(t, PacMeta{Pac: 0}, decodeMeta[PacMeta](t, tr))
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
	assert.Equal(t, SwitchMeta{Pac: 0, Type: "PAPER"}, decodeMeta[SwitchMeta](t, tr))
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
	assert.Equal(t, TraceEat, g.traces[0].Type)
	assert.Equal(t, EatMeta{Pac: 0, Coord: [2]int{3, 1}, Cost: 1}, decodeMeta[EatMeta](t, g.traces[0]))
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
	out[0].Type = "MUTATED"
	assert.NotEqual(t, "MUTATED", g.traces[0].Type, "TurnTraces returned a copy")
}

func TestRefereeTurnTracesNilWhenEmpty(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	r := NewReferee(g)

	asArena := []arena.Player{g.Players[0], g.Players[1]}
	assert.Nil(t, r.TurnTraces(0, asArena), "no traces yet → nil")
}
