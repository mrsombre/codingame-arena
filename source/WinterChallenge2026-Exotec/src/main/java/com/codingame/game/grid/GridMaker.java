
package com.codingame.game.grid;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedList;
import java.util.List;
import java.util.Random;
import java.util.Set;

import javax.imageio.ImageIO;

import com.google.inject.Singleton;

@Singleton
public class GridMaker {
    Random random;
    Grid grid;
    int leagueLevel;

    private static final int COLOR_WALL = 0xFF000000;
    private static final int COLOR_FOOD = 0xFFFF0000;
    private static final int COLOR_SPAWN = 0xFF0000FF;
    private static final int COLOR_DEBUG_SPAWN = 0xFF00FF00;

    public static final int MIN_GRID_HEIGHT = 10;
    public static final int MAX_GRID_HEIGHT = 24;

    public static final float ASPECT_RATIO = 1.8f;

    public GridMaker(Random random, int leagueLevel) {
        this.random = random;
        this.leagueLevel = leagueLevel;
    }

    public static int SPAWN_HEIGHT = 3;
    public static int DESIRED_SPAWNS = 4;

    public Grid makeFromImage() {
        BufferedImage img = loadImage("/levels/level_gravity_test.bmp");
        parsePixels(img);
        return grid;
    }

    public Grid make() {
        double skew;
        // Will need adjusting if wood leagues are introduced
        if (leagueLevel == 1) {
            skew = 2; // bronze
        } else if (leagueLevel == 2) {
            skew = 1; // silver
        } else if (leagueLevel == 3) {
            skew = 0.8; // gold
        } else {
            skew = 0.3; // legend
        }

        double rand = random.nextDouble();

        int height = MIN_GRID_HEIGHT +
            (int) Math.round(
                Math.pow(rand, skew) *
                    (MAX_GRID_HEIGHT - MIN_GRID_HEIGHT)
            );

        int width = Math.round(height * ASPECT_RATIO);
        if (width % 2 != 0) width += 1; // Make even
        grid = new Grid(width, height);

        double b = 5 + random.nextDouble() * 10;

        for (int x = 0; x < width; x++) {
            grid.get(x, height - 1).setType(Tile.TYPE_WALL);
        }

        for (int y = height - 2; y >= 0; y--) {
            // Row by row from bottom to top

            double yNorm = (double) (height - 1 - y) / (height - 1);
            double blockChanceEl = 1 / (yNorm + 0.1) / b;

            for (int x = 0; x < width; x++) {
                if (random.nextDouble() < blockChanceEl) {
                    grid.get(x, y).setType(Tile.TYPE_WALL);
                }
            }
        }

        for (int y = 0; y < height; ++y) {
            for (int x = 0; x < width; ++x) {
                Coord c = new Coord(x, y);
                Coord opp = grid.opposite(c);
                grid.get(opp).setType(grid.get(c).getType());
            }
        }

        // Fill islands
        List<Set<Coord>> islands = grid.detectAirPockets();
        for (Set<Coord> island : islands) {
            if (island.size() < 10) {
                for (Coord c : island) {
                    grid.get(c).setType(Tile.TYPE_WALL);
                }
            }
        }

        boolean somethingDestroyed = true;
        while (somethingDestroyed) {
            somethingDestroyed = false;
            for (Coord c : grid.cells.keySet()) {
                if (grid.get(c).getType() == Tile.TYPE_WALL) continue;
                List<Coord> neighbours = grid.getNeighbours(c);
                List<Coord> neighbourWalls = neighbours.stream().filter(n -> grid.get(n).getType() == Tile.TYPE_WALL).toList();
                if (neighbourWalls.size() >= 3) {
                    List<Coord> destroyable = new ArrayList<>(neighbourWalls.stream().filter(n -> n.y <= c.y).toList());
                    Collections.shuffle(destroyable, random);
                    grid.get(destroyable.get(0)).setType(Tile.TYPE_EMPTY);
                    grid.get(grid.opposite(destroyable.get(0))).setType(Tile.TYPE_EMPTY);
                    somethingDestroyed = true;
                }
            }
        }

        // Sink lowest island
        List<Coord> island = grid.detectLowestIsland();
        int lowerBy = 0;
        boolean canLower = true;
        while (canLower) {
            for (int x = 0; x < width; x++) {
                Coord c = new Coord(x, height - 1 - (lowerBy + 1));
                if (!island.contains(c)) {
                    canLower = false;
                    break;
                }
            }
            if (canLower) lowerBy++;
        }

        if (lowerBy >= 2) {
            lowerBy = random.nextInt(2, lowerBy + 1);
        }

        for (Coord c : island) {
            grid.get(c).setType(Tile.TYPE_EMPTY);
            grid.get(grid.opposite(c)).setType(Tile.TYPE_EMPTY);
        }
        for (Coord c : island) {
            Coord lowered = new Coord(c.x, c.y + lowerBy);
            if (grid.get(lowered).isValid()) {
                grid.get(lowered).setType(Tile.TYPE_WALL);
                grid.get(grid.opposite(lowered)).setType(Tile.TYPE_WALL);
            }
        }

        //Spawn a few apples
        for (int y = 0; y < height; ++y) {
            for (int x = 0; x < width / 2; ++x) {
                Coord c = new Coord(x, y);
                if (grid.get(c).getType() == Tile.TYPE_EMPTY && random.nextDouble() < 0.025) {
                    grid.apples.add(c);
                    grid.apples.add(grid.opposite(c));
                }
            }
        }

        // Previous technique could spawn no apples, so without changing the result of too many existing seeds, here is a better generation method:
        if (grid.apples.size() < 8) {
            grid.apples.clear();
            LinkedList<Coord> freeTiles = new LinkedList<>();
            freeTiles = new LinkedList<>(grid.cells.keySet().stream().filter(c -> grid.get(c).getType() == Tile.TYPE_EMPTY).toList());
            Collections.shuffle(freeTiles, random);

            //Spawn a few apples
            int minApplesCoords = Math.max(4, (int) (0.025 * freeTiles.size()));
            while (grid.apples.size() < minApplesCoords * 2 && !freeTiles.isEmpty()) {
                Coord c = freeTiles.poll();
                grid.apples.add(c);
                grid.apples.add(grid.opposite(c));
                freeTiles.remove(grid.opposite(c));
            }
        }

        // Convert lone walls to apples
        for (Coord c : grid.cells.keySet()) {
            if (grid.get(c).getType() == Tile.TYPE_EMPTY) {
                continue;
            }
            long neighbourWallCount = grid.getNeighbours(c, Grid.ADJACENCY_8).stream().filter(n -> grid.get(n).getType() == Tile.TYPE_WALL).count();
            if (neighbourWallCount == 0) {
                grid.get(c).setType(Tile.TYPE_EMPTY);
                grid.get(grid.opposite(c)).setType(Tile.TYPE_EMPTY);
                grid.apples.add(c);
                grid.apples.add(grid.opposite(c));
            }
        }

        List<Coord> potentialSpawns = new ArrayList<>(grid.cells.keySet().stream().filter(c -> {
            if (grid.get(c).getType() != Tile.TYPE_WALL)
                return false;

            List<Coord> aboves = getFreeAbove(c, SPAWN_HEIGHT);
            if (aboves.size() < SPAWN_HEIGHT)
                return false;

            return true;
        }).toList());

        Collections.shuffle(potentialSpawns, random);
        int desiredSpawns = DESIRED_SPAWNS;
        if (height <= 15) {
            desiredSpawns--;
        }
        if (height <= 10) {
            desiredSpawns--;
        }
        while (desiredSpawns > 0 && !potentialSpawns.isEmpty()) {
            Coord spawn = potentialSpawns.remove(0);

            List<Coord> spawnLoc = getFreeAbove(spawn, SPAWN_HEIGHT);
            // Check not neighbouring other spawn (or center line)
            boolean tooClose = false;
            for (Coord c : spawnLoc) {
                if (c.x == width / 2 - 1 || c.x == width / 2) {
                    tooClose = true;
                    break;
                }
                for (Coord n : grid.getNeighbours(c, Grid.ADJACENCY_8)) {
                    if (grid.spawns.contains(n) || grid.spawns.contains(grid.opposite(n))) {
                        tooClose = true;
                        break;
                    }
                }
                if (tooClose) break;
            }
            if (tooClose) continue;

            spawnLoc.forEach(ac -> {
                grid.spawns.add(ac);
                grid.apples.remove(ac);
                grid.apples.remove(grid.opposite(ac));

            });
            desiredSpawns--;
        }

        // Assert every apple is unique and not on a wall
        checkGrid();

        return grid;

    }

    private void checkGrid() {
        for (Coord c : grid.apples) {
            if (grid.get(c).getType() != Tile.TYPE_EMPTY) {
                throw new RuntimeException("Apple on wall at " + c);
            }
        }
        if (grid.apples.size() != grid.apples.stream().distinct().count()) {
            throw new RuntimeException("Duplicate apples");
        }
    }

    private List<Coord> getFreeAbove(Coord c, int by) {
        List<Coord> result = new ArrayList<>();
        for (int i = 1; i <= by; i++) {
            Coord above = new Coord(c.x, c.y - i);
            if (grid.get(above).isValid() && grid.get(above).getType() == Tile.TYPE_EMPTY) {
                result.add(above);
            } else {
                break;
            }
        }
        return result;
    }

    private static BufferedImage loadImage(String resourcePath) {
        try (InputStream is = GridMaker.class.getResourceAsStream(resourcePath)) {
            return ImageIO.read(is);
        } catch (IOException e) {
            e.printStackTrace();
            return null;
        }
    }

    private void parsePixels(BufferedImage img) {
        int width = img.getWidth();
        int height = img.getHeight();
        grid = new Grid(width, height);

        for (int y = 0; y < height; y++) {
            for (int x = 0; x < width; x++) {
                int rgb = img.getRGB(x, y);
                Coord coord = new Coord(x, y);

                if (rgb == COLOR_WALL) {
                    grid.get(coord).setType(Tile.TYPE_WALL);

                } else if (rgb == COLOR_FOOD) {
                    grid.apples.add(coord);

                } else if (rgb == COLOR_SPAWN) {
                    grid.spawns.add(coord);
                } else if (rgb == COLOR_DEBUG_SPAWN) {
                    grid.spawns.add(grid.opposite(coord));
                } else if (rgb != 0xFFFFFFFF) {
                    System.out.println("error " + Integer.toHexString(rgb) + " at " + x + "," + y);
                }
            }
        }
    }
}
