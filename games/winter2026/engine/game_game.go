// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java
package engine

import (
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:31-43

@Singleton
public class Game {
    List<Player> players;
    Random random;
    Grid grid;
    Map<Integer, Bird> birdByIdCache = new TreeMap<>();
    public int turn;
    public int[] losses = new int[] { 0, 0 };
*/

type Game struct {
	Players       []*Player
	Grid          *Grid
	BirdByIDCache map[int]*Bird
	Turn          int
	Losses        [2]int
	ended         bool
	summary       []string
	traces        []arena.TurnTrace
}

func NewGame(seed int64, leagueLevel int) *Game {
	g := &Game{
		BirdByIDCache: make(map[int]*Bird),
	}
	g.initGrid(seed, leagueLevel)
	return g
}

func (g *Game) Init(players []*Player) {
	g.Players = players
	g.InitPlayers()
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:54-82

private void initPlayers() {
    int birdId = 0;
    List<List<Coord>> spawnLocations = findSpawnLocations();
    for (Player p : players) {
        p.init();
        for (List<Coord> spawn : spawnLocations) {
            Bird bird = new Bird(birdId++, p);
            p.birds.add(bird);
            for (Coord c : spawn) {
                if (p.getIndex() == 1) c = grid.opposite(c);
                bird.body.add(c);
                // if head is enclosed by walls on both sides, clear one side
            }
        }
    }
}
*/

func (g *Game) InitPlayers() {
	birdID := 0
	spawnLocations := g.FindSpawnLocations()
	for _, p := range g.Players {
		p.Init()
		for _, spawn := range spawnLocations {
			bird := NewBird(birdID, p)
			birdID++
			p.Birds = append(p.Birds, bird)
			for _, c := range spawn {
				if p.GetIndex() == 1 {
					c = g.Grid.Opposite(c)
				}
				bird.Body = append(bird.Body, c)
				if len(bird.Body) == 1 {
					left := g.Grid.Get(c.AddXY(-1, 0))
					right := g.Grid.Get(c.AddXY(1, 0))
					if left.IsWall() && right.IsWall() {
						g.Grid.Get(c.AddXY(-1, 0)).Clear()
						g.Grid.Get(g.Grid.Opposite(c.AddXY(-1, 0))).Clear()
					}
				}
			}
			g.BirdByIDCache[bird.ID] = bird
		}
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:97-100

private List<List<Coord>> findSpawnLocations() {
    List<Set<Coord>> islands = grid.detectSpawnIslands();
    return islands.stream().map(s -> s.stream().sorted().toList()).toList();
}
*/

func (g *Game) FindSpawnLocations() [][]Coord {
	islands := g.Grid.DetectSpawnIslands()
	result := make([][]Coord, 0, len(islands))
	for _, island := range islands {
		result = append(result, SortedCoordsFromSet(island))
	}
	return result
}

func (g *Game) initGrid(seed int64, leagueLevel int) {
	gridMaker := NewGridMaker(sha1prng.New(seed), leagueLevel)
	g.Grid = gridMaker.Make()
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:107-110

public void resetGameTurnData() {
    animation.reset();
    players.stream().forEach(Player::reset);
}
*/

func (g *Game) ResetGameTurnData() {
	for _, p := range g.Players {
		p.Reset()
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:112-133

private void doMoves() {
    for (Player p : players) {
        for (Bird bird : p.birds) {
            if (!bird.alive) continue;
            if (bird.direction == null || bird.direction == Direction.UNSET)
                bird.direction = bird.getFacing();
            Coord newHead = bird.getHeadPos().add(bird.direction.coord);
            boolean willEatApple = grid.apples.contains(newHead);
            if (!willEatApple) bird.body.pollLast();
            bird.body.addFirst(newHead);
        }
    }
}
*/

func (g *Game) DoMoves() {
	for _, p := range g.Players {
		for _, bird := range p.Birds {
			if !bird.Alive {
				continue
			}
			if !bird.HasMove || bird.Direction == DirUnset {
				bird.Direction = bird.Facing()
			}

			newHead := bird.HeadPos().Add(bird.Direction.Coord())
			willEatApple := coordSliceContains(g.Grid.Apples, newHead)
			if !willEatApple {
				bird.Body = bird.Body[:len(bird.Body)-1]
			}
			bird.Body = append([]Coord{newHead}, bird.Body...)
		}
	}
}

func (g *Game) AllBirds() []*Bird {
	all := make([]*Bird, 0)
	for _, p := range g.Players {
		all = append(all, p.Birds...)
	}
	return all
}

func (g *Game) LiveBirds() []*Bird {
	live := make([]*Bird, 0)
	for _, b := range g.AllBirds() {
		if b.Alive {
			live = append(live, b)
		}
	}
	return live
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:145-176

private void doBeheadings() {
    for (Bird bird : getLiveBirds()) {
        boolean isInWall = grid.get(bird.getHeadPos()).getType() == Tile.TYPE_WALL;
        List<Bird> intersectingBirds = allBirds.get()
            .filter(b -> b.alive && b.body.contains(bird.getHeadPos())).toList();
        boolean isInBird = intersectingBirds.stream().anyMatch(b ->
            b.id != bird.id || b.body.subList(1, b.body.size()).contains(b.getHeadPos()));
        if (isInWall || isInBird) birdsToBehead.add(bird);
    }
    birdsToBehead.forEach(b -> {
        if (b.body.size() <= 3) { b.alive = false; losses[b.owner.getIndex()] += b.body.size(); }
        else { b.body.pollFirst(); losses[b.owner.getIndex()]++; }
    });
}
*/

func (g *Game) DoBeheadings() {
	birdsToBehead := make([]*Bird, 0)
	for _, bird := range g.LiveBirds() {
		isInWall := g.Grid.Get(bird.HeadPos()).IsWall()
		intersectingBirds := make([]*Bird, 0)
		for _, b := range g.AllBirds() {
			if b.Alive && coordSliceContains(b.Body, bird.HeadPos()) {
				intersectingBirds = append(intersectingBirds, b)
			}
		}

		isInEnemy, isInSelf := false, false
		for _, b := range intersectingBirds {
			if b.ID != bird.ID {
				isInEnemy = true
			} else if coordSliceContains(b.Body[1:], b.HeadPos()) {
				isInSelf = true
			}
		}

		head := bird.HeadPos()
		if isInWall {
			g.trace(arena.MakeTurnTrace(TraceHitWall, BirdCoordMeta{Bird: bird.ID, Coord: coordPair(head)}))
		}
		if isInEnemy {
			g.trace(arena.MakeTurnTrace(TraceHitEnemy, BirdCoordMeta{Bird: bird.ID, Coord: coordPair(head)}))
		}
		if isInSelf {
			g.trace(arena.MakeTurnTrace(TraceHitSelf, BirdCoordMeta{Bird: bird.ID, Coord: coordPair(head)}))
		}

		if isInWall || isInEnemy || isInSelf {
			birdsToBehead = append(birdsToBehead, bird)
		}
	}

	for _, b := range birdsToBehead {
		if len(b.Body) <= 3 {
			b.Alive = false
			g.Losses[b.Owner.GetIndex()] += len(b.Body)
			g.trace(arena.MakeTurnTrace(TraceDead, BirdMeta{Bird: b.ID}))
		} else {
			b.Body = b.Body[1:]
			g.Losses[b.Owner.GetIndex()]++
		}
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:178-189

private void doEats() {
    Set<Coord> applesEatenThisTurn = new HashSet<>();
    for (Player p : players) {
        for (Bird bird : p.birds) {
            if (bird.alive && grid.apples.contains(bird.getHeadPos())) {
                applesEatenThisTurn.add(bird.getHeadPos());
            }
        }
    }
    grid.apples.removeAll(applesEatenThisTurn);
}
*/

func (g *Game) DoEats() {
	eaten := make(map[Coord]struct{})
	for _, p := range g.Players {
		for _, bird := range p.Birds {
			if bird.Alive && coordSliceContains(g.Grid.Apples, bird.HeadPos()) {
				g.trace(arena.MakeTurnTrace(TraceEat, BirdCoordMeta{Bird: bird.ID, Coord: coordPair(bird.HeadPos())}))
				eaten[bird.HeadPos()] = struct{}{}
			}
		}
	}
	if len(eaten) == 0 {
		return
	}
	apples := g.Grid.Apples[:0]
	for _, apple := range g.Grid.Apples {
		if _, ok := eaten[apple]; !ok {
			apples = append(apples, apple)
		}
	}
	g.Grid.Apples = apples
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:191-203

private boolean hasTileOrAppleUnder(Coord c) {
    Coord below = c.add(0, 1);
    if (grid.get(below).getType() == Tile.TYPE_WALL) return true;
    for (Coord a : grid.apples) {
        if (a.equals(below)) return true;
    }
    return false;
}
*/

func (g *Game) HasTileOrAppleUnder(c Coord) bool {
	below := c.AddXY(0, 1)
	if g.Grid.Get(below).IsWall() {
		return true
	}
	return coordSliceContains(g.Grid.Apples, below)
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:227-232

private boolean isGrounded(Coord c, Set<Bird> frozenBirds) {
    Coord under = c.add(0, 1);
    return hasTileOrAppleUnder(c) ||
        frozenBirds.stream().anyMatch(b -> b.body.contains(under));
}
*/

func (g *Game) IsGrounded(c Coord, frozenBirds map[*Bird]struct{}) bool {
	under := c.AddXY(0, 1)
	if g.HasTileOrAppleUnder(c) {
		return true
	}
	for b := range frozenBirds {
		if coordSliceContains(b.Body, under) {
			return true
		}
	}
	return false
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:234-285

private void doFalls() {
    boolean somethingFell = true;
    Set<Bird> airborneBirds = new HashSet<>(getLiveBirds());
    Set<Bird> groundedBirds = new HashSet<>();
    while (somethingFell) {
        somethingFell = false;
        boolean somethingGotGrounded = true;
        while (somethingGotGrounded) {
            somethingGotGrounded = false;
            for (Bird bird : airborneBirds) {
                boolean isGrounded = bird.body.stream().anyMatch(c -> isGrounded(c, groundedBirds));
                if (isGrounded) { groundedBirds.add(bird); somethingGotGrounded = true; }
            }
            airborneBirds.removeAll(groundedBirds);
        }
        for (Bird bird : airborneBirds) {
            somethingFell = true;
            bird.body = new LinkedList<>(bird.body.stream().map(c -> c.add(0, 1)).toList());
            if (bird.body.stream().allMatch(part -> part.getY() >= grid.height + 1)) {
                bird.alive = false; outOfBounds.add(bird);
            }
        }
        airborneBirds.removeAll(outOfBounds);
    }
}
*/

func (g *Game) DoFalls() {
	somethingFell := true
	outOfBounds := make([]*Bird, 0)
	airborneBirds := make(map[*Bird]struct{})
	for _, b := range g.LiveBirds() {
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
					if g.IsGrounded(c, groundedBirds) {
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
				if part.Y < g.Grid.Height+1 {
					allOut = false
					break
				}
			}
			if allOut {
				bird.Alive = false
				outOfBounds = append(outOfBounds, bird)
				g.trace(arena.MakeTurnTrace(TraceFall, BirdSegmentsMeta{Bird: bird.ID, Segments: len(bird.Body)}))
			}
		}
		for _, bird := range outOfBounds {
			delete(airborneBirds, bird)
		}
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:384-403

public void performGameUpdate(int turn) {
    this.turn = turn;
    doMoves();
    doEats();
    doBeheadings();
    doFalls();
    if (isGameOver()) {
        gameManager.endGame();
    }
}
*/

func (g *Game) PerformGameUpdate(turn int) {
	g.Turn = turn
	g.traces = g.traces[:0]
	g.DoMoves()
	g.DoEats()
	g.DoBeheadings()
	g.DoFalls()
	if g.IsGameOver() {
		g.ended = true
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:405-409

public boolean isGameOver() {
    boolean noApples = grid.apples.isEmpty();
    boolean playerDed = players.stream().anyMatch(p -> p.birds.stream().noneMatch(Bird::isAlive));
    return noApples || playerDed;
}
*/

func (g *Game) IsGameOver() bool {
	noApples := len(g.Grid.Apples) == 0
	playerDead := false
	for _, p := range g.Players {
		hasLiveBird := false
		for _, b := range p.Birds {
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

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:411-452

public void onEnd() {
    for (Player p : players) {
        if (!p.isActive()) { p.setScore(-1); }
        else { p.setScore(p.birds.stream().filter(Bird::isAlive).mapToInt(b -> b.body.size()).sum()); }
    }
    if (players.get(0).getScore() == players.get(1).getScore() && players.get(0).getScore() != -1) {
        // tie-breaker: subtract losses from score
        players.forEach(p -> p.setScore(p.getScore() - losses[p.getIndex()]));
    }
}
*/

func (g *Game) OnEnd() {
	for _, p := range g.Players {
		if p.IsDeactivated() {
			p.SetScore(-1)
			continue
		}
		score := 0
		for _, b := range p.Birds {
			if b.Alive {
				score += len(b.Body)
			}
		}
		p.SetScore(score)
	}

	if len(g.Players) >= 2 && g.Players[0].GetScore() == g.Players[1].GetScore() && g.Players[0].GetScore() != -1 {
		for _, p := range g.Players {
			p.SetScore(p.GetScore() - g.Losses[p.GetIndex()])
		}
	}
}

func (g *Game) ShouldSkipPlayerTurn(player *Player) bool {
	return false
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Game.java:536-538

public static String getExpected(String command) {
    return "MESSAGE text";
}
*/

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

