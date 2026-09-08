// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java
package engine

import (
	"fmt"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:45-72

public static final int MIN_GRID_HEIGHT = 14;
public static final int MAX_GRID_HEIGHT = 20;
public static final float ASPECT_RATIO = 1.5f;
public static final int MIN_TOWN_DISTANCE = 4;
public static final int AVERAGE_TILES_PER_ZONE_COEFF_TO_GRID_HEIGHT = 2;
public static final int AVERAGE_TILES_PER_TOWN = 50;
public static final float TRAIN_TO_TOWN_RATIO = 0.4f;
public static final float RIVER_SPLIT_PROBA = 0.055f;
public static final float RIVER_TO_LAND_MIN_RATIO = 0.07f;
public static final int MIN_MOUNTAINS = 2;
public static final float MOUNTAIN_TO_CELL_RATIO = 0.04f;
public static final int MIN_RIVER_LENGTH = 3;
public static final int BASE_RAIL_COST = 1;
public static final int PASSIVE_INCOME = 3;
public static final int GRASS_COST_MULTIPLIER = 1;
public static final int RIVER_COST_MULTIPLIER = 2;
public static final int MOUNTAIN_COST_MULTIPLIER = 3;
public static final int POI_COST_MULTIPLIER = 3;
public static final int STARTING_DOSH = 0;
public static final int BLOT_POINTS_PER_TURN = 1;
public static final int MAX_TURNS = 100;
public static final int INSTABILITY_THRESHOLD_BASE = 4;
public static final int INSTABILITY_THRESHOLD_INCREASE = 0;
*/

const (
	MIN_GRID_HEIGHT = 14
	MAX_GRID_HEIGHT = 20

	ASPECT_RATIO      = float32(1.5)
	MIN_TOWN_DISTANCE = 4

	AVERAGE_TILES_PER_ZONE_COEFF_TO_GRID_HEIGHT = 2
	AVERAGE_TILES_PER_TOWN                      = 50

	TRAIN_TO_TOWN_RATIO     = float32(0.4)
	RIVER_SPLIT_PROBA       = float32(0.055)
	RIVER_TO_LAND_MIN_RATIO = float32(0.07)
	MIN_MOUNTAINS           = 2
	MOUNTAIN_TO_CELL_RATIO  = float32(0.04)
	MIN_RIVER_LENGTH        = 3

	BASE_RAIL_COST           = 1
	PASSIVE_INCOME           = 3
	GRASS_COST_MULTIPLIER    = 1
	RIVER_COST_MULTIPLIER    = 2
	MOUNTAIN_COST_MULTIPLIER = 3
	POI_COST_MULTIPLIER      = 3
	STARTING_DOSH            = 0
	BLOT_POINTS_PER_TURN     = 1
	MAX_TURNS                = 100

	INSTABILITY_THRESHOLD_BASE     = 4
	INSTABILITY_THRESHOLD_INCREASE = 0
)

// DEFAULT_LEAGUE is the highest league, the full game. Leagues 1-2 are
// tutorials routed through TutorialManager and are not modelled yet.
const DEFAULT_LEAGUE = 5

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:74-106

public boolean canBuy = false;
public boolean buyOnlyAdjacentToTrack = false;
public boolean buyOnlyContainingTrack = true;
public boolean freezeTrackPieceValue = false;
public boolean enableSideQuest = false;
public boolean showSideQuest = false;
...
List<Player> players;
Random random;
Grid grid;
int turn;
int instabilityThreshold;
int leagueLevel;
boolean inTutorial;
*/

// Game owns all match state. The Java class is a Guice singleton with the
// same fields; nothing here is package-global, so repeated simulations in one
// process cannot leak state into each other.
type Game struct {
	Players []*Player
	Random  *sha1prng.Random
	Grid    *Grid
	Turn    int

	InstabilityThreshold int
	LeagueLevel          int
	InTutorial           bool

	// EnableSideQuest is hardcoded false upstream, which makes every POI and
	// side-quest code path inert. Kept as a field to match the source.
	EnableSideQuest bool
	ShowSideQuest   bool

	// Summary collects the lines Java sends to gameManager.addToGameSummary.
	Summary []string

	// ended mirrors gameManager.endGame() having been called.
	ended bool
}

func NewGame(random *sha1prng.Random, leagueLevel int) *Game {
	return &Game{
		Random:      random,
		LeagueLevel: leagueLevel,
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:108-140

public void init() {
    this.leagueLevel = gameManager.getLeagueLevel();
    ...
    players = gameManager.getPlayers();
    random = gameManager.getRandom();
    initPlayers();
    initGrid(random);
    inTutorial = tutorialManager.initTutorial();
    turn = 0;
    instabilityThreshold = INSTABILITY_THRESHOLD_BASE;
}

private void initPlayers() { players.forEach(Player::init); }

private void initGrid(Random random) {
    gridMaker.init(random, enableSideQuest);
    grid = gridMaker.make();
}
*/

func (g *Game) Init(players []*Player) {
	g.Players = players
	for _, p := range g.Players {
		p.Init()
	}
	// The real GridMaker (and with it seed parity) lands with the map
	// generation work; until then Init builds the fixed placeholder grid.
	g.Grid = MakePlaceholderGrid()

	g.Turn = 0
	g.InstabilityThreshold = INSTABILITY_THRESHOLD_BASE
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:142-145

public void resetGameTurnData() {
    this.players.forEach(Player::reset);
    animation.reset();
}
*/

func (g *Game) ResetGameTurnData() {
	for _, p := range g.Players {
		p.Reset()
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:147-165

public void performGameUpdate(int turn) {
    this.turn++;
    logExecutionTime();
    doIncome();
    computeAutobuilds();
    doActions();
    doInstabilityCheck();
    moveTrains();
    computeTileStates();
    checkSideQuest();
    if (isGameOver()) gameManager.endGame();
    computeEvents();
}
*/

// PerformGameUpdate runs one turn. Autobuild expansion, track placement,
// disruption, instability and scoring are staged in by later work; the turn
// counter, the income reset and the end check are live.
func (g *Game) PerformGameUpdate(_ int) {
	g.Turn++

	g.DoIncome()

	if g.IsGameOver() {
		g.EndGame()
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:442-447

private void doIncome() {
    for (Player player : players) {
        player.dosh = PASSIVE_INCOME;
        player.blotPoints = BLOT_POINTS_PER_TURN;
    }
}
*/

// DoIncome assigns, it does not accumulate — unspent paint is lost.
func (g *Game) DoIncome() {
	for _, player := range g.Players {
		player.Dosh = PASSIVE_INCOME
		player.BlotPoints = BLOT_POINTS_PER_TURN
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:649-665

public int getRailCost(Coord c) { return getRailCost(grid.get(c)); }

public static int getRailCost(Tile t) {
    int cost = Game.BASE_RAIL_COST;
    if (t.isMountain()) return cost * Game.MOUNTAIN_COST_MULTIPLIER;
    if (t.isWater()) return cost * Game.RIVER_COST_MULTIPLIER;
    if (t.getType() == Tile.TYPE_POI) return cost * Game.POI_COST_MULTIPLIER;
    return cost * Game.GRASS_COST_MULTIPLIER;
}
*/

func (g *Game) RailCostAt(c Coord) int { return RailCost(g.Grid.Get(c)) }

func RailCost(t *Tile) int {
	cost := BASE_RAIL_COST
	switch {
	case t.IsMountain():
		return cost * MOUNTAIN_COST_MULTIPLIER
	case t.IsWater():
		return cost * RIVER_COST_MULTIPLIER
	case t.GetType() == TYPE_POI:
		return cost * POI_COST_MULTIPLIER
	}
	return cost * GRASS_COST_MULTIPLIER
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:708-714

public boolean isGameOver() {
    if (inTutorial) return tutorialManager.objectiveComplete() || this.turn >= MAX_TURNS;
    boolean anySchedulePossible = isAnyConnectionStillPossible();
    return !anySchedulePossible || this.turn >= MAX_TURNS;
}
*/

// IsGameOver currently reports only the turn cap. The tutorial objective and
// the "no desired connection is reachable any more" early exit both need
// pathfinding and arrive with the end-condition work.
func (g *Game) IsGameOver() bool {
	return g.Turn >= MAX_TURNS
}

func (g *Game) EndGame() { g.ended = true }

func (g *Game) Ended() bool { return g.ended }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:728-750

public void onEnd() {
    ...
    for (Player p : players) {
        if (!p.isActive()) { p.setScore(-1); scoreTexts[...] = "-"; }
        else scoreTexts[...] = p.getScore() + " point" + (p.getScore() > 1 ? "s" : "");
    }
    ...
}
*/

// OnEnd overwrites a deactivated player's score with -1. The score texts and
// the end screen are viewer-only; the tutorial branch arrives with league
// support.
func (g *Game) OnEnd() {
	for _, p := range g.Players {
		if p.IsDeactivated() {
			p.SetScore(-1)
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:787-804

public boolean shouldSkipPlayerTurn(Player player) { return false; }

public static String getExpected(String command) {
    if (command.startsWith("AUTOPLACE")) return "AUTOPLACE x1 y1 x2 y2";
    else if (command.startsWith("PLACE_TRACK")) return "PLACE_TRACK x y";
    else if (command.startsWith("DISRUPT")) return "DISRUPT zoneId";
    else if (command.startsWith("MESSAGE")) return "MESSAGE text";
    else if (command.startsWith("WAIT")) return "WAIT";
    return "AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT";
}
*/

func (g *Game) ShouldSkipPlayerTurn(_ *Player) bool { return false }

// GetExpected names the syntax a malformed command was closest to. Unlike
// the patterns themselves, these prefix checks are case-sensitive.
func GetExpected(command string) string {
	switch {
	case strings.HasPrefix(command, "AUTOPLACE"):
		return "AUTOPLACE x1 y1 x2 y2"
	case strings.HasPrefix(command, "PLACE_TRACK"):
		return "PLACE_TRACK x y"
	case strings.HasPrefix(command, "DISRUPT"):
		return "DISRUPT zoneId"
	case strings.HasPrefix(command, "MESSAGE"):
		return "MESSAGE text"
	case strings.HasPrefix(command, "WAIT"):
		return "WAIT"
	}
	return "AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT"
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:449-453

private void reportPlayerError(Player player, String message) {
    gameManager.addToGameSummary(GameManager.formatErrorMessage(player.getNicknameToken() + " " + message));
}
*/

func (g *Game) ReportPlayerError(player *Player, message string) {
	g.AddToGameSummary(formatErrorMessage(playerNickname(player) + " " + message))
}

func (g *Game) AddToGameSummary(line string) {
	g.Summary = append(g.Summary, line)
}

// formatErrorMessage mirrors GameManager.formatErrorMessage; the markers are
// the CodinGame viewer's red-text delimiters.
func formatErrorMessage(message string) string {
	return fmt.Sprintf("¤RED¤%s§RED§", message)
}

// playerNickname stands in for AbstractPlayer.getNicknameToken(). The arena
// has no nickname source, so summaries identify a side by index.
func playerNickname(player *Player) string {
	return fmt.Sprintf("Player %d", player.GetIndex())
}
