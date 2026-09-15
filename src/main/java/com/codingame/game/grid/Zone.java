package com.codingame.game.grid;

import java.util.ArrayList;
import java.util.List;

public class Zone {
    private List<Coord> coords;
    private List<Integer> neighbours;
    private List<Town> containedTowns;
    public final int id;

    public static final int FOR_SALE = -1;
    public static final int NEUTRAL = 2;

    public int instability = 0;
    public boolean inked;

    public Zone(int id, List<Coord> coords) {
        this.id = id;
        this.setCoords(coords);
        this.containedTowns = new ArrayList<>(1);
        this.inked = false;
    }

    public int getId() {
        return id;
    }

    public int getCost() {
        return getCoords().size() * 2;
    }

    public List<Coord> getCoords() {
        return coords;
    }

    public void setCoords(List<Coord> coords) {
        this.coords = coords;
    }

    public List<Integer> getNeighbours() {
        return neighbours;
    }

    public void setNeighbours(List<Integer> neighbours) {
        this.neighbours = neighbours;

    }

    public List<Town> getContainedTowns() {
        return containedTowns;
    }

    public void addTown(Town t) {
        containedTowns.add(t);
    }

}
