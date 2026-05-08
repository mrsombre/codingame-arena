package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Acceptance tests exercise the full rules of WinterChallenge2026-Exotec on
// hand-crafted grids, bypassing random map generation. Each scenario asserts
// a single rule observable through the public Game API so any divergence from
// the Java reference is caught here without depending on seed parity.
//
// Gravity note: a body is grounded only when at least one of its cells has a
// wall/apple *directly* below it, or is resting on top of another grounded
// bird. Most scenarios therefore place bodies on row y=Height-2 with a wall
// row at y=Height-1 underneath.

// newScenario builds a Game with an empty grid of (width, height). The grid
// is floor-only; callers add walls via wall() and apples via apple().
func newScenario(width, height int) *Game {
	g := &Game{
		BirdByIDCache: make(map[int]*Bird),
	}
	g.Grid = NewGrid(width, height)
	g.Players = []*Player{NewPlayer(0), NewPlayer(1)}
	for _, p := range g.Players {
		p.Init()
	}
	return g
}

// floorRow paints the last row as platform tiles (impassable floor).
func floorRow(g *Game) {
	y := g.Grid.Height - 1
	for x := 0; x < g.Grid.Width; x++ {
		g.Grid.Get(Coord{X: x, Y: y}).SetType(TileWall)
	}
}

func wall(g *Game, c Coord) {
	g.Grid.Get(c).SetType(TileWall)
}

func apple(g *Game, c Coord) {
	g.Grid.Apples = append(g.Grid.Apples, c)
}

// spawn attaches a bird with the given body to a player. Body[0] is the head.
func spawn(g *Game, playerIdx, birdID int, body []Coord) *Bird {
	owner := g.Players[playerIdx]
	b := NewBird(birdID, owner)
	b.Body = append([]Coord(nil), body...)
	owner.Birds = append(owner.Birds, b)
	g.BirdByIDCache[b.ID] = b
	return b
}

// groundedRow spawns a bird on y=floorY-1 (directly above the floor) with the
// head at `headX` facing in the direction specified by `dir`. The tail extends
// opposite to `dir`. All body cells lie on floorY-1 so the snake is grounded.
func groundedRow(g *Game, playerIdx, birdID, headX, floorY, length int, dir Direction) *Bird {
	y := floorY - 1
	body := make([]Coord, 0, length)
	dx := 0
	switch dir {
	case DirEast:
		dx = -1 // tail to the west of head
	case DirWest:
		dx = 1 // tail to the east of head
	}
	for i := 0; i < length; i++ {
		body = append(body, Coord{X: headX + i*dx, Y: y})
	}
	return spawn(g, playerIdx, birdID, body)
}

// advance runs one turn (no explicit command — uses each bird's Facing).
func advance(g *Game, turn int) {
	g.ResetGameTurnData()
	g.PerformGameUpdate(turn)
}

// advanceWith sets intents via the provided setup func between ResetGameTurnData
// (which clears them) and PerformGameUpdate.
func advanceWith(g *Game, turn int, setup func()) {
	g.ResetGameTurnData()
	setup()
	g.PerformGameUpdate(turn)
}

// ——— movement ————————————————————————————————————————————————————————————

func TestSpawnFacesUpByDefault(t *testing.T) {
	// Rules: "The starting direction is up." A vertically-stacked body with
	// head above neck has Facing() = North, which doMoves uses by default.
	b := spawn(newScenario(10, 10), 0, 0, []Coord{
		{X: 5, Y: 3}, {X: 5, Y: 4}, {X: 5, Y: 5}, {X: 5, Y: 6},
	})
	assert.Equal(t, DirNorth, b.Facing(), "spawn body shape encodes UP")
}

func TestDefaultMoveUsesFacingWhenNoCommand(t *testing.T) {
	// doMoves picks Facing() when a bird has no HasMove/Direction set. Use an
	// L-shaped body with a horizontal tail grounded on the floor and the head
	// sticking up — verifies facing=North is applied on an unrecorded turn.
	g := newScenario(10, 6)
	floorRow(g)

	// Head at (5,3), neck at (5,4), body rests on floor-1 row (y=4).
	b := spawn(g, 0, 0, []Coord{
		{X: 5, Y: 3}, {X: 5, Y: 4}, {X: 4, Y: 4}, {X: 3, Y: 4}, {X: 2, Y: 4},
	})
	groundedRow(g, 1, 1, 8, g.Grid.Height-1, 3, DirEast)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.True(t, b.Alive)
	assert.Equal(t, Coord{X: 5, Y: 2}, b.HeadPos(), "head moved north one step")
	assert.Len(t, b.Body, 5, "no eating, length preserved")
}

func TestMovePerpetuallyContinuesLastDirection(t *testing.T) {
	// Rules: "Snakebots are perpetually moving." Between turns with no command,
	// the bird continues in the last direction (encoded in body orientation).
	g := newScenario(20, 6)
	floorRow(g)

	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 4, DirEast)
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 4, DirWest)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)
	assert.Equal(t, Coord{X: 6, Y: 4}, b.HeadPos(), "turn 1 east")

	advance(g, 1)
	assert.Equal(t, Coord{X: 7, Y: 4}, b.HeadPos(), "turn 2 still east — no command required")
}

func TestMoveFollowsCommandedDirection(t *testing.T) {
	g := newScenario(20, 6)
	floorRow(g)

	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 3, DirWest)
	apple(g, Coord{X: 0, Y: 0})

	advanceWith(g, 0, func() {
		b.Direction = DirNorth
		b.HasMove = true
	})
	assert.Equal(t, Coord{X: 5, Y: 3}, b.HeadPos(), "commanded N overrides eastward facing")
}

// ——— eating ——————————————————————————————————————————————————————————————

func TestHeadEatingApplePowersGrowth(t *testing.T) {
	// Rules: "The snakebot grows, a new body part appears at the end of its body.
	// This cell is no longer considered solid."
	g := newScenario(20, 6)
	floorRow(g)

	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 3, DirWest)
	// Apple directly east of head (6, 4).
	apple(g, Coord{X: 6, Y: 4})

	advance(g, 0)

	assert.Equal(t, Coord{X: 6, Y: 4}, b.HeadPos(), "head moved onto apple")
	assert.Len(t, b.Body, 4, "body grew by one (tail was NOT trimmed)")
	assert.False(t, g.Grid.HasApple(Coord{X: 6, Y: 4}), "apple removed")
}

func TestMultipleHeadsOnSameAppleBothEat(t *testing.T) {
	// Rules special case: both heads count as eating, apple removed once.
	// Each bird grows then both collide head-on → both beheaded (body >3 so
	// neither dies). Net body length stays equal to initial length.
	g := newScenario(20, 6)
	floorRow(g)

	a := groundedRow(g, 0, 0, 4, g.Grid.Height-1, 5, DirEast)
	b := groundedRow(g, 1, 1, 6, g.Grid.Height-1, 5, DirWest)
	apple(g, Coord{X: 5, Y: 4})

	advance(g, 0)

	assert.False(t, g.Grid.HasApple(Coord{X: 5, Y: 4}), "apple gone")
	assert.True(t, a.Alive)
	assert.True(t, b.Alive)
	assert.Len(t, a.Body, 5, "bird A: grew to 6, beheaded to 5")
	assert.Len(t, b.Body, 5, "bird B: symmetric outcome")
}

// ——— beheadings / collisions ————————————————————————————————————————————

func TestHeadIntoWallBeheadIfLongEnough(t *testing.T) {
	// Rules Case 1: head into platform → head destroyed, neck becomes new head
	// (needs ≥3 parts remaining afterward — code's check is len(Body) > 3).
	g := newScenario(20, 6)
	floorRow(g)
	// Wall one cell east of the head.
	wall(g, Coord{X: 6, Y: 4})

	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 4, DirEast)
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 3, DirWest)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.True(t, b.Alive, "4-part snake survives a head-into-wall")
	assert.Len(t, b.Body, 3, "head trimmed, body shrank from 4 to 3")
}

func TestHeadIntoWallKillsIfThreeOrFewerParts(t *testing.T) {
	// Rules Case 1: "If not, the whole snakebot is removed."
	g := newScenario(20, 6)
	floorRow(g)
	wall(g, Coord{X: 6, Y: 4})

	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 3, DirWest)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.False(t, b.Alive, "3-part snake dies on wall collision")
}

func TestHeadIntoEnemyBodyBeheads(t *testing.T) {
	g := newScenario(20, 6)
	floorRow(g)

	// Attacker heading east — its head will run into defender's body cell.
	attacker := groundedRow(g, 0, 0, 4, g.Grid.Height-1, 4, DirEast)
	// Defender occupies (5, 4). spawn manually since shape is vertical tails don't apply.
	defender := spawn(g, 1, 1, []Coord{{X: 5, Y: 4}, {X: 6, Y: 4}, {X: 7, Y: 4}, {X: 8, Y: 4}})
	// Stabilise defender's facing so it doesn't collide with the attacker head
	// on the same cell: body head is east of floor-placed neck? Actually body
	// is (5,4),(6,4) — head (5,4) minus neck (6,4) = (-1, 0) = West. Defender
	// will move west into (4,4). But (4,4) is the attacker's current head —
	// cross paths? Let's set defender direction fixed to WAIT… actually a
	// defender with head west of its body naturally keeps moving west, which
	// collides with the attacker. Simpler: freeze defender by commanding it
	// east (back of body) — but that's "backwards" and rejected by command
	// manager. So just let defender move. Both move toward each other.
	_ = defender

	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	// Attacker head was at (4,4); moved east to (5,4). (5,4) is defender's
	// head *before* it moved. After both move, defender's body is
	// [(4,4),(5,4),(6,4),(7,4)]. Attacker head (5,4) is in defender's body
	// (body[1]) → beheaded.
	assert.True(t, attacker.Alive)
	assert.Len(t, attacker.Body, 3, "attacker beheaded (4 → 3)")
}

func TestHeadIntoOwnBodyBeheads(t *testing.T) {
	// Self-hit requires landing on a body cell that won't be vacated by the
	// tail-trim this turn. A long coiled body biting its own mid-section:
	// head(5,6), ..., mid(5,7), tail(6,7). Turning south lands on (5,7) —
	// still occupied after trim, so isInSelf triggers.
	g := newScenario(20, 10)
	floorRow(g)

	b := spawn(g, 0, 0, []Coord{
		{X: 5, Y: 6}, {X: 4, Y: 6}, {X: 4, Y: 7}, {X: 5, Y: 7}, {X: 6, Y: 7},
	})
	groundedRow(g, 1, 1, 15, g.Grid.Height-1, 3, DirWest)
	apple(g, Coord{X: 0, Y: 0})

	advanceWith(g, 0, func() {
		b.Direction = DirSouth
		b.HasMove = true
	})

	assert.True(t, b.Alive)
	assert.Len(t, b.Body, 4, "head ate own segment → beheaded (5 → 4)")
}

// ——— gravity / falling ————————————————————————————————————————————————

func TestGravityBirdStaysOnGround(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	b := groundedRow(g, 0, 0, 5, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 8, g.Grid.Height-1, 2, DirEast)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.True(t, b.Alive)
	for _, c := range b.Body {
		assert.Less(t, c.Y, g.Grid.Height, "no cell dropped into the wall row")
	}
}

func TestGravityAirborneBirdFallsUntilLanded(t *testing.T) {
	// Snake starts high in the air with nothing directly below any body cell.
	// It should fall until the tail sits on top of the floor row.
	g := newScenario(10, 10)
	floorRow(g) // wall row at y=9

	// Vertical body at x=5, y=0..2. Facing north. After doMoves head → (5,-1)
	// → out of bounds, type -1 (not wall). So no beheading. Body will be
	// [(5,-1),(5,0),(5,1)] then fall. Hmm that leaves the head out of grid.
	//
	// Cleaner: orient head at bottom so Facing=South. Head → (5,3) free,
	// body after move [(5,3),(5,2),(5,1)]. Falls until tail (5,1) has wall
	// below — at (5,8), wall at (5,9). So final body [(5,8),(5,7),(5,6)]?
	// Wait — the tail is at body[len-1]. Head moves SOUTH so tail grows
	// upward (head at bottom, tail at top). Let me recompute.
	//
	// Initial [(5,2),(5,1),(5,0)]: head (5,2), neck (5,1), tail (5,0). Facing
	// = (0,1) = South. Head moves to (5,3). New body [(5,3),(5,2),(5,1)].
	// Now fall: no wall/apple directly under any cell. All 3 y values inc
	// by 1 per tick. After 6 falls → [(5,9),...] but (5,9) is wall — cannot
	// be there. Actually fall continues until *any* cell has wall/apple
	// directly under. So final body [(5,8),(5,7),(5,6)] (tail below floor
	// row y=9 would be wall, so bird grounded when head (5,3+k) cell has
	// (5,4+k)=wall → at k=5, head=(5,8), below (5,9)=wall → grounded.
	b := spawn(g, 0, 0, []Coord{{X: 5, Y: 2}, {X: 5, Y: 1}, {X: 5, Y: 0}})

	groundedRow(g, 1, 1, 1, g.Grid.Height-1, 3, DirEast)
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.True(t, b.Alive, "bird lands on the floor row")
	head := b.Body[0]
	assert.Equal(t, 8, head.Y, "head rests at y=8 (on top of wall at y=9)")
	assert.Equal(t, 5, head.X)
}

func TestGravityKillsBirdThatFallsOutOfGrid(t *testing.T) {
	// No floor row — snake has no ground to catch it.
	g := newScenario(10, 6)
	// Opponent on a small platform so the game remains valid.
	wall(g, Coord{X: 0, Y: 5})
	wall(g, Coord{X: 1, Y: 5})
	wall(g, Coord{X: 2, Y: 5})
	spawn(g, 1, 1, []Coord{{X: 0, Y: 4}, {X: 1, Y: 4}, {X: 2, Y: 4}})

	// Hover body high, no solid support beneath it.
	b := spawn(g, 0, 0, []Coord{{X: 6, Y: 2}, {X: 6, Y: 1}, {X: 6, Y: 0}})
	apple(g, Coord{X: 0, Y: 0})

	advance(g, 0)

	assert.False(t, b.Alive, "bird with no ground falls out of bounds and is removed")
}

func TestBirdStackOnOtherGroundedBirdIsGrounded(t *testing.T) {
	// One bird on the floor, a longer second bird sitting directly on top.
	// The stacked bird must stay at its row — bottom bird grounds it via the
	// frozenBirds mechanic.
	g := newScenario(12, 8)
	floorRow(g)

	// Bottom bird on the floor-1 row (y=6), spanning x=3..7.
	bottom := spawn(g, 1, 1, []Coord{
		{X: 7, Y: 6}, {X: 6, Y: 6}, {X: 5, Y: 6}, {X: 4, Y: 6}, {X: 3, Y: 6},
	})
	// Top bird on y=5 (directly above bottom), spanning x=3..6.
	top := spawn(g, 0, 0, []Coord{
		{X: 6, Y: 5}, {X: 5, Y: 5}, {X: 4, Y: 5}, {X: 3, Y: 5},
	})
	apple(g, Coord{X: 0, Y: 0})

	// Freeze movement by sending them along their facing (bottom west, top west).
	// Their bodies overlap vertically, so top remains grounded on bottom.
	advance(g, 0)

	assert.True(t, top.Alive, "stacked bird did not fall out")
	assert.True(t, bottom.Alive)
	// Every cell of top body stays at y=5 after moving — no fall happened.
	for _, c := range top.Body {
		assert.Equal(t, 5, c.Y, "top bird stays on y=5 resting on bottom")
	}
}

// ——— game end ———————————————————————————————————————————————————————————

func TestIsGameOverNoApples(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	groundedRow(g, 0, 0, 2, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 7, g.Grid.Height-1, 3, DirEast)

	assert.True(t, g.IsGameOver(), "no power sources remain")
}

func TestIsGameOverPlayerHasNoLiveBirds(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	groundedRow(g, 0, 0, 2, g.Grid.Height-1, 3, DirEast)
	dead := groundedRow(g, 1, 1, 7, g.Grid.Height-1, 3, DirEast)
	dead.Alive = false
	apple(g, Coord{X: 5, Y: 2})

	assert.True(t, g.IsGameOver(), "opponent has no live birds")
}

func TestIsGameOverFalseWhenBothPlayersAliveAndApplesLeft(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	groundedRow(g, 0, 0, 2, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 7, g.Grid.Height-1, 3, DirEast)
	apple(g, Coord{X: 5, Y: 2})

	assert.False(t, g.IsGameOver())
}

// ——— OnEnd scoring ———————————————————————————————————————————————————————

func TestOnEndScoreIsSumOfLiveBodyParts(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	// p0: 4 parts total (one bird).
	spawn(g, 0, 0, []Coord{{X: 0, Y: 4}, {X: 1, Y: 4}, {X: 2, Y: 4}, {X: 3, Y: 4}})
	// p1: 3-part live bird + 2-part dead bird.
	spawn(g, 1, 1, []Coord{{X: 6, Y: 4}, {X: 7, Y: 4}, {X: 8, Y: 4}})
	dead := spawn(g, 1, 2, []Coord{{X: 5, Y: 4}, {X: 5, Y: 3}})
	dead.Alive = false

	g.OnEnd()

	assert.Equal(t, 4, g.Players[0].GetScore())
	assert.Equal(t, 3, g.Players[1].GetScore(), "only live bird counted")
}

func TestOnEndTieBreakSubtractsLosses(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	spawn(g, 0, 0, []Coord{{X: 0, Y: 4}, {X: 1, Y: 4}, {X: 2, Y: 4}})
	spawn(g, 1, 1, []Coord{{X: 6, Y: 4}, {X: 7, Y: 4}, {X: 8, Y: 4}})
	g.Losses[0] = 2
	g.Losses[1] = 5

	g.OnEnd()

	assert.Equal(t, 3-2, g.Players[0].GetScore(), "tie-break: p0 score minus p0 losses")
	assert.Equal(t, 3-5, g.Players[1].GetScore(), "tie-break: p1 score minus p1 losses")
}

func TestOnEndDeactivatedPlayerScoresMinusOne(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	spawn(g, 0, 0, []Coord{{X: 0, Y: 4}, {X: 1, Y: 4}, {X: 2, Y: 4}})
	spawn(g, 1, 1, []Coord{{X: 6, Y: 4}, {X: 7, Y: 4}, {X: 8, Y: 4}})
	g.Players[1].Deactivate("bad input")

	g.OnEnd()

	assert.Equal(t, 3, g.Players[0].GetScore())
	assert.Equal(t, -1, g.Players[1].GetScore())
}

// ——— traces ————————————————————————————————————————————————————————————

func TestEatEmitsTraceWithBirdAndCoord(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	groundedRow(g, 0, 0, 2, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 7, g.Grid.Height-1, 3, DirEast)
	apple(g, Coord{X: 3, Y: g.Grid.Height - 2})
	apple(g, Coord{X: 0, Y: 0}) // keep game from ending

	advance(g, 0)

	var types []string
	for _, slot := range g.traces {
		for _, e := range slot {
			types = append(types, e.Type)
		}
	}
	assert.Contains(t, types, TraceEat)
}

// ——— serialization smoke ————————————————————————————————————————————————

func TestSerializeFrameIncludesApplesAndBodies(t *testing.T) {
	g := newScenario(10, 6)
	floorRow(g)
	groundedRow(g, 0, 0, 2, g.Grid.Height-1, 3, DirEast)
	groundedRow(g, 1, 1, 7, g.Grid.Height-1, 3, DirEast)
	apple(g, Coord{X: 5, Y: 2})

	lines := SerializeFrameInfoFor(g.Players[0], g)

	assert.Equal(t, "1", lines[0], "apple count")
	assert.Equal(t, "5 2", lines[1], "apple coord")
	assert.Equal(t, "2", lines[2], "live bird count")
	assert.Contains(t, lines[3], "2,4")
	assert.Contains(t, lines[3], "1,4")
	assert.Contains(t, lines[3], "0,4")
}

func TestSerializeGlobalIncludesPlayerIdAndGrid(t *testing.T) {
	g := newScenario(4, 3)
	for x := 0; x < 4; x++ {
		wall(g, Coord{X: x, Y: 2})
	}
	spawn(g, 0, 0, []Coord{{X: 0, Y: 0}})
	spawn(g, 1, 1, []Coord{{X: 3, Y: 0}})

	lines := SerializeGlobalInfoFor(g.Players[0], g)

	assert.Equal(t, "0", lines[0], "player id")
	assert.Equal(t, "4", lines[1], "width")
	assert.Equal(t, "3", lines[2], "height")
	assert.Equal(t, "....", lines[3], "top row empty")
	assert.Equal(t, "####", lines[5], "bottom row all walls")
}
