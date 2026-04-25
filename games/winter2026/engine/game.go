// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java
package engine

import (
	"sort"
	"strconv"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

type Game struct {
	players       []*Player
	grid          *grid.Grid
	birdByIDCache map[int]*Bird
	turn          int
	losses        [2]int
	ended         bool
	summary       []string
	events        []arena.TurnEvent
}

func NewGame(seed int64, leagueLevel int) *Game {
	g := &Game{
		birdByIDCache: make(map[int]*Bird),
	}
	g.initGrid(seed, leagueLevel)
	return g
}

func (g *Game) Init(players []*Player) {
	g.players = players
	g.initPlayers()
}

func (g *Game) initPlayers() {
	birdID := 0
	spawnLocations := g.findSpawnLocations()
	for _, p := range g.players {
		p.Init()
		for _, spawn := range spawnLocations {
			bird := NewBird(birdID, p)
			birdID++
			p.birds = append(p.birds, bird)
			for _, c := range spawn {
				if p.GetIndex() == 1 {
					c = g.grid.Opposite(c)
				}
				bird.Body = append(bird.Body, c)
				if len(bird.Body) == 1 {
					left := g.grid.Get(c.AddXY(-1, 0))
					right := g.grid.Get(c.AddXY(1, 0))
					if left.Type == grid.TileWall && right.Type == grid.TileWall {
						g.grid.Get(c.AddXY(-1, 0)).Clear()
						g.grid.Get(g.grid.Opposite(c.AddXY(-1, 0))).Clear()
					}
				}
			}
			g.birdByIDCache[bird.ID] = bird
		}
	}
}

func (g *Game) findSpawnLocations() [][]grid.Coord {
	islands := g.grid.DetectSpawnIslands()
	result := make([][]grid.Coord, 0, len(islands))
	for _, island := range islands {
		result = append(result, grid.SortedCoordsFromSet(island))
	}
	return result
}

func (g *Game) initGrid(seed int64, leagueLevel int) {
	gridMaker := grid.NewGridMaker(sha1prng.New(seed), leagueLevel)
	g.grid = gridMaker.Make()
}

func (g *Game) ResetGameTurnData() {
	for _, p := range g.players {
		p.Reset()
	}
}

func (g *Game) doMoves() {
	for _, p := range g.players {
		for _, bird := range p.birds {
			if !bird.Alive {
				continue
			}
			if !bird.HasMove || bird.Direction == grid.DirUnset {
				bird.Direction = bird.Facing()
			}

			newHead := bird.HeadPos().Add(bird.Direction.Coord())
			willEatApple := coordSliceContainsWinter(g.grid.Apples, newHead)
			if !willEatApple && len(bird.Body) > 0 {
				bird.Body = bird.Body[:len(bird.Body)-1]
			}
			bird.Body = append([]grid.Coord{newHead}, bird.Body...)
		}
	}
}

func (g *Game) allBirds() []*Bird {
	all := make([]*Bird, 0)
	for _, p := range g.players {
		all = append(all, p.birds...)
	}
	return all
}

func (g *Game) liveBirds() []*Bird {
	live := make([]*Bird, 0)
	for _, b := range g.allBirds() {
		if b.Alive {
			live = append(live, b)
		}
	}
	return live
}

func (g *Game) doBeheadings() {
	birdsToBehead := make([]*Bird, 0)
	for _, bird := range g.liveBirds() {
		isInWall := g.grid.Get(bird.HeadPos()).Type == grid.TileWall
		intersectingBirds := make([]*Bird, 0)
		for _, b := range g.allBirds() {
			if b.Alive && coordSliceContainsWinter(b.Body, bird.HeadPos()) {
				intersectingBirds = append(intersectingBirds, b)
			}
		}

		isInEnemy, isInSelf := false, false
		for _, b := range intersectingBirds {
			if b.ID != bird.ID {
				isInEnemy = true
			} else if coordSliceContainsWinter(b.Body[1:], b.HeadPos()) {
				isInSelf = true
			}
		}

		if isInWall {
			g.emit(EventHitWall, eventBirdCoordPayload(bird.ID, bird.HeadPos()))
		}
		if isInEnemy {
			g.emit(EventHitEnemy, eventBirdCoordPayload(bird.ID, bird.HeadPos()))
		}
		if isInSelf {
			g.emit(EventHitSelf, eventBirdCoordPayload(bird.ID, bird.HeadPos()))
		}

		if isInWall || isInEnemy || isInSelf {
			birdsToBehead = append(birdsToBehead, bird)
		}
	}

	for _, b := range birdsToBehead {
		if len(b.Body) <= 3 {
			b.Alive = false
			g.losses[b.Owner.GetIndex()] += len(b.Body)
			g.emit(EventDead, eventBirdPayload(b.ID))
		} else {
			b.Body = b.Body[1:]
			g.losses[b.Owner.GetIndex()]++
		}
	}
}

func (g *Game) doEats() {
	eaten := make(map[grid.Coord]struct{})
	for _, p := range g.players {
		for _, bird := range p.birds {
			if bird.Alive && coordSliceContainsWinter(g.grid.Apples, bird.HeadPos()) {
				g.emit(EventEat, eventBirdCoordPayload(bird.ID, bird.HeadPos()))
				eaten[bird.HeadPos()] = struct{}{}
			}
		}
	}
	if len(eaten) == 0 {
		return
	}
	apples := g.grid.Apples[:0]
	for _, apple := range g.grid.Apples {
		if _, ok := eaten[apple]; !ok {
			apples = append(apples, apple)
		}
	}
	g.grid.Apples = apples
}

func (g *Game) hasTileOrAppleUnder(c grid.Coord) bool {
	below := c.AddXY(0, 1)
	if g.grid.Get(below).Type == grid.TileWall {
		return true
	}
	return coordSliceContainsWinter(g.grid.Apples, below)
}

func (g *Game) isGrounded(c grid.Coord, frozenBirds map[*Bird]struct{}) bool {
	under := c.AddXY(0, 1)
	if g.hasTileOrAppleUnder(c) {
		return true
	}
	for b := range frozenBirds {
		if coordSliceContainsWinter(b.Body, under) {
			return true
		}
	}
	return false
}

func (g *Game) doFalls() {
	somethingFell := true
	outOfBounds := make([]*Bird, 0)
	airborneBirds := make(map[*Bird]struct{})
	for _, b := range g.liveBirds() {
		airborneBirds[b] = struct{}{}
	}
	groundedBirds := make(map[*Bird]struct{})

	for somethingFell {
		somethingFell = false
		somethingGotGrounded := true
		for somethingGotGrounded {
			somethingGotGrounded = false
			for _, bird := range sortedBirdSet(airborneBirds) {
				isGrounded := false
				for _, c := range bird.Body {
					if g.isGrounded(c, groundedBirds) {
						isGrounded = true
						break
					}
				}
				if isGrounded {
					groundedBirds[bird] = struct{}{}
					somethingGotGrounded = true
				}
			}
			for bird := range groundedBirds {
				delete(airborneBirds, bird)
			}
		}

		for _, bird := range sortedBirdSet(airborneBirds) {
			somethingFell = true
			for i, c := range bird.Body {
				bird.Body[i] = c.AddXY(0, 1)
			}
			allOut := true
			for _, part := range bird.Body {
				if part.Y < g.grid.Height+1 {
					allOut = false
					break
				}
			}
			if allOut {
				bird.Alive = false
				outOfBounds = append(outOfBounds, bird)
				g.emit(EventFall, eventBirdPayload(bird.ID))
			}
		}
		for _, bird := range outOfBounds {
			delete(airborneBirds, bird)
		}
	}
}

func (g *Game) PerformGameUpdate(turn int) {
	g.turn = turn
	g.events = g.events[:0]
	g.doMoves()
	g.doEats()
	g.doBeheadings()
	g.doFalls()
	if g.IsGameOver() {
		g.ended = true
	}
}

func (g *Game) IsGameOver() bool {
	noApples := len(g.grid.Apples) == 0
	playerDead := false
	for _, p := range g.players {
		hasLiveBird := false
		for _, b := range p.birds {
			if b.Alive {
				hasLiveBird = true
				break
			}
		}
		if !hasLiveBird {
			playerDead = true
			break
		}
	}
	return noApples || playerDead
}

func (g *Game) EndGame() {
	g.ended = true
}

func (g *Game) OnEnd() {
	scoreTexts := make([]string, len(g.players))
	for _, p := range g.players {
		if p.IsDeactivated() {
			p.SetScore(-1)
			scoreTexts[p.GetIndex()] = "-"
		} else {
			score := 0
			for _, b := range p.birds {
				if b.Alive {
					score += len(b.Body)
				}
			}
			p.SetScore(score)
		}
	}

	if len(g.players) >= 2 && g.players[0].GetScore() == g.players[1].GetScore() && g.players[0].GetScore() != -1 {
		for _, p := range g.players {
			scoreTexts[p.GetIndex()] = scoreTextWithLosses(p.GetScore(), g.losses[0] == g.losses[1], g.losses[p.GetIndex()])
		}
		for _, p := range g.players {
			p.SetScore(p.GetScore() - g.losses[p.GetIndex()])
		}
	} else {
		for _, p := range g.players {
			if p.GetScore() > -1 {
				scoreTexts[p.GetIndex()] = scoreText(p.GetScore())
			} else {
				scoreTexts[p.GetIndex()] = "-"
			}
		}
	}
}

func (g *Game) ShouldSkipPlayerTurn(player *Player) bool {
	return false
}

func GetExpected(command string) string {
	return "MESSAGE text"
}

func sortedBirdSet(set map[*Bird]struct{}) []*Bird {
	birds := make([]*Bird, 0, len(set))
	for b := range set {
		birds = append(birds, b)
	}
	sort.Slice(birds, func(i, j int) bool {
		return birds[i].ID < birds[j].ID
	})
	return birds
}

func coordSliceContainsWinter(coords []grid.Coord, target grid.Coord) bool {
	for _, c := range coords {
		if c == target {
			return true
		}
	}
	return false
}

func scoreText(score int) string {
	suffix := ""
	if score != 1 {
		suffix = "s"
	}
	return strconv.Itoa(score) + " point" + suffix
}

func scoreTextWithLosses(score int, equalLosses bool, losses int) string {
	join := "but"
	if equalLosses {
		join = "and"
	}
	lossSuffix := ""
	if losses != 1 {
		lossSuffix = "es"
	}
	return scoreText(score) + " (" + join + " " + strconv.Itoa(losses) + " loss" + lossSuffix + ")"
}
