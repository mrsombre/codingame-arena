package com.codingame.game.grid;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.Queue;
import java.util.Set;

import com.google.common.collect.Table.Cell;

public class Grid {
    public static final Coord[] ADJACENCY = new Coord[] { Direction.NORTH.coord, Direction.EAST.coord, Direction.SOUTH.coord, Direction.WEST.coord };

    public static final Coord[] ADJACENCY_8 = new Coord[] {
        Direction.NORTH.coord,
        Direction.EAST.coord,
        Direction.SOUTH.coord,
        Direction.WEST.coord,
        new Coord(-1, -1),
        new Coord(1, 1),
        new Coord(1, -1),
        new Coord(-1, 1)
    };

    public int width, height;
    public LinkedHashMap<Coord, Tile> cells;
    boolean ySymetry;
    public List<Coord> spawns;
    public List<Coord> apples;

    public Grid(int width, int height) {
        this(width, height, false);
    }

    public Grid(int width, int height, boolean ySymetry) {
        this.width = width;
        this.height = height;
        this.ySymetry = ySymetry;
        spawns = new ArrayList<>();
        apples = new ArrayList<>();

        cells = new LinkedHashMap<>();

        for (int y = 0; y < height; ++y) {
            for (int x = 0; x < width; ++x) {
                Coord coord = new Coord(x, y);
                Tile cell = new Tile(coord);
                cells.put(coord, cell);
            }
        }
    }

    public Tile get(int x, int y) {
        return cells.getOrDefault(new Coord(x, y), Tile.NO_TILE);
    }

    public List<Coord> getNeighbours(Coord pos, Coord[] adjacency) {
        List<Coord> neighs = new ArrayList<>();
        for (Coord delta : adjacency) {
            Coord n = new Coord(pos.getX() + delta.getX(), pos.getY() + delta.getY());
            if (get(n) != Tile.NO_TILE) {
                neighs.add(n);
            }
        }
        return neighs;
    }

    public List<Coord> getNeighbours(Coord pos) {
        return getNeighbours(pos, ADJACENCY);
    }

    public Tile get(Coord n) {
        return get(n.getX(), n.getY());
    }

    public <T extends Coord> List<T> getClosestTargets(Coord from, List<T> targets) {
        List<T> closest = new ArrayList<>();
        int closestBy = 0;
        for (T targ : targets) {
            Coord neigh = targ;
            int distance = from.manhattanTo(neigh);
            if (closest.isEmpty() || closestBy > distance) {
                closest.clear();
                closest.add(targ);
                closestBy = distance;
            } else if (!closest.isEmpty() && closestBy == distance) {
                closest.add(targ);
            }
        }
        return closest;
    }

    public List<Coord> getCoords() {
        return cells.keySet().stream().toList();
    }

    public Coord opposite(Coord c) {
        return new Coord(width - c.x - 1, ySymetry ? (height - c.y - 1) : c.y);
    }

    public boolean isYSymetric() {
        return ySymetry;
    }

    public List<Set<Coord>> detectIslands2() {
        List<Set<Coord>> islands = new ArrayList<>();
        Set<Coord> computed = new HashSet<>();
        Set<Coord> current = new HashSet<>();

        //        for (Coord p : cells.keySet()) {
        while (true) {
            Coord p = new Coord(24, 23);
            Tile curCell = get(p);
            if (curCell.getType() == Tile.TYPE_WALL) {
                computed.add(p);
                continue;
            }
            if (!computed.contains(p)) {
                Queue<Coord> fifo = new LinkedList<>();
                fifo.add(p);
                computed.add(p);

                while (!fifo.isEmpty()) {
                    Coord e = fifo.poll();
                    for (Coord delta : ADJACENCY) {
                        Coord n = e.add(delta);
                        Tile cell = get(n);
                        if (cell.isValid() && !computed.contains(n) && cell.getType() != Tile.TYPE_WALL) {
                            fifo.add(n);
                            computed.add(n);
                        }
                    }
                    current.add(e);
                }
                islands.add(new HashSet<>(current));
                current.clear();
            }
            //        }

            return islands;
        }
    }

    public List<Set<Coord>> detectAirPockets() {
        List<Set<Coord>> islands = new ArrayList<>();
        Set<Coord> computed = new HashSet<>();
        Set<Coord> current = new HashSet<>();

        for (Coord p : cells.keySet()) {
            Tile curCell = get(p);
            if (curCell.getType() == Tile.TYPE_WALL) {
                computed.add(p);
                continue;
            }
            if (!computed.contains(p)) {
                Queue<Coord> fifo = new LinkedList<>();
                fifo.add(p);
                computed.add(p);

                while (!fifo.isEmpty()) {
                    Coord e = fifo.poll();
                    for (Coord delta : ADJACENCY) {
                        Coord n = e.add(delta);
                        Tile cell = get(n);
                        if (cell.isValid() && !computed.contains(n) && cell.getType() != Tile.TYPE_WALL) {
                            fifo.add(n);
                            computed.add(n);
                        }
                    }
                    current.add(e);
                }
                islands.add(new HashSet<>(current));
                current.clear();
            }
        }

        return islands;
    }

    public List<Set<Coord>> detectSpawnIslands() {
        List<Set<Coord>> islands = new ArrayList<>();
        Set<Coord> computed = new HashSet<>();
        Set<Coord> current = new HashSet<>();

        for (Coord p : spawns) {
            if (!computed.contains(p)) {
                Queue<Coord> fifo = new LinkedList<>();
                fifo.add(p);
                computed.add(p);

                while (!fifo.isEmpty()) {
                    Coord e = fifo.poll();
                    for (Coord delta : ADJACENCY) {
                        Coord n = e.add(delta);
                        Tile cell = get(n);
                        if (cell.isValid() && !computed.contains(n) && spawns.contains(n)) {
                            fifo.add(n);
                            computed.add(n);
                        }
                    }
                    current.add(e);
                }
                islands.add(new HashSet<>(current));
                current.clear();
            }
        }

        return islands;
    }

    public List<Coord> detectLowestIsland() {
        // flood fill walls from bottom coorner
        Coord start = new Coord(0, height - 1);
        if (get(start).getType() != Tile.TYPE_WALL) {
            return Collections.emptyList();
        }
        Set<Coord> computed = new HashSet<>();
        Queue<Coord> fifo = new LinkedList<>();
        List<Coord> lowest = new ArrayList<>();
        fifo.add(start);
        computed.add(start);
        lowest.add(start);
        while (!fifo.isEmpty()) {
            Coord e = fifo.poll();
            for (Coord delta : ADJACENCY) {
                Coord n = e.add(delta);
                Tile cell = get(n);
                if (cell.isValid() && !computed.contains(n) && cell.getType() == Tile.TYPE_WALL) {
                    fifo.add(n);
                    computed.add(n);
                    lowest.add(n);
                }
            }
        }
        return lowest;
        
    }

}
