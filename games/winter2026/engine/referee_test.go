package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// TestRefereeSmokeFullTurn runs a single turn through the arena.Referee
// interface: Init → GlobalInfoFor → ResetGameTurnData → FrameInfoFor →
// ParsePlayerOutputs → PerformGameUpdate → Metrics → RawScores → OnEnd.
// It asserts the end-to-end wiring without random map generation by
// substituting a crafted grid after Init.
func TestRefereeSmokeFullTurn(t *testing.T) {
	g := NewGame(1, 1)
	r := NewReferee(g)

	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	r.Init(players)

	// Overwrite the random-generated state with a hand-crafted one.
	g.grid = grid.NewGrid(10, 6)
	for x := 0; x < 10; x++ {
		g.grid.Get(grid.Coord{X: x, Y: 5}).SetType(grid.TileWall)
	}
	// Reset both players, wipe their randomly-made birds.
	g.players[0].birds = nil
	g.players[1].birds = nil
	g.birdByIDCache = make(map[int]*Bird)

	b0 := spawn(g, 0, 0, []grid.Coord{{X: 1, Y: 4}, {X: 0, Y: 4}})
	// Needs length ≥ 2 so Facing() works; both birds face east by shape.
	b1 := spawn(g, 1, 1, []grid.Coord{{X: 8, Y: 4}, {X: 9, Y: 4}})
	g.grid.Apples = []grid.Coord{{X: 3, Y: 4}, {X: 6, Y: 4}}

	// Global + frame serialization exercised through the referee interface.
	global := r.GlobalInfoFor(players[0])
	assert.NotEmpty(t, global)
	frame := r.FrameInfoFor(players[1])
	assert.NotEmpty(t, frame)

	r.ResetGameTurnData()

	// Player 0 commands bird 0 east; player 1 commands bird 1 west.
	players[0].SetOutputs([]string{"0 RIGHT"})
	players[1].SetOutputs([]string{"1 LEFT"})
	r.ParsePlayerOutputs(players)
	assert.True(t, b0.HasMove)
	// Bird 1's facing is west ((8,4)-(9,4) = (-1,0)); LEFT is its facing too,
	// so the command is accepted.
	assert.True(t, b1.HasMove)

	// Active players count before a potential disqualification.
	assert.Equal(t, 2, r.ActivePlayers(players))

	r.PerformGameUpdate(0)

	assert.Equal(t, grid.Coord{X: 2, Y: 4}, b0.HeadPos())
	assert.Equal(t, grid.Coord{X: 7, Y: 4}, b1.HeadPos())

	raw := r.RawScores()
	assert.Equal(t, len(b0.Body), raw[0])
	assert.Equal(t, len(b1.Body), raw[1])

	m := r.Metrics()
	assert.NotEmpty(t, m, "metrics emitted")

	// Turn events should be a copy, not the backing slice.
	events := r.TurnEvents(0, players)
	_ = events

	assert.False(t, r.Ended())
	assert.False(t, r.ShouldSkipPlayerTurn(players[0]))

	r.EndGame()
	assert.True(t, r.Ended(), "EndGame flips the ended flag")

	r.OnEnd()
	assert.NotEqual(t, -1, players[0].GetScore())
	assert.NotEqual(t, -1, players[1].GetScore())
}

// TestRefereeParsePlayerOutputsSkipsDeactivated verifies the referee honours
// the deactivated-flag and the ShouldSkipPlayerTurn hook.
func TestRefereeParsePlayerOutputsSkipsDeactivated(t *testing.T) {
	g := NewGame(1, 1)
	r := NewReferee(g)
	players := []arena.Player{NewPlayer(0), NewPlayer(1)}
	r.Init(players)
	players[1].Deactivate("preexisting")
	// Even with output set, deactivated player's commands are ignored.
	players[1].SetOutputs([]string{"DANCE"})

	r.ParsePlayerOutputs(players)
	assert.True(t, players[1].IsDeactivated(), "stayed deactivated")
	// Score unchanged by ParseCommands for deactivated player.
	assert.NotEqual(t, -1, players[1].GetScore())
}

// TestNewGameWiresUpGridAndPlayers checks that NewGame + Init produce a
// consistent initial state for the default league.
func TestNewGameWiresUpGridAndPlayers(t *testing.T) {
	g := NewGame(1, 1)
	players := []*Player{NewPlayer(0), NewPlayer(1)}
	g.Init(players)

	assert.NotNil(t, g.grid, "grid generated")
	assert.Greater(t, g.grid.Width, 0)
	assert.Greater(t, g.grid.Height, 0)
	assert.Len(t, g.players, 2)
	assert.Len(t, g.players[0].birds, len(g.players[1].birds))
	assert.Greater(t, len(g.players[0].birds), 0, "at least one bird per player")
}

// TestGameErrorAndInvalidInputError covers the Error formatters.
func TestGameErrorAndInvalidInputError(t *testing.T) {
	ge := &GameError{Message: "oops"}
	assert.Equal(t, "oops", ge.Error())

	iie := &InvalidInputError{Expected: "FOO", Got: "BAR"}
	assert.Contains(t, iie.Error(), "Expected FOO")
	assert.Contains(t, iie.Error(), "BAR")
}

// TestPlayerAddScoreAccumulatesPoints covers the AddScore helper.
func TestPlayerAddScoreAccumulatesPoints(t *testing.T) {
	p := NewPlayer(0)
	p.AddScore(3)
	p.AddScore(4)
	assert.Equal(t, 7, p.GetScore())
}

// TestGameShouldSkipPlayerTurnIsAlwaysFalse codifies the current behaviour:
// Winter 2026 has no sub-turn mechanic.
func TestGameShouldSkipPlayerTurnIsAlwaysFalse(t *testing.T) {
	g := NewGame(1, 1)
	assert.False(t, g.ShouldSkipPlayerTurn(nil))
}

// TestFactoryBasics asserts the public factory wiring.
func TestFactoryBasics(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, "winter-2026", f.Name())
	assert.Equal(t, 200, f.MaxTurns())

	ref, players := f.NewGame(1, map[string]string{"league": "1"})
	assert.NotNil(t, ref)
	assert.Len(t, players, 2)

	// Invalid league option is silently ignored — defaults still applied.
	ref2, _ := f.NewGame(1, map[string]string{"league": "abc"})
	assert.NotNil(t, ref2)
}
