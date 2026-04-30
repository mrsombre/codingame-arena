package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newEndReasonScenario(t *testing.T) (*Referee, []arena.Player) {
	t.Helper()
	g := newScenario(10, 6)
	floorRow(g)
	spawn(g, 0, 0, []Coord{{X: 2, Y: 4}, {X: 1, Y: 4}, {X: 0, Y: 4}})
	spawn(g, 1, 1, []Coord{{X: 7, Y: 4}, {X: 8, Y: 4}, {X: 9, Y: 4}})
	apple(g, Coord{X: 5, Y: 2})
	r := NewReferee(g)
	players := []arena.Player{g.Players[0], g.Players[1]}
	return r, players
}

func TestEndReasonTimeoutStartOnTurnZero(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[0].Deactivate("Timeout!")

	got := r.EndReason(0, players, [2]int{0, -1})
	assert.Equal(t, arena.EndReasonTimeoutStart, got)
}

func TestEndReasonTimeoutOnLaterTurn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[1].Deactivate("Timeout!")

	got := r.EndReason(42, players, [2]int{-1, 42})
	assert.Equal(t, arena.EndReasonTimeout, got)
}

func TestEndReasonInvalidForBadCommand(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[0].Deactivate(`Expected MESSAGE text; got "DANCE"`)

	got := r.EndReason(5, players, [2]int{5, -1})
	assert.Equal(t, arena.EndReasonInvalid, got)
}

func TestEndReasonEliminatedWhenAllBirdsDead(t *testing.T) {
	r, players := newEndReasonScenario(t)
	for _, b := range r.Game.Players[1].Birds {
		b.Alive = false
	}

	got := r.EndReason(20, players, [2]int{-1, -1})
	assert.Equal(t, arena.EndReasonEliminated, got)
}

func TestEndReasonScoreWhenAllApplesConsumed(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Grid.Apples = nil

	got := r.EndReason(15, players, [2]int{-1, -1})
	assert.Equal(t, arena.EndReasonScore, got)
}

func TestEndReasonTurnsOutWhenCapHit(t *testing.T) {
	r, players := newEndReasonScenario(t)
	// Both players still active, apples remain, both birds alive → IsGameOver
	// is false. Loop must have exited via turn cap.
	got := r.EndReason(200, players, [2]int{-1, -1})
	assert.Equal(t, arena.EndReasonTurnsOut, got)
}
