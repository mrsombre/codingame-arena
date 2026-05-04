package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Acceptance tests exercise the rules of SpringChallenge2021 on hand-prepared
// state, bypassing random board holes and the auto-WAIT round 0. Each scenario
// asserts a single rule observable through the public Game API.

// newScenario builds a Game with a generated 37-cell board (no holes), two
// fresh players, and zeroed turn state. League 4 (Bronze+) enables every rule.
// The board is generated once with the seed-0 RNG; with ENABLE_HOLES=false no
// random call happens, so the layout is deterministic and matches the engine's
// canonical 37-cell hex grid.
func newScenario(leagueLevel int) *Game {
	g := NewGame(0, leagueLevel)
	g.Players = []*Player{NewPlayer(0), NewPlayer(1)}
	g.Board = NewBoardGenerator().Generate(g.random, g.Cfg, false)
	g.Trees = make(map[int]*Tree)
	g.AvailableSun = []int{0, 0}
	g.Sun = Sun{}
	g.Shadows = make(map[int]int)
	g.Nutrients = g.Cfg.STARTING_NUTRIENTS
	return g
}

// runActionTurn mirrors the runner's per-turn loop: ResetGameTurnData copies
// p.Sun → AvailableSun and clears actions; the caller then attaches fresh
// actions; performActionUpdate consumes them. Skipping ResetGameTurnData
// would leave AvailableSun at zero so cost checks would falsely fail.
func runActionTurn(g *Game, setup func()) {
	g.ResetGameTurnData()
	setup()
	g.performActionUpdate()
}

// ——— starting state ————————————————————————————————————————————————————————

// Rule: "Players start the game with two size 1 trees placed randomly along
// the edge of the grid."
func TestStartingTreesAreSize1OnEdgeRing(t *testing.T) {
	g := NewGame(testArenaPositiveSeed, 4)
	g.Init([]*Player{NewPlayer(0), NewPlayer(1)})
	require.Len(t, g.Trees, 4, "two trees per player")
	for idx, tree := range g.Trees {
		assert.Equal(t, TREE_SMALL, tree.Size, "starting tree size 1")
		assert.GreaterOrEqualf(t, idx, 19, "starting tree on edge ring (cell %d)", idx)
	}
}

// ——— GROW action ——————————————————————————————————————————————————————————

func TestGrowSucceedsConsumingSunAndSettingDormant(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 5
	tree := g.placeTree(p0, 0, TREE_SMALL) // grow cost = 3

	runActionTurn(g, func() {
		p0.SetAction(NewGrowAction(0))
	})

	assert.Equal(t, 2, p0.Sun)
	assert.Equal(t, TREE_MEDIUM, tree.Size)
	assert.True(t, tree.Dormant)
}

// Rule: "Trees can grow up to size 3."
func TestGrowRejectedOnTallTree(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	tree := g.placeTree(p0, 0, TREE_TALL)

	runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

	assert.Equal(t, TREE_TALL, tree.Size)
	assert.True(t, p0.IsWaiting(), "invalid action puts player to sleep")
}

func TestGrowRejectedWithoutEnoughSun(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 0
	tree := g.placeTree(p0, 0, TREE_SMALL)

	runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

func TestGrowRejectedOnOpponentTree(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 100
	tree := g.placeTree(p1, 0, TREE_SMALL)

	runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

// Rule: "A dormant tree cannot be the subject of an action."
func TestGrowRejectedOnDormantTree(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	tree := g.placeTree(p0, 0, TREE_SMALL)
	tree.SetDormant()

	runActionTurn(g, func() { p0.SetAction(NewGrowAction(0)) })

	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

// ——— SEED action ———————————————————————————————————————————————————————————

// Rule: "Performing this action impacts both the source tree and the planted
// seed. Meaning both trees will be dormant until the next day."
func TestSeedSucceedsPlantingDormantSeedAndDormantingSource(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 5
	src := g.placeTree(p0, 0, TREE_SMALL)

	runActionTurn(g, func() { p0.SetAction(NewSeedAction(0, 1)) })

	assert.Equal(t, 5, p0.Sun, "no seeds owned → cost 0")
	assert.True(t, src.Dormant)
	planted, ok := g.Trees[1]
	require.True(t, ok)
	assert.Equal(t, TREE_SEED, planted.Size)
	assert.True(t, planted.Dormant)
	assert.Equal(t, 0, planted.FatherIndex, "father is the source cell")
}

// Rule: "Eject a seed onto a cell within distance equal to the tree's size."
func TestSeedDistanceLimitedBySourceSize(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	src := g.placeTree(p0, 0, TREE_SMALL) // size 1 → max distance 1

	runActionTurn(g, func() { p0.SetAction(NewSeedAction(0, 7)) }) // distance 2

	assert.True(t, p0.IsWaiting())
	assert.False(t, src.Dormant, "rejected action does not consume source")
	_, planted := g.Trees[7]
	assert.False(t, planted)
}

// Rule: "You may not send a seed onto an unusable cell."
func TestSeedRejectedOnUnusableCell(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	g.Board.Cells[1].SetRichness(RICHNESS_NULL)
	g.placeTree(p0, 0, TREE_SMALL)

	runActionTurn(g, func() { p0.SetAction(NewSeedAction(0, 1)) })

	assert.True(t, p0.IsWaiting())
	_, planted := g.Trees[1]
	assert.False(t, planted)
}

// Rule: "...or a cell already containing a tree."
func TestSeedRejectedOnOccupiedCell(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_SMALL)
	g.placeTree(p0, 1, TREE_SEED)

	runActionTurn(g, func() { p0.SetAction(NewSeedAction(0, 1)) })

	assert.True(t, p0.IsWaiting())
	tree, ok := g.Trees[1]
	require.True(t, ok)
	assert.Equal(t, TREE_SEED, tree.Size)
}

// A seed (size 0) cannot itself emit another seed.
func TestSeedRejectedFromSeedSource(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_SEED)

	runActionTurn(g, func() { p0.SetAction(NewSeedAction(0, 1)) })

	assert.True(t, p0.IsWaiting())
}

// Rule: "If both players send a seed to the same place on the same turn,
// neither seed is planted and the sun points are refunded. The source tree,
// however, still becomes dormant."
func TestSeedConflictRefundsSunButKeepsSourcesDormant(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 10
	p1.Sun = 10
	// Extra seeds give each player a non-zero seed cost so the refund is observable.
	g.placeTree(p0, 2, TREE_SEED)
	g.placeTree(p1, 8, TREE_SEED)
	src0 := g.placeTree(p0, 0, TREE_SMALL) // cell 0 → cell 1, distance 1
	src1 := g.placeTree(p1, 7, TREE_SMALL) // cell 7 → cell 1, distance 1

	runActionTurn(g, func() {
		p0.SetAction(NewSeedAction(0, 1))
		p1.SetAction(NewSeedAction(7, 1))
	})

	assert.Equal(t, 10, p0.Sun, "sun refunded on conflict")
	assert.Equal(t, 10, p1.Sun, "sun refunded on conflict")
	assert.True(t, src0.Dormant)
	assert.True(t, src1.Dormant)
	_, planted := g.Trees[1]
	assert.False(t, planted, "no seed planted on conflict")
}

// ——— COMPLETE action ——————————————————————————————————————————————————————

// Rule: "Completing a tree's lifecycle requires 4 sun points... awards
// nutrients value + richness bonus."
func TestCompleteAwardsNutrientsAndLushBonus(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 4
	g.placeTree(p0, 0, TREE_TALL) // cell 0 = lush → +4

	runActionTurn(g, func() { p0.SetAction(NewCompleteAction(0)) })

	assert.Equal(t, 20+RICHNESS_BONUS_LUSH, p0.GetScore())
	assert.Equal(t, 0, p0.Sun)
	assert.Equal(t, 19, g.Nutrients, "nutrients drop by 1")
	_, exists := g.Trees[0]
	assert.False(t, exists, "tree removed from forest")
}

// Rule: bonuses are 0/+2/+4 for poor/medium/lush.
func TestCompleteRichnessBonuses(t *testing.T) {
	cases := []struct {
		name    string
		cellIdx int
		want    int
	}{
		{"poor", 19, 20 + 0},
		{"medium", 7, 20 + RICHNESS_BONUS_OK},
		{"lush", 0, 20 + RICHNESS_BONUS_LUSH},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newScenario(4)
			p0 := g.Players[0]
			p0.Sun = 4
			g.placeTree(p0, tc.cellIdx, TREE_TALL)
			runActionTurn(g, func() { p0.SetAction(NewCompleteAction(tc.cellIdx)) })
			assert.Equal(t, tc.want, p0.GetScore())
		})
	}
}

// Rule: "You can only complete the lifecycle of a size 3 tree."
func TestCompleteRejectedOnNonTallTree(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_MEDIUM)

	runActionTurn(g, func() { p0.SetAction(NewCompleteAction(0)) })

	_, exists := g.Trees[0]
	assert.True(t, exists)
	assert.True(t, p0.IsWaiting())
}

func TestCompleteRejectedWithoutEnoughSun(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 3
	g.placeTree(p0, 0, TREE_TALL)

	runActionTurn(g, func() { p0.SetAction(NewCompleteAction(0)) })

	_, exists := g.Trees[0]
	assert.True(t, exists)
	assert.Equal(t, 0, p0.GetScore())
	assert.True(t, p0.IsWaiting())
}

// Rule: "If both players complete a lifecycle on the same turn, they both
// receive full points and the nutrient value is decreased by two."
func TestBothPlayersCompleteSameTurnNutrientsDropByTwo(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 4
	p1.Sun = 4
	g.placeTree(p0, 0, TREE_TALL) // lush +4
	g.placeTree(p1, 7, TREE_TALL) // medium +2

	runActionTurn(g, func() {
		p0.SetAction(NewCompleteAction(0))
		p1.SetAction(NewCompleteAction(7))
	})

	assert.Equal(t, 20+RICHNESS_BONUS_LUSH, p0.GetScore())
	assert.Equal(t, 20+RICHNESS_BONUS_OK, p1.GetScore())
	assert.Equal(t, 18, g.Nutrients)
}

// Rule: "The nutrients value cannot drop below 0."
func TestNutrientsClampedAtZero(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	p0.Sun = 4
	g.Nutrients = 0
	g.placeTree(p0, 19, TREE_TALL)

	runActionTurn(g, func() { p0.SetAction(NewCompleteAction(19)) })

	assert.Equal(t, 0, g.Nutrients)
	assert.Equal(t, 0, p0.GetScore(), "0 nutrients + 0 poor bonus")
}

// ——— sun gathering & shadows ——————————————————————————————————————————————

// Rule: gathered sun = tree size when not under a spooky shadow.
func TestGiveSunByTreeSizeWhenNotShadowed(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	g.placeTree(p0, 22, TREE_SEED)
	g.placeTree(p0, 25, TREE_SMALL)
	g.placeTree(p0, 28, TREE_MEDIUM)
	g.placeTree(p0, 30, TREE_TALL)

	g.giveSun() // shadows map empty → all trees gather

	assert.Equal(t, 0+1+2+3, p0.Sun)
}

// Rule: "lesser spirits will find the shadow on a cell spooky if any of the
// trees casting a shadow is of equal or greater size than the tree on that
// cell."
func TestSpookyShadowsBlockSunGathering(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	g.Sun.SetOrientation(0)
	g.placeTree(p0, 0, TREE_TALL)  // shadows cells 1,7,19 with size 3
	g.placeTree(p0, 7, TREE_SMALL) // shadow 3 >= 1 → spooky
	g.placeTree(p0, 19, TREE_TALL) // shadow 3 >= 3 → spooky
	g.placeTree(p0, 4, TREE_SMALL) // not shadowed, but shadows cell 0 with size 1

	g.calculateShadows()
	g.giveSun()

	// Cell 0 (size 3) shadowed by size 1 → 1 < 3 → +3.
	// Cell 4 (size 1) → +1. Cells 7 and 19 spooky → 0.
	assert.Equal(t, 4, p0.Sun)
}

// Rule: "in between each day, the sun moves to point towards the next
// direction... the sun's direction will always be equal to the current day
// modulo 6."
func TestPerformSunMoveAdvancesRoundAndRecomputesShadows(t *testing.T) {
	g := newScenario(4)
	g.placeTree(g.Players[0], 0, TREE_SMALL)
	g.calculateShadows()
	require.Equal(t, TREE_SMALL, g.Shadows[1])

	g.performSunMoveUpdate()

	assert.Equal(t, 1, g.Round)
	assert.Equal(t, 1, g.Sun.Orientation)
	assert.Equal(t, TREE_SMALL, g.Shadows[2], "shadow now in new direction")
	_, oldStill := g.Shadows[1]
	assert.False(t, oldStill, "old shadow cleared after sun move")
}

// Each new day starts with all trees no longer dormant.
func TestSunGatheringResetsDormantTrees(t *testing.T) {
	g := newScenario(4)
	p0 := g.Players[0]
	tree := g.placeTree(p0, 25, TREE_MEDIUM)
	tree.SetDormant()

	g.performSunGatheringUpdate()

	assert.False(t, tree.Dormant)
	assert.Equal(t, 2, p0.Sun)
}

// ——— end of game —————————————————————————————————————————————————————————

// Rule: "The game lasts the time it takes for the sun to circle around the
// board 4 times. This means players have 24 days to play."
func TestGameOverAfterMaxRounds(t *testing.T) {
	g := newScenario(4)
	assert.False(t, g.IsGameOver())
	g.Round = g.MAX_ROUNDS
	assert.True(t, g.IsGameOver())
}

// Rule: "Players gain an extra 1 point for every 3 sun points they have at
// the end of the game."
func TestOnEndAddsFloorSunOverThree(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(10)
	p1.SetScore(20)
	p0.Sun = 7
	p1.Sun = 9

	g.OnEnd()

	assert.Equal(t, 12, p0.GetScore())
	assert.Equal(t, 23, p1.GetScore())
}

// Rule: "If players have the same score, the winner is the player with the
// most trees in the forest. Note that a seed is also considered a tree."
func TestOnEndTiebreakAddsBonusPerTree(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(10)
	p1.SetScore(10)
	g.placeTree(p0, 1, TREE_SMALL)
	g.placeTree(p0, 2, TREE_SEED) // seed counts as a tree
	g.placeTree(p1, 3, TREE_SMALL)

	g.OnEnd()

	assert.Equal(t, 12, p0.GetScore())
	assert.Equal(t, 11, p1.GetScore())
	assert.Equal(t, 2, p0.BonusScore)
	assert.Equal(t, 1, p1.BonusScore)
}

func TestOnEndNoTiebreakWhenScoresDiffer(t *testing.T) {
	g := newScenario(4)
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(15)
	p1.SetScore(10)
	g.placeTree(p0, 0, TREE_SMALL)
	g.placeTree(p1, 1, TREE_SMALL)

	g.OnEnd()

	assert.Equal(t, 15, p0.GetScore())
	assert.Equal(t, 10, p1.GetScore())
	assert.Equal(t, 0, p0.BonusScore)
	assert.Equal(t, 0, p1.BonusScore)
}
