package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newEndReasonScenario(t *testing.T) (*Referee, []arena.Player) {
	t.Helper()
	g := newScenario(4)
	r := NewReferee(g)
	players := []arena.Player{g.Players[0], g.Players[1]}
	return r, players
}

// Spring 2021's loop turn 0 is the GATHERING phase (engine-only no-op for
// players); the first prompted turn is loop turn 1. A timeout on that first
// prompt must classify as TIMEOUT_START even though the raw turn index is 1.
func TestEndReasonTimeoutStartOnFirstOutputTurn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[0].Deactivate("Timeout!")

	got := r.EndReason(1, players, [2]int{1, -1}, [2]int{1, 1})
	assert.Equal(t, arena.EndReasonTimeoutStart, got)
}

func TestEndReasonTimeoutOnLaterTurn(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[1].Deactivate("Timeout!")

	got := r.EndReason(42, players, [2]int{-1, 42}, [2]int{1, 1})
	assert.Equal(t, arena.EndReasonTimeout, got)
}

func TestEndReasonInvalidForBadCommand(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Players[0].Deactivate("invalid input")

	got := r.EndReason(5, players, [2]int{5, -1}, [2]int{1, 1})
	assert.Equal(t, arena.EndReasonInvalid, got)
}

func TestEndReasonScoreWhenRoundCapReached(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Round = r.Game.MAX_ROUNDS

	got := r.EndReason(99, players, [2]int{-1, -1}, [2]int{1, 1})
	assert.Equal(t, arena.EndReasonScore, got)
}

func TestEndReasonTurnsOutWhenRoundCapNotReached(t *testing.T) {
	r, players := newEndReasonScenario(t)
	r.Game.Round = 5

	got := r.EndReason(MaxTurns, players, [2]int{-1, -1}, [2]int{1, 1})
	assert.Equal(t, arena.EndReasonTurnsOut, got)
}

func TestRawScoresReadsPlayerScoreBeforeOnEnd(t *testing.T) {
	r, _ := newEndReasonScenario(t)
	r.Game.Players[0].SetScore(7)
	r.Game.Players[1].SetScore(11)

	assert.Equal(t, [2]int{7, 11}, r.RawScores())
}

// Rule: OnEnd layers floor(sun/3) onto each player's score plus an
// all-equal tree-count tiebreaker. RawScores must capture the pre-OnEnd
// value, so trace.Scores reflects raw in-match score-points rather than
// the bonus-adjusted final value CG reports as gameResult.scores.
func TestRawScoresDivergesFromFinalAfterOnEndBonus(t *testing.T) {
	r, _ := newEndReasonScenario(t)
	r.Game.Players[0].SetScore(20)
	r.Game.Players[0].SetSun(9) // floor(9/3) = 3 bonus
	r.Game.Players[1].SetScore(10)
	r.Game.Players[1].SetSun(6) // floor(6/3) = 2 bonus

	raw := r.RawScores()
	r.OnEnd()
	final := [2]int{r.Game.Players[0].GetScore(), r.Game.Players[1].GetScore()}

	assert.Equal(t, [2]int{20, 10}, raw)
	assert.Equal(t, [2]int{23, 12}, final)
}
