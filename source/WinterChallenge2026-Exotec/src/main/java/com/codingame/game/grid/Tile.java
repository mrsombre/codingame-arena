package com.codingame.game.grid;

public class Tile {
    public static final Tile NO_TILE = new Tile(new Coord(-1, -1), -1);

    public static int TYPE_EMPTY = 0;
    public static int TYPE_WALL = 1;

    private int type;
    Coord coord;

    public Tile(Coord coord) {
        this.coord = coord;
    }

    public Tile(Coord coord, int type) {
        this.coord = coord;
        this.type = type;
    }

    public void setType(int type) {
        if (this == NO_TILE) {
            System.out.println("NOPE");
        }
        this.type = type;
    }

    public int getType() {
        return type;
    }

    public void clear() {
        type = TYPE_EMPTY;
    }

    public boolean isValid() {
        return this != NO_TILE;
    }

    public boolean isAccessible() {
        return true;
    }

}
