package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Acceptance tests exercise the full rules of SpringChallenge2020 on
// hand-crafted grids, bypassing random map generation. Each scenario asserts
// a single rule observable through the public Game API so any divergence from
// the Java reference is caught here without depending on seed parity.

// newScenario builds a Game pre-populated with a crafted grid. The caller
// then attaches pacmen and drives turns via runTurn.
func newScenario(leagueLevel int, rows []string, mapWraps bool) *Game {
	g := NewGame(0, leagueLevel)
	// Override MapWraps on the config to match the grid we're building.
	g.Config.MAP_WRAPS = mapWraps
	g.Grid = NewGridFromRows(rows, mapWraps)
	g.Players = []*Player{NewPlayer(0), NewPlayer(1)}
	return g
}

// spawn creates and registers a pacman at pos for player idx.
func spawn(g *Game, idx, number int, t PacmanType, pos Coord) *Pacman {
	owner := g.Players[idx]
	pac := NewPacman(len(g.Pacmen), number, owner, t)
	pac.Position = pos
	owner.AddPacman(pac)
	g.Pacmen = append(g.Pacmen, pac)
	return pac
}

// runTurn ticks cooldowns/durations, applies the provided setup to set fresh
// intents, then runs one main turn. PerformGameUpdate folds any speed
// sub-step into the same call.
func runTurn(g *Game, setup func()) {
	g.ResetGameTurnData()
	setup()
	g.PerformGameUpdate()
}

// ——— movement / pellets ————————————————————————————————————————————————————

func TestMoveFirstStepOfShortestPathOnly(t *testing.T) {
	// Rules: "MOVE gives a target; pac moves the first step of the shortest route."
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})

	assert.Equal(t, Coord{X: 2, Y: 1}, pac.Position, "one step toward target")
	assert.Equal(t, 1, g.Players[0].Pellets, "pellet at destination eaten for 1 pt")
}

func TestSuperPelletIsWorthTenPoints(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	// NewGridFromRows treats 'o' as a regular pellet; cherries are marked via
	// Cell.HasCherry directly.
	g.Grid.Get(Coord{X: 2, Y: 1}).HasCherry = true
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 3, Y: 1})
	})

	assert.Equal(t, Coord{X: 2, Y: 1}, pac.Position)
	assert.Equal(t, CHERRY_SCORE, g.Players[0].Pellets)
	assert.False(t, g.Grid.Get(Coord{X: 2, Y: 1}).HasCherry)
}

func TestTwoFriendlyPacsStackedCreditOwnerOnce(t *testing.T) {
	// Rules: a pellet is eaten once; multiple pacs of the same player on the
	// same cell do NOT score the pellet twice.
	g := newScenario(4, []string{
		"#####",
		"#...#",
		"#####",
	}, false)
	// Both pacs sit on top of the pellet at (2,1) — hand-place, bypass spawn.
	pac0 := spawn(g, 0, 0, TypeRock, Coord{X: 2, Y: 1})
	pac1 := spawn(g, 0, 1, TypePaper, Coord{X: 2, Y: 1})

	runTurn(g, func() {
		pac0.Intent = NoAction
		pac1.Intent = NoAction
	})

	assert.Equal(t, 1, g.Players[0].Pellets, "one pellet → one point")
	assert.False(t, g.Grid.Get(Coord{X: 2, Y: 1}).HasPellet)
}

// ——— wrapping ————————————————————————————————————————————————————————————

func TestHorizontalWrappingShortestPath(t *testing.T) {
	// Rules: "pacs can wrap around the map and appear on the other side."
	// Path via wrap from x=0 to x=8 is 1 step instead of 8.
	g := newScenario(4, []string{
		"#########",
		"         ",
		"#########",
	}, true)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 0, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 8, Y: 1})
	})

	assert.Equal(t, Coord{X: 8, Y: 1}, pac.Position, "wrap to the right edge")
}

// ——— rock-paper-scissors combat ———————————————————————————————————————————

func rpsKillCase(t *testing.T, attackerType, victimType PacmanType) {
	t.Helper()
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	attacker := spawn(g, 0, 0, attackerType, Coord{X: 1, Y: 1})
	victim := spawn(g, 1, 0, victimType, Coord{X: 3, Y: 1})

	runTurn(g, func() {
		attacker.Intent = NewMoveAction(Coord{X: 3, Y: 1})
		victim.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	})

	// Both aimed at one-step destinations that collide at (2,1).
	assert.Equal(t, Coord{X: 2, Y: 1}, attacker.Position, "attacker advances")
	assert.True(t, victim.Dead, "victim killed by attacker type")
	assert.True(t, g.Players[1].IsDeactivated(), "no pacs → deactivated")
}

func TestRockEatsScissors(t *testing.T)  { rpsKillCase(t, TypeRock, TypeScissors) }
func TestPaperEatsRock(t *testing.T)     { rpsKillCase(t, TypePaper, TypeRock) }
func TestScissorsEatsPaper(t *testing.T) { rpsKillCase(t, TypeScissors, TypePaper) }

func TestSameTypeCollisionBothBlocked(t *testing.T) {
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

	assert.Equal(t, Coord{X: 1, Y: 1}, a.Position, "same type is body-blocked")
	assert.Equal(t, Coord{X: 3, Y: 1}, b.Position)
	assert.False(t, a.Dead)
	assert.False(t, b.Dead)
}

func TestFriendlyPacsCannotOccupySameCell(t *testing.T) {
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

	assert.Equal(t, Coord{X: 1, Y: 1}, a.Position, "friendly body-block")
	assert.Equal(t, Coord{X: 3, Y: 1}, b.Position)
}

// ——— SWITCH ability ——————————————————————————————————————————————————————

func TestSwitchChangesTypeAndSetsCooldown(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypePaper)
		pac.AbilityToUse = AbilitySetPaper
		pac.HasAbilityToUse = true
	})

	assert.Equal(t, TypePaper, pac.Type)
	assert.Equal(t, g.Config.ABILITY_COOLDOWN, pac.AbilityCooldown)
}

func TestSwitchBlockedByCooldown(t *testing.T) {
	g := newScenario(4, []string{"###", "# #", "###"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	// First switch: ROCK → PAPER. Cooldown → 10.
	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypePaper)
		pac.AbilityToUse = AbilitySetPaper
		pac.HasAbilityToUse = true
	})
	assert.Equal(t, TypePaper, pac.Type)

	// Second switch attempt on the next turn: cooldown was 10, ticks to 9 in
	// TurnReset, still > 0 at executePacmenAbilities so it gets ignored.
	runTurn(g, func() {
		pac.Intent = NewSwitchAction(TypeScissors)
		pac.AbilityToUse = AbilitySetScissors
		pac.HasAbilityToUse = true
	})

	assert.Equal(t, TypePaper, pac.Type, "second switch ignored during cooldown")
	assert.Equal(t, 9, pac.AbilityCooldown, "cooldown ticks once per turn")
}

// ——— SPEED ability ———————————————————————————————————————————————————————

func TestSpeedActivationSetsDurationAndCooldown(t *testing.T) {
	g := newScenario(4, []string{"#######", "#     #", "#######"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewSpeedAction()
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	})

	assert.Equal(t, g.Config.SPEED_BOOST, pac.Speed, "speed boosted")
	assert.Equal(t, g.Config.ABILITY_DURATION, pac.AbilityDuration)
	assert.Equal(t, g.Config.ABILITY_COOLDOWN, pac.AbilityCooldown)
	// The SPEED action itself is not a MOVE — pac stays put.
	assert.Equal(t, Coord{X: 1, Y: 1}, pac.Position)
}

func TestSpeedSubTurnDelivers2Steps(t *testing.T) {
	g := newScenario(4, []string{"#######", "#     #", "#######"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	// Turn 1: activate SPEED.
	runTurn(g, func() {
		pac.Intent = NewSpeedAction()
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	})

	// Turn 2: MOVE — both movement steps resolve in the same PerformGameUpdate.
	runTurn(g, func() {
		pac.Intent = NewMoveAction(Coord{X: 5, Y: 1})
	})
	assert.Equal(t, Coord{X: 3, Y: 1}, pac.Position, "two steps in one turn")
	assert.False(t, g.IsSpeedTurn(), "no follow-up sub-turn pending")
}

func TestSpeedExpiresAfterDurationTicks(t *testing.T) {
	// Duration starts at 6 (AbilityDur). After that many cooldown ticks the
	// pac's speed reverts to base.
	g := newScenario(4, []string{"#######", "#     #", "#######"}, false)
	pac := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})

	runTurn(g, func() {
		pac.Intent = NewSpeedAction()
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	})
	assert.Equal(t, 2, pac.Speed)

	// Tick through AbilityDur WAIT turns.
	for i := 0; i < g.Config.ABILITY_DURATION; i++ {
		runTurn(g, func() {
			pac.Intent = NoAction
		})
	}

	assert.Equal(t, g.Config.PACMAN_BASE_SPEED, pac.Speed, "speed reset after duration")
	assert.Equal(t, 0, pac.AbilityDuration)
}

// ——— game over ——————————————————————————————————————————————————————————

func TestGameOverAwardsRemainingPelletsToSurvivor(t *testing.T) {
	// Rules: "If all of a player's pacs are dead, all remaining pellets are
	// automatically scored by any surviving pacs."
	g := newScenario(4, []string{
		"#######",
		"#     #",
		"# ... #",
		"#######",
	}, false)
	attacker := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	victim := spawn(g, 1, 0, TypeScissors, Coord{X: 3, Y: 1})

	runTurn(g, func() {
		attacker.Intent = NewMoveAction(Coord{X: 3, Y: 1})
		victim.Intent = NewMoveAction(Coord{X: 1, Y: 1})
	})

	assert.True(t, victim.Dead)
	assert.True(t, g.Players[1].IsDeactivated())
	assert.True(t, g.IsGameOver())

	g.PerformGameOver()

	// Row 2 has three pellets → awarded wholesale to the surviving player.
	assert.Equal(t, 3, g.Players[0].Pellets)
}

func TestRefereeStopsAfterTwoHundredMainTurns(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"#   #",
		"#####",
	}, false)
	g.Grid.Get(Coord{X: 2, Y: 1}).HasPellet = true
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeRock, Coord{X: 3, Y: 1})
	r := NewReferee(g)

	for turn := 0; turn < MaxMainTurns; turn++ {
		r.ResetGameTurnData()
		r.PerformGameUpdate(turn)
	}

	assert.False(t, r.Ended())
	assert.True(t, r.GameOverFrame)
	assert.Equal(t, MaxMainTurns, r.MainTurns)

	r.PerformGameUpdate(MaxMainTurns)
	assert.True(t, r.Ended())
}

func TestRefereeEndGameAwardsRemainingPelletsAfterEarlyArenaStop(t *testing.T) {
	// The arena runner may call EndGame immediately after a player is
	// deactivated. Spring 2020 still needs the Java performGameOver absorption.
	g := newScenario(4, []string{
		"#####",
		"#...#",
		"#####",
	}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeScissors, Coord{X: 3, Y: 1})
	g.Players[1].Deactivate("bad command")

	r := NewReferee(g)
	r.EndGame()
	r.EndGame()

	assert.True(t, r.Ended())
	assert.Equal(t, 3, g.Players[0].Pellets, "remaining pellets awarded exactly once")
}

func TestCanImproveRankingEndsGameWhenPelletsCannotCloseGap(t *testing.T) {
	g := newScenario(4, []string{"####", "#  #", "####"}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeRock, Coord{X: 2, Y: 1})

	// No pellets remain (grid has only spaces). Ranking cannot change.
	assert.True(t, g.IsGameOver())

	// With a pellet introduced, ranking can still change → not over.
	g.Grid.Get(Coord{X: 1, Y: 1}).HasPellet = true
	assert.False(t, g.IsGameOver())
}

// ——— fog of war / line of sight ———————————————————————————————————————————

func TestFogOfWarHidesPelletBehindWall(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"# # #",
		"#####",
	}, false)
	// Put a pellet in the far cell, then a wall blocks line of sight.
	g.Grid.Get(Coord{X: 3, Y: 1}).HasPellet = true
	// Attack pacs must be set up for both players for serialization to work.
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeRock, Coord{X: 3, Y: 1})

	lines := SerializeFrameInfoFor(g.Players[0], g)
	joined := strings.Join(lines, "\n")
	assert.NotContains(t, joined, "3 1 1", "pellet behind wall must not appear")
}

func TestFogOfWarCherriesAlwaysVisible(t *testing.T) {
	// Super-pellets are "bright" — visible even through walls in the rules,
	// and the engine achieves that by returning them via AllCherries regardless
	// of line-of-sight.
	g := newScenario(4, []string{
		"#####",
		"# # #",
		"#####",
	}, false)
	g.Grid.Get(Coord{X: 3, Y: 1}).HasCherry = true
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypeRock, Coord{X: 3, Y: 1})

	lines := SerializeFrameInfoFor(g.Players[0], g)
	joined := strings.Join(lines, "\n")
	assert.Contains(t, joined, "3 1 10", "cherry visible through walls")
}

func TestFogOfWarEnemyPacInvisibleBehindWall(t *testing.T) {
	g := newScenario(4, []string{
		"#####",
		"# # #",
		"#####",
	}, false)
	spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	spawn(g, 1, 0, TypePaper, Coord{X: 3, Y: 1})

	lines := SerializeFrameInfoFor(g.Players[0], g)
	// Line count: "0 0", "<visible pac count>", then that many pac lines.
	// Only own pac is visible.
	assert.Equal(t, "1", lines[1], "only one pac visible")
}

// ——— serialization / referee smoke test ————————————————————————————————————

func TestReferePerformGameUpdateAdvancesScoreViaCommandParse(t *testing.T) {
	g := newScenario(4, []string{
		"#######",
		"#.....#",
		"#######",
	}, false)
	// Directly craft initial state without Init.
	pac0 := spawn(g, 0, 0, TypeRock, Coord{X: 1, Y: 1})
	pac1 := spawn(g, 1, 0, TypePaper, Coord{X: 5, Y: 1})
	_ = pac1

	r := NewReferee(g)

	asArena := []arena.Player{g.Players[0], g.Players[1]}

	// Simulate a match loop tick: ResetGameTurnData → ParseCommands → PerformGameUpdate.
	r.ResetGameTurnData()
	g.Players[0].SetOutputs([]string{"MOVE 0 5 1"})
	g.Players[1].SetOutputs([]string{"MOVE 0 1 1"})
	r.ParsePlayerOutputs(asArena)

	// Before update, no movement.
	assert.Equal(t, Coord{X: 1, Y: 1}, pac0.Position)

	r.PerformGameUpdate(0)

	assert.Equal(t, Coord{X: 2, Y: 1}, pac0.Position)
	assert.Equal(t, Coord{X: 4, Y: 1}, pac1.Position)
	assert.Equal(t, 1, g.Players[0].Pellets)
	assert.Equal(t, 1, g.Players[1].Pellets)
}
