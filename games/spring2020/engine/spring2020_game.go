// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java
package engine

import (
	"fmt"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/javarand"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:34-48

@Singleton
public class Game {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject TetrisBasedMapGenerator mapGenerator;

    List<Pacman> pacmen;
    Random random;
    Grid grid;
    private int totalPacmen;
    private int pacmenPerPlayer;
    private int currentStep = 0;
}
*/

// Game holds all simulation state.
type Game struct {
	Config  Config
	Players []*Player
	Pacmen  []*Pacman
	Grid    *Grid

	Random      *javarand.Random
	PacsPer     int
	TotalPacs   int
	CurrentStep int

	EndedFlag         bool
	GameOverProcessed bool
	Summary           []string
	traces            [2][]arena.TurnTrace
}

// NewGame sets up a fresh simulation with the given seed and league level.
// Players must be added via Init before running turns.
func NewGame(seed int64, leagueLevel int) *Game {
	rules := LeagueRulesFromIndex(leagueLevel)
	cfg := NewConfig(rules)
	g := &Game{
		Config: cfg,
		Random: javarand.New(seed),
	}
	return g
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:115-204

public void init(long seed) {
    random = new Random(seed);
    if (grid == null) {
        grid = generateGrid(seed);
    }
    if (pacmenPerPlayer == 0) {
        pacmenPerPlayer = randomInt(Config.MIN_PACS_PER_PLAYER, Config.MAX_PACS_PER_PLAYER);
    }
    totalPacmen = pacmenPerPlayer * gameManager.getPlayerCount();
    if (pacmen == null) {
        generatePacmen();
        // shuffle free cells, place pacmen, place cherries
    }
    // generate pellets on every floor cell that is not a cherry and is not occupied
}
*/

// Init seeds grid, pacmen, pellets, and cherries. Must be called before turns.
func (g *Game) Init(players []*Player) {
	g.Players = players
	g.Grid = g.GenerateGrid()

	if g.PacsPer == 0 {
		g.PacsPer = g.RandomInt(g.Config.MIN_PACS_PER_PLAYER, g.Config.MAX_PACS_PER_PLAYER)
	}
	g.TotalPacs = g.PacsPer * len(g.Players)

	if g.Pacmen == nil {
		g.GeneratePacmen()
		g.PlacePacmenAndCherries()
	}

	g.SpawnPellets()
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:50-52

private int randomInt(int min, int max) {
    return min + random.nextInt(max - min + 1);
}
*/

func (g *Game) RandomInt(min, max int) int {
	return min + g.Random.NextInt(max-min+1)
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:54-105

private Grid generateGrid(long seed) {
    int width = randomInt(Config.MAP_MIN_WIDTH, Config.MAP_MAX_WIDTH);
    int height = randomInt(Config.MAP_MIN_HEIGHT, Config.MAP_MAX_HEIGHT);
    if (gameManager.getPlayerCount() == 2 && width % 2 == 0) width++;

    Grid grid = new Grid(width, height);
    if (gameManager.getPlayerCount() == 2) {
        mapGenerator.generateWithHorizontalSymetry(grid, random);
    } else {
        mapGenerator.generate(grid, random);
    }

    if (Config.MAP_WRAPS) {
        Grid bigGrid = new Grid(width + 2, height + 2);
        for (int y = 0; y < height + 2; ++y) {
            for (int x = 0; x < width + 2; ++x) {
                Coord gridPos = new Coord(x - 1, y - 1);
                if (isOuterBorder(x, y, width, height)) {
                    if (x == 0 && isTunnelExit(grid, gridPos.x + 1, gridPos.y)) {
                        bigGrid.get(x, y).setType(CellType.FLOOR);
                    } else if (x == width + 1 && isTunnelExit(grid, gridPos.x - 1, gridPos.y)) {
                        bigGrid.get(x, y).setType(CellType.FLOOR);
                    } else {
                        bigGrid.get(x, y).setType(CellType.WALL);
                    }
                } else {
                    bigGrid.get(x, y).copy(grid.get(x - 1, y - 1));
                }
            }
        }
        grid = bigGrid;
    }
    return grid;
}
*/

func (g *Game) GenerateGrid() *Grid {
	width := g.RandomInt(g.Config.MAP_MIN_WIDTH, g.Config.MAP_MAX_WIDTH)
	height := g.RandomInt(g.Config.MAP_MIN_HEIGHT, g.Config.MAP_MAX_HEIGHT)
	gen := NewTetrisBasedMapGenerator()

	if len(g.Players) == 2 && width%2 == 0 {
		width++
	}

	innerMapWraps := g.Config.MAP_WRAPS
	inner := NewGrid(width, height, innerMapWraps)
	if len(g.Players) == 2 {
		gen.GenerateWithHorizontalSymmetry(inner, g.Random)
	} else {
		gen.Generate(inner, g.Random)
	}

	if !g.Config.MAP_WRAPS {
		return inner
	}

	// Wrap: surround with a 1-cell tunnel border.
	bigW, bigH := width+2, height+2
	big := NewGrid(bigW, bigH, true)
	for y := 0; y < bigH; y++ {
		for x := 0; x < bigW; x++ {
			gridPos := Coord{X: x - 1, Y: y - 1}
			if IsOuterBorder(x, y, width, height) {
				switch {
				case x == 0 && IsTunnelExit(inner, gridPos.X+1, gridPos.Y):
					big.GetXY(x, y).Type = CellFloor
				case x == width+1 && IsTunnelExit(inner, gridPos.X-1, gridPos.Y):
					big.GetXY(x, y).Type = CellFloor
				default:
					big.GetXY(x, y).Type = CellWall
				}
			} else {
				big.GetXY(x, y).Copy(inner.GetXY(x-1, y-1))
			}
		}
	}
	return big
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:107-113

private boolean isTunnelExit(Grid grid, int x, int y) {
    return grid.get(x, y).isFloor() && !grid.get(x, y - 1).isFloor() && !grid.get(x, y + 1).isFloor();
}
private boolean isOuterBorder(int x, int y, int width, int height) {
    return x == 0 || y == 0 || x == width + 1 || y == height + 1;
}
*/

func IsTunnelExit(g *Grid, x, y int) bool {
	return g.GetXY(x, y).IsFloor() && !g.GetXY(x, y-1).IsFloor() && !g.GetXY(x, y+1).IsFloor()
}

func IsOuterBorder(x, y, width, height int) bool {
	return x == 0 || y == 0 || x == width+1 || y == height+1
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:285-302

private void generatePacmen() {
    pacmen = new ArrayList<>(totalPacmen);
    int pacmanIndex = 0;
    int typeIndex = 0;
    while (pacmanIndex < totalPacmen) {
        for (Player player : gameManager.getPlayers()) {
            if (pacmanIndex < totalPacmen) {
                PacmanType type = Config.SWITCH_ABILITY_AVAILABLE ? PacmanType.values()[typeIndex % 3] : PacmanType.NEUTRAL;
                Pacman pac = new Pacman(pacmanIndex, player.getPacmen().size(), player, type);
                player.addPacman(pac);
                pacmanIndex++;
                pacmen.add(pac);
            }
        }
        typeIndex++;
    }
}
*/

func (g *Game) GeneratePacmen() {
	g.Pacmen = make([]*Pacman, 0, g.TotalPacs)
	pacmanIndex := 0
	typeIndex := 0
	rotation := [3]PacmanType{TypeRock, TypePaper, TypeScissors}
	for pacmanIndex < g.TotalPacs {
		for _, player := range g.Players {
			if pacmanIndex >= g.TotalPacs {
				break
			}
			t := TypeNeutral
			if g.Config.SWITCH_ABILITY_AVAILABLE {
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:134-162

List<Coord> freeCells = grid.cells.entrySet()
    .stream()
    .filter(entry -> entry.getValue().getType() == CellType.FLOOR)
    .filter(entry -> entry.getKey().getX() != grid.getWidth() / 2)
    .map(entry -> entry.getKey())
    .collect(Collectors.toList());

Collections.shuffle(freeCells, random);

if (gameManager.getPlayerCount() == 2) {
    List<Coord> leftCells = freeCells.stream()
        .filter(c -> c.getX() <= grid.getWidth() / 2)
        .collect(Collectors.toList());
    int i;
    for (i = 0; i < pacmenPerPlayer; ++i) {
        Coord leftCell = leftCells.get(i);
        Coord rightCell = new Coord(grid.getWidth() - 1 - leftCell.getX(), leftCell.getY());
        int leftPlayer = random.nextInt(2);
        int rightPlayer = (leftPlayer + 1) % 2;
        gameManager.getPlayer(leftPlayer).getPacmen().get(i).setPosition(leftCell);
        gameManager.getPlayer(rightPlayer).getPacmen().get(i).setPosition(rightCell);
    }
    for (int j = 0; j < Config.NUMBER_OF_CHERRIES / 2; j++) {
        Coord leftCell = leftCells.get(i + j);
        Coord rightCell = new Coord(grid.getWidth() - 1 - leftCell.getX(), leftCell.getY());
        grid.get(leftCell).setHasCherry(true);
        grid.get(rightCell).setHasCherry(true);
    }
}
*/

func (g *Game) PlacePacmenAndCherries() {
	freeCells := make([]Coord, 0)
	halfX := g.Grid.Width / 2
	for _, cell := range g.Grid.Cells {
		if cell.Type == CellFloor && cell.Coord.X != halfX {
			freeCells = append(freeCells, cell.Coord)
		}
	}
	Shuffle(freeCells, g.Random)

	if len(g.Players) == 2 {
		leftCells := make([]Coord, 0)
		for _, c := range freeCells {
			if c.X <= halfX {
				leftCells = append(leftCells, c)
			}
		}
		var i int
		for i = 0; i < g.PacsPer; i++ {
			left := leftCells[i]
			right := Coord{X: g.Grid.Width - 1 - left.X, Y: left.Y}
			leftPlayer := g.Random.NextInt(2)
			rightPlayer := (leftPlayer + 1) % 2
			g.Players[leftPlayer].Pacmen[i].Position = left
			g.Players[rightPlayer].Pacmen[i].Position = right
		}
		for j := 0; j < g.Config.NUMBER_OF_CHERRIES/2; j++ {
			left := leftCells[i+j]
			right := Coord{X: g.Grid.Width - 1 - left.X, Y: left.Y}
			g.Grid.Get(left).HasCherry = true
			g.Grid.Get(right).HasCherry = true
		}
	} else {
		for i := 0; i < g.TotalPacs; i++ {
			g.Pacmen[i].Position = freeCells[i]
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:190-200

grid.cells.entrySet().stream().forEach(entry -> {
    Coord coord = entry.getKey();
    Cell cell = entry.getValue();
    boolean spawnPellet = cell.isFloor() && !cell.hasCherry() &&
        pacmen.stream().filter(pac -> pac.getPosition().equals(coord)).count() == 0;
    if (spawnPellet) {
        cell.setHasPellet(true);
    }
});
*/

func (g *Game) SpawnPellets() {
	for _, cell := range g.Grid.Cells {
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:407-410

public void resetGameTurnData() {
    currentStep = 0;
    gameManager.getPlayers().stream().forEach(Player::turnReset);
}
*/

// ResetGameTurnData is invoked at the start of each main (non-speed) turn.
func (g *Game) ResetGameTurnData() {
	g.CurrentStep = 0
	for _, p := range g.Players {
		p.TurnReset()
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:1004-1011

public boolean isSpeedTurn() {
    int numSteps = pacmen.stream()
        .mapToInt(p -> p.getIntendedPath().size())
        .max()
        .getAsInt() - 1;
    return currentStep > 0 && currentStep < numSteps;
}
*/

// IsSpeedTurn reports whether the next update should run as a speed sub-turn.
func (g *Game) IsSpeedTurn() bool {
	numSteps := 0
	for _, p := range g.Pacmen {
		if len(p.IntendedPath) > numSteps {
			numSteps = len(p.IntendedPath)
		}
	}
	numSteps--
	return g.CurrentStep > 0 && g.CurrentStep < numSteps
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:412-424

public void performGameUpdate() {
    executePacmenAbilities();
    updateAbilityModifiers();
    processPacmenIntent();
    resolveMovement();
}

The Java referee splits the speed sub-turn (a second movement step for SPEED
pacs) into a separate `gameTurn` frame that does not read player input. We
fold that loop into a single PerformGameUpdate: bots are read once, all
movement steps for the turn (1 or 2) resolve before the next read. Mechanics
are identical because Java never reads input between steps either —
performGameSpeedUpdate only re-runs resolveMovement against the already
sliced intendedPath.

Java checks isGameOver() after every frame and sets gameOverFrame; once
gameOverFrame is true the next gameTurn skips straight to the post-game
branch, so a sub-turn queued after a game-ending main step never runs. We
mirror that by guarding the inner loop with !IsGameOver — without it a
SPEED pac can take a second step on the final turn and eat a pellet that
Java's referee would have left on the grid.
*/

// PerformGameUpdate runs one full main turn including any speed sub-steps.
func (g *Game) PerformGameUpdate() {
	g.traces = [2][]arena.TurnTrace{}
	g.ExecutePacmenAbilities()
	g.UpdateAbilityModifiers()
	g.ProcessPacmenIntent()
	g.ResolveMovement()
	for g.IsSpeedTurn() && !g.IsGameOver() {
		g.PerformGameSpeedUpdate()
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:426-451

public void performGameSpeedUpdate() {
    for (Pacman pac : pacmen) {
        if (pac.getSpeed() <= 1 || pac.getIntent().getActionType() != ActionType.MOVE) {
            List<Coord> path = new ArrayList<>();
            path.add(pac.getPosition());
            pac.setIntendedPath(path);
        }
        pac.setBlocked(false);
    }
    resolveMovement();
}
*/

// PerformGameSpeedUpdate runs one extra speed sub-turn (SPEED ability step 2+).
func (g *Game) PerformGameSpeedUpdate() {
	for _, pac := range g.Pacmen {
		if pac.Speed <= 1 || pac.Intent.Type != ActionMove {
			pac.IntendedPath = []Coord{pac.Position}
		}
		pac.Blocked = false
	}
	g.ResolveMovement()
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:682-752

private void executePacmenAbilities() {
    for (Pacman pac : pacmen) {
        if (pac.getAbilityToUse() != null) {
            Ability.Type ability = pac.getAbilityToUse();
            if (pac.getAbilityCooldown() != 0) continue;
            if (!Config.SPEED_ABILITY_AVAILABLE && ability == Ability.Type.SPEED
                || !Config.SWITCH_ABILITY_AVAILABLE && ability != Ability.Type.SPEED) continue;

            if (ability == Type.SET_ROCK || ability == Type.SET_PAPER || ability == Type.SET_SCISSORS) {
                pac.setType(getPacmanTypeFromAbility(ability));
            } else if (ability == Type.SPEED) {
                pac.setSpeed(Config.SPEED_BOOST);
                pac.setAbilityDuration(Config.ABILITY_DURATION);
            }
            pac.setAbilityCooldown(Config.ABILITY_COOLDOWN);
        }
    }
}
*/

func (g *Game) ExecutePacmenAbilities() {
	for _, pac := range g.Pacmen {
		if !pac.HasAbilityToUse {
			continue
		}
		ability := pac.AbilityToUse

		if pac.AbilityCooldown != 0 {
			continue
		}
		if (!g.Config.SPEED_ABILITY_AVAILABLE && ability == AbilitySpeed) ||
			(!g.Config.SWITCH_ABILITY_AVAILABLE && ability != AbilitySpeed) {
			continue
		}

		switch ability {
		case AbilitySetRock, AbilitySetPaper, AbilitySetScissors:
			pac.Type = PacTypeFromAbility(ability)
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceSwitch, SwitchMeta{Pac: pac.ID, Type: pac.Type.Name()}))
		case AbilitySpeed:
			pac.Speed = g.Config.SPEED_BOOST
			pac.AbilityDuration = g.Config.ABILITY_DURATION
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceSpeed, PacMeta{Pac: pac.ID}))
		}
		pac.AbilityCooldown = g.Config.ABILITY_COOLDOWN
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:559-566

private void updateAbilityModifiers() {
    for (Pacman pac : pacmen) {
        if (pac.getSpeed() > 1 && pac.getAbilityDuration() == 0) {
            pac.setSpeed(Config.PACMAN_BASE_SPEED);
        }
    }
}
*/

func (g *Game) UpdateAbilityModifiers() {
	for _, pac := range g.Pacmen {
		if pac.Speed > 1 && pac.AbilityDuration == 0 {
			pac.Speed = g.Config.PACMAN_BASE_SPEED
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:669-680

private void processPacmenIntent() {
    for (Pacman pac : pacmen) {
        if (pac.getIntent().getActionType() == ActionType.MOVE) {
            pac.setIntendedPath(computeIntendedPath(pac));
        } else {
            ArrayList<Coord> intendedPath = new ArrayList<>();
            intendedPath.add(pac.getPosition());
            pac.setIntendedPath(intendedPath);
        }
    }
}
*/

func (g *Game) ProcessPacmenIntent() {
	for _, pac := range g.Pacmen {
		if pac.Intent.Type == ActionMove {
			pac.IntendedPath = g.ComputeIntendedPath(pac)
		} else {
			pac.IntendedPath = []Coord{pac.Position}
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:767-810

private List<Coord> computeIntendedPath(Pacman pac) {
    MoveAction intent = (MoveAction) pac.getIntent();
    PathFinderResult pfr = pathfinder.setGrid(grid).from(pac.getPosition()).to(intent.getTarget()).findPath();
    List<Coord> wholePath = pfr.path;
    if (pfr.isNearest) {
        // build a "target unreachable" warning message
    } else {
        pac.setWarningPathMessage(null);
    }

    List<Coord> pathThisTurn = new ArrayList<>();
    if (wholePath.size() > 1) {
        int stepsThisTurn = Math.min(pac.getSpeed(), wholePath.size() - 1);
        pathThisTurn = wholePath.subList(0, stepsThisTurn + 1);
    } else {
        pathThisTurn.add(pac.getPosition());
    }
    return pathThisTurn;
}
*/

func (g *Game) ComputeIntendedPath(pac *Pacman) []Coord {
	target := pac.Intent.Target
	pfr := FindPath(g.Grid, pac.Position, target, nil)
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
	return []Coord{pac.Position}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:453-537

private void resolveMovement() {
    Set<Pacman> pacmenToKill = new HashSet<>();
    MovementResolution resolution = resolvePacmenMovement();
    for (Pacman pac : pacmen) {
        otherPacmen(pac).forEach(other -> {
            if (canEat(pac, other) && pacmenHaveCollided(pac, other)) {
                pacmenToKill.add(other);
            }
        });
    }
    killPacmen(pacmenToKill);
    eatPellets();
    eatCherries();
    currentStep++;
}
*/

func (g *Game) ResolveMovement() {
	type pendingKill struct {
		victim *Pacman
		killer *Pacman
	}
	var kills []pendingKill
	seen := make(map[*Pacman]struct{})

	resolution := g.ResolvePacmenMovement()

	for _, pac := range resolution.BlockedPacmen {
		blocker := resolution.BlockerOf(pac)
		if blocker == nil {
			continue
		}
		event := arena.MakeTurnTrace(TraceCollideSelf, PacMeta{Pac: pac.ID})
		if blocker.Owner != pac.Owner {
			event = arena.MakeTurnTrace(TraceCollideEnemy, PacMeta{Pac: pac.ID})
			g.traceBoth(event)
			continue
		}
		g.tracePlayer(pac.Owner.Index, event)
	}

	for _, pac := range g.Pacmen {
		for _, other := range g.Pacmen {
			if pac == other {
				continue
			}
			if g.CanEat(pac, other) && g.PacmenHaveCollided(pac, other) {
				if _, ok := seen[other]; !ok {
					seen[other] = struct{}{}
					kills = append(kills, pendingKill{victim: other, killer: pac})
				}
			}
		}
	}

	pacmenToKill := make([]*Pacman, len(kills))
	for i, k := range kills {
		pacmenToKill[i] = k.victim
		g.tracePlayer(k.victim.Owner.Index, arena.MakeTurnTrace(TraceKilled, KilledMeta{
			Pac:    k.victim.ID,
			Coord:  coordPair(k.victim.Position),
			Killer: k.killer.ID,
		}))
	}

	g.KillPacmen(pacmenToKill)

	g.EatPellets()
	g.EatCherries()

	g.CurrentStep++
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:304-318

public boolean canEat(Pacman pac1, Pacman pac2) {
    if (pac1.getOwner().equals(pac2.getOwner())) return false;
    if (pac1.getType() == PacmanType.PAPER)    return pac2.getType() == PacmanType.ROCK;
    if (pac1.getType() == PacmanType.ROCK)     return pac2.getType() == PacmanType.SCISSORS;
    if (pac1.getType() == PacmanType.SCISSORS) return pac2.getType() == PacmanType.PAPER;
    return false;
}
*/

func (g *Game) CanEat(a, b *Pacman) bool {
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:568-574

private boolean pacmenHaveCollided(Pacman a, Pacman b) {
    Coord fromA = getIntendedPositionAtStep(a, a.getPreviousPathStep());
    Coord fromB = getIntendedPositionAtStep(b, b.getPreviousPathStep());
    Coord toA = getIntendedPositionAtStep(a, a.getCurrentPathStep());
    Coord toB = getIntendedPositionAtStep(b, b.getCurrentPathStep());
    return toA.equals(toB) || (toA.equals(fromB) && toB.equals(fromA));
}
*/

func (g *Game) PacmenHaveCollided(a, b *Pacman) bool {
	fromA := g.IntendedPositionAtStep(a, a.PreviousPathStep)
	fromB := g.IntendedPositionAtStep(b, b.PreviousPathStep)
	toA := g.IntendedPositionAtStep(a, a.CurrentPathStep)
	toB := g.IntendedPositionAtStep(b, b.CurrentPathStep)
	if toA == toB {
		return true
	}
	return toA == fromB && toB == fromA
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:584-588

private static Comparator<? super Pacman> byDistanceTo(Coord position) {
    return (a, b) -> {
        return a.getPosition().manhattanTo(position) - b.getPosition().manhattanTo(position);
    };
}
*/

// Plain Manhattan distance (no map wrap), matching Java's byDistanceTo
// comparator. Do not switch to Grid.CalculateDistance — that wraps and
// produces a different blocker ranking on toroidal maps.

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:590-627

private MovementResolution resolvePacmenMovement() {
    MovementResolution resolution = new MovementResolution();
    List<Pacman> pacmenToResolve = pacmen.stream()
        .filter(pac -> pac.getIntent().getActionType() == ActionType.MOVE)
        .filter(pac -> !pac.moveFinished())
        .filter(getPacFastEnoughFilter(currentStep))
        .collect(Collectors.toList());

    List<Pacman> resolvedPacmen = new ArrayList<>();
    do {
        resolvedPacmen.clear();
        for (Pacman pac : pacmenToResolve) {
            Optional<Pacman> blockedBy = otherPacmen(pac)
                .filter(other -> isBodyBlockedBy(pac, other))
                .min(byDistanceTo(pac.getPosition()));
            if (blockedBy.isPresent()) {
                resolvedPacmen.add(pac);
                resolution.addBlockedPacmen(pac);
                resolution.blockedBy.put(pac, blockedBy.get());
            }
        }
        resolvedPacmen.forEach(pac -> pac.setBlocked(true));
        pacmenToResolve.removeAll(resolvedPacmen);
    } while (!resolvedPacmen.isEmpty());

    for (Pacman pac : pacmenToResolve) {
        movePacman(pac);
        resolution.addMovedPacman(pac);
    }
    return resolution;
}
*/

func (g *Game) ResolvePacmenMovement() *MovementResolution {
	res := NewMovementResolution()
	pacmenToResolve := make([]*Pacman, 0)
	for _, pac := range g.Pacmen {
		if pac.Intent.Type == ActionMove && !pac.MoveFinished() && pac.FastEnoughToMoveAt(g.CurrentStep) {
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
				if !g.IsBodyBlockedBy(pac, other) {
					continue
				}
				d := pac.Position.ManhattanTo(other.Position)
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
		pacmenToResolve = RemoveAll(pacmenToResolve, resolved)
		if len(resolved) == 0 {
			break
		}
	}

	for _, pac := range pacmenToResolve {
		g.MovePacman(pac)
		res.AddMoved(pac)
	}
	return res
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:636-656

private boolean isBodyBlockedBy(Pacman pac, Pacman other) {
    if (!Config.BODY_BLOCK) return false;
    if (!Config.FRIENDLY_BODY_BLOCK && pac.getOwner() == other.getOwner()) return false;
    // Never blocked against something pac can eat
    if (canEat(pac, other)) return false;
    // If beaten, can go to same coord (we only block crossing in that case)
    if (canEat(other, pac) && pacmenWillShareSameCoord(pac, other)) return false;
    return pacmenWillCollide(pac, other);
}
*/

func (g *Game) IsBodyBlockedBy(pac, other *Pacman) bool {
	if !g.Config.BODY_BLOCK {
		return false
	}
	if !g.Config.FRIENDLY_BODY_BLOCK && pac.Owner == other.Owner {
		return false
	}
	if g.CanEat(pac, other) {
		return false
	}
	if g.CanEat(other, pac) && g.PacmenWillShareSameCoord(pac, other) {
		return false
	}
	return g.PacmenWillCollide(pac, other)
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:658-670

private boolean pacmenWillShareSameCoord(Pacman a, Pacman b) {
    Coord toA = getIntendedPositionAtStep(a, a.gotBlocked() ? a.getCurrentPathStep() : a.getCurrentPathStep() + 1);
    Coord toB = getIntendedPositionAtStep(b, b.gotBlocked() ? b.getCurrentPathStep() : b.getCurrentPathStep() + 1);
    return toA.equals(toB);
}

private boolean pacmenWillCollide(Pacman a, Pacman b) {
    Coord fromA = getIntendedPositionAtStep(a, a.getCurrentPathStep());
    Coord fromB = getIntendedPositionAtStep(b, b.getCurrentPathStep());
    Coord toA = a.gotBlocked() ? fromA : getIntendedPositionAtStep(a, a.getCurrentPathStep() + 1);
    Coord toB = b.gotBlocked() ? fromB : getIntendedPositionAtStep(b, b.getCurrentPathStep() + 1);
    return toA.equals(toB) || (toA.equals(fromB) && toB.equals(fromA));
}
*/

func (g *Game) PacmenWillShareSameCoord(a, b *Pacman) bool {
	aStep := a.CurrentPathStep + 1
	if a.Blocked {
		aStep = a.CurrentPathStep
	}
	bStep := b.CurrentPathStep + 1
	if b.Blocked {
		bStep = b.CurrentPathStep
	}
	return g.IntendedPositionAtStep(a, aStep) == g.IntendedPositionAtStep(b, bStep)
}

func (g *Game) PacmenWillCollide(a, b *Pacman) bool {
	fromA := g.IntendedPositionAtStep(a, a.CurrentPathStep)
	fromB := g.IntendedPositionAtStep(b, b.CurrentPathStep)
	toA := fromA
	if !a.Blocked {
		toA = g.IntendedPositionAtStep(a, a.CurrentPathStep+1)
	}
	toB := fromB
	if !b.Blocked {
		toB = g.IntendedPositionAtStep(b, b.CurrentPathStep+1)
	}
	if toA == toB {
		return true
	}
	return toA == fromB && toB == fromA
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:672-674

private Coord getIntendedPositionAtStep(Pacman pac, int step) {
    return pac.getIntendedPath().get(Math.min(step, pac.getIntendedPath().size() - 1));
}
*/

func (g *Game) IntendedPositionAtStep(pac *Pacman, step int) Coord {
	if step > len(pac.IntendedPath)-1 {
		step = len(pac.IntendedPath) - 1
	}
	if step < 0 {
		step = 0
	}
	return pac.IntendedPath[step]
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:819-841

private void movePacman(Pacman pac) {
    pac.setCurrentPathStep(pac.getCurrentPathStep() + 1);
    Coord to = getIntendedPositionAtStep(pac, pac.getCurrentPathStep());
    pac.setPosition(to);
}
*/

func (g *Game) MovePacman(pac *Pacman) {
	pac.SetCurrentPathStep(pac.CurrentPathStep + 1)
	to := g.IntendedPositionAtStep(pac, pac.CurrentPathStep)
	pac.Position = to
	pac.WarnPathMsg = ""
	pac.HasWarnPathMsg = false
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:843-866

private void killPacmen(Collection<Pacman> pacmenToKill) {
    for (Pacman pac : pacmenToKill) {
        pacmen.remove(pac);
        pac.setDead();
        Player pacOwner = pac.getOwner();
        if (!pacOwner.getAlivePacmen().findAny().isPresent()) {
            pacOwner.deactivate();
        }
    }
}
*/

func (g *Game) KillPacmen(pacmenToKill []*Pacman) {
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:868-911

private void eatItem(Function<Cell, Boolean> hasItem, int pelletValue) {
    Map<Coord, List<Pacman>> eatenBy = new HashMap<>();
    for (Pacman pac : pacmen) {
        Cell cell = grid.get(pac.getPosition());
        if (hasItem.apply(cell)) {
            // collect: dedupe per (coord, owner) — only first pac per owner is credited per cell
        }
    }
    eatenBy.forEach((coord, list) -> {
        list.stream().map(Pacman::getOwner).distinct().forEach(player -> {
            player.pellets += pelletValue;
        });
        grid.get(coord).setHasPellet(false);
        grid.get(coord).setHasCherry(false);
    });
}
private void eatPellets()  { eatItem(Cell::hasPellet, 1); }
private void eatCherries() { eatItem(Cell::hasCherry, Config.CHERRY_SCORE); }
*/

func (g *Game) EatPellets() {
	g.EatItem(func(c *Cell) bool { return c.HasPellet }, 1)
}

func (g *Game) EatCherries() {
	g.EatItem(func(c *Cell) bool { return c.HasCherry }, CHERRY_SCORE)
}

func (g *Game) EatItem(hasItem func(*Cell) bool, pelletValue int) {
	type eater struct {
		coord Coord
		pacs  []*Pacman
	}
	byCoord := make(map[Coord]*eater)
	order := make([]Coord, 0)
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
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceEat, EatMeta{
				Pac:   pac.ID,
				Coord: coordPair(coord),
				Cost:  pelletValue,
			}))
		}
		cell := g.Grid.Get(coord)
		cell.HasPellet = false
		cell.HasCherry = false
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:921-930

public boolean isGameOver() {
    List<Player> activePlayers = gameManager.getActivePlayers();
    if (activePlayers.size() <= 1) return true;
    return gameManager.getActivePlayers().stream().noneMatch(this::canImproveRanking);
}
*/

// IsGameOver mirrors Game.isGameOver.
func (g *Game) IsGameOver() bool {
	activePlayers := g.ActivePlayers()
	if len(activePlayers) <= 1 {
		return true
	}
	for _, p := range activePlayers {
		if g.CanImproveRanking(p) {
			return false
		}
	}
	return true
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:932-965

public void performGameOver() {
    List<Player> activePlayers = gameManager.getActivePlayers();
    if (activePlayers.size() <= 1) {
        if (activePlayers.size() == 1) {
            activePlayers.get(0).pellets += getRemainingPellets();
        }
        return;
    }
}
*/

// PerformGameOver absorbs remaining pellets when only one player remains.
func (g *Game) PerformGameOver() {
	if g.GameOverProcessed {
		return
	}
	activePlayers := g.ActivePlayers()
	if len(activePlayers) == 1 {
		activePlayers[0].Pellets += g.RemainingPellets()
	}
	if len(activePlayers) <= 1 {
		g.GameOverProcessed = true
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:913-919

private boolean canImproveRanking(Player player) {
    int remainingPellets = getRemainingPellets();
    return remainingPellets > 0 && gameManager.getPlayers().stream()
        .filter(p -> p != player && p.pellets >= player.pellets)
        .anyMatch(p -> player.pellets + remainingPellets >= p.pellets);
}
*/

func (g *Game) CanImproveRanking(player *Player) bool {
	remaining := g.RemainingPellets()
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

func (g *Game) ActivePlayers() []*Player {
	out := make([]*Player, 0, len(g.Players))
	for _, p := range g.Players {
		if !p.IsDeactivated() {
			out = append(out, p)
		}
	}
	return out
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:967-972

private int getRemainingPellets() {
    return grid.getCells().values().stream()
        .filter(cell -> cell.hasPellet() || cell.hasCherry())
        .mapToInt(cell -> cell.hasPellet() ? 1 : Config.CHERRY_SCORE)
        .sum();
}
*/

func (g *Game) RemainingPellets() int {
	sum := 0
	for _, cell := range g.Grid.Cells {
		switch {
		case cell.HasPellet:
			sum++
		case cell.HasCherry:
			sum += CHERRY_SCORE
		}
	}
	return sum
}

// EndGame, Ended, OnEnd are arena-lifecycle helpers without a direct Java
// counterpart — Java's Referee.onEnd plus end-of-game scoring is split across
// these three Go methods.

// EndGame flags the game as ended.
func (g *Game) EndGame() { g.EndedFlag = true }

// Ended reports whether the simulation has finished.
func (g *Game) Ended() bool { return g.EndedFlag }

// OnEnd sets the final score as the pellet count for each player.
func (g *Game) OnEnd() {
	for _, p := range g.Players {
		p.SetScore(p.Pellets)
	}
}

// RemoveAll removes every element of toRemove from list, in-place. Mirrors
// Java's List.removeAll(Collection) used in resolvePacmenMovement.
func RemoveAll[T comparable](list, toRemove []T) []T {
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
