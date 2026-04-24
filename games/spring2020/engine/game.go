// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java
package engine

import (
	"fmt"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/action"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/maps"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/pathfinder"
	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

// Game holds all simulation state.
type Game struct {
	Config  Config
	Players []*Player
	Pacmen  []*Pacman
	Grid    *grid.Grid

	random      *javarand.Random
	pacsPer     int
	totalPacs   int
	currentStep int

	ended             bool
	gameOverProcessed bool
	summary           []string
}

// NewGame sets up a fresh simulation with the given seed and league level.
// Players must be added via Init before running turns.
func NewGame(seed int64, leagueLevel int) *Game {
	rules := LeagueRulesFromIndex(leagueLevel)
	cfg := NewConfig(rules)
	g := &Game{
		Config: cfg,
		random: javarand.New(seed),
	}
	return g
}

// Init seeds grid, pacmen, pellets, and cherries. Must be called before turns.
func (g *Game) Init(players []*Player) {
	g.Players = players
	g.Grid = g.generateGrid()

	if g.pacsPer == 0 {
		g.pacsPer = g.randomInt(g.Config.MinPacs, g.Config.MaxPacs)
	}
	g.totalPacs = g.pacsPer * len(g.Players)

	if g.Pacmen == nil {
		g.generatePacmen()
		g.placePacmenAndCherries()
	}

	g.spawnPellets()
}

func (g *Game) randomInt(min, max int) int {
	return min + g.random.NextInt(max-min+1)
}

func (g *Game) generateGrid() *grid.Grid {
	width := g.randomInt(g.Config.MapMinWidth, g.Config.MapMaxWidth)
	height := g.randomInt(g.Config.MapMinHeight, g.Config.MapMaxHeight)
	gen := maps.NewGenerator()

	if len(g.Players) == 2 && width%2 == 0 {
		width++
	}

	innerMapWraps := g.Config.MapWraps
	inner := grid.NewGrid(width, height, innerMapWraps)
	if len(g.Players) == 2 {
		gen.GenerateWithHorizontalSymmetry(inner, g.random)
	} else {
		gen.Generate(inner, g.random)
	}

	if !g.Config.MapWraps {
		return inner
	}

	// Wrap: surround with a 1-cell tunnel border.
	bigW, bigH := width+2, height+2
	big := grid.NewGrid(bigW, bigH, true)
	for y := 0; y < bigH; y++ {
		for x := 0; x < bigW; x++ {
			gridPos := grid.Coord{X: x - 1, Y: y - 1}
			if isOuterBorder(x, y, width, height) {
				switch {
				case x == 0 && isTunnelExit(inner, gridPos.X+1, gridPos.Y):
					big.GetXY(x, y).Type = grid.CellFloor
				case x == width+1 && isTunnelExit(inner, gridPos.X-1, gridPos.Y):
					big.GetXY(x, y).Type = grid.CellFloor
				default:
					big.GetXY(x, y).Type = grid.CellWall
				}
			} else {
				big.GetXY(x, y).Copy(inner.GetXY(x-1, y-1))
			}
		}
	}
	return big
}

func isTunnelExit(g *grid.Grid, x, y int) bool {
	return g.GetXY(x, y).IsFloor() && !g.GetXY(x, y-1).IsFloor() && !g.GetXY(x, y+1).IsFloor()
}

func isOuterBorder(x, y, width, height int) bool {
	return x == 0 || y == 0 || x == width+1 || y == height+1
}

func (g *Game) generatePacmen() {
	g.Pacmen = make([]*Pacman, 0, g.totalPacs)
	pacmanIndex := 0
	typeIndex := 0
	rotation := [3]PacmanType{TypeRock, TypePaper, TypeScissors}
	for pacmanIndex < g.totalPacs {
		for _, player := range g.Players {
			if pacmanIndex >= g.totalPacs {
				break
			}
			t := TypeNeutral
			if g.Config.SwitchAvail {
				t = rotation[typeIndex%3]
			}
			pac := NewPacman(pacmanIndex, len(player.Pacmen), player, t)
			player.AddPacman(pac)
			g.Pacmen = append(g.Pacmen, pac)
			pacmanIndex++
		}
		typeIndex++
	}
}

func (g *Game) placePacmenAndCherries() {
	freeCells := make([]grid.Coord, 0)
	halfX := g.Grid.Width / 2
	for _, cell := range g.Grid.Cells() {
		if cell.Type == grid.CellFloor && cell.Coord.X != halfX {
			freeCells = append(freeCells, cell.Coord)
		}
	}
	shuffleCoords(freeCells, g.random)

	if len(g.Players) == 2 {
		leftCells := make([]grid.Coord, 0)
		for _, c := range freeCells {
			if c.X <= halfX {
				leftCells = append(leftCells, c)
			}
		}
		var i int
		for i = 0; i < g.pacsPer; i++ {
			left := leftCells[i]
			right := grid.Coord{X: g.Grid.Width - 1 - left.X, Y: left.Y}
			leftPlayer := g.random.NextInt(2)
			rightPlayer := (leftPlayer + 1) % 2
			g.Players[leftPlayer].Pacmen[i].Position = left
			g.Players[rightPlayer].Pacmen[i].Position = right
		}
		for j := 0; j < g.Config.NumCherries/2; j++ {
			left := leftCells[i+j]
			right := grid.Coord{X: g.Grid.Width - 1 - left.X, Y: left.Y}
			g.Grid.Get(left).HasCherry = true
			g.Grid.Get(right).HasCherry = true
		}
	} else {
		for i := 0; i < g.totalPacs; i++ {
			g.Pacmen[i].Position = freeCells[i]
		}
	}
}

func (g *Game) spawnPellets() {
	for _, cell := range g.Grid.Cells() {
		if !cell.IsFloor() || cell.HasCherry {
			continue
		}
		occupied := false
		for _, pac := range g.Pacmen {
			if pac.Position == cell.Coord {
				occupied = true
				break
			}
		}
		if !occupied {
			cell.HasPellet = true
		}
	}
}

// ResetGameTurnData is invoked at the start of each main (non-speed) turn.
func (g *Game) ResetGameTurnData() {
	g.currentStep = 0
	for _, p := range g.Players {
		p.TurnReset()
	}
}

// IsSpeedTurn reports whether the next update should run as a speed sub-turn.
func (g *Game) IsSpeedTurn() bool {
	numSteps := 0
	for _, p := range g.Pacmen {
		if len(p.IntendedPath) > numSteps {
			numSteps = len(p.IntendedPath)
		}
	}
	numSteps--
	return g.currentStep > 0 && g.currentStep < numSteps
}

// PerformGameUpdate runs one normal turn of the simulation.
func (g *Game) PerformGameUpdate() {
	g.executePacmenAbilities()
	g.updateAbilityModifiers()
	g.processPacmenIntent()
	g.resolveMovement()
}

// PerformGameSpeedUpdate runs one extra speed sub-turn (SPEED ability step 2+).
func (g *Game) PerformGameSpeedUpdate() {
	for _, pac := range g.Pacmen {
		if pac.Speed <= 1 || pac.Intent.Type != action.ActionMove {
			pac.IntendedPath = []grid.Coord{pac.Position}
		}
		pac.Blocked = false
	}
	g.resolveMovement()
}

func (g *Game) executePacmenAbilities() {
	for _, pac := range g.Pacmen {
		if !pac.HasAbilityToUse {
			continue
		}
		ability := pac.AbilityToUse

		if pac.AbilityCooldown != 0 {
			continue
		}
		if (!g.Config.SpeedAvail && ability == AbilitySpeed) ||
			(!g.Config.SwitchAvail && ability != AbilitySpeed) {
			continue
		}

		switch ability {
		case AbilitySetRock, AbilitySetPaper, AbilitySetScissors:
			pac.Type = pacTypeFromAbility(ability)
		case AbilitySpeed:
			pac.Speed = g.Config.SpeedBoost
			pac.AbilityDuration = g.Config.AbilityDur
		}
		pac.AbilityCooldown = g.Config.AbilityCool
	}
}

func (g *Game) updateAbilityModifiers() {
	for _, pac := range g.Pacmen {
		if pac.Speed > 1 && pac.AbilityDuration == 0 {
			pac.Speed = g.Config.PacmanBase
		}
	}
}

func (g *Game) processPacmenIntent() {
	for _, pac := range g.Pacmen {
		if pac.Intent.Type == action.ActionMove {
			pac.IntendedPath = g.computeIntendedPath(pac)
		} else {
			pac.IntendedPath = []grid.Coord{pac.Position}
		}
	}
}

func (g *Game) computeIntendedPath(pac *Pacman) []grid.Coord {
	target := pac.Intent.Target
	pfr := pathfinder.FindPath(g.Grid, pac.Position, target, nil)
	wholePath := pfr.Path
	if pfr.IsNearest {
		if len(pfr.Path) > 1 {
			newTarget := pfr.Path[len(pfr.Path)-1]
			pac.WarnPathMsg = fmt.Sprintf(
				"Warning: target (%d, %d) is unreachable, going to (%d, %d) instead.",
				target.X, target.Y, newTarget.X, newTarget.Y,
			)
		} else {
			pac.WarnPathMsg = fmt.Sprintf(
				"Warning: target (%d, %d) is unreachable. Staying here!",
				target.X, target.Y,
			)
		}
		pac.HasWarnPathMsg = true
	} else {
		pac.WarnPathMsg = ""
		pac.HasWarnPathMsg = false
	}

	if len(wholePath) > 1 {
		stepsThisTurn := pac.Speed
		if stepsThisTurn > len(wholePath)-1 {
			stepsThisTurn = len(wholePath) - 1
		}
		return wholePath[:stepsThisTurn+1]
	}
	return []grid.Coord{pac.Position}
}

func (g *Game) resolveMovement() {
	pacmenToKill := make([]*Pacman, 0)
	seen := make(map[*Pacman]struct{})

	resolution := g.resolvePacmenMovement()

	for _, pac := range g.Pacmen {
		for _, other := range g.Pacmen {
			if pac == other {
				continue
			}
			if g.canEat(pac, other) && g.pacmenHaveCollided(pac, other) {
				if _, ok := seen[other]; !ok {
					seen[other] = struct{}{}
					pacmenToKill = append(pacmenToKill, other)
				}
			}
		}
	}

	g.killPacmen(pacmenToKill)

	g.eatPellets()
	g.eatCherries()

	_ = resolution
	g.currentStep++
}

func (g *Game) canEat(a, b *Pacman) bool {
	if a.Owner == b.Owner {
		return false
	}
	switch a.Type {
	case TypePaper:
		return b.Type == TypeRock
	case TypeRock:
		return b.Type == TypeScissors
	case TypeScissors:
		return b.Type == TypePaper
	}
	return false
}

func (g *Game) pacmenHaveCollided(a, b *Pacman) bool {
	fromA := g.intendedPositionAtStep(a, a.PreviousPathStep)
	fromB := g.intendedPositionAtStep(b, b.PreviousPathStep)
	toA := g.intendedPositionAtStep(a, a.CurrentPathStep)
	toB := g.intendedPositionAtStep(b, b.CurrentPathStep)
	if toA == toB {
		return true
	}
	return toA == fromB && toB == fromA
}

func (g *Game) resolvePacmenMovement() *MovementResolution {
	res := NewMovementResolution()
	pacmenToResolve := make([]*Pacman, 0)
	for _, pac := range g.Pacmen {
		if pac.Intent.Type == action.ActionMove && !pac.MoveFinished() && pac.FastEnoughToMoveAt(g.currentStep) {
			pacmenToResolve = append(pacmenToResolve, pac)
		}
	}

	for {
		resolved := make([]*Pacman, 0)
		for _, pac := range pacmenToResolve {
			var blocker *Pacman
			bestDist := 0
			for _, other := range g.Pacmen {
				if pac == other {
					continue
				}
				if !g.isBodyBlockedBy(pac, other) {
					continue
				}
				d := g.Grid.CalculateDistance(pac.Position, other.Position)
				if blocker == nil || d < bestDist {
					blocker = other
					bestDist = d
				}
			}
			if blocker != nil {
				resolved = append(resolved, pac)
				res.AddBlocked(pac)
				res.BlockedBy[pac] = blocker
			}
		}
		for _, pac := range resolved {
			pac.Blocked = true
		}
		pacmenToResolve = removeAll(pacmenToResolve, resolved)
		if len(resolved) == 0 {
			break
		}
	}

	for _, pac := range pacmenToResolve {
		g.movePacman(pac)
		res.AddMoved(pac)
	}
	return res
}

func (g *Game) isBodyBlockedBy(pac, other *Pacman) bool {
	if !g.Config.BodyBlock {
		return false
	}
	if !g.Config.FriendlyBlock && pac.Owner == other.Owner {
		return false
	}
	if g.canEat(pac, other) {
		return false
	}
	if g.canEat(other, pac) && g.pacmenWillShareSameCoord(pac, other) {
		return false
	}
	return g.pacmenWillCollide(pac, other)
}

func (g *Game) pacmenWillShareSameCoord(a, b *Pacman) bool {
	aStep := a.CurrentPathStep + 1
	if a.Blocked {
		aStep = a.CurrentPathStep
	}
	bStep := b.CurrentPathStep + 1
	if b.Blocked {
		bStep = b.CurrentPathStep
	}
	return g.intendedPositionAtStep(a, aStep) == g.intendedPositionAtStep(b, bStep)
}

func (g *Game) pacmenWillCollide(a, b *Pacman) bool {
	fromA := g.intendedPositionAtStep(a, a.CurrentPathStep)
	fromB := g.intendedPositionAtStep(b, b.CurrentPathStep)
	toA := fromA
	if !a.Blocked {
		toA = g.intendedPositionAtStep(a, a.CurrentPathStep+1)
	}
	toB := fromB
	if !b.Blocked {
		toB = g.intendedPositionAtStep(b, b.CurrentPathStep+1)
	}
	if toA == toB {
		return true
	}
	return toA == fromB && toB == fromA
}

func (g *Game) intendedPositionAtStep(pac *Pacman, step int) grid.Coord {
	if step > len(pac.IntendedPath)-1 {
		step = len(pac.IntendedPath) - 1
	}
	if step < 0 {
		step = 0
	}
	return pac.IntendedPath[step]
}

func (g *Game) movePacman(pac *Pacman) {
	pac.SetCurrentPathStep(pac.CurrentPathStep + 1)
	to := g.intendedPositionAtStep(pac, pac.CurrentPathStep)
	pac.Position = to
	pac.WarnPathMsg = ""
	pac.HasWarnPathMsg = false
}

func (g *Game) killPacmen(pacmenToKill []*Pacman) {
	if len(pacmenToKill) == 0 {
		return
	}
	// Remove dead pacs from the "alive" master list while keeping Player.Pacmen.
	kill := make(map[*Pacman]struct{}, len(pacmenToKill))
	for _, p := range pacmenToKill {
		kill[p] = struct{}{}
		p.Dead = true
	}
	alive := g.Pacmen[:0]
	for _, p := range g.Pacmen {
		if _, dead := kill[p]; !dead {
			alive = append(alive, p)
		}
	}
	g.Pacmen = alive

	for _, pac := range pacmenToKill {
		owner := pac.Owner
		if len(owner.AlivePacmen()) == 0 {
			owner.Deactivate("all pacmen dead")
		}
	}
}

func (g *Game) eatPellets() {
	g.eatItem(func(c *grid.Cell) bool { return c.HasPellet }, 1)
}

func (g *Game) eatCherries() {
	g.eatItem(func(c *grid.Cell) bool { return c.HasCherry }, CherryScore)
}

func (g *Game) eatItem(hasItem func(*grid.Cell) bool, pelletValue int) {
	type eater struct {
		coord grid.Coord
		pacs  []*Pacman
	}
	byCoord := make(map[grid.Coord]*eater)
	order := make([]grid.Coord, 0)
	for _, pac := range g.Pacmen {
		cell := g.Grid.Get(pac.Position)
		if !hasItem(cell) {
			continue
		}
		e, ok := byCoord[pac.Position]
		if !ok {
			e = &eater{coord: pac.Position}
			byCoord[pac.Position] = e
			order = append(order, pac.Position)
		}
		dupOwner := false
		for _, existing := range e.pacs {
			if existing.Owner == pac.Owner {
				dupOwner = true
				break
			}
		}
		if !dupOwner {
			e.pacs = append(e.pacs, pac)
		}
	}

	for _, coord := range order {
		e := byCoord[coord]
		credited := make(map[*Player]struct{})
		for _, pac := range e.pacs {
			if _, ok := credited[pac.Owner]; ok {
				continue
			}
			credited[pac.Owner] = struct{}{}
			pac.Owner.Pellets += pelletValue
		}
		cell := g.Grid.Get(coord)
		cell.HasPellet = false
		cell.HasCherry = false
	}
}

// IsGameOver mirrors Game.isGameOver.
func (g *Game) IsGameOver() bool {
	activePlayers := g.activePlayers()
	if len(activePlayers) <= 1 {
		return true
	}
	for _, p := range activePlayers {
		if g.canImproveRanking(p) {
			return false
		}
	}
	return true
}

// PerformGameOver absorbs remaining pellets when only one player remains.
func (g *Game) PerformGameOver() {
	if g.gameOverProcessed {
		return
	}
	activePlayers := g.activePlayers()
	if len(activePlayers) == 1 {
		activePlayers[0].Pellets += g.remainingPellets()
	}
	if len(activePlayers) <= 1 {
		g.gameOverProcessed = true
	}
}

func (g *Game) canImproveRanking(player *Player) bool {
	remaining := g.remainingPellets()
	if remaining == 0 {
		return false
	}
	for _, other := range g.Players {
		if other == player {
			continue
		}
		if other.Pellets < player.Pellets {
			continue
		}
		if player.Pellets+remaining >= other.Pellets {
			return true
		}
	}
	return false
}

func (g *Game) activePlayers() []*Player {
	out := make([]*Player, 0, len(g.Players))
	for _, p := range g.Players {
		if !p.IsDeactivated() {
			out = append(out, p)
		}
	}
	return out
}

func (g *Game) remainingPellets() int {
	sum := 0
	for _, cell := range g.Grid.Cells() {
		switch {
		case cell.HasPellet:
			sum++
		case cell.HasCherry:
			sum += CherryScore
		}
	}
	return sum
}

// EndGame flags the game as ended.
func (g *Game) EndGame() { g.ended = true }

// Ended reports whether the simulation has finished.
func (g *Game) Ended() bool { return g.ended }

// OnEnd sets the final score as the pellet count for each player.
func (g *Game) OnEnd() {
	for _, p := range g.Players {
		p.SetScore(p.Pellets)
	}
}

// shuffleCoords applies Java's Collections.shuffle semantics.
func shuffleCoords(list []grid.Coord, r *javarand.Random) {
	for i := len(list); i > 1; i-- {
		j := r.NextInt(i)
		list[i-1], list[j] = list[j], list[i-1]
	}
}

func removeAll[T comparable](list, toRemove []T) []T {
	if len(toRemove) == 0 {
		return list
	}
	rm := make(map[T]struct{}, len(toRemove))
	for _, v := range toRemove {
		rm[v] = struct{}{}
	}
	out := list[:0]
	for _, v := range list {
		if _, ok := rm[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}
