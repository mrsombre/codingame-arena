package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rail cost is a multiple of BASE_RAIL_COST chosen by terrain, and the tests
// name the multiplier rather than the literal so a rebalance moves one place.
func TestRailCostScalesWithTerrain(t *testing.T) {
	cases := map[string]struct {
		typ  int
		want int
	}{
		"plains":   {TYPE_GRASS, BASE_RAIL_COST * GRASS_COST_MULTIPLIER},
		"river":    {TYPE_WATER, BASE_RAIL_COST * RIVER_COST_MULTIPLIER},
		"mountain": {TYPE_MOUNTAIN, BASE_RAIL_COST * MOUNTAIN_COST_MULTIPLIER},
		"poi":      {TYPE_POI, BASE_RAIL_COST * POI_COST_MULTIPLIER},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tile := NewTile(Coord{0, 0})
			tile.SetType(tc.typ)

			assert.Equal(t, tc.want, RailCost(tile))
		})
	}
}

// Off the grid Grid.Get returns nil, and every predicate RailCost consults is
// nil-tolerant, so it answers the plains price rather than panicking. Nothing
// charges it — doActions rejects the placement first — but the arithmetic has
// to stay total.
func TestRailCostOfAnOffGridTileIsThePlainsPrice(t *testing.T) {
	game := NewGame(nil, DEFAULT_LEAGUE)
	game.Init([]*Player{NewPlayer(0), NewPlayer(1)})

	assert.Equal(t, BASE_RAIL_COST*GRASS_COST_MULTIPLIER, game.RailCostAt(Coord{-1, -1}))
}
