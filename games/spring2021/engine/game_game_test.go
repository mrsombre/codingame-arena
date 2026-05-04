package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rule: "you must pay sun points equal to the number of seeds you already
// own in the forest."
func TestSeedCostScalesWithOwnSeedCount(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	assert.Equal(t, 0, g.getSeedCost(p0))
	g.placeTree(p0, 1, TREE_SEED)
	g.placeTree(p0, 2, TREE_SEED)
	assert.Equal(t, 2, g.getSeedCost(p0))
	g.placeTree(p1, 3, TREE_SEED)
	assert.Equal(t, 2, g.getSeedCost(p0), "opponent seeds don't count")
	assert.Equal(t, 1, g.getSeedCost(p1))
}

// Rule: grow cost = base[targetSize] + own count of trees of targetSize.
func TestGrowCostScalesWithOwnTargetSizeCount(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	subject := g.placeTree(p0, 0, TREE_SMALL) // target = MEDIUM, base 3
	assert.Equal(t, 3, g.getGrowthCost(subject))
	g.placeTree(p0, 1, TREE_MEDIUM)
	assert.Equal(t, 4, g.getGrowthCost(subject))
	g.placeTree(p0, 2, TREE_MEDIUM)
	assert.Equal(t, 5, g.getGrowthCost(subject))
	g.placeTree(g.Players[1], 3, TREE_MEDIUM)
	assert.Equal(t, 5, g.getGrowthCost(subject), "opponent trees don't count")
}

func TestGrowCostBaseValuesPerSize(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	cases := []struct {
		size int
		want int
	}{
		{TREE_SEED, 1},   // → small, base 1
		{TREE_SMALL, 3},  // → medium, base 3
		{TREE_MEDIUM, 7}, // → tall, base 7
	}
	for _, tc := range cases {
		tree := g.placeTree(p0, 0, tc.size)
		assert.Equal(t, tc.want, g.getGrowthCost(tree))
		g.removeTree(0)
	}
}

// Rule: completing a lifecycle costs LIFECYCLE_END_COST (4 sun).
func TestGrowCostForTallEqualsLifecycleEndConstant(t *testing.T) {
	g := newScenario(4)
	tall := g.placeTree(g.Players[0], 0, TREE_TALL)
	assert.Equal(t, LIFECYCLE_END_COST, g.getGrowthCost(tall))
}

func TestSeedsAreConflictingDetectsDuplicateTarget(t *testing.T) {
	g := newScenario(4)
	g.SentSeeds = []Seed{
		{Owner: 0, SourceCell: 1, TargetCell: 5},
		{Owner: 1, SourceCell: 7, TargetCell: 5},
	}
	assert.True(t, g.seedsAreConflicting())
	g.SentSeeds = []Seed{
		{Owner: 0, SourceCell: 1, TargetCell: 5},
		{Owner: 1, SourceCell: 7, TargetCell: 6},
	}
	assert.False(t, g.seedsAreConflicting())
}

func TestPlaceTreeKeepsTreeOrderSorted(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	g.placeTree(p0, 30, TREE_SEED)
	g.placeTree(p0, 5, TREE_SEED)
	g.placeTree(p0, 17, TREE_SEED)
	assert.Equal(t, []int{5, 17, 30}, g.TreeOrder)
	g.removeTree(17)
	assert.Equal(t, []int{5, 30}, g.TreeOrder)
}

// Rule: shadow length = caster size, in the sun direction.
func TestCalculateShadowsLengthEqualsSize(t *testing.T) {
	g := newScenario(4)
	g.Sun.SetOrientation(0)
	g.placeTree(g.Players[0], 0, TREE_TALL) // shadows cells 1, 7, 19
	g.calculateShadows()
	assert.Equal(t, TREE_TALL, g.Shadows[1])
	assert.Equal(t, TREE_TALL, g.Shadows[7])
	assert.Equal(t, TREE_TALL, g.Shadows[19])
	_, hasCenter := g.Shadows[0]
	assert.False(t, hasCenter, "tree never shadows its own cell")
}

// Rule: when multiple trees shadow the same cell, the largest casting size wins.
func TestCalculateShadowsKeepsMaxAcrossCasters(t *testing.T) {
	g := newScenario(4)
	g.Sun.SetOrientation(0)
	g.placeTree(g.Players[0], 0, TREE_MEDIUM) // shadows 1, 7
	g.placeTree(g.Players[0], 1, TREE_SMALL)  // shadows 7
	g.calculateShadows()
	assert.Equal(t, TREE_MEDIUM, g.Shadows[7])
}

func TestAllPlayersAreWaiting(t *testing.T) {
	g := newScenario(4)
	assert.False(t, g.allPlayersAreWaiting())
	g.Players[0].SetWaiting(true)
	assert.False(t, g.allPlayersAreWaiting())
	g.Players[1].SetWaiting(true)
	assert.True(t, g.allPlayersAreWaiting())
}

func TestGameOverWhenRoundReachesMax(t *testing.T) {
	g := newScenario(4)
	g.Round = g.MAX_ROUNDS - 1
	assert.False(t, g.IsGameOver())
	g.Round = g.MAX_ROUNDS
	assert.True(t, g.IsGameOver())
}

func TestGameOverWhenOnlyOneActivePlayerRemains(t *testing.T) {
	g := newScenario(4)
	g.Players[1].Deactivate("bad input")
	assert.True(t, g.IsGameOver())
}

// Rule: GetExpected reports legal action grammar for the current league.
func TestGetExpectedReflectsLeagueFlags(t *testing.T) {
	g := newScenario(4)
	assert.Equal(t, "SEED <from> <to> | GROW <idx> | COMPLETE <idx> | WAIT", g.GetExpected())

	g.ENABLE_SEED = false
	assert.Equal(t, "GROW <idx> | COMPLETE <idx> | WAIT", g.GetExpected())

	g.ENABLE_GROW = false
	assert.Equal(t, "COMPLETE <idx> | WAIT", g.GetExpected())
}
