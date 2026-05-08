package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// findTrace scans both player slots for a trace of typ. Tests that don't
// care which side owned the event use this — the per-side specifics are
// covered by dedicated assertions where they matter.
func findTrace(traces [2][]arena.TurnTrace, typ string) (arena.TurnTrace, bool) {
	for _, slot := range traces {
		for _, tr := range slot {
			if tr.Type == typ {
				return tr, true
			}
		}
	}
	return arena.TurnTrace{}, false
}

func countTraces(traces [2][]arena.TurnTrace, typ string) int {
	n := 0
	for _, slot := range traces {
		for _, tr := range slot {
			if tr.Type == typ {
				n++
			}
		}
	}
	return n
}

func totalTraces(traces [2][]arena.TurnTrace) int {
	return len(traces[0]) + len(traces[1])
}

func decodeMeta[T any](t *testing.T, tr arena.TurnTrace) T {
	t.Helper()
	v, err := arena.DecodeData[T](tr)
	require.NoError(t, err)
	return v
}

// SerializeTraceFrameInfo must include every pacman and pellet on the
// board, regardless of fog-of-war visibility — that's the whole point of
// the trace variant. The grid uses two parallel corridors separated by a
// solid wall row so each pac's cone-of-sight cannot reach the other; the
// fog-aware serializer must hide the opponent, the trace variant must not.
func TestSerializeTraceFrameInfoIncludesAllPacsAndPelletsUnderFog(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
		"#.....#",
		"#######",
	}, false)
	require.True(t, g.Config.FOG_OF_WAR, "league 4 enables fog of war")

	pacMine := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	pacOpp := spawn(g, 1, 0, TypePaper, Coord{X: 1, Y: 3})

	// Sanity check: standard FrameInfoFor under fog hides the opponent's pac
	// from side 0's perspective (separated by a wall row).
	fogLines := SerializeFrameInfoFor(g.Players[0], g)
	require.NotContains(t, fogLines, PacmanLine(g.Players[0], pacOpp), "fog should hide opponent across walls")

	// Trace variant must include both pacs and the full pellet count.
	traceLines := SerializeTraceFrameInfo(g)
	assert.Contains(t, traceLines, PacmanLine(g.Players[0], pacMine), "trace lists own pac")
	assert.Contains(t, traceLines, PacmanLine(g.Players[0], pacOpp), "trace lists opponent pac despite fog")

	// Pellet count line: line 0 = scores, line 1 = pac count, then 2 pac
	// rows, then the pellet count line. Trace variant must report ALL
	// pellets/cherries on the board (10 in this 5×7 grid: 5 per corridor).
	require.GreaterOrEqual(t, len(traceLines), 5)
	assert.Equal(t, "10", traceLines[4], "trace pellet count covers both corridors")
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
	assert.Empty(t, g.traces[1], "events for pac on side 0 must not leak into side 1")
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

	// Each blocked pac emits its own COLLIDE_ENEMY, mirrored into both slots.
	// Two pacs collide → 2 events × 2 mirrors = 4 events total.
	assert.Equal(t, 4, countTraces(g.traces, TraceCollideEnemy))
	assert.Equal(t, 0, countTraces(g.traces, TraceCollideSelf))
	assert.Equal(t, 2, len(g.traces[0]), "side 0 sees both mirrored events")
	assert.Equal(t, 2, len(g.traces[1]), "side 1 sees both mirrored events")
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
	assert.Empty(t, g.traces[1], "same-team collisions stay in the owner's slot only")
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
	assert.Equal(t, 1, len(g.traces[1]), "KILLED lives in the victim's owner slot (side 1)")
	assert.Equal(t, 0, len(g.traces[0]), "killer's slot does not receive the KILLED event")
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
	require.Greater(t, totalTraces(g.traces), 0, "first turn produced traces")

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})
	// One EAT for the new cell stepped onto, no leftover from the first turn.
	require.Equal(t, 1, totalTraces(g.traces))
	assert.Equal(t, TraceEat, g.traces[0][0].Type)
	assert.Equal(t, EatMeta{Pac: 0, Coord: [2]int{3, 1}, Cost: 1}, decodeMeta[EatMeta](t, g.traces[0][0]))
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
	require.Greater(t, totalTraces(g.traces), 0)

	asArena := []arena.Player{g.Players[0], g.Players[1]}
	out := r.TurnTraces(0, asArena)
	require.Greater(t, totalTraces(out), 0)
	// Mutate the copy and verify the engine slot is unaffected.
	out[0][0].Type = "MUTATED"
	assert.NotEqual(t, "MUTATED", g.traces[0][0].Type, "TurnTraces returned a copy")
}

func TestRefereeTurnTracesEmptyWhenNoEvents(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	r := NewReferee(g)

	asArena := []arena.Player{g.Players[0], g.Players[1]}
	out := r.TurnTraces(0, asArena)
	assert.Equal(t, 0, totalTraces(out), "no traces yet → both slots empty")
}
