package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newEndReasonScenario(t *testing.T) (*Referee, []arena.Player) {
	t.Helper()
	g := newScenario(4, []string{
		"#####",
		"#...#",
		"#####",
	}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeScissors, Coord{X: 3, Y: 1})
	r := NewReferee(g)
	players := []arena.Player{g.Players[0], g.Players[1]}
	return r, players
}

func TestEndReasonTimeoutStartOnFirstOutputTurn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	g.Players[0].Deactivate("Timeout!")

	got := r.EndReason(0, players, [2]int{0, -1}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonTimeoutStart, got)
}

func TestEndReasonTimeoutOnLaterTurn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	g.Players[1].Deactivate("Timeout!")

	got := r.EndReason(42, players, [2]int{-1, 42}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonTimeout, got)
}

func TestEndReasonInvalidForBadCommand(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	g.Players[0].Deactivate("Pac 0 cannot be commanded twice!")

	got := r.EndReason(5, players, [2]int{5, -1}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonInvalid, got)
}

func TestEndReasonEliminatedWhenAllPacmenDead(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	g.Players[1].Deactivate("all pacmen dead")

	got := r.EndReason(20, players, [2]int{-1, 20}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonEliminated, got)
}

func TestEndReasonScoreWhenAllPelletsConsumed(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	for _, cell := range g.Grid.Cells {
		cell.HasPellet = false
		cell.HasCherry = false
	}

	got := r.EndReason(15, players, [2]int{-1, -1}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonScore, got)
}

func TestEndReasonScoreEarlyOnMathLockIn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	g := r.Game
	// One pellet remains; p0 leads by 5, only 1 reachable → no improvement
	// possible despite remaining pellets.
	for _, cell := range g.Grid.Cells {
		cell.HasPellet = false
		cell.HasCherry = false
	}
	g.Grid.Get(Coord{X: 2, Y: 1}).HasPellet = true
	g.Players[0].Pellets = 5
	g.Players[1].Pellets = 0

	got := r.EndReason(30, players, [2]int{-1, -1}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonScoreEarly, got)
}

func TestEndReasonTurnsOutWhenCapHit(t *testing.T) {
	r, players := newEndReasonScenario(t)
	// Both players still active, pellets remain, scores tied → game can still
	// improve so IsGameOver is false. Loop must have exited via turn cap.
	got := r.EndReason(MaxMainTurns, players, [2]int{-1, -1}, [2]int{0, 0})
	assert.Equal(t, arena.EndReasonTurnsOut, got)
}
