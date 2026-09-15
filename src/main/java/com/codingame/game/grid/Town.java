package com.codingame.game.grid;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

public class Town {
    public int id;
    public Coord coord;
    public List<Town> desiredConnections;
    public List<Town> activeConnections;
    public Map<Integer, List<Coord>> paths;

    public Town(int id, Coord coord) {
        this.id = id;
        this.coord = coord;
        this.activeConnections = new ArrayList<>();
        this.paths = new TreeMap<>();
    }

    public int getId() {
        return id;
    }
}
