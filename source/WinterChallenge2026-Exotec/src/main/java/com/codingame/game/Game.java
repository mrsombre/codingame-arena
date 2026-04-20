
package com.codingame.game;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.Iterator;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.Set;
import java.util.TreeMap;
import java.util.function.Supplier;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.codingame.event.Animation;
import com.codingame.event.EventData;
import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.GridMaker;
import com.codingame.game.grid.Tile;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.codingame.gameengine.module.endscreen.EndScreenModule;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class Game {

    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject Animation animation;
    @Inject EndScreenModule endScreenModule;

    List<Player> players;
    Random random;
    Grid grid;
    Map<Integer, Bird> birdByIdCache = new TreeMap<>();
    public int turn;

    public int[] losses = new int[] { 0, 0 };

    public void init() {
        players = gameManager.getPlayers();
        random = gameManager.getRandom();

        initGrid(random);
        initPlayers();

    }

    private void initPlayers() {
        int birdId = 0;
        List<List<Coord>> spawnLocations = findSpawnLocations();

        for (Player p : players) {
            p.init();

            for (List<Coord> spawn : spawnLocations) {
                Bird bird = new Bird(birdId++, p);
                p.birds.add(bird);
                for (Coord c : spawn) {
                    if (p.getIndex() == 1) {
                        c = grid.opposite(c);
                    }
                    bird.body.add(c);
                    if (bird.body.size() == 1) {
                        // Is the head enclosed?
                        Tile left = grid.get(c.add(-1, 0));
                        Tile right = grid.get(c.add(1, 0));
                        if (left.getType() == Tile.TYPE_WALL && right.getType() == Tile.TYPE_WALL) {
                            grid.get(c.add(-1, 0)).clear();
                            grid.get(grid.opposite(c.add(-1, 0))).clear();
                        }
                    }
                }
            }

        }
    }

    private int countNeighboursIn(Coord coord, Set<Coord> coords) {
        int total = 0;
        for (Coord x : coords) {
            if (x == coord) {
                continue;
            }
            if (coord.manhattanTo(x) == 1) {
                total++;
            }
        }
        return total;
    }

    private List<List<Coord>> findSpawnLocations() {
        List<Set<Coord>> islands = grid.detectSpawnIslands();
        return islands.stream().map(s -> s.stream().sorted().toList()).toList();
    }

    private void initGrid(Random random) {
        GridMaker gridMaker = new GridMaker(random, gameManager.getLeagueLevel());
        this.grid = gridMaker.make();
    }

    public void resetGameTurnData() {
        animation.reset();
        players.stream().forEach(Player::reset);
    }

    private void doMoves() {
        for (Player p : players) {
            for (Bird bird : p.birds) {
                if (!bird.alive) {
                    continue;
                }
                if (bird.direction == null || bird.direction == Direction.UNSET) {
                    bird.direction = bird.getFacing();
                }

                Coord newHead = bird.getHeadPos().add(bird.direction.coord);

                boolean willEatApple = grid.apples.contains(newHead);

                if (!willEatApple) {
                    bird.body.pollLast();
                }
                bird.body.addFirst(newHead);
                launchMoveEvent(bird, willEatApple);
            }
        }
    }

    Supplier<Stream<Bird>> allBirds = () -> players.stream().flatMap(p -> p.birds.stream());

    List<Bird> getAllBirds() {
        return allBirds.get().toList();
    }

    List<Bird> getLiveBirds() {
        return allBirds.get().filter(Bird::isAlive).toList();
    }

    private void doBeheadings() {
        List<Bird> birdsToBehead = new ArrayList<>();

        for (Bird bird : getLiveBirds()) {
            boolean isInWall = grid.get(bird.getHeadPos()).getType() == Tile.TYPE_WALL;
            List<Bird> intersectingBirds = allBirds.get().filter(b -> b.alive && b.body.contains(bird.getHeadPos())).toList();

            boolean isInBird = intersectingBirds.stream().anyMatch(b ->
            // head intersects with another bird
            b.id != bird.id
                // head intersects with same bird on a pos that is not it's head
                || b.body.subList(1, b.body.size()).contains(b.getHeadPos())
            );

            if (isInWall || isInBird) {
                birdsToBehead.add(bird);
            }
        }

        birdsToBehead.forEach(b -> {
            if (b.body.size() <= 3) {
                launchDeathEvent(b);
                b.alive = false;
                losses[b.owner.getIndex()] += b.body.size();
            } else {
                // Behead
                b.body.pollFirst();
                launchBeheadEvent(b);
                losses[b.owner.getIndex()]++;
            }
        });
    }

    private void doEats() {
        Set<Coord> applesEatenThisTurn = new HashSet<>();
        for (Player p : players) {
            for (Bird bird : p.birds) {
                if (bird.alive && grid.apples.contains(bird.getHeadPos())) {
                    launchEatEvent(bird, bird.getHeadPos(), 1);
                    applesEatenThisTurn.add(bird.getHeadPos());
                }
            }
        }
        grid.apples.removeAll(applesEatenThisTurn);
    }

    private boolean hasTileOrAppleUnder(Coord c) {
        Coord below = c.add(0, 1);
        if (grid.get(below).getType() == Tile.TYPE_WALL) {
            return true;
        }

        for (Coord a : grid.apples) {
            if (a.equals(below)) {
                return true;
            }
        }
        return false;
    }

    private boolean somethingSolidUnder(Coord c, List<Coord> ignoreBody) {
        Coord below = c.add(0, 1);
        if (ignoreBody.contains(below)) {
            return false;
        }

        if (grid.get(below).getType() == Tile.TYPE_WALL) {
            return true;
        }
        for (Bird b : getLiveBirds()) {
            if (b.body.contains(below)) {
                return true;
            }
        }
        for (Coord a : grid.apples) {
            if (a.equals(below)) {
                return true;
            }
        }
        return false;
    }

    private boolean isGrounded(Coord c, Set<Bird> frozenBirds) {
        // Check if a coordinate touches the ground or a frozen bird
        Coord under = c.add(0, 1);
        return hasTileOrAppleUnder(c) ||
            frozenBirds.stream().anyMatch(b -> b.body.contains(under));
    }

    private void doFalls() {
        boolean somethingFell = true;
        Map<Integer, Integer> fallDistances = new TreeMap<>();

        List<Bird> outOfBounds = new ArrayList<>();
        Set<Bird> airborneBirds = new HashSet<>(getLiveBirds());
        Set<Bird> groundedBirds = new HashSet<>();

        while (somethingFell) {
            somethingFell = false;
            boolean somethingGotGrounded = true;

            while (somethingGotGrounded) {
                somethingGotGrounded = false;
                for (Bird bird : airborneBirds) {
                    boolean isGrounded = bird.body.stream().anyMatch(c -> isGrounded(c, groundedBirds));
                    if (isGrounded) {
                        groundedBirds.add(bird);
                        somethingGotGrounded = true;
                    }
                }
                airborneBirds.removeAll(groundedBirds);
            }

            for (Bird bird : airborneBirds) {

                somethingFell = true;
                bird.body = new LinkedList<>(bird.body.stream().map(c -> c.add(0, 1)).toList());
                fallDistances.compute(bird.id, (k, v) -> v == null ? 1 : v + 1);

                // check out of bounds
                if (bird.body.stream().allMatch(part -> part.getY() >= grid.height + 1)) {
                    bird.alive = false;
                    outOfBounds.add(bird);
                }

            }

            airborneBirds.removeAll(outOfBounds);
        }

        for (Integer birdId : fallDistances.keySet()) {
            Bird bird = getBirdById(birdId);
            int numberOfCells = fallDistances.get(birdId);
            if (numberOfCells > 0) {
                launchFallEvent(bird, numberOfCells);
            }
        }
        for (Bird bird : outOfBounds) {
            launchDeathEvent(bird);
        }
    }

    private List<List<Bird>> getIntercoiledBirds() {
        List<List<Bird>> intercoiledGroups = new ArrayList<>();
        List<Bird> allBirds = getLiveBirds();
        Set<Bird> visited = new HashSet<>();

        // Restrict to birds that are not on solid ground
        allBirds = allBirds.stream()
            .filter(b -> b.body.stream().noneMatch(c -> hasTileOrAppleUnder(c))).toList();

        for (Bird bird : allBirds) {
            if (visited.contains(bird)) {
                continue;
            }
            List<Bird> group = new ArrayList<>();
            LinkedList<Bird> toVisit = new LinkedList<>();
            toVisit.add(bird);
            while (!toVisit.isEmpty()) {
                Bird current = toVisit.poll();
                if (visited.contains(current)) {
                    continue;
                }

                visited.add(current);
                group.add(current);
                for (Bird other : allBirds) {
                    if (current == other) {
                        continue;
                    }
                    if (visited.contains(other)) {
                        continue;
                    }

                    // Are they adjacent?
                    boolean adj = birdsAreTouchingVertically(current, other);
                    if (adj) {
                        toVisit.add(other);
                    }
                }
            }
            if (group.size() > 1) {
                intercoiledGroups.add(group);
            }
        }
        return intercoiledGroups;
    }

    private boolean doIntercoiledFalls(Map<Integer, Integer> fallDistances, List<Bird> outOfBounds) {

        boolean somethingFellAtSomePoint = false;

        boolean somethingFell = true;
        while (somethingFell) {
            somethingFell = false;
            List<List<Bird>> intercoiledBirds = getIntercoiledBirds();

            for (List<Bird> birds : intercoiledBirds) {
                List<Coord> metaBody = birds.stream().flatMap(b -> b.body.stream()).toList();
                boolean canFall = metaBody.stream().noneMatch(c -> somethingSolidUnder(c, metaBody));
                if (canFall) {
                    somethingFell = true;
                    somethingFellAtSomePoint = true;
                    birds.forEach(bird -> {
                        bird.body = new LinkedList<>(bird.body.stream().map(c -> c.add(0, 1)).toList());
                        fallDistances.compute(bird.id, (k, v) -> v == null ? 1 : v + 1);
                        // check out of bounds
                        if (bird.body.stream().allMatch(part -> part.getY() >= grid.height + 1)) {
                            bird.alive = false;
                            outOfBounds.add(bird);
                        }
                    });

                }
            }
        }
        return somethingFellAtSomePoint;

    }

    private boolean birdsAreTouchingVertically(Bird bird, Bird other) {
        for (Coord c1 : bird.body) {
            for (Coord c2 : other.body) {
                if (c1.manhattanTo(c2) == 1 && c1.getX() == c2.getX()) {
                    return true;
                }
            }
        }
        return false;
    }

    private Bird getBirdById(Integer birdId) {
        if (birdByIdCache.containsKey(birdId)) {
            return birdByIdCache.get(birdId);
        }
        return allBirds.get().filter(b -> b.id == birdId).findFirst().orElseThrow();

    }

    public void performGameUpdate(int turn) {
        this.turn = turn;

        int currentTime = animation.getFrameTime();
        doMoves();
        animation.setFrameTime(currentTime);
        doEats();
        animation.catchUp();
        doBeheadings();
        animation.catchUp();
        doFalls();
        animation.catchUp();

        computeEvents();

        if (isGameOver()) {
            gameManager.endGame();
        }
    }

    public boolean isGameOver() {
        boolean noApples = grid.apples.isEmpty();
        boolean playerDed = players.stream().anyMatch(p -> p.birds.stream().noneMatch(Bird::isAlive));
        return noApples || playerDed;
    }

    public void onEnd() {
        String[] scoreTexts = new String[players.size()];
        for (Player p : players) {
            if (!p.isActive()) {
                p.setScore(-1);
                scoreTexts[p.getIndex()] = "-";
            } else {
                p.setScore(
                    p.birds.stream()
                        .filter(Bird::isAlive)
                        .mapToInt(b -> b.body.size())
                        .sum()
                );
            }
        }

        if (players.get(0).getScore() == players.get(1).getScore() && players.get(0).getScore() != -1) {
            // tie breaker: least losses
            players.forEach(p -> {
                scoreTexts[p.getIndex()] = String.format(
                    "%s point%s (%s %d loss%s)",
                    p.getScore(),
                    p.getScore() != 1 ? "s" : "",
                    losses[0] == losses[1] ? "and" : "but",
                    losses[p.getIndex()],
                    losses[p.getIndex()] != 1 ? "es" : ""
                );
            });
            for (Player p : players) {
                p.setScore(p.getScore() - losses[p.getIndex()]);
            }
        } else {
            players.forEach(p -> {
                scoreTexts[p.getIndex()] = p.getScore() > -1 ? String.format(
                    "%s point%s",
                    p.getScore(),
                    p.getScore() != 1 ? "s" : ""
                ) : "-";
            });
        }

        int[] scores = players.stream().mapToInt(Player::getScore).toArray();
        endScreenModule.setScores(scores, scoreTexts);
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

    private void launchMoveEvent(Bird bird, boolean willEatApple) {
        EventData e = new EventData();
        e.type = EventData.MOVE;
        e.params = new int[5];
        int idx = 0;
        e.params[idx++] = bird.id;
        e.params[idx++] = bird.owner.getIndex();
        e.params[idx++] = bird.getHeadPos().getX();
        e.params[idx++] = bird.getHeadPos().getY();
        e.params[idx++] = willEatApple ? 1 : 0;

        animation.startAnim(e, Animation.HALF);
    }

    private void launchBeheadEvent(Bird bird) {
        EventData e = new EventData();
        e.type = EventData.BEHEAD;
        e.params = new int[2];
        int idx = 0;
        e.params[idx++] = bird.id;
        e.params[idx++] = bird.owner.getIndex();

        animation.startAnim(e, Animation.HALF);
    }

    private void launchEatEvent(Bird bird, Coord coord, int growth) {
        EventData e = new EventData();
        e.type = EventData.EAT;
        e.params = new int[5];
        int idx = 0;
        e.params[idx++] = bird.id;
        e.params[idx++] = bird.owner.getIndex();
        e.params[idx++] = growth;
        e.params[idx++] = coord.getX();
        e.params[idx++] = coord.getY();

        animation.startAnim(e, Animation.HALF);
    }

    private void launchFallEvent(Bird bird, int numberOfCells) {
        EventData e = new EventData();
        e.type = EventData.FALL;
        e.params = new int[3];
        int idx = 0;
        e.params[idx++] = bird.id;
        e.params[idx++] = bird.owner.getIndex();
        e.params[idx++] = numberOfCells;

        animation.startAnim(e, 200);
    }

    private void launchDeathEvent(Bird bird) {
        EventData e = new EventData();
        e.type = EventData.DEATH;
        e.params = new int[2];
        e.params[0] = bird.id;
        e.params[1] = bird.owner.getIndex();
        animation.startAnim(e, Animation.HALF);
    }

    public List<EventData> getViewerEvents() {
        return animation.getViewerEvents();
    }

    public static String getExpected(String command) {
        return "MESSAGE text";
    }

}
