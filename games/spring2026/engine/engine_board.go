// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Board.java
package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:19-30

public class Board {
    private int width, height;
    private Random random;
    private Cell[][] grid;
    private ArrayList<Player> players = new ArrayList<>();
    private ArrayList<Unit> units = new ArrayList<>();
    private ArrayList<Plant> plants = new ArrayList<>();
*/

// Board owns the per-match simulation state. It mirrors Java engine.Board:
// random, grid, players, units, plants — plus runner-side bookkeeping (Turn,
// League, ended, noTreeCounter, Summary).
type Board struct {
	Width   int
	Height  int
	Grid    [][]*Cell // column-major: Grid[x][y], matching Java
	Random  *javarand.Random
	Players []*Player
	Units   []*Unit
	Plants  []*Plant

	League         int
	Turn           int
	Seed           int64
	noTreeCounter  int
	ended          bool
	stalled        bool
	Summary        []string
}

// NewBoard creates a fresh empty grid. CreateMap runs map generation on top.
// It is private in Java; we keep it package-private here too (lowercase).
func newBoard(width, height int, rng *javarand.Random) *Board {
	b := &Board{Width: width, Height: height, Random: rng}
	b.Grid = make([][]*Cell, width)
	for x := 0; x < width; x++ {
		b.Grid[x] = make([]*Cell, height)
		for y := 0; y < height; y++ {
			b.Grid[x][y] = NewCell(x, y)
		}
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			b.Grid[x][y].InitNeighbors(b)
		}
	}
	return b
}

func (b *Board) GetCell(x, y int) *Cell { return b.Grid[x][y] }
func (b *Board) AddPlant(p *Plant)       { b.Plants = append(b.Plants, p) }
func (b *Board) GetPlants() []*Plant    { return b.Plants }
func (b *Board) GetPlayers() []*Player  { return b.Players }
func (b *Board) AddUnit(u *Unit)         { b.Units = append(b.Units, u) }

func (b *Board) GetUnit(id int) *Unit {
	for _, u := range b.Units {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (b *Board) GetUnitsByPlayerID(playerID int) []*Unit {
	out := make([]*Unit, 0)
	for _, u := range b.Units {
		if u.Player.GetIndex() == playerID {
			out = append(out, u)
		}
	}
	return out
}

func (b *Board) GetUnitsByCell(c *Cell) []*Unit {
	out := make([]*Unit, 0)
	for _, u := range b.Units {
		if u.Cell == c {
			out = append(out, u)
		}
	}
	return out
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:95-155

public static Board createMap(List<Player> players, Random random, int league, ...) {
    while (true) {
        int height = random.nextInt(MAP_MAX_HEIGHT - MAP_MIN_HEIGHT + 1) + MAP_MIN_HEIGHT;
        if (league <= 2) height = MAP_MIN_HEIGHT;
        int width = 2 * height;
        Board board = new Board(width, height, random);
        // rivers (league > 2), shacks, inventory, terrain, trees, validate, retry
    }
}
*/

// CreateMap reproduces Java Board.createMap. Players are mutated in place
// (init + setInventory) so the returned Board's Players slice points at the
// same objects the caller already owns.
func CreateMap(players []*Player, rng *javarand.Random, league int) *Board {
	for {
		height := rng.NextInt(MAP_MAX_HEIGHT-MAP_MIN_HEIGHT+1) + MAP_MIN_HEIGHT
		if league <= 2 {
			height = MAP_MIN_HEIGHT
		}
		width := 2 * height
		board := newBoard(width, height, rng)
		board.League = league

		// Rivers (league 3+).
		if league > 2 {
			maxTotalRiver := width*height -
				2*(MAP_MAX_ROCK+PLANT_MAX_SIZE*4+MAP_MAX_IRON+1)*4/5
			riversCount := rng.NextInt(MAP_MAX_RIVER-MAP_MIN_RIVER+1) + MAP_MIN_RIVER
			for i := 0; i < riversCount; i++ {
				river := board.getRandomCell()
				for j := 0; j < 10 && river.IsNearEdge(); j++ {
					river = board.getRandomCell()
				}
				for river != nil && maxTotalRiver > 0 {
					board.setCellType(river, CellWATER)
					river = river.GetNeighbor(rng.NextInt(4))
					maxTotalRiver -= 2
				}
			}
		}

		// Shacks + starting inventory.
		inventory := make([]int, int(ItemIRON)+1)
		if league > 1 {
			for i := range inventory {
				inventory[i] = MIN_STARTING_RESOURCE +
					rng.NextInt(MAX_STARTING_RESOURCE-MIN_STARTING_RESOURCE+1)
			}
		}
		if league < 3 {
			inventory[ItemIRON] = 0
		}
		shack := board.Grid[rng.NextInt(width/2)][rng.NextInt(height)]
		for shack.Type == CellWATER {
			shack = board.Grid[rng.NextInt(width/2)][rng.NextInt(height)]
		}
		board.setCellType(shack, CellSHACK)

		UnitIDCounter = 0
		shacks := [2]*Cell{
			shack,
			board.Grid[width-1-shack.X][height-1-shack.Y],
		}
		for i, player := range players {
			player.InitForGame(shacks[i], league)
			board.Players = append(board.Players, player)
			board.Units = append(board.Units, player.Units[0])
			player.SetInventory(inventory)
			player.RecomputeScore()
		}

		// Terrain (rocks + iron) — league 3+.
		if league > 2 {
			board.placeTerrain(CellIRON, MAP_MIN_IRON, MAP_MAX_IRON)
			board.placeTerrain(CellROCK, MAP_MIN_ROCK, MAP_MAX_ROCK)
		}

		// Trees.
		board.placeTree(ItemPLUM, MAP_MIN_TREE, MAP_MAX_TREE)
		board.placeTree(ItemLEMON, MAP_MIN_TREE, MAP_MAX_TREE)
		board.placeTree(ItemAPPLE, MAP_MIN_TREE, MAP_MAX_TREE)
		board.placeTree(ItemBANANA, MAP_MIN_TREE, MAP_MAX_TREE)

		if board.isValid(league) {
			return board
		}
		// Retry: drop this board (its Players reference will be reassigned on
		// the next pass) and reuse the shared RNG so state continues to
		// advance.
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:187-202

private void placeTree(Item type, int min, int max) {
    int count = random.nextInt(max - min + 1) + min;
    for (int i = 0; i < count; i++) {
        Cell cell = getRandomCell();
        cell.setPlant(new Plant(cell, type));
        int ticks = random.nextInt(1, cell.getPlant().getGrowthCooldown() * (Constants.PLANT_MAX_SIZE + Constants.PLANT_MAX_RESOURCES));
        for (int t = 0; t < ticks; t++) cell.getPlant().tick(true);
        addPlant(cell.getPlant());
        // mirror cell + tree to symmetric location (return early if same cell)
    }
}
*/

func (b *Board) placeTree(kind Item, min, max int) {
	count := b.Random.NextInt(max-min+1) + min
	for i := 0; i < count; i++ {
		cell := b.getRandomCell()
		cell.SetPlant(NewPlant(cell, kind))
		bound := cell.Plant.GetGrowthCooldown() * (PLANT_MAX_SIZE + PLANT_MAX_RESOURCES)
		ticks := b.Random.NextIntRange(1, bound)
		for t := 0; t < ticks; t++ {
			cell.Plant.Tick(true)
		}
		b.AddPlant(cell.Plant)

		mirror := cell
		cell = b.Grid[b.Width-1-cell.X][b.Height-1-cell.Y]
		if cell == mirror {
			return
		}
		cell.SetPlant(NewPlant(cell, kind))
		for t := 0; t < ticks; t++ {
			cell.Plant.Tick(true)
		}
		b.AddPlant(cell.Plant)
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:204-211

private void placeTerrain(Cell.Type type, int min, int max) {
    int count = random.nextInt(max - min + 1) + min;
    for (int i = 0; i < count; i++) {
        Cell cell = getRandomCell();
        setCellType(cell, type);
    }
}
*/

func (b *Board) placeTerrain(t CellType, min, max int) {
	count := b.Random.NextInt(max-min+1) + min
	for i := 0; i < count; i++ {
		cell := b.getRandomCell()
		b.setCellType(cell, t)
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:212-218

private Cell getRandomCell() {
    while (true) {
        Cell cell = grid[random.nextInt(width)][random.nextInt(height)];
        if (cell.getType() == Cell.Type.GRASS && cell.getPlant() == null) return cell;
    }
}
*/

func (b *Board) getRandomCell() *Cell {
	for {
		cell := b.Grid[b.Random.NextInt(b.Width)][b.Random.NextInt(b.Height)]
		if cell.Type == CellGRASS && cell.Plant == nil {
			return cell
		}
	}
}

func (b *Board) setCellType(cell *Cell, t CellType) {
	cell.SetType(t)
	b.Grid[b.Width-1-cell.X][b.Height-1-cell.Y].SetType(t)
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:224-261

private boolean isValid(int league) {
    if (players.get(0).getShack().isNearIron()) return false;
    // shack has at least one walkable neighbour
    // for league > 2: at least one iron has a walkable neighbour
    // all walkable cells are reachable from one of them (BFS)
    // shack within MAP_MAX_OPP_DIST walking distance of opp's shack neighbourhood
    // league < 3: at least one plant has resources > 0
}
*/

func (b *Board) isValid(league int) bool {
	if b.Players[0].Shack.IsNearIron() {
		return false
	}
	shackReachable := false
	for _, c := range b.Players[0].Shack.Neighbors {
		if c != nil && c.IsWalkable() {
			shackReachable = true
		}
	}
	if !shackReachable {
		return false
	}

	walkables := make([]*Cell, 0)
	irons := make([]*Cell, 0)
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			c := b.Grid[x][y]
			if c.IsWalkable() {
				walkables = append(walkables, c)
			}
			if c.Type == CellIRON {
				irons = append(irons, c)
			}
		}
	}
	canReachIron := false
	for _, iron := range irons {
		for _, c := range iron.Neighbors {
			if c != nil && c.IsWalkable() {
				canReachIron = true
			}
		}
	}
	if !canReachIron && league > 2 {
		return false
	}

	dist := b.GetDistances(walkables[0])
	for _, c := range walkables {
		if dist[c.X][c.Y] == -1 {
			return false
		}
	}

	shackDist := b.GetDistances(b.Players[0].Shack)
	oppReachable := false
	for _, c := range b.Players[1].Shack.Neighbors {
		if c != nil && shackDist[c.X][c.Y] >= 0 && shackDist[c.X][c.Y] < MAP_MAX_OPP_DIST {
			oppReachable = true
		}
	}
	if !oppReachable {
		return false
	}

	if league < 3 {
		hasFruit := false
		for _, p := range b.Plants {
			if p.Resources > 0 {
				hasFruit = true
				break
			}
		}
		if !hasFruit {
			return false
		}
	}
	return true
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:267-303

public Cell getNextCell(Cell current, Cell target, int speed) {
    int[][] targetDist = getDistances(target);
    int[][] sourceDist = getDistances(current);
    if (sourceDist[target.x][target.y] >= 0 && sourceDist[target.x][target.y] <= speed) return target;
    if (sourceDist[target.x][target.y] == -1) {
        // collect cells reachable from source that minimise manhattan(target)
        // re-run BFS from those as a multi-source start
        targetDist = getDistances(closest);
    }
    // collect cells within speed of source minimising targetDist; pick random tie-break
}
*/

func (b *Board) GetNextCell(unit *Unit, target *Cell) *Cell {
	return b.getNextCell(unit.Cell, target, unit.MovementSpeed)
}

func (b *Board) getNextCell(current, target *Cell, speed int) *Cell {
	targetDist := b.GetDistances(target)
	sourceDist := b.GetDistances(current)
	if sourceDist[target.X][target.Y] >= 0 && sourceDist[target.X][target.Y] <= speed {
		return target
	}
	if sourceDist[target.X][target.Y] == -1 {
		closest := make([]*Cell, 0)
		best := b.Width * b.Height
		for x := 0; x < b.Width; x++ {
			for y := 0; y < b.Height; y++ {
				if sourceDist[x][y] == -1 {
					continue
				}
				d := target.Manhattan(b.Grid[x][y])
				if d < best {
					best = d
					closest = closest[:0]
				}
				if d == best {
					closest = append(closest, b.Grid[x][y])
				}
			}
		}
		targetDist = b.getDistancesMulti(closest)
	}

	closest := make([]*Cell, 0)
	best := b.Width * b.Height
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			if sourceDist[x][y] > speed || sourceDist[x][y] == -1 {
				continue
			}
			d := targetDist[x][y]
			if d >= 0 && d < best {
				best = d
				closest = closest[:0]
			}
			if d == best {
				closest = append(closest, b.Grid[x][y])
			}
		}
	}
	return closest[b.Random.NextInt(len(closest))]
}

// GetDistances mirrors Java Board.getDistances(Cell): BFS distance from
// `start` over walkable cells. Unreachable cells are -1.
func (b *Board) GetDistances(start *Cell) [][]int {
	return b.getDistancesMulti([]*Cell{start})
}

func (b *Board) getDistancesMulti(starts []*Cell) [][]int {
	result := make([][]int, b.Width)
	for x := 0; x < b.Width; x++ {
		result[x] = make([]int, b.Height)
		for y := 0; y < b.Height; y++ {
			result[x][y] = -1
		}
	}
	queue := make([]*Cell, 0)
	for _, c := range starts {
		queue = append(queue, c)
		result[c.X][c.Y] = 0
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range c.Neighbors {
			if n != nil && n.IsWalkable() && result[n.X][n.Y] == -1 {
				result[n.X][n.Y] = result[c.X][c.Y] + 1
				queue = append(queue, n)
			}
		}
	}
	return result
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:335-358

public void tick(int turn, TaskManager taskManager, GameManager gameManager) {
    while (taskManager.hasTasks()) {
        ArrayList<Task> tasks = taskManager.popTasks();
        for (Task t : tasks) t.apply(this, tasks);
    }
    for (Plant plant : plants) plant.tick(true);
    plants = plants.stream().filter(p -> !p.isDead()).collect(...);
    for (Player player : players) player.recomputeScore();
}
*/

// Tick advances the simulation by one turn: drain the task manager bucket by
// bucket (lowest priority first), grow every surviving plant, drop dead
// plants from the live set, and refresh scores.
func (b *Board) Tick(turn int, taskManager *TaskManager) {
	for taskManager.HasTasks() {
		tasks := taskManager.PopTasks()
		for _, t := range tasks {
			t.Apply(b, tasks)
		}
	}
	for _, plant := range b.Plants {
		plant.Tick(true)
	}
	alive := b.Plants[:0]
	for _, plant := range b.Plants {
		if !plant.IsDead() {
			alive = append(alive, plant)
		}
	}
	b.Plants = alive
	for _, p := range b.Players {
		p.RecomputeScore()
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:361-380

public ArrayList<String> getInitialInputs(int id) {
    // width height
    // height rows of width characters: '.', '~', '+', '#', '0'/'1' for shacks
}
*/

func (b *Board) GetInitialInputs(id int) []string {
	result := make([]string, 0, 1+b.Height)
	result = append(result, itoa(b.Width)+" "+itoa(b.Height))
	for y := 0; y < b.Height; y++ {
		var line strings.Builder
		for x := 0; x < b.Width; x++ {
			cell := b.Grid[x][y]
			switch cell.Type {
			case CellGRASS:
				line.WriteByte('.')
			case CellWATER:
				line.WriteByte('~')
			case CellIRON:
				line.WriteByte('+')
			case CellROCK:
				line.WriteByte('#')
			case CellSHACK:
				ownerIdx := -1
				for _, p := range b.Players {
					if p.Shack == cell {
						ownerIdx = p.GetIndex()
						break
					}
				}
				line.WriteString(strconv.Itoa((ownerIdx - id + len(b.Players)) % len(b.Players)))
			}
		}
		result = append(result, line.String())
	}
	return result
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:382-393

public ArrayList<String> getTurnInputs(int id) {
    // inventory of recipient first, opponent second
    // tree count + lines
    // troll count + lines
}
*/

func (b *Board) GetTurnInputs(id int) []string {
	result := make([]string, 0)
	for i := 0; i < len(b.Players); i++ {
		p := b.Players[(i+id)%len(b.Players)]
		result = append(result, p.Inv.GetInputLine())
	}
	result = append(result, itoa(len(b.Plants)))
	for _, plant := range b.Plants {
		result = append(result, plant.GetInputLine())
	}
	result = append(result, itoa(len(b.Units)))
	for _, u := range b.Units {
		result = append(result, u.GetInputLine(id, len(b.Players)))
	}
	return result
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Board.java:402-424

private int noTreeCounter = 0;
public boolean hasStalled() {
    if (plants.size() > 0) { noTreeCounter = 0; return false; }
    noTreeCounter++;
    if (noTreeCounter == STALL_LIMIT) return true;
    boolean[] playerStuck = { true, true };
    for (Unit unit : units) if (unit.getInventory().getTotal() > unit.getInventory().getItemCount(Item.IRON)) playerStuck[unit.getPlayer().getIndex()] = false;
    for (Player player : players)
        for (int i = 0; i <= Item.BANANA.ordinal(); i++)
            if (player.getInventory().getItemCount(i) > 0) playerStuck[player.getIndex()] = false;
    if (playerStuck[0] && playerStuck[1]) return true;
    if (playerStuck[0] && players.get(0).getScore() < players.get(1).getScore()) return true;
    if (playerStuck[1] && players.get(1).getScore() < players.get(0).getScore()) return true;
    return false;
}
*/

func (b *Board) HasStalled() bool {
	if len(b.Plants) > 0 {
		b.noTreeCounter = 0
		return false
	}
	b.noTreeCounter++
	if b.noTreeCounter == STALL_LIMIT {
		return true
	}
	var playerStuck [2]bool
	playerStuck[0] = true
	playerStuck[1] = true
	for _, u := range b.Units {
		if u.Inv.GetTotal() > u.Inv.GetItemCount(ItemIRON) {
			playerStuck[u.Player.GetIndex()] = false
		}
	}
	for _, p := range b.Players {
		for i := 0; i <= int(ItemBANANA); i++ {
			if p.Inv.GetItemCount(Item(i)) > 0 {
				playerStuck[p.GetIndex()] = false
			}
		}
	}
	if playerStuck[0] && playerStuck[1] {
		return true
	}
	if playerStuck[0] && b.Players[0].GetScore() < b.Players[1].GetScore() {
		return true
	}
	if playerStuck[1] && b.Players[1].GetScore() < b.Players[0].GetScore() {
		return true
	}
	return false
}

// Ended reports whether the engine has resolved a terminal condition. The
// Referee checks this after PerformGameUpdate and stops the loop accordingly.
func (b *Board) Ended() bool { return b.ended }

// MainTurnsForLeague mirrors Java Referee.init: leagues 1-2 cap at 100 turns,
// otherwise 300.
func MainTurnsForLeague(league int) int {
	if league > 2 {
		return GAME_TURNS
	}
	return GAME_TURNS_LOW_LEAGUE
}

