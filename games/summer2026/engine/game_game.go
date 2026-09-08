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

	// Per-player counters, indexed by player index. Java publishes them as
	// match metadata; here they feed the arena's MetricsProvider.
	PlacedTracks            [2]int
	TracksPlacedOnPlains    [2]int
	TracksPlacedOnRiver     [2]int
	TracksPlacedOnMountains [2]int
	AutobuildCalled         [2]int

	// Ink attribution counters. A region is credited to a player only if that
	// player's last successful disruption this turn named it, so a region that
	// tips over on somebody else's blot is credited to nobody.
	ZonesInked          [2]int
	OwnTracksInkedOut   [2]int
	EnemyTracksInkedOut [2]int

	// ExtraTilesInConnection sums, over every scoring event, how much longer
	// the connection's path was than the straight-line distance between the
	// two towns — a measure of how far a player's rails detour.
	ExtraTilesInConnection [2]int
	// TrackOwnershipPercentagePerActiveConnection sums the player's share of
	// each connection it scored on, and ...Total counts those events, so the
	// average is the ratio of the two. float32, as in Java.
	TrackOwnershipPercentagePerActiveConnection      [2]float32
	TrackOwnershipPercentagePerActiveConnectionTotal [2]int

	// SuccessfulBlotsThisTurn maps a player index to the region its last
	// honoured disruption named. Java (spelling its own field
	// `succesfulBlotsThisTurn`) builds it in init and never clears it, so an
	// entry outlives the turn that wrote it and can still credit an inking on
	// a later turn. Faithful to the source; only the counters above depend on
	// it, so the divergence cannot reach a score.
	SuccessfulBlotsThisTurn map[int]int

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
	g.SuccessfulBlotsThisTurn = map[int]int{}
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

// PerformGameUpdate runs one turn. Everything but the side quest — inert in
// this build — and the tutorial objective is live.
func (g *Game) PerformGameUpdate(_ int) {
	g.Turn++

	g.DoIncome()
	g.ComputeAutobuilds()
	g.DoActions()
	g.DoInstabilityCheck()

	g.MoveTrains()
	g.ComputeTileStates()

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
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:667-706

private void computeAutobuilds() {
    for (Player player : players) {
        boolean autoBuildUsed = false;
        List<Action> resolvedIntents = new ArrayList<>();
        for (Action intent : player.intents) {
            if (intent.isAutobuild()) {
                if (autoBuildUsed) {
                    reportPlayerError(player, "Only one autobuild action allowed per turn.");
                    continue;
                }
                List<Action> actions = resolveAutobuild(player, intent);
                resolvedIntents.addAll(actions);
                autoBuildUsed = true;
                autobuildCalled[player.getIndex()]++;
            } else {
                resolvedIntents.add(intent);
            }
        }
        player.intents = resolvedIntents;
    }
}

private List<Action> resolveAutobuild(Player player, Action intent) {
    Coord from = intent.getFrom();
    Coord to = intent.getTo();
    AutobuildAStar planner = new AutobuildAStar(grid, player, from, to);
    Optional<List<AutobuildState>> path = planner.search();
    if (!path.isPresent()) return List.of();
    return path.get().stream().map(s -> s.action).filter(v -> !Objects.isNull(v)).toList();
}
*/

// ComputeAutobuilds rewrites each player's intent list in place, replacing
// the turn's one AUTOPLACE with the placements it expands to. Later AUTOPLACEs
// are reported and dropped; the counter still only counts the honoured one.
//
// A plan that cannot be paid for in full is not trimmed here — DoActions runs
// the expanded placements against the budget and interrupts the rest.
func (g *Game) ComputeAutobuilds() {
	for _, player := range g.Players {
		autoBuildUsed := false
		resolvedIntents := make([]*Action, 0, len(player.Intents))
		for _, intent := range player.Intents {
			if !intent.IsAutobuild() {
				resolvedIntents = append(resolvedIntents, intent)
				continue
			}
			if autoBuildUsed {
				g.ReportPlayerError(player, "Only one autobuild action allowed per turn.")
				continue
			}
			resolvedIntents = append(resolvedIntents, g.resolveAutobuild(player, intent)...)
			autoBuildUsed = true
			g.AutobuildCalled[player.GetIndex()]++
		}
		player.Intents = resolvedIntents
	}
}

// resolveAutobuild keeps only the steps of the winning plan that build
// something; the steps that merely walk the cursor over existing track carry
// no action and drop out.
func (g *Game) resolveAutobuild(player *Player, intent *Action) []*Action {
	path, ok := NewAutobuildAStar(g.Grid, player, intent.From, intent.To).Search()
	if !ok {
		return nil
	}

	actions := make([]*Action, 0, len(path))
	for _, s := range path {
		if s.Action != nil {
			actions = append(actions, s.Action)
		}
	}
	return actions
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:455-541

private void doActions() {
    Map<Coord, List<Integer>> tracksPlaced = new TreeMap<>();
    List<Coord> placementOrder = new ArrayList<>();
    for (Player player : players) {
        boolean interruptAutobuild = false;
        for (Action action : player.intents) {
            if (interruptAutobuild && action.isGeneratedByAutobuild()) continue;
            if (action.isPlaceTrack()) {
                Coord coord = action.getCoord();
                Tile tile = grid.get(coord);
                if (!tile.isValid()) { reportPlayerError(player, "Not part of grid: " + coord + ""); continue; }
                if (tile.isTown()) { reportPlayerError(player, "Cannot place tracks on a town at " + coord); continue; }
                Zone zone = grid.zones.get(tile.zoneId);
                if (!canPlaceTrackInZone(zone, player)) { reportPlayerError(player, "Cannot build in region " + zone.getId()); continue; }
                if (!isFreeOfTracks(coord, player, tracksPlaced)) { reportPlayerError(player, "Cannot place tracks on existing tracks at " + coord); continue; }
                int railCost = getRailCost(tile);
                if (player.getDosh() < railCost) {
                    if (action.isGeneratedByAutobuild()) {
                        reportPlayerError(player, "Autobuild interrupted: not enough track points to build a track at " + coord + ".");
                        interruptAutobuild = true;
                    } else {
                        reportPlayerError(player, "Not enough track points to build a track at " + coord + ".");
                    }
                    continue;
                }
                tracksPlaced.computeIfAbsent(coord, k -> new ArrayList<>(2)).add(player.getIndex());
                placementOrder.remove(coord);
                placementOrder.add(coord);
                player.pay(railCost);
                placedTracks[player.getIndex()]++;
                if (tile.isMountain()) tracksPlacedOnMountains[player.getIndex()]++;
                else if (tile.isWater()) tracksPlacedOnRiver[player.getIndex()]++;
                else if (tile.isPlains()) tracksPlacedOnPlains[player.getIndex()]++;
            }
        }
    }
    for (Coord coord : placementOrder) {
        List<Integer> playerIdxs = tracksPlaced.get(coord);
        Tile tile = grid.get(coord);
        if (playerIdxs.size() == 1) tile.track = playerIdxs.get(0);
        else tile.track = Tile.TRACK_NEUTRAL;
    }
    ... disruptions, quoted above doDisruptions ...
}
*/

// DoActions resolves both players' intents for the turn. Ownership is only
// written after both players have been charged, so a cell claimed by both in
// one turn goes neutral and neither side gets a refund. Illegal placements
// are reported and skipped; only unparseable output disqualifies, and that
// already happened in CommandManager. Disruptions run last, in the same
// player order, over the board the placements have already settled.
func (g *Game) DoActions() {
	// Java keys a TreeMap by Coord; nothing iterates it, so a plain map is
	// equivalent. placementOrder is what ordering the resolution pass sees.
	tracksPlaced := map[Coord][]int{}
	var placementOrder []Coord

	for _, player := range g.Players {
		interruptAutobuild := false
		for _, action := range player.Intents {
			if interruptAutobuild && action.GeneratedByAutobuild {
				continue
			}
			if !action.IsPlaceTrack() {
				continue
			}

			coord := action.Coord
			tile := g.Grid.Get(coord)
			if !tile.IsValid() {
				g.ReportPlayerError(player, "Not part of grid: "+coord.String())
				continue
			}
			if tile.IsTown() {
				g.ReportPlayerError(player, "Cannot place tracks on a town at "+coord.String())
				continue
			}

			zone := g.Grid.Zones[tile.ZoneID]
			if !canPlaceTrackInZone(zone) {
				g.ReportPlayerError(player, fmt.Sprintf("Cannot build in region %d", zone.ID))
				continue
			}

			if !g.isFreeOfTracks(coord, player, tracksPlaced) {
				g.ReportPlayerError(player, "Cannot place tracks on existing tracks at "+coord.String())
				continue
			}

			railCost := RailCost(tile)
			if player.GetDosh() < railCost {
				if action.GeneratedByAutobuild {
					g.ReportPlayerError(player, "Autobuild interrupted: not enough track points to build a track at "+coord.String()+".")
					interruptAutobuild = true
				} else {
					g.ReportPlayerError(player, "Not enough track points to build a track at "+coord.String()+".")
				}
				continue
			}

			tracksPlaced[coord] = append(tracksPlaced[coord], player.GetIndex())
			placementOrder = moveToBack(placementOrder, coord)

			player.Pay(railCost)
			g.PlacedTracks[player.GetIndex()]++
			switch {
			case tile.IsMountain():
				g.TracksPlacedOnMountains[player.GetIndex()]++
			case tile.IsWater():
				g.TracksPlacedOnRiver[player.GetIndex()]++
			case tile.IsPlains():
				g.TracksPlacedOnPlains[player.GetIndex()]++
			}
		}
	}

	for _, coord := range placementOrder {
		tile := g.Grid.Get(coord)
		if playerIdxs := tracksPlaced[coord]; len(playerIdxs) == 1 {
			tile.Track = playerIdxs[0]
		} else {
			tile.Track = TRACK_NEUTRAL
		}
	}

	g.doDisruptions()
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:544-582

// Disruptions
for (Player player : players) {
    for (Action action : player.intents) {
        if (action.getType() == ActionType.DISRUPT_ALT) {
            // Alternate disrupt command, using coord instead of zone id
            Tile tile = grid.get(action.getCoord());
            if (!tile.isValid()) { reportPlayerError(player, "Not part of grid: " + action.getCoord() + ""); continue; }
            action.setZoneId(tile.zoneId);
            action.setType(ActionType.DISRUPT);
        }
        if (action.isDisrupt()) {
            if (player.blotPoints <= 0) { reportPlayerError(player, "Not enough disruption points."); continue; }
            if (action.getZoneId() < 0 || action.getZoneId() >= grid.zones.size()) { reportPlayerError(player, "Invalid region id: " + action.getZoneId()); continue; }
            Zone zone = grid.zones.get(action.getZoneId());
            if (zone.inked) { reportPlayerError(player, "Cannot disrupt region" + zone.getId() + ". Already inked out."); continue; }
            if (!zone.getContainedTowns().isEmpty()) { reportPlayerError(player, "Cannot disrupt region" + zone.getId() + ". It contains a town."); continue; }
            player.blotPoints--;
            zone.instability++;
            launchDisruptEvent(player, zone);
            succesfulBlotsThisTurn.put(player.getIndex(), zone.id);
        }
    }
}
*/

// doDisruptions raises instability by one per honoured blot. A player gets
// BLOT_POINTS_PER_TURN of them a turn, so naming the same region twice in one
// line only lands once. The DISRUPT_ALT coordinate form is rewritten in place
// into the region form before the shared checks run, which is also why an
// out-of-bounds coordinate is the one rejection reported before the budget is
// even consulted.
func (g *Game) doDisruptions() {
	for _, player := range g.Players {
		for _, action := range player.Intents {
			if action.Type == ACTION_DISRUPT_ALT {
				tile := g.Grid.Get(action.Coord)
				if !tile.IsValid() {
					g.ReportPlayerError(player, "Not part of grid: "+action.Coord.String())
					continue
				}
				action.SetZoneID(tile.ZoneID)
				action.SetType(ACTION_DISRUPT)
			}
			if !action.IsDisrupt() {
				continue
			}

			if player.BlotPoints <= 0 {
				g.ReportPlayerError(player, "Not enough disruption points.")
				continue
			}
			if action.GetZoneID() < 0 || action.GetZoneID() >= len(g.Grid.Zones) {
				g.ReportPlayerError(player, fmt.Sprintf("Invalid region id: %d", action.GetZoneID()))
				continue
			}

			zone := g.Grid.Zones[action.GetZoneID()]
			// Both messages are missing the space Java's concatenation never
			// added; kept verbatim so summaries match the server's.
			if zone.Inked {
				g.ReportPlayerError(player, fmt.Sprintf("Cannot disrupt region%d. Already inked out.", zone.ID))
				continue
			}
			if len(zone.GetContainedTowns()) > 0 {
				g.ReportPlayerError(player, fmt.Sprintf("Cannot disrupt region%d. It contains a town.", zone.ID))
				continue
			}

			player.BlotPoints--
			zone.Instability++
			g.SuccessfulBlotsThisTurn[player.GetIndex()] = zone.ID
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:236-286

private void doInstabilityCheck() {
    List<Zone> toInk = new ArrayList<>();
    for (Zone zone : grid.zones) {
        if (zone.inked) continue;
        if (!zone.getContainedTowns().isEmpty()) continue;
        if (zone.instability >= instabilityThreshold) toInk.add(zone);
    }

    for (Zone zone : toInk) {
        instabilityThreshold += INSTABILITY_THRESHOLD_INCREASE;
        zone.inked = true;
        List<Coord> coordsWithTrackPiece = zone.getCoords().stream().filter(c -> grid.get(c).isTrack()).toList();

        int rekt[] = new int[] { 0, 0, 0 };
        for (Coord coord : coordsWithTrackPiece) {
            int trackOwner = grid.get(coord).track;
            if (trackOwner > -1) rekt[trackOwner]++;
            grid.get(coord).track = Tile.TRACK_NONE;
        }
        launchInkEvent(zone, coordsWithTrackPiece);

        boolean player2GotRekt = rekt[1] > 0;
        if (succesfulBlotsThisTurn.getOrDefault(0, -1) == zone.id && player2GotRekt)
            tutorialManager.setPlayerOneHasInkedEnemyTrack();

        for (Player p : players) {
            if (succesfulBlotsThisTurn.getOrDefault(p.getIndex(), -1) == zone.id) {
                zonesInked[p.getIndex()]++;
                ownTracksInkedOut[p.getIndex()] += rekt[p.getIndex()];
                enemyTracksInkedOut[p.getIndex()] += rekt[1 - p.getIndex()];
            }
        }
    }
}
*/

// DoInstabilityCheck inks every region that has reached the threshold, wiping
// the tracks inside it. Inking is permanent: an inked region can never be
// built in nor disrupted again.
//
// A region holding a town is exempt — it cannot be disrupted in the first
// place, and the check skips it again in case its instability was raised
// before the town existed.
//
// INSTABILITY_THRESHOLD_INCREASE is 0 in this build, so the threshold this
// bumps per inking never actually moves.
func (g *Game) DoInstabilityCheck() {
	var toInk []*Zone
	for _, zone := range g.Grid.Zones {
		if zone.Inked || len(zone.GetContainedTowns()) > 0 {
			continue
		}
		if zone.Instability >= g.InstabilityThreshold {
			toInk = append(toInk, zone)
		}
	}

	for _, zone := range toInk {
		g.InstabilityThreshold += INSTABILITY_THRESHOLD_INCREASE
		zone.Inked = true

		// rekt is indexed by the track sentinel itself, so slot 2 collects the
		// contested tracks nobody is credited for.
		var rekt [3]int
		for _, coord := range zone.Coords {
			tile := g.Grid.Get(coord)
			if !tile.IsTrack() {
				continue
			}
			rekt[tile.Track]++
			tile.Track = TRACK_NONE
		}

		for _, p := range g.Players {
			idx := p.GetIndex()
			if blotted, ok := g.SuccessfulBlotsThisTurn[idx]; !ok || blotted != zone.ID {
				continue
			}
			g.ZonesInked[idx]++
			g.OwnTracksInkedOut[idx] += rekt[idx]
			g.EnemyTracksInkedOut[idx] += rekt[1-idx]
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:586-601

private boolean canPlaceTrackInZone(Zone zone, Player player) { return !zone.inked; }

private boolean isFreeOfTracks(Coord coord, Player player, Map<Coord, List<Integer>> tracksPlaced) {
    Tile t = grid.get(coord);
    return t.track == Tile.TRACK_NONE && !tracksPlaced.getOrDefault(coord, List.of()).contains(player.getIndex());
}
*/

// canPlaceTrackInZone ignores the player Java passes it.
func canPlaceTrackInZone(zone *Zone) bool { return !zone.Inked }

// isFreeOfTracks rejects a cell that already carried a track before the turn,
// and a cell this same player has already claimed this turn — but not one the
// opponent claimed, which is how a contested cell arises.
func (g *Game) isFreeOfTracks(coord Coord, player *Player, tracksPlaced map[Coord][]int) bool {
	if g.Grid.Get(coord).Track != TRACK_NONE {
		return false
	}
	for _, idx := range tracksPlaced[coord] {
		if idx == player.GetIndex() {
			return false
		}
	}
	return true
}

// moveToBack reproduces `list.remove(coord); list.add(coord)` — the coord ends
// up last, and a cell claimed by both players is resolved at the position of
// the second claim.
func moveToBack(order []Coord, coord Coord) []Coord {
	for i, c := range order {
		if c == coord {
			order = append(order[:i], order[i+1:]...)
			break
		}
	}
	return append(order, coord)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:301-341

private void moveTrains() {
    // We're not moving anything, were detecting connected cities and scoring them.
    for (Town town : grid.towns) {
        List<Town> previousConnections = new ArrayList<>(town.activeConnections);
        Map<Integer, List<Coord>> prevTownPaths = town.paths;

        town.paths = new TreeMap<>();
        town.activeConnections.clear();
        for (Town other : town.desiredConnections) {
            List<Coord> path = new TrainBFS(grid, town, other).search();

            if (!path.isEmpty()) {
                town.activeConnections.add(other);
                town.paths.put(other.id, path);

                int[] pointsPerPlayer = new int[] { 0, 0 };
                for (Coord coord : path) {
                    Tile tile = grid.get(coord);
                    for (Player p : players)
                        if (tile.track == p.getIndex()) pointsPerPlayer[p.getIndex()]++;
                }
                for (Player p : players) {
                    if (pointsPerPlayer[p.getIndex()] == 0) continue;
                    int points = pointsPerPlayer[p.getIndex()];
                    p.addScore(points);
                    launchEarnPointsEvent(p, points, town, other);
                    extraTilesInConnection[p.getIndex()] += path.size() - town.coord.manhattanTo(other.coord);
                    trackOwnershipPercentagePerActiveConnection[p.getIndex()] += (float) points / (float) path.size();
                    trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()]++;
                }
                ... newConnections bookkeeping ...
            }
        }
        ... removals ...
    }
    ... animations ...
}
*/

// MoveTrains moves nothing. It recomputes, for every town, which of its
// desired counterparts are currently reachable over track, and pays out one
// point per track the player owns on each such path — every turn the
// connection stands, not once when it is made.
//
// Scoring walks the whole path, towns included, but a town tile carries
// TRACK_NONE and a contested tile carries TRACK_NEUTRAL, so neither matches a
// player index and neither scores for anybody.
//
// A connection is directional: town A wanting B and town B wanting A are two
// separate connections and both pay out.
//
// Java's newConnections set, the pathHasChanged comparison against the
// previous turn's paths, and the removal walk over previousConnections exist
// only to fire viewer animation events and are not ported.
func (g *Game) MoveTrains() {
	for _, town := range g.Grid.Towns {
		town.Paths = map[int][]Coord{}
		town.ActiveConnections = town.ActiveConnections[:0]

		for _, other := range town.DesiredConnections {
			path := NewTrainBFS(g.Grid, town, other).Search()
			if len(path) == 0 {
				continue
			}

			town.ActiveConnections = append(town.ActiveConnections, other)
			town.Paths[other.ID] = path

			var pointsPerPlayer [2]int
			for _, coord := range path {
				tile := g.Grid.Get(coord)
				for _, p := range g.Players {
					if tile.Track == p.GetIndex() {
						pointsPerPlayer[p.GetIndex()]++
					}
				}
			}

			for _, p := range g.Players {
				idx := p.GetIndex()
				points := pointsPerPlayer[idx]
				if points == 0 {
					continue
				}
				p.AddScore(points)
				g.ExtraTilesInConnection[idx] += len(path) - town.Coord.ManhattanTo(other.Coord)
				g.TrackOwnershipPercentagePerActiveConnection[idx] += float32(points) / float32(len(path))
				g.TrackOwnershipPercentagePerActiveConnectionTotal[idx]++
			}
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:219-234

private void computeTileStates() {
    grid.cells.values().forEach(tile -> { tile.activeConnections.clear(); });

    for (Town t : grid.towns) {
        t.paths.forEach(
            (Integer toTownId, List<Coord> path) -> {
                path.forEach(c -> grid.get(c).activeConnections.add(new ScheduleStep(t.id, toTownId)));
            }
        );
    }
}
*/

// ComputeTileStates stamps every cell with the connections whose path crosses
// it, which is what the per-turn serialization reports. Town.Paths is a
// TreeMap in Java, so the stamps land in ascending destination-town order;
// the serializer sorts them anyway, but the order is preserved regardless.
func (g *Game) ComputeTileStates() {
	for _, tile := range g.Grid.Cells {
		tile.ActiveConnections = tile.ActiveConnections[:0]
	}

	for _, t := range g.Grid.Towns {
		for _, toTownID := range t.SortedPathTownIDs() {
			for _, c := range t.Paths[toTownID] {
				tile := g.Grid.Get(c)
				tile.ActiveConnections = append(tile.ActiveConnections, ScheduleStep{
					FromTownID: t.ID,
					ToTownID:   toTownID,
				})
			}
		}
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
