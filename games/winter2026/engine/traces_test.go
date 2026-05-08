package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeState calls Game.DecorateTraceTurn and unmarshals the result
// into a typed TraceTurnState so per-field assertions can run without
// re-implementing the wire format inside each test.
func decodeState(t *testing.T, g *Game) TraceTurnState {
	t.Helper()
	raw := g.DecorateTraceTurn(0, nil)
	require.NotEmpty(t, raw, "DecorateTraceTurn returned empty bytes")
	var state TraceTurnState
	require.NoError(t, json.Unmarshal(raw, &state))
	return state
}

// makeStateGame builds a minimal Game with two players, four snakes
// (two per side, ids 0..3), and a custom apple list. Snakes start at
// the cells passed in; the body is auto-filled with two trailing
// segments below the head so Bird.Facing() works.
func makeStateGame(applesXY [][2]int, heads [4][2]int) *Game {
	g := &Game{Grid: NewGrid(20, 10)}
	g.Players = []*Player{NewPlayer(0), NewPlayer(1)}
	for i, p := range g.Players {
		p.Init()
		for j := 0; j < 2; j++ {
			id := i*2 + j
			b := NewBird(id, p)
			head := heads[id]
			b.Body = []Coord{
				{X: head[0], Y: head[1]},
				{X: head[0], Y: head[1] + 1},
				{X: head[0], Y: head[1] + 2},
			}
			p.Birds = append(p.Birds, b)
		}
	}
	for _, c := range applesXY {
		g.Grid.Apples = append(g.Grid.Apples, Coord{X: c[0], Y: c[1]})
	}
	return g
}

func TestDecorateTraceTurnState(t *testing.T) {
	g := makeStateGame(
		[][2]int{{1, 1}, {2, 2}, {3, 3}},
		[4][2]int{{5, 5}, {6, 5}, {7, 5}, {8, 5}},
	)

	state := decodeState(t, g)

	assert.Equal(t, 3, state.Apples)
	assert.Equal(t, []TraceSnake{
		{ID: 0, Size: 3, Head: [2]int{5, 5}},
		{ID: 1, Size: 3, Head: [2]int{6, 5}},
	}, state.Snakes[0])
	assert.Equal(t, []TraceSnake{
		{ID: 2, Size: 3, Head: [2]int{7, 5}},
		{ID: 3, Size: 3, Head: [2]int{8, 5}},
	}, state.Snakes[1])
}

func TestDecorateTraceTurnDropsDeadSnakes(t *testing.T) {
	g := makeStateGame(
		[][2]int{{1, 1}},
		[4][2]int{{5, 5}, {6, 5}, {7, 5}, {8, 5}},
	)
	g.Players[0].Birds[1].Alive = false // bird 1 dead
	g.Players[1].Birds[0].Alive = false // bird 2 dead

	state := decodeState(t, g)

	assert.Equal(t, 1, state.Apples)
	require.Len(t, state.Snakes[0], 1)
	assert.Equal(t, 0, state.Snakes[0][0].ID, "side 0's surviving snake")
	require.Len(t, state.Snakes[1], 1)
	assert.Equal(t, 3, state.Snakes[1][0].ID, "side 1's surviving snake")
}

func TestDecorateTraceTurnSnakeSizeReflectsBodyGrowth(t *testing.T) {
	g := makeStateGame(
		nil,
		[4][2]int{{5, 5}, {6, 5}, {7, 5}, {8, 5}},
	)
	// Simulate bird 0 having eaten an apple — body length grows beyond
	// the spawn 3.
	g.Players[0].Birds[0].Body = append(
		g.Players[0].Birds[0].Body,
		Coord{X: 5, Y: 8},
		Coord{X: 5, Y: 9},
	)

	state := decodeState(t, g)

	assert.Equal(t, 5, state.Snakes[0][0].Size, "size tracks current body length")
	assert.Equal(t, 0, state.Apples, "no apples on the field")
}

func TestDecorateTraceTurnSnakesSortedByID(t *testing.T) {
	g := makeStateGame(
		nil,
		[4][2]int{{5, 5}, {6, 5}, {7, 5}, {8, 5}},
	)
	// Reverse the spawn order in side 0's roster: ids 1, 0. Output
	// must still be sorted ascending.
	p0 := g.Players[0]
	p0.Birds = []*Bird{p0.Birds[1], p0.Birds[0]}

	state := decodeState(t, g)

	require.Len(t, state.Snakes[0], 2)
	assert.Equal(t, 0, state.Snakes[0][0].ID)
	assert.Equal(t, 1, state.Snakes[0][1].ID)
}
