package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGame builds a Game with a generated board (no holes), two players,
// and zeroed simulation state so individual rules can be exercised without
// going through initStartingTrees randomness. League 4 is used so all flags
// (SEED, GROW, SHADOW, HOLES) are enabled.
func newTestGame() *Game {
	g := NewGame(0, 4)
	g.Players = []*Player{NewPlayer(0), NewPlayer(1)}
	g.Board = NewBoardGenerator().Generate(g.random, g.Cfg, false)
	g.Trees = make(map[int]*Tree)
	g.TreeOrder = nil
	g.AvailableSun = []int{0, 0}
	g.Sun = Sun{}
	g.Shadows = make(map[int]int)
	g.Nutrients = g.Cfg.STARTING_NUTRIENTS
	return g
}

// --- Board generation rules -------------------------------------------------

// Rule: "The forest is made up of 37 hexagonal cells".
func TestBoardHas37Cells(t *testing.T) {
	g := newTestGame()
	assert.Len(t, g.Board.Cells, 37)
	assert.Len(t, g.Board.Coords, 37)
	assert.Len(t, g.Board.Map, 37)
}

// Rule: richness rings — center+inner=lush, middle=ok, outer=poor.
func TestBoardRichnessRings(t *testing.T) {
	g := newTestGame()
	for i := 0; i <= 6; i++ {
		assert.Equalf(t, RICHNESS_LUSH, g.Board.Cells[i].GetRichness(), "cell %d (center/inner)", i)
	}
	for i := 7; i <= 18; i++ {
		assert.Equalf(t, RICHNESS_OK, g.Board.Cells[i].GetRichness(), "cell %d (middle)", i)
	}
	for i := 19; i <= 36; i++ {
		assert.Equalf(t, RICHNESS_POOR, g.Board.Cells[i].GetRichness(), "cell %d (outer)", i)
	}
}

// Rule: starting trees are size 1 placed on the edge ring (default league).
func TestStartingTreesAreSize1OnEdgeRing(t *testing.T) {
	g := NewGame(testArenaPositiveSeed, 4)
	g.Init([]*Player{NewPlayer(0), NewPlayer(1)})
	require.Len(t, g.Trees, 4, "two trees per player")
	for idx, tree := range g.Trees {
		assert.Equal(t, TREE_SMALL, tree.Size, "starting tree size 1")
		assert.GreaterOrEqualf(t, idx, 19, "starting tree on edge ring (cell %d)", idx)
	}
}

// --- Sun & shadow rules -----------------------------------------------------

// Rule: "the sun's direction will always be equal to the current day modulo 6".
func TestSunOrientationWrapsModSix(t *testing.T) {
	s := Sun{}
	s.SetOrientation(0)
	for i := 0; i < 6; i++ {
		s.Move()
	}
	assert.Equal(t, 0, s.Orientation, "after 6 moves wraps to 0")
	s.Move()
	assert.Equal(t, 1, s.Orientation, "7th move == 7 mod 6")
}

// Rule: "size N tree casts a shadow N cells long" in the sun direction.
func TestShadowFromTallTreeAtCenter(t *testing.T) {
	g := newTestGame()
	g.Sun.SetOrientation(0) // direction 0 → cell 0 → 1 → 7 → 19
	g.placeTree(g.Players[0], 0, TREE_TALL)
	g.calculateShadows()
	assert.Equal(t, TREE_TALL, g.Shadows[1])
	assert.Equal(t, TREE_TALL, g.Shadows[7])
	assert.Equal(t, TREE_TALL, g.Shadows[19])
	_, hasCenter := g.Shadows[0]
	assert.False(t, hasCenter, "tree never shadows its own cell")
}

// Rule: shadow value is the *largest* casting size — not the sum.
func TestShadowKeepsMaxAcrossCasters(t *testing.T) {
	g := newTestGame()
	g.Sun.SetOrientation(0)
	// Both reach cell 7: medium tree at 0 (dist 2) and small tree at 1 (dist 1).
	g.placeTree(g.Players[0], 0, TREE_MEDIUM)
	g.placeTree(g.Players[0], 1, TREE_SMALL)
	g.calculateShadows()
	assert.Equal(t, TREE_MEDIUM, g.Shadows[7], "max caster size wins")
}

// Rule: trees not under a spooky shadow gain sun = their size.
func TestGiveSunByTreeSizeWhenNotShadowed(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	// No shadows calculated → all trees gain.
	g.placeTree(p0, 22, TREE_SEED)
	g.placeTree(p0, 25, TREE_SMALL)
	g.placeTree(p0, 28, TREE_MEDIUM)
	g.placeTree(p0, 30, TREE_TALL)
	g.giveSun()
	assert.Equal(t, 0+1+2+3, p0.Sun)
}

// Rule: "spooky shadow" = caster size >= target tree size → no sun.
func TestGiveSunSkipsSpookyTrees(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	g.Sun.SetOrientation(0)
	// Tall tree at cell 0 shadows cells 1, 7, 19 with shadow=3.
	g.placeTree(p0, 0, TREE_TALL)    // shadowed by cell 4 (size 1) below: 1<3 → gains 3
	g.placeTree(p0, 7, TREE_SMALL)   // shadow 3 >= 1 → spooky
	g.placeTree(p0, 19, TREE_TALL)   // shadow 3 >= 3 → spooky
	g.placeTree(p0, 4, TREE_SMALL)   // unshadowed, also shadows cell 0 with size 1
	g.calculateShadows()
	g.giveSun()
	// 3 (cell 0) + 1 (cell 4); cells 7 and 19 spooky → 0.
	assert.Equal(t, 4, p0.Sun)
}

// Rule: SUN_MOVE phase rotates sun by one and recomputes shadows.
func TestPerformSunMoveAdvancesRoundAndRecomputesShadows(t *testing.T) {
	g := newTestGame()
	g.Sun.SetOrientation(0)
	g.placeTree(g.Players[0], 0, TREE_SMALL)
	g.calculateShadows()
	require.Equal(t, TREE_SMALL, g.Shadows[1])

	g.Round = 0
	g.performSunMoveUpdate()
	assert.Equal(t, 1, g.Round)
	assert.Equal(t, 1, g.Sun.Orientation)
	// After rotation cell 0's small tree now shadows neighbor at orientation 1 = cell 2.
	assert.Equal(t, TREE_SMALL, g.Shadows[2])
	_, oldStill := g.Shadows[1]
	assert.False(t, oldStill, "old shadow cleared after sun move")
}

// --- Cost rules -------------------------------------------------------------

// Rule: seed cost = number of seeds you already own.
func TestSeedCostScalesWithOwnSeedCount(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	assert.Equal(t, 0, g.getSeedCost(p0))
	g.placeTree(p0, 1, TREE_SEED)
	g.placeTree(p0, 2, TREE_SEED)
	assert.Equal(t, 2, g.getSeedCost(p0))
	g.placeTree(p1, 3, TREE_SEED)
	assert.Equal(t, 2, g.getSeedCost(p0), "opponent seeds don't count")
	assert.Equal(t, 1, g.getSeedCost(p1))
}

// Rule: grow cost = base + same-size count of own trees of the *target* size.
func TestGrowCostScalesWithTargetSizeCount(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	subject := g.placeTree(p0, 0, TREE_SMALL) // grow target = MEDIUM, base=3
	assert.Equal(t, 3, g.getGrowthCost(subject))
	g.placeTree(p0, 1, TREE_MEDIUM)
	assert.Equal(t, 4, g.getGrowthCost(subject))
	g.placeTree(p0, 2, TREE_MEDIUM)
	assert.Equal(t, 5, g.getGrowthCost(subject))
	// Opponent trees of target size do not affect cost.
	g.placeTree(g.Players[1], 3, TREE_MEDIUM)
	assert.Equal(t, 5, g.getGrowthCost(subject))
}

func TestGrowCostBaseValuesPerSize(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	seed := g.placeTree(p0, 0, TREE_SEED)
	assert.Equal(t, 1, g.getGrowthCost(seed), "grow seed → small base 1")
	g.removeTree(0)

	small := g.placeTree(p0, 0, TREE_SMALL)
	assert.Equal(t, 3, g.getGrowthCost(small), "grow small → medium base 3")
	g.removeTree(0)

	medium := g.placeTree(p0, 0, TREE_MEDIUM)
	assert.Equal(t, 7, g.getGrowthCost(medium), "grow medium → tall base 7")
}

// Rule: completing a lifecycle costs LIFECYCLE_END_COST (4 sun).
func TestCompleteCostIsLifecycleEndConstant(t *testing.T) {
	g := newTestGame()
	tall := g.placeTree(g.Players[0], 0, TREE_TALL)
	assert.Equal(t, LIFECYCLE_END_COST, g.getGrowthCost(tall))
}

// --- Action: GROW -----------------------------------------------------------

func TestGrowSucceedsAndConsumesSunAndSetsDormant(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 5
	tree := g.placeTree(p0, 0, TREE_SMALL) // cost = 3
	g.ResetGameTurnData()
	p0.SetAction(NewGrowAction(0))
	g.performActionUpdate()
	assert.Equal(t, 2, p0.Sun)
	assert.Equal(t, TREE_MEDIUM, tree.Size)
	assert.True(t, tree.Dormant)
}

// Rule: trees can grow up to size 3.
func TestGrowFailsOnTallTree(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	tree := g.placeTree(p0, 0, TREE_TALL)
	g.ResetGameTurnData()
	p0.SetAction(NewGrowAction(0))
	g.performActionUpdate()
	assert.Equal(t, TREE_TALL, tree.Size, "tall tree cannot grow further")
	assert.True(t, p0.IsWaiting(), "invalid action puts player to sleep")
}

// Rule: cannot afford → action rejected.
func TestGrowFailsWithoutEnoughSun(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 0
	tree := g.placeTree(p0, 0, TREE_SMALL)
	g.ResetGameTurnData()
	p0.SetAction(NewGrowAction(0))
	g.performActionUpdate()
	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

// Rule: can only act on your own trees.
func TestGrowFailsOnOpponentTree(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 100
	tree := g.placeTree(p1, 0, TREE_SMALL)
	g.ResetGameTurnData()
	p0.SetAction(NewGrowAction(0))
	g.performActionUpdate()
	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

// Rule: dormant trees cannot be the subject of an action.
func TestGrowFailsOnDormantTree(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	tree := g.placeTree(p0, 0, TREE_SMALL)
	tree.SetDormant()
	g.ResetGameTurnData()
	p0.SetAction(NewGrowAction(0))
	g.performActionUpdate()
	assert.Equal(t, TREE_SMALL, tree.Size)
	assert.True(t, p0.IsWaiting())
}

// --- Action: SEED -----------------------------------------------------------

// Rule: seed action plants a dormant seed and the source becomes dormant too.
func TestSeedSucceedsPlantsDormantSeedAndDormantsSource(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 5
	src := g.placeTree(p0, 0, TREE_SMALL)
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 1))
	g.performActionUpdate()
	assert.Equal(t, 5, p0.Sun, "no seeds yet → cost 0")
	assert.True(t, src.Dormant, "source tree is dormant")
	planted, ok := g.Trees[1]
	require.True(t, ok)
	assert.Equal(t, TREE_SEED, planted.Size)
	assert.True(t, planted.Dormant, "planted seed is dormant the rest of the day")
	assert.Equal(t, 0, planted.FatherIndex)
}

// Rule: seed range = source tree's size.
func TestSeedDistanceLimitedBySourceSize(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	src := g.placeTree(p0, 0, TREE_SMALL) // distance limit = 1
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 7)) // cell 7 is distance 2 from cell 0
	g.performActionUpdate()
	assert.True(t, p0.IsWaiting())
	assert.False(t, src.Dormant, "rejected action does not consume source")
	_, planted := g.Trees[7]
	assert.False(t, planted)
}

// Rule: cannot seed onto an unusable cell.
func TestSeedRejectedOnUnusableCell(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	g.Board.Cells[1].SetRichness(RICHNESS_NULL)
	g.placeTree(p0, 0, TREE_SMALL)
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 1))
	g.performActionUpdate()
	assert.True(t, p0.IsWaiting())
	_, planted := g.Trees[1]
	assert.False(t, planted)
}

// Rule: cannot seed onto a cell that already holds a tree.
func TestSeedRejectedOnOccupiedCell(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_SMALL)
	g.placeTree(p0, 1, TREE_SEED) // already taken
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 1))
	g.performActionUpdate()
	assert.True(t, p0.IsWaiting())
	tree, ok := g.Trees[1]
	require.True(t, ok)
	assert.Equal(t, TREE_SEED, tree.Size, "existing tree untouched")
}

// Rule: a seed (size 0) cannot be the source of a SEED action.
func TestSeedRejectedFromSeedSource(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_SEED)
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 1))
	g.performActionUpdate()
	assert.True(t, p0.IsWaiting())
}

// Rule: when both players seed the same cell on the same turn, neither plants
// and sun is refunded — but each source tree still becomes dormant.
func TestSeedConflictRefundsSunButKeepsSourcesDormant(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 10
	p1.Sun = 10
	// Add an extra seed each so cost > 0 → refund is observable.
	g.placeTree(p0, 2, TREE_SEED)
	g.placeTree(p1, 8, TREE_SEED)
	src0 := g.placeTree(p0, 0, TREE_SMALL) // cell 0 → cell 1 dist 1
	src1 := g.placeTree(p1, 7, TREE_SMALL) // cell 7 → cell 1 dist 1
	g.ResetGameTurnData()
	p0.SetAction(NewSeedAction(0, 1))
	p1.SetAction(NewSeedAction(7, 1))
	g.performActionUpdate()
	assert.Equal(t, 10, p0.Sun, "sun refunded on conflict")
	assert.Equal(t, 10, p1.Sun, "sun refunded on conflict")
	assert.True(t, src0.Dormant, "source still dormant")
	assert.True(t, src1.Dormant, "source still dormant")
	_, planted := g.Trees[1]
	assert.False(t, planted, "no seed planted on conflict")
}

// Helper / unit on the conflict detector.
func TestSeedsAreConflictingDetectsSameTarget(t *testing.T) {
	g := newTestGame()
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

// --- Action: COMPLETE -------------------------------------------------------

// Rule: complete pays 4 sun and awards nutrients + richness bonus.
func TestCompleteAwardsNutrientsAndRichnessBonusLush(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 4
	g.placeTree(p0, 0, TREE_TALL) // cell 0 = lush → +4
	g.ResetGameTurnData()
	p0.SetAction(NewCompleteAction(0))
	g.performActionUpdate()
	assert.Equal(t, 20+RICHNESS_BONUS_LUSH, p0.GetScore())
	assert.Equal(t, 0, p0.Sun)
	assert.Equal(t, 19, g.Nutrients, "nutrients drop by 1")
	_, exists := g.Trees[0]
	assert.False(t, exists, "tree removed from forest")
}

// Rule: richness bonuses are 0 / +2 / +4.
func TestCompleteRichnessBonusesPerCellQuality(t *testing.T) {
	cases := []struct {
		name     string
		cellIdx  int
		expected int
	}{
		{"poor outer ring", 19, 20 + 0},
		{"medium ring", 7, 20 + RICHNESS_BONUS_OK},
		{"lush center", 0, 20 + RICHNESS_BONUS_LUSH},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGame()
			p0 := g.Players[0]
			p0.Sun = 4
			g.placeTree(p0, tc.cellIdx, TREE_TALL)
			g.ResetGameTurnData()
			p0.SetAction(NewCompleteAction(tc.cellIdx))
			g.performActionUpdate()
			assert.Equal(t, tc.expected, p0.GetScore())
		})
	}
}

// Rule: only size-3 trees can be completed.
func TestCompleteRejectedOnNonTallTree(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 100
	g.placeTree(p0, 0, TREE_MEDIUM)
	g.ResetGameTurnData()
	p0.SetAction(NewCompleteAction(0))
	g.performActionUpdate()
	_, exists := g.Trees[0]
	assert.True(t, exists)
	assert.True(t, p0.IsWaiting())
}

// Rule: not enough sun → action rejected.
func TestCompleteRejectedWithoutEnoughSun(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 3 // < 4
	g.placeTree(p0, 0, TREE_TALL)
	g.ResetGameTurnData()
	p0.SetAction(NewCompleteAction(0))
	g.performActionUpdate()
	_, exists := g.Trees[0]
	assert.True(t, exists, "tree not removed when action rejected")
	assert.Equal(t, 0, p0.GetScore())
	assert.True(t, p0.IsWaiting())
}

// Rule: when both players complete the same turn each gets full points and
// nutrients drop by 2.
func TestBothPlayersCompleteSameTurnNutrientsDropByTwo(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.Sun = 4
	p1.Sun = 4
	g.placeTree(p0, 0, TREE_TALL) // lush +4
	g.placeTree(p1, 7, TREE_TALL) // medium +2
	g.ResetGameTurnData()
	p0.SetAction(NewCompleteAction(0))
	p1.SetAction(NewCompleteAction(7))
	g.performActionUpdate()
	assert.Equal(t, 20+RICHNESS_BONUS_LUSH, p0.GetScore())
	assert.Equal(t, 20+RICHNESS_BONUS_OK, p1.GetScore())
	assert.Equal(t, 18, g.Nutrients, "nutrients dropped by 2 (one per dying tree)")
}

// Rule: nutrients value cannot drop below 0.
func TestNutrientsClampedAtZero(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	p0.Sun = 4
	g.Nutrients = 0
	g.placeTree(p0, 19, TREE_TALL)
	g.ResetGameTurnData()
	p0.SetAction(NewCompleteAction(19))
	g.performActionUpdate()
	assert.Equal(t, 0, g.Nutrients, "stays clamped at 0")
	assert.Equal(t, 0, p0.GetScore(), "0 nutrients + 0 poor bonus")
}

// --- Game lifecycle ---------------------------------------------------------

// Rule: game ends after MAX_ROUNDS days.
func TestGameOverAfterMaxRounds(t *testing.T) {
	g := newTestGame()
	assert.False(t, g.IsGameOver())
	g.Round = g.MAX_ROUNDS
	assert.True(t, g.IsGameOver())
}

// Rule: end-of-game bonus: +1 score per 3 sun.
func TestOnEndAddsFloorSunOverThree(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(10)
	p1.SetScore(20)
	p0.Sun = 7
	p1.Sun = 9
	g.OnEnd()
	assert.Equal(t, 12, p0.GetScore(), "10 + floor(7/3)")
	assert.Equal(t, 23, p1.GetScore(), "20 + floor(9/3)")
}

// Rule: tiebreak by tree count when scores are equal (seeds count as trees).
func TestOnEndTiebreakAddsBonusPerTree(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(10)
	p0.Sun = 0
	p1.SetScore(10)
	p1.Sun = 0
	g.placeTree(p0, 1, TREE_SMALL)
	g.placeTree(p0, 2, TREE_SEED) // seed counts as a tree
	g.placeTree(p1, 3, TREE_SMALL)
	g.OnEnd()
	assert.Equal(t, 12, p0.GetScore())
	assert.Equal(t, 11, p1.GetScore())
	assert.Equal(t, 2, p0.BonusScore)
	assert.Equal(t, 1, p1.BonusScore)
}

// Negative case: no tiebreak when scores differ after the sun bonus.
func TestOnEndNoTiebreakWhenScoresDiffer(t *testing.T) {
	g := newTestGame()
	p0, p1 := g.Players[0], g.Players[1]
	p0.SetScore(15)
	p0.Sun = 0
	p1.SetScore(10)
	p1.Sun = 0
	g.placeTree(p0, 0, TREE_SMALL)
	g.placeTree(p1, 1, TREE_SMALL)
	g.OnEnd()
	assert.Equal(t, 15, p0.GetScore())
	assert.Equal(t, 10, p1.GetScore())
	assert.Equal(t, 0, p0.BonusScore)
	assert.Equal(t, 0, p1.BonusScore)
}

// --- Frame transitions ------------------------------------------------------

// At the start of a new day, dormant trees are reset before sun is given.
func TestSunGatheringResetsDormantTrees(t *testing.T) {
	g := newTestGame()
	p0 := g.Players[0]
	tree := g.placeTree(p0, 25, TREE_MEDIUM) // far enough that nothing shadows it
	tree.SetDormant()
	g.performSunGatheringUpdate()
	assert.False(t, tree.Dormant, "trees reset between days")
	assert.Equal(t, 2, p0.Sun, "tree gathered sun = its size")
}
