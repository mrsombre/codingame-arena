
package com.codingame.game;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import java.util.Random;
import java.util.Set;
import java.util.TreeMap;

import com.codingame.event.Animation;
import com.codingame.event.EventData;
import com.codingame.game.action.Action;
import com.codingame.game.action.ActionType;
import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.GridMaker;
import com.codingame.game.grid.ScheduleStep;
import com.codingame.game.grid.Tile;
import com.codingame.game.grid.Town;
import com.codingame.game.grid.Zone;
import com.codingame.game.grid.pathfinding.AutobuildAStar;
import com.codingame.game.grid.pathfinding.AutobuildState;
import com.codingame.game.grid.pathfinding.TerrainAStar;
import com.codingame.game.grid.pathfinding.TrainBFS;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.codingame.gameengine.module.endscreen.EndScreenModule;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class Game {

    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private GridMaker gridMaker;
    @Inject private Animation animation;
    @Inject private EndScreenModule endScreenModule;
    @Inject private TutorialManager tutorialManager;

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

    public static boolean CAN_BUY = false;
    public static boolean BUY_ONLY_ADJACENT_TO_TRACK = false;
    public static boolean BUY_ONLY_CONTAINING_TRACK = true;
    public static boolean FREEZE_TRACK_PIECE_VALUE = false;

    int[] placedTracks = new int[] { 0, 0 };
    int[] zonesInked = new int[] { 0, 0 };
    int[] extraTilesInConnection = new int[] { 0, 0 };
    int[] autobuildCalled = new int[] { 0, 0 };
    int[] tracksPlacedOnPlains = new int[] { 0, 0 };
    int[] tracksPlacedOnRiver = new int[] { 0, 0 };
    int[] tracksPlacedOnMountains = new int[] { 0, 0 };
    int[] ownTracksInkedOut = new int[] { 0, 0 };
    int[] enemyTracksInkedOut = new int[] { 0, 0 };
    float[] trackOwnershipPercentagePerActiveConnection = new float[] { 0, 0 };
    int[] trackOwnershipPercentagePerActiveConnectionTotal = new int[] { 0, 0 };

    long[] executionTimeMs = new long[] { 0, 0 };

    List<Player> players;
    Random random;
    Grid grid;
    int turn;

    int instabilityThreshold;
    int leagueLevel;
    boolean inTutorial;
    Map<Integer, Integer> succesfulBlotsThisTurn;
    
    Map<Integer, Set<Coord>> poisConnected;

    public void init() {
        this.leagueLevel = gameManager.getLeagueLevel();

        succesfulBlotsThisTurn = new TreeMap<>();
        players = gameManager.getPlayers();
        random = gameManager.getRandom();
        initPlayers();
        initGrid(random);

        inTutorial = tutorialManager.initTutorial();

        turn = 0;
        instabilityThreshold = INSTABILITY_THRESHOLD_BASE;
    }

    private void initPlayers() {
        players.forEach(Player::init);
        poisConnected = new LinkedHashMap<>();
        for (Player p : players) {
            poisConnected.put(p.getIndex(), new HashSet<>());
        }
    }

    private void initGrid(Random random) {
        gridMaker.init(random);
        grid = gridMaker.make();
    }

    public void resetGameTurnData() {
        this.players.forEach(Player::reset);
        animation.reset();
    }

    public void performGameUpdate(int turn) {
        this.turn++;

        logExecutionTime();

        doIncome();
        computeAutobuilds();
        doActions();

        doInstabilityCheck();

        moveTrains();
        computeTileStates();
        if (isGameOver()) {
            gameManager.endGame();
        }

        computeEvents();
    }

    private void logExecutionTime() {
        for (Player player : players) {
            executionTimeMs[player.getIndex()] += player.getLastExectionTimeMs();
        }

    }

    private boolean isConnectedToTownByPlayer(Coord coord, Player p) {
        LinkedList<Coord> fifo = new LinkedList<>();
        Set<Coord> visited = new HashSet<>();
        fifo.add(coord);
        visited.add(coord);
        
        // quick fix
        Tile poiT = grid.get(coord);
        boolean poiCoordOwnedByPoi = poiT.track == p.getIndex() || poiT.track == Tile.TRACK_NEUTRAL;
        if (!poiCoordOwnedByPoi) {
            return false;
        }
        
        while (!fifo.isEmpty()) {
            Coord current = fifo.poll();
            if (grid.get(current).isTown()) {
                return true;
            }
            List<Coord> neighbours = grid.getNeighbours(current);
            for (Coord n : neighbours) {
                Tile t = grid.get(n);
                boolean ownedByPlayer = t.track == p.getIndex() || t.track == Tile.TRACK_NEUTRAL;
                if (!visited.contains(n) && (ownedByPlayer || t.isTown())) {
                    fifo.add(n);
                    visited.add(n);
                }
            }

        }
        return false;

    }

    private void computeTileStates() {
        grid.cells.values().forEach(tile -> {
            tile.activeConnections.clear();
        });

        for (Town t : grid.towns) {
            t.paths.forEach(
                (Integer toTownId, List<Coord> path) -> {
                    path.forEach(
                        c -> grid.get(c).activeConnections.add(new ScheduleStep(t.id, toTownId))
                    );
                }
            );
        }

    }

    private void doInstabilityCheck() {
        List<Zone> toInk = new ArrayList<>();
        for (Zone zone : grid.zones) {
            if (zone.inked) {
                continue;
            }
            if (!zone.getContainedTowns().isEmpty()) {
                continue;
            }

            if (zone.instability >= instabilityThreshold) {
                toInk.add(zone);
            }
        }

        for (Zone zone : toInk) {
            instabilityThreshold += INSTABILITY_THRESHOLD_INCREASE;
            zone.inked = true;
            List<Coord> coordsWithTrackPiece = zone.getCoords().stream()
                .filter(c -> grid.get(c).isTrack())
                .toList();

            int rekt[] = new int[] { 0, 0, 0 };

            for (Coord coord : coordsWithTrackPiece) {
                int trackOwner = grid.get(coord).track;
                if (trackOwner > -1) {
                    rekt[trackOwner]++;
                }
                grid.get(coord).track = Tile.TRACK_NONE;
            }

            launchInkEvent(zone, coordsWithTrackPiece);

            // For tutorial, was this zone disrupted by player one?
            boolean player2GotRekt = rekt[1] > 0;
            if (succesfulBlotsThisTurn.getOrDefault(0, -1) == zone.id && player2GotRekt) {
                tutorialManager.setPlayerOneHasInkedEnemyTrack();
            }

            for (Player p : players) {
                if (succesfulBlotsThisTurn.getOrDefault(p.getIndex(), -1) == zone.id) {
                    zonesInked[p.getIndex()]++;
                    ownTracksInkedOut[p.getIndex()] += rekt[p.getIndex()];
                    enemyTracksInkedOut[p.getIndex()] += rekt[1 - p.getIndex()];

                }
            }
        }
        animation.catchUp();
    }

    private void launchInkEvent(Zone zone, List<Coord> coords) {
        EventData e = new EventData();
        e.type = EventData.INK;
        e.params = new int[1 + coords.size() * 2];
        e.params[0] = zone.id;
        for (int i = 0; i < coords.size(); i++) {
            e.params[1 + i * 2] = coords.get(i).getX();
            e.params[2 + i * 2] = coords.get(i).getY();
        }
        animation.startAnim(e, Animation.WHOLE);

    }

    private void moveTrains() {
        Set<ScheduleStep> newConnections = new LinkedHashSet<>();

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
                        // Score
                        for (Player p : players) {
                            if (tile.track == p.getIndex()) {
                                pointsPerPlayer[p.getIndex()]++;
                            }
                        }
                    }
                    for (Player p : players) {
                        if (pointsPerPlayer[p.getIndex()] == 0) {
                            continue;
                        }

                        int points = pointsPerPlayer[p.getIndex()];
                        p.addScore(points);
                        launchEarnPointsEvent(p, points, town, other);
                        extraTilesInConnection[p.getIndex()] += path.size() - town.coord.manhattanTo(other.coord);
                        trackOwnershipPercentagePerActiveConnection[p.getIndex()] += (float) points / (float) path.size();
                        trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()]++;

                    }

                    boolean pathHasChanged = false;

                    if (previousConnections.contains(other)) {
                        List<Coord> previousPath = prevTownPaths.get(other.id);
                        if (previousPath.size() != path.size()) {
                            pathHasChanged = true;
                        } else {
                            for (int i = 0; i < path.size(); i++) {
                                if (!previousPath.get(i).equals(path.get(i))) {
                                    pathHasChanged = true;
                                    break;
                                }
                            }
                        }
                    }

                    if (!previousConnections.contains(other) || pathHasChanged) {
                        if (pathHasChanged) {
                            launchConnectionLostEvent(town, other);
                        }
                        newConnections.add(new ScheduleStep(town.id, other.id));
                    }
                }
            }
            // Removals?
            for (Town other : previousConnections) {
                if (!town.activeConnections.contains(other)) {
                    launchConnectionLostEvent(town, other);
                }
            }
        }

        animation.catchUp();

        // Animations
        boolean skipBumpAnimation = newConnections.stream().mapToInt(
            ss -> {
                Town from = grid.towns.get(ss.fromTownId());
                List<Coord> path = from.paths.get(ss.toTownId());
                return path.size();
            }
        ).sum() > 100;

        for (ScheduleStep ss : newConnections) {
            Town from = grid.towns.get(ss.fromTownId());
            Town to = grid.towns.get(ss.toTownId());
            List<Coord> path = from.paths.get(ss.toTownId());
            launchConnectionGainedEvent(from, to, path, skipBumpAnimation);
            animation.catchUp();
        }
    }

    private void launchConnectionLostEvent(Town town, Town other) {
        EventData e = new EventData();
        e.type = EventData.CONNECTION_LOST;
        e.params = new int[] {
            town.id,
            other.id
        };
        animation.startAnim(e, Animation.THIRD);

    }

    private void launchConnectionGainedEvent(Town town, Town other, List<Coord> path, boolean skipBumpAnimation) {
        int t = animation.getFrameTime();
        EventData e = new EventData();
        e.type = EventData.CONNECTION_GAINED;
        e.params = new int[path.size() * 2 + 2];
        e.params[0] = town.id;
        e.params[1] = other.id;
        for (int i = 0; i < path.size(); i++) {
            e.params[2 + i * 2] = path.get(i).getX();
            e.params[3 + i * 2] = path.get(i).getY();
        }
        ;
        animation.startAnim(e, Animation.THIRD);

        if (skipBumpAnimation) {
            return;
        }

        for (int i = 0; i < path.size(); i++) {
            Coord step = path.get(i);
            EventData bump = new EventData();
            Tile tile = grid.get(step);
            if (tile.isTown()) {
                continue;
            }
            bump.type = EventData.BUMP;
            bump.params = new int[2];
            bump.params[0] = step.getX();
            bump.params[1] = step.getY();
            animation.startAnim(bump, Animation.THIRD);
            animation.wait(Animation.TENTH);
        }
        animation.setFrameTime(t);

    }

    private void doIncome() {
        for (Player player : players) {
            player.dosh = PASSIVE_INCOME;
            player.blotPoints = BLOT_POINTS_PER_TURN;
        }
    }

    private void reportPlayerError(Player player, String message) {
        gameManager.addToGameSummary(
            GameManager.formatErrorMessage(player.getNicknameToken() + " " + message)
        );
    }

    private void doActions() {
        // For each player's action list, we perform the actions in order. If player two's actions come into conflict
        // with player one's, we unresolve player one's and replace with a conflict action
        Map<Coord, List<Integer>> tracksPlaced = new TreeMap<>();
        List<Coord> placementOrder = new ArrayList<>();

        for (Player player : players) {

            boolean interruptAutobuild = false;
            for (Action action : player.intents) {
                if (interruptAutobuild && action.isGeneratedByAutobuild()) {
                    continue;
                }
                if (action.isPlaceTrack()) {
                    Coord coord = action.getCoord();
                    Tile tile = grid.get(coord);
                    if (!tile.isValid()) {
                        reportPlayerError(player, "Not part of grid: " + coord + "");
                        continue;
                    }
                    if (tile.isTown()) {
                        reportPlayerError(player, "Cannot place tracks on a town at " + coord);
                        continue;
                    }

                    Zone zone = grid.zones.get(tile.zoneId);
                    if (!canPlaceTrackInZone(zone, player)) {
                        reportPlayerError(player, "Cannot build in region " + zone.getId());
                        continue;
                    }

                    if (!isFreeOfTracks(coord, player, tracksPlaced)) {
                        reportPlayerError(player, "Cannot place tracks on existing tracks at " + coord);
                        continue;
                    }
                    int railCost = getRailCost(tile);
                    if (player.getDosh() < railCost) {
                        if (action.isGeneratedByAutobuild()) {
                            reportPlayerError(player, "Autobuild interrupted: not enough track points to build a track at " + coord + ".");
                            // Interrupt auto actions on lack of money
                            interruptAutobuild = true;
                        } else {
                            reportPlayerError(player, "Not enough track points to build a track at " + coord + ".");
                        }
                        continue;
                    }
                    // Record purchase
                    tracksPlaced.computeIfAbsent(coord, k -> new ArrayList<>(2)).add(player.getIndex());

                    placementOrder.remove(coord);
                    placementOrder.add(coord);

                    player.pay(railCost);
                    placedTracks[player.getIndex()]++;
                    if (tile.isMountain()) {
                        tracksPlacedOnMountains[player.getIndex()]++;
                    } else if (tile.isWater()) {
                        tracksPlacedOnRiver[player.getIndex()]++;
                    } else if (tile.isPlains()) {
                        tracksPlacedOnPlains[player.getIndex()]++;
                    }
                }
            }
        }

        int[] animTimes = new int[] { animation.getFrameTime(), animation.getFrameTime() };
        for (Coord coord : placementOrder) {

            List<Integer> playerIdxs = tracksPlaced.get(coord);
            Tile tile = grid.get(coord);
            if (playerIdxs.size() == 1) {
                int playerIndex = playerIdxs.get(0);
                tile.track = playerIndex;
                animation.setFrameTime(animTimes[playerIndex]);
                launchBuildEvent(players.get(playerIndex), coord);
                animation.wait(Animation.THIRD);
                animTimes[playerIndex] = animation.getFrameTime();
            } else {
                tile.track = Tile.TRACK_NEUTRAL;
                animation.setFrameTime(Math.max(animTimes[0], animTimes[1]));
                launchNeutralBuild(coord);
                animation.wait(Animation.THIRD);
                animTimes[0] = animation.getFrameTime();
                animTimes[1] = animation.getFrameTime();

            }
        }
        animation.catchUp();

        // Disruptions
        for (Player player : players) {
            for (Action action : player.intents) {
                if (action.getType() == ActionType.DISRUPT_ALT) {
                    // Alternate disrupt command, using coord instead of zone id
                    Tile tile = grid.get(action.getCoord());
                    if (!tile.isValid()) {
                        reportPlayerError(player, "Not part of grid: " + action.getCoord() + "");
                        continue;
                    }
                    action.setZoneId(tile.zoneId);
                    action.setType(ActionType.DISRUPT);
                }
                if (action.isDisrupt()) {
                    if (player.blotPoints <= 0) {
                        reportPlayerError(player, "Not enough disruption points.");
                        continue;
                    }
                    if (action.getZoneId() < 0 || action.getZoneId() >= grid.zones.size()) {
                        reportPlayerError(player, "Invalid region id: " + action.getZoneId());
                        continue;
                    }
                    Zone zone = grid.zones.get(action.getZoneId());
                    if (zone.inked) {
                        reportPlayerError(player, "Cannot disrupt region" + zone.getId() + ". Already inked out.");
                        continue;
                    }
                    if (!zone.getContainedTowns().isEmpty()) {
                        reportPlayerError(player, "Cannot disrupt region" + zone.getId() + ". It contains a town.");
                        continue;
                    }
                    player.blotPoints--;
                    zone.instability++;
                    launchDisruptEvent(player, zone);

                    succesfulBlotsThisTurn.put(player.getIndex(), zone.id);
                }
            }
        }
        animation.catchUp();
    }

    private boolean canPlaceTrackInZone(
        Zone zone, Player player
    ) {
        return !zone.inked;
    }

    private boolean isFreeOfTracks(
        Coord coord, Player player, Map<Coord, List<Integer>> tracksPlaced
    ) {
        Tile t = grid.get(coord);
        boolean tileCanHaveTrack = t.track == Tile.TRACK_NONE
            && !tracksPlaced.getOrDefault(coord, List.of()).contains(player.getIndex());

        return tileCanHaveTrack;

    }

    private void launchDisruptEvent(Player player, Zone zone) {
        EventData e = new EventData();
        e.type = EventData.DISRUPT;
        e.params = new int[] {
            player.getIndex(),
            zone.id,
        };
        animation.startAnim(e, Animation.HALF + Animation.TENTH);
    }

    private void launchBuildEvent(Player player, Coord coord) {
        EventData e = new EventData();
        e.type = EventData.BUILD;
        e.params = new int[] {
            player.getIndex(),
            coord.getX(),
            coord.getY(),
            getRailCost(coord)
        };
        animation.startAnim(e, Animation.HALF);
    }

    private void launchNeutralBuild(Coord coord) {
        EventData e = new EventData();
        e.type = EventData.BUILD;
        e.params = new int[] {
            Tile.TRACK_NEUTRAL,
            coord.getX(),
            coord.getY(),
            getRailCost(coord)
        };
        animation.startAnim(e, Animation.HALF);
    }

    private void launchEarnPointsEvent(Player player, int score, Town from, Town to) {
        EventData e = new EventData();
        e.type = EventData.EARN_POINTS;
        e.params = new int[] {
            player.getIndex(),
            score,
            from.id,
            to.id
        };
        animation.startAnim(e, Animation.HALF);
    }

    public int getRailCost(Coord c) {
        return getRailCost(grid.get(c));
    }

    public static int getRailCost(Tile t) {
        int cost = Game.BASE_RAIL_COST;
        if (t.isMountain()) {
            return cost * Game.MOUNTAIN_COST_MULTIPLIER;
        }
        if (t.isWater()) {
            return cost * Game.RIVER_COST_MULTIPLIER;
        }
        if (t.getType() == Tile.TYPE_POI) {
            return cost * Game.POI_COST_MULTIPLIER;
        }
        return cost * Game.GRASS_COST_MULTIPLIER;
    }

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

        if (!path.isPresent()) {
            return List.of();
        }

        return path.get().stream()
            .map(s -> s.action)
            .filter(v -> !Objects.isNull(v))
            .toList();

    }

    public boolean isGameOver() {
        if (inTutorial) {
            return tutorialManager.objectiveComplete() || this.turn >= MAX_TURNS;
        }
        boolean anySchedulePossible = isAnyConnectionStillPossible();
        return !anySchedulePossible || this.turn >= MAX_TURNS;
    }

    private boolean isAnyConnectionStillPossible() {
        for (Town town : grid.towns) {
            for (Town other : town.desiredConnections) {
                Optional<List<Coord>> path = new TerrainAStar(grid, town, other).search();
                if (path.isPresent()) {
                    return true;
                }
            }
        }
        return false;
    }

    public void onEnd() {
        String[] scoreTexts = new String[players.size()];
        if (inTutorial) {
            tutorialManager.handleEnd(scoreTexts);
        } else {
            for (Player p : players) {
                if (!p.isActive()) {
                    p.setScore(-1);
                    scoreTexts[p.getIndex()] = "-";
                } else {
                    scoreTexts[p.getIndex()] = p.getScore() + " point" + (p.getScore() > 1 ? "s" : "");
                }
            }
        }

        int[] scores = players.stream().mapToInt(Player::getScore).toArray();
        endScreenModule.setScores(scores, scoreTexts);

        writeMetadata();
    }

    private void writeMetadata() {
        for (Player p : players) {
            gameManager.putMetadata("tracksPlaced_" + p.getIndex(), placedTracks[p.getIndex()]);
            gameManager.putMetadata("tracksPlacedOnPlains_" + p.getIndex(), tracksPlacedOnPlains[p.getIndex()]);
            gameManager.putMetadata("tracksPlacedOnRiver_" + p.getIndex(), tracksPlacedOnRiver[p.getIndex()]);
            gameManager.putMetadata("tracksPlacedOnMountains_" + p.getIndex(), tracksPlacedOnMountains[p.getIndex()]);
            gameManager.putMetadata("zonesInked_" + p.getIndex(), zonesInked[p.getIndex()]);
            gameManager.putMetadata("ownTracksInkedOut_" + p.getIndex(), ownTracksInkedOut[p.getIndex()]);
            gameManager.putMetadata("enemyTracksInkedOut_" + p.getIndex(), enemyTracksInkedOut[p.getIndex()]);
            gameManager.putMetadata("extraTilesInConnection_" + p.getIndex(), extraTilesInConnection[p.getIndex()]);
            gameManager.putMetadata("autobuildCalled_" + p.getIndex(), autobuildCalled[p.getIndex()]);
            gameManager.putMetadata(
                "averageTrackOwnershipPercentagePerActiveConnection_" + p.getIndex(),
                trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()] == 0 ? 0f
                    : trackOwnershipPercentagePerActiveConnection[p.getIndex()]
                        / (float) trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()]
            );
            gameManager.putMetadata("executionTimeMs_" + p.getIndex(), executionTimeMs[p.getIndex()]);
            gameManager.putMetadata("poisServed_" + p.getIndex(), poisConnected.get(p.getIndex()).size());
        }
        gameManager.putMetadata("totalPois", grid.pois.size());
    }

    private void computeEvents() {
        int minTime = 500;

        animation.catchUp();

        int frameTime = Math.max(
            animation.getFrameTime(),
            minTime
        );
        gameManager.setFrameDuration(frameTime);
    }

    public boolean shouldSkipPlayerTurn(Player player) {
        return false;
    }

    public static String getExpected(String command) {
        if (command.startsWith("AUTOPLACE")) {
            return "AUTOPLACE x1 y1 x2 y2";
        } else if (command.startsWith("PLACE_TRACK")) {
            return "PLACE_TRACK x y";
        } else if (command.startsWith("DISRUPT")) {
            return "DISRUPT zoneId";
        } else if (command.startsWith("MESSAGE")) {
            return "MESSAGE text";
        } else if (command.startsWith("WAIT")) {
            return "WAIT";
        }
        return "AUTOPLACE | PLACE_TRACK | DISRUPT | MESSAGE | WAIT";
    }

    public List<EventData> getViewerEvents() {
        return animation.getViewerEvents();
    }

}
