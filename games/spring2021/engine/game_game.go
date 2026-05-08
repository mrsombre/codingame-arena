// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Game.java
package engine

import (
	"fmt"
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:32-60

@Singleton
public class Game {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private GameSummaryManager gameSummaryManager;
    Integer nutrients = Config.STARTING_NUTRIENTS;

    public static boolean ENABLE_SEED;
    public static boolean ENABLE_GROW;
    public static boolean ENABLE_SHADOW;
    public static boolean ENABLE_HOLES;
    public static int MAX_ROUNDS;
    public static int STARTING_TREE_COUNT;
    public static int STARTING_TREE_SIZE;
    public static int STARTING_TREE_DISTANCE;
    public static boolean STARTING_TREES_ON_EDGES;

    Board board;
    Map<Integer, Tree> trees;
    List<CubeCoord> dyingTrees;
    List<Integer> availableSun;
    List<Seed> sentSeeds;
    Sun sun;
    Map<Integer, Integer> shadows;
    int round = 0;
    int turn = 0;
    FrameType currentFrameType = FrameType.INIT;
    FrameType nextFrameType = FrameType.GATHERING;
}
*/

// Game holds simulation state for one Spring Challenge 2021 match. The Java
// "static" league flags become per-Game fields so repeated simulations don't
// leak state between games.
type Game struct {
	Cfg     Config
	Players []*Player
	Summary *GameSummaryManager

	// League flags (Java statics).
	ENABLE_SEED             bool
	ENABLE_GROW             bool
	ENABLE_SHADOW           bool
	ENABLE_HOLES            bool
	MAX_ROUNDS              int
	STARTING_TREE_COUNT     int
	STARTING_TREE_SIZE      int
	STARTING_TREE_DISTANCE  int
	STARTING_TREES_ON_EDGES bool

	Board        *Board
	Trees        map[int]*Tree
	TreeOrder    []int // index keys in ascending order — TreeMap traversal parity.
	DyingTrees   []CubeCoord
	AvailableSun []int
	SentSeeds    []Seed
	Sun          Sun
	Shadows      map[int]int
	Nutrients    int

	Round            int
	Turn             int
	DayActionIndex   int
	CurrentFrameType FrameType
	NextFrameType    FrameType

	random *javarand.Random
	league int
	ended  bool

	// traces accumulates per-turn structured events for the current
	// PerformGameUpdate call, partitioned by player. Reset at the top of
	// each call; drained by Referee.TurnTraces after the call returns.
	traces [2][]arena.TurnTrace
	// seedConflictCell is set when both players' seeds collide on the same
	// target this action frame. The decorator copies it to the trace turn
	// root field SeedConflictCell. Reset at the top of each game update.
	seedConflictCell *int
}

func NewGame(seed int64, leagueLevel int) *Game {
	cfg := NewConfig()
	g := &Game{
		Cfg:              cfg,
		Summary:          NewGameSummaryManager(cfg),
		random:           javarand.New(seed),
		league:           leagueLevel,
		Nutrients:        cfg.STARTING_NUTRIENTS,
		CurrentFrameType: FrameInit,
		NextFrameType:    FrameGathering,
	}
	g.applyLeagueFlags(leagueLevel)
	// Refresh GameSummaryManager once league-driven MAX_ROUNDS is known.
	g.Summary.cfg.MAX_ROUNDS = g.MAX_ROUNDS
	return g
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:62-100

switch (gameManager.getLeagueLevel()) {
case 1: ... Wood 2 ...
case 2: ... Wood 1 ...
default: ... Bronze+ ...
}
*/

func (g *Game) applyLeagueFlags(level int) {
	switch level {
	case 1:
		g.MAX_ROUNDS = 1
		g.ENABLE_SEED = false
		g.ENABLE_GROW = false
		g.ENABLE_SHADOW = false
		g.ENABLE_HOLES = false
		g.STARTING_TREE_COUNT = 6
		g.STARTING_TREE_SIZE = TREE_TALL
		g.STARTING_TREE_DISTANCE = 0
		g.STARTING_TREES_ON_EDGES = false
	case 2:
		g.MAX_ROUNDS = 6
		g.ENABLE_SEED = false
		g.ENABLE_GROW = true
		g.ENABLE_SHADOW = false
		g.ENABLE_HOLES = false
		g.STARTING_TREE_COUNT = 4
		g.STARTING_TREE_SIZE = TREE_SMALL
		g.STARTING_TREE_DISTANCE = 1
		g.STARTING_TREES_ON_EDGES = false
	default:
		g.MAX_ROUNDS = g.Cfg.MAX_ROUNDS
		g.ENABLE_SEED = true
		g.ENABLE_GROW = true
		g.ENABLE_SHADOW = true
		g.ENABLE_HOLES = true
		g.STARTING_TREE_COUNT = STARTING_TREE_COUNT
		g.STARTING_TREE_SIZE = TREE_SMALL
		g.STARTING_TREE_DISTANCE = 2
		g.STARTING_TREES_ON_EDGES = true
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:62-119

public void init(long seed) {
    // ... league flags ...
    random = new Random(seed);
    board = BoardGenerator.generate(random);
    trees = new TreeMap<>();
    dyingTrees = new ArrayList<>();
    availableSun = new ArrayList<>(gameManager.getPlayerCount());
    sentSeeds = new ArrayList<>();
    initStartingTrees();
    sun = new Sun(); shadows = new HashMap<>(); sun.setOrientation(0);
    round = 0;
    if (ENABLE_SHADOW) calculateShadows();
}
*/

func (g *Game) Init(players []*Player) {
	g.Players = players
	for _, p := range g.Players {
		p.SetSun(g.Cfg.STARTING_SUN)
		p.Action = NoAction
	}

	g.Board = NewBoardGenerator().Generate(g.random, g.Cfg, g.ENABLE_HOLES)
	g.Trees = make(map[int]*Tree)
	g.TreeOrder = nil
	g.DyingTrees = nil
	g.AvailableSun = make([]int, len(g.Players))
	g.SentSeeds = nil

	g.initStartingTrees()

	g.Sun = Sun{}
	g.Shadows = make(map[int]int)
	g.Sun.SetOrientation(0)
	g.Round = 0
	if g.ENABLE_SHADOW {
		g.calculateShadows()
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:121-129

public static String getExpected() {
    if (!ENABLE_GROW && !ENABLE_SEED) return "COMPLETE <idx> | WAIT";
    if (!ENABLE_SEED && ENABLE_GROW)  return "GROW <idx> | COMPLETE <idx> | WAIT";
    return "SEED <from> <to> | GROW <idx> | COMPLETE <idx> | WAIT";
}
*/

func (g *Game) GetExpected() string {
	switch {
	case !g.ENABLE_GROW && !g.ENABLE_SEED:
		return "COMPLETE <idx> | WAIT"
	case !g.ENABLE_SEED && g.ENABLE_GROW:
		return "GROW <idx> | COMPLETE <idx> | WAIT"
	default:
		return "SEED <from> <to> | GROW <idx> | COMPLETE <idx> | WAIT"
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:141-168

private void initStartingTrees() {
    List<CubeCoord> startingCoords = ...;
    if (STARTING_TREES_ON_EDGES) startingCoords = getBoardEdges();
    else startingCoords = new ArrayList<>(board.coords); // minus centre
    startingCoords.removeIf(coord -> board.map.get(coord).getRichness() == RICHNESS_NULL);
    while (validCoords.size() < STARTING_TREE_COUNT * 2) {
        validCoords = tryInitStartingTrees(startingCoords);
    }
    for (int i = 0; i < STARTING_TREE_COUNT; i++) {
        placeTree(players.get(0), validCoords.get(2*i),   STARTING_TREE_SIZE);
        placeTree(players.get(1), validCoords.get(2*i+1), STARTING_TREE_SIZE);
    }
}
*/

func (g *Game) initStartingTrees() {
	var startingCoords []CubeCoord
	if g.STARTING_TREES_ON_EDGES {
		startingCoords = g.boardEdges()
	} else {
		startingCoords = make([]CubeCoord, 0, len(g.Board.Coords))
		for _, c := range g.Board.Coords {
			if c.X == 0 && c.Y == 0 && c.Z == 0 {
				continue
			}
			startingCoords = append(startingCoords, c)
		}
	}

	startingCoords = filterCoords(startingCoords, func(c CubeCoord) bool {
		return g.Board.Map[c].GetRichness() != RICHNESS_NULL
	})

	var validCoords []CubeCoord
	for len(validCoords) < g.STARTING_TREE_COUNT*2 {
		validCoords = g.tryInitStartingTrees(startingCoords)
	}

	for i := 0; i < g.STARTING_TREE_COUNT; i++ {
		g.placeTree(g.Players[0], g.Board.Map[validCoords[2*i]].GetIndex(), g.STARTING_TREE_SIZE)
		g.placeTree(g.Players[1], g.Board.Map[validCoords[2*i+1]].GetIndex(), g.STARTING_TREE_SIZE)
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:170-189

private List<CubeCoord> tryInitStartingTrees(List<CubeCoord> startingCoords) {
    List<CubeCoord> coordinates = new ArrayList<>();
    List<CubeCoord> availableCoords = new ArrayList<>(startingCoords);
    for (int i = 0; i < STARTING_TREE_COUNT; i++) {
        if (availableCoords.isEmpty()) return coordinates;
        int r = random.nextInt(availableCoords.size());
        CubeCoord normalCoord = availableCoords.get(r);
        CubeCoord oppositeCoord = normalCoord.getOpposite();
        availableCoords.removeIf(coord ->
            coord.distanceTo(normalCoord) <= STARTING_TREE_DISTANCE ||
            coord.distanceTo(oppositeCoord) <= STARTING_TREE_DISTANCE);
        coordinates.add(normalCoord);
        coordinates.add(oppositeCoord);
    }
    return coordinates;
}
*/

func (g *Game) tryInitStartingTrees(startingCoords []CubeCoord) []CubeCoord {
	coordinates := make([]CubeCoord, 0, g.STARTING_TREE_COUNT*2)
	available := append([]CubeCoord(nil), startingCoords...)
	for i := 0; i < g.STARTING_TREE_COUNT; i++ {
		if len(available) == 0 {
			return coordinates
		}
		r := g.random.NextInt(len(available))
		normal := available[r]
		opposite := normal.Opposite()
		available = filterCoords(available, func(c CubeCoord) bool {
			return c.DistanceTo(normal) > g.STARTING_TREE_DISTANCE &&
				c.DistanceTo(opposite) > g.STARTING_TREE_DISTANCE
		})
		coordinates = append(coordinates, normal, opposite)
	}
	return coordinates
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:191-205

private void calculateShadows() {
    shadows.clear();
    trees.forEach((index, tree) -> {
        CubeCoord coord = board.coords.get(index);
        for (int i = 1; i <= tree.getSize(); i++) {
            CubeCoord temp = coord.neighbor(sun.getOrientation(), i);
            if (board.map.containsKey(temp)) {
                shadows.compute(board.map.get(temp).getIndex(), (k, v) ->
                    v == null ? tree.getSize() : Math.max(v, tree.getSize()));
            }
        }
    });
}
*/

func (g *Game) calculateShadows() {
	for k := range g.Shadows {
		delete(g.Shadows, k)
	}
	for _, idx := range g.TreeOrder {
		tree := g.Trees[idx]
		coord := g.Board.Coords[idx]
		for i := 1; i <= tree.Size; i++ {
			temp := coord.NeighborAt(g.Sun.Orientation, i)
			if cell, ok := g.Board.Map[temp]; ok {
				cellIdx := cell.GetIndex()
				existing, has := g.Shadows[cellIdx]
				if !has || tree.Size > existing {
					g.Shadows[cellIdx] = tree.Size
				}
			}
		}
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:207-212

private List<CubeCoord> getBoardEdges() {
    CubeCoord centre = new CubeCoord(0, 0, 0);
    return board.coords.stream()
        .filter(coord -> coord.distanceTo(centre) == Config.MAP_RING_COUNT)
        .collect(Collectors.toList());
}
*/

func (g *Game) boardEdges() []CubeCoord {
	centre := NewCubeCoord(0, 0, 0)
	out := make([]CubeCoord, 0, len(g.Board.Coords))
	for _, c := range g.Board.Coords {
		if c.DistanceTo(centre) == g.Cfg.MAP_RING_COUNT {
			out = append(out, c)
		}
	}
	return out
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:382-391

public void resetGameTurnData() {
    dyingTrees.clear();
    availableSun.clear();
    sentSeeds.clear();
    for (Player p : gameManager.getPlayers()) {
        availableSun.add(p.getSun());
        p.reset();
    }
    currentFrameType = nextFrameType;
}
*/

func (g *Game) ResetGameTurnData() {
	g.DyingTrees = g.DyingTrees[:0]
	g.SentSeeds = g.SentSeeds[:0]
	g.AvailableSun = g.AvailableSun[:0]
	for _, p := range g.Players {
		g.AvailableSun = append(g.AvailableSun, p.GetSun())
		p.Reset()
	}
	g.CurrentFrameType = g.NextFrameType
}

// placeTree mirrors Java's Game.placeTree: insert a Tree into the trees map
// and refresh TreeOrder so iteration stays in ascending-index order (Java uses
// a TreeMap).
func (g *Game) placeTree(player *Player, index, size int) *Tree {
	tree := NewTree()
	tree.Size = size
	tree.Owner = player
	g.Trees[index] = tree
	g.refreshTreeOrder()
	return tree
}

func (g *Game) removeTree(index int) {
	delete(g.Trees, index)
	g.refreshTreeOrder()
}

func (g *Game) refreshTreeOrder() {
	if cap(g.TreeOrder) < len(g.Trees) {
		g.TreeOrder = make([]int, 0, len(g.Trees))
	} else {
		g.TreeOrder = g.TreeOrder[:0]
	}
	for k := range g.Trees {
		g.TreeOrder = append(g.TreeOrder, k)
	}
	sort.Ints(g.TreeOrder)
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:393-403

private int getGrowthCost(Tree targetTree) {
    int targetSize = targetTree.getSize() + 1;
    if (targetSize > Constants.TREE_TALL) return Constants.LIFECYCLE_END_COST;
    return getCostFor(targetSize, targetTree.getOwner());
}
private int getSeedCost(Player player) { return getCostFor(0, player); }
*/

func (g *Game) getGrowthCost(t *Tree) int {
	targetSize := t.Size + 1
	if targetSize > TREE_TALL {
		return LIFECYCLE_END_COST
	}
	return g.getCostFor(targetSize, t.Owner)
}

func (g *Game) getSeedCost(p *Player) int {
	return g.getCostFor(0, p)
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:466-473

private int getCostFor(int size, Player owner) {
    int baseCost = Constants.TREE_BASE_COST[size];
    int sameTreeCount = (int) trees.values().stream()
        .filter(t -> t.getSize() == size && t.getOwner() == owner)
        .count();
    return (baseCost + sameTreeCount);
}
*/

func (g *Game) getCostFor(size int, owner *Player) int {
	base := TREE_BASE_COST[size]
	same := 0
	for _, t := range g.Trees {
		if t.Size == size && t.Owner == owner {
			same++
		}
	}
	return base + same
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:405-434

private void doGrow(Player player, Action action) throws GameException { ... }
*/

func (g *Game) doGrow(player *Player, action Action, debug string) error {
	idx := action.GetTargetID()
	cell := g.Board.CellByIndex(idx)
	if !cell.IsValid() {
		return NewCellNotFoundException(idx)
	}
	tree, ok := g.Trees[cell.GetIndex()]
	if !ok {
		return NewTreeNotFoundException(cell.GetIndex())
	}
	if tree.Owner != player {
		return NewNotOwnerOfTreeException(cell.GetIndex(), tree.Owner)
	}
	if tree.Dormant {
		return NewAlreadyActivatedTree(cell.GetIndex())
	}
	if tree.Size >= TREE_TALL {
		return NewTreeAlreadyTallException(cell.GetIndex())
	}
	cost := g.getGrowthCost(tree)
	current := g.AvailableSun[player.GetIndex()]
	if current < cost {
		return NewNotEnoughSunException(cost, player.GetSun())
	}
	g.AvailableSun[player.GetIndex()] = current - cost
	tree.Grow()
	g.Summary.AddGrowTree(player, cell)
	g.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceGrow, GrowData{Cell: cell.GetIndex(), Cost: cost, Debug: debug}))
	tree.SetDormant()
	return nil
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:436-464

private void doComplete(Player player, Action action) throws GameException { ... }
*/

func (g *Game) doComplete(player *Player, action Action, debug string) error {
	idx := action.GetTargetID()
	cell := g.Board.CellByIndex(idx)
	if !cell.IsValid() {
		return NewCellNotFoundException(idx)
	}
	tree, ok := g.Trees[cell.GetIndex()]
	if !ok {
		return NewTreeNotFoundException(cell.GetIndex())
	}
	if tree.Owner != player {
		return NewNotOwnerOfTreeException(cell.GetIndex(), tree.Owner)
	}
	if tree.Size < TREE_TALL {
		return NewTreeNotTallException(cell.GetIndex())
	}
	if tree.Dormant {
		return NewAlreadyActivatedTree(cell.GetIndex())
	}
	cost := g.getGrowthCost(tree)
	current := g.AvailableSun[player.GetIndex()]
	if current < cost {
		return NewNotEnoughSunException(cost, player.GetSun())
	}
	g.AvailableSun[player.GetIndex()] = current - cost
	coord, _ := g.Board.CoordByIndex(cell.GetIndex())
	g.DyingTrees = append(g.DyingTrees, coord)
	tree.SetDormant()
	// Emit the COMPLETE trace eagerly (before any DEBUG message in the same
	// action loop). Score-award and tree-removal still happen later in
	// removeDyingTrees, but nutrients and richness don't change between here
	// and there, so the points computed now equal the points awarded then —
	// including the simultaneous-COMPLETE case where both players get full
	// points before updateNutrients drops the pool.
	points := g.Nutrients
	switch cell.GetRichness() {
	case RICHNESS_OK:
		points += RICHNESS_BONUS_OK
	case RICHNESS_LUSH:
		points += RICHNESS_BONUS_LUSH
	}
	g.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceComplete, CompleteData{Cell: cell.GetIndex(), Points: points, Cost: cost, Debug: debug}))
	return nil
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:475-525

private void doSeed(Player player, Action action) throws GameException { ... }
*/

func (g *Game) doSeed(player *Player, action Action, debug string) error {
	targetCell := g.Board.CellByIndex(action.GetTargetID())
	if !targetCell.IsValid() {
		return NewCellNotFoundException(action.GetTargetID())
	}
	sourceCell := g.Board.CellByIndex(action.GetSourceID())
	if !sourceCell.IsValid() {
		return NewCellNotFoundException(action.GetSourceID())
	}
	targetCoord, _ := g.Board.CoordByIndex(targetCell.GetIndex())
	sourceCoord, _ := g.Board.CoordByIndex(sourceCell.GetIndex())

	if g.aTreeIsOn(targetCell) {
		return NewCellNotEmptyException(targetCell.GetIndex())
	}
	sourceTree, ok := g.Trees[sourceCell.GetIndex()]
	if !ok {
		return NewTreeNotFoundException(sourceCell.GetIndex())
	}
	if sourceTree.Size == TREE_SEED {
		return NewTreeIsSeedException(sourceCell.GetIndex())
	}
	if sourceTree.Owner != player {
		return NewNotOwnerOfTreeException(sourceCell.GetIndex(), sourceTree.Owner)
	}
	if sourceTree.Dormant {
		return NewAlreadyActivatedTree(sourceCell.GetIndex())
	}

	distance := sourceCoord.DistanceTo(targetCoord)
	if distance > sourceTree.Size {
		return NewTreeTooFarException(sourceCell.GetIndex(), targetCell.GetIndex())
	}
	if targetCell.GetRichness() == RICHNESS_NULL {
		return NewCellNotValidException(targetCell.GetIndex())
	}

	cost := g.getSeedCost(player)
	current := g.AvailableSun[player.GetIndex()]
	if current < cost {
		return NewNotEnoughSunException(cost, player.GetSun())
	}
	g.AvailableSun[player.GetIndex()] = current - cost
	sourceTree.SetDormant()
	g.SentSeeds = append(g.SentSeeds, Seed{
		Owner:      player.GetIndex(),
		SourceCell: sourceCell.GetIndex(),
		TargetCell: targetCell.GetIndex(),
	})
	g.Summary.AddPlantSeed(player, targetCell, sourceCell)
	g.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceSeed, SeedData{Source: sourceCell.GetIndex(), Target: targetCell.GetIndex(), Cost: cost, Debug: debug}))
	return nil
}

func (g *Game) aTreeIsOn(cell *Cell) bool {
	_, ok := g.Trees[cell.GetIndex()]
	return ok
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:527-542

private void giveSun() {
    int[] givenToPlayer = new int[2];
    trees.forEach((index, tree) -> {
        if (!shadows.containsKey(index) || shadows.get(index) < tree.getSize()) {
            tree.getOwner().addSun(tree.getSize());
            givenToPlayer[tree.getOwner().getIndex()] += tree.getSize();
        }
    });
    // summary lines
}
*/

func (g *Game) giveSun() {
	given := [2]int{}
	for _, idx := range g.TreeOrder {
		tree := g.Trees[idx]
		shadow, has := g.Shadows[idx]
		gathered := 0
		if !has || shadow < tree.Size {
			tree.Owner.AddSun(tree.Size)
			given[tree.Owner.GetIndex()] += tree.Size
			gathered = tree.Size
		}
		g.tracePlayer(tree.Owner.GetIndex(), arena.MakeTurnTrace(TraceGather, GatherData{Cell: idx, Sun: gathered}))
	}
	for _, p := range g.Players {
		v := given[p.GetIndex()]
		if v > 0 {
			g.Summary.AddGather(p, v)
		}
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:544-565

private void removeDyingTrees() {
    dyingTrees.forEach(coord -> {
        Cell cell = board.map.get(coord);
        int points = nutrients;
        if (cell.getRichness() == RICHNESS_OK)   points += RICHNESS_BONUS_OK;
        else if (cell.getRichness() == RICHNESS_LUSH) points += RICHNESS_BONUS_LUSH;
        Player player = trees.get(cell.getIndex()).getOwner();
        player.addScore(points);
        trees.remove(cell.getIndex());
    });
}
*/

func (g *Game) removeDyingTrees() {
	for _, coord := range g.DyingTrees {
		cell := g.Board.Map[coord]
		points := g.Nutrients
		switch cell.GetRichness() {
		case RICHNESS_OK:
			points += RICHNESS_BONUS_OK
		case RICHNESS_LUSH:
			points += RICHNESS_BONUS_LUSH
		}
		player := g.Trees[cell.GetIndex()].Owner
		player.AddScore(points)
		g.Summary.AddCutTree(player, cell, points)
		g.removeTree(cell.GetIndex())
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:567-571

private void updateNutrients() {
    dyingTrees.forEach(coord -> { nutrients = Math.max(0, nutrients - 1); });
}
*/

func (g *Game) updateNutrients() {
	for range g.DyingTrees {
		g.Nutrients--
		if g.Nutrients < 0 {
			g.Nutrients = 0
		}
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:573-607

public void performGameUpdate() {
    turn++;
    switch (currentFrameType) {
    case GATHERING: ... performSunGatheringUpdate(); nextFrameType = ACTIONS; break;
    case ACTIONS:   ... performActionUpdate();
                    if (allPlayersAreWaiting()) nextFrameType = SUN_MOVE; break;
    case SUN_MOVE:  ... performSunMoveUpdate(); nextFrameType = GATHERING; break;
    }
    if (gameOver()) gameManager.endGame();
}
*/

func (g *Game) PerformGameUpdate(turn int) {
	g.Turn++
	g.traces = [2][]arena.TurnTrace{}
	g.seedConflictCell = nil

	switch g.CurrentFrameType {
	case FrameGathering:
		g.DayActionIndex = 0
		g.Summary.AddRound(g.Round)
		g.performSunGatheringUpdate()
		g.NextFrameType = FrameActions
	case FrameActions:
		g.Summary.AddRound(g.Round)
		g.performActionUpdate()
		g.DayActionIndex++
		if g.allPlayersAreWaiting() {
			g.NextFrameType = FrameSunMove
		}
	case FrameSunMove:
		g.Summary.AddRoundTransition(g.Round)
		g.performSunMoveUpdate()
		g.NextFrameType = FrameGathering
	default:
		// FrameInit shouldn't reach PerformGameUpdate; the runner calls
		// ResetGameTurnData first which transitions to NextFrameType.
	}

	if g.gameOver() {
		g.ended = true
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:609-618

public void performSunMoveUpdate() {
    round++;
    if (round < MAX_ROUNDS) {
        sun.move();
        if (ENABLE_SHADOW) calculateShadows();
    }
}
*/

func (g *Game) performSunMoveUpdate() {
	g.Round++
	if g.Round < g.MAX_ROUNDS {
		g.Sun.Move()
		if g.ENABLE_SHADOW {
			g.calculateShadows()
		}
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:620-632

public void performSunGatheringUpdate() {
    gameManager.getPlayers().forEach(p -> p.setWaiting(false));
    trees.forEach((index, tree) -> tree.reset());
    giveSun();
}
*/

func (g *Game) performSunGatheringUpdate() {
	for _, p := range g.Players {
		p.SetWaiting(false)
	}
	for _, t := range g.Trees {
		t.Reset()
	}
	g.giveSun()
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:634-673

public void performActionUpdate() {
    gameManager.getPlayers().stream().filter(p -> !p.isWaiting()).forEach(player -> {
        try {
            Action action = player.getAction();
            if (action.isGrow())     doGrow(player, action);
            else if (action.isSeed()) doSeed(player, action);
            else if (action.isComplete()) doComplete(player, action);
            else { player.setWaiting(true); summary.addWait(player); }
        } catch (GameException e) {
            summary.addError(...); player.setWaiting(true);
        }
    });
    if (seedsAreConflicting()) summary.addSeedConflict(sentSeeds.get(0));
    else { sentSeeds.forEach(seed -> plantSeed(...)); for (Player p : ...) p.setSun(availableSun.get(p.getIndex())); }
    removeDyingTrees();
    updateNutrients();
}
*/

func (g *Game) performActionUpdate() {
	for _, player := range g.Players {
		if player.IsWaiting() {
			continue
		}
		action := player.GetAction()
		debug := ""
		if player.HasMessage {
			debug = player.Message
		}
		var err error
		switch {
		case action.IsGrow():
			err = g.doGrow(player, action, debug)
		case action.IsSeed():
			err = g.doSeed(player, action, debug)
		case action.IsComplete():
			err = g.doComplete(player, action, debug)
		default:
			player.SetWaiting(true)
			g.Summary.AddWait(player)
			if debug != "" {
				g.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceWait, WaitData{Debug: debug}))
			} else {
				g.tracePlayer(player.GetIndex(), arena.TurnTrace{Type: TraceWait})
			}
		}
		if err != nil {
			g.Summary.AddError(fmt.Sprintf("%s: %s", player.NicknameToken(), err.Error()))
			player.SetWaiting(true)
		}
	}

	if g.seedsAreConflicting() {
		g.Summary.AddSeedConflict(g.SentSeeds[0])
		cell := g.SentSeeds[0].TargetCell
		g.seedConflictCell = &cell
	} else {
		for _, seed := range g.SentSeeds {
			g.plantSeed(g.Players[seed.Owner], seed.TargetCell, seed.SourceCell)
		}
		for _, p := range g.Players {
			p.SetSun(g.AvailableSun[p.GetIndex()])
		}
	}
	g.removeDyingTrees()
	g.updateNutrients()
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:675-684

private boolean seedsAreConflicting() {
    return sentSeeds.size() != sentSeeds.stream()
        .map(seed -> seed.getTargetCell()).distinct().count();
}
*/

func (g *Game) seedsAreConflicting() bool {
	seen := make(map[int]struct{}, len(g.SentSeeds))
	for _, s := range g.SentSeeds {
		seen[s.TargetCell] = struct{}{}
	}
	return len(seen) != len(g.SentSeeds)
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:686-690

private boolean allPlayersAreWaiting() {
    return gameManager.getPlayers().stream().filter(Player::isWaiting).count() == playerCount;
}
*/

func (g *Game) allPlayersAreWaiting() bool {
	for _, p := range g.Players {
		if !p.IsWaiting() {
			return false
		}
	}
	return true
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:692-704

private void plantSeed(Player player, int index, int fatherIndex) {
    Tree seed = placeTree(player, index, Constants.TREE_SEED);
    seed.setDormant();
    seed.setFatherIndex(fatherIndex);
}
*/

func (g *Game) plantSeed(player *Player, index, fatherIndex int) {
	seed := g.placeTree(player, index, TREE_SEED)
	seed.SetDormant()
	seed.FatherIndex = fatherIndex
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:706-723

public void onEnd() {
    gameManager.getActivePlayers().forEach(p -> p.addScore((int) Math.floor(p.getSun() / 3)));
    if (gameManager.getActivePlayers().stream().map(Player::getScore).distinct().count() == 1) {
        trees.forEach((index, tree) -> {
            if (tree.getOwner().isActive()) { tree.getOwner().addBonusScore(1); tree.getOwner().addScore(1); }
        });
    }
}
*/

func (g *Game) OnEnd() {
	for _, p := range g.Players {
		if !p.IsActive() {
			continue
		}
		p.AddScore(p.GetSun() / 3)
	}

	scoresEqual := true
	var firstScore int
	first := true
	for _, p := range g.Players {
		if !p.IsActive() {
			continue
		}
		if first {
			firstScore = p.GetScore()
			first = false
			continue
		}
		if p.GetScore() != firstScore {
			scoresEqual = false
			break
		}
	}

	if !first && scoresEqual {
		for _, idx := range g.TreeOrder {
			tree := g.Trees[idx]
			if tree.Owner.IsActive() {
				tree.Owner.AddBonusScore(1)
				tree.Owner.AddScore(1)
			}
		}
	}
}

func (g *Game) EndGame() {
	g.ended = true
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:738-740

private boolean gameOver() {
    return gameManager.getActivePlayers().size() <= 1 || round >= MAX_ROUNDS;
}
*/

func (g *Game) gameOver() bool {
	active := 0
	for _, p := range g.Players {
		if p.IsActive() {
			active++
		}
	}
	return active <= 1 || g.Round >= g.MAX_ROUNDS
}

func (g *Game) IsGameOver() bool { return g.gameOver() }
func (g *Game) Ended() bool      { return g.ended }

func (g *Game) ShouldSkipPlayerTurn(p *Player) bool {
	// Java only solicits input from non-waiting players in the ACTIONS frame.
	return g.CurrentFrameType != FrameActions || p.IsWaiting()
}

func filterCoords(in []CubeCoord, keep func(CubeCoord) bool) []CubeCoord {
	out := in[:0]
	for _, c := range in {
		if keep(c) {
			out = append(out, c)
		}
	}
	return out
}
