package com.codingame.game.grid;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class Tile {
    public static final Tile NO_TILE = new Tile(new Coord(-1, -1), -1);

    public static final int TYPE_GRASS = 0;
    public static final int TYPE_WATER = 1;
    public static final int TYPE_MOUNTAIN = 2;
    public static final int TYPE_POI = 3;

    public static final int TRACK_NONE = -1;
    public static final int TRACK_NEUTRAL = 2;
    public static final int TOWN_NONE = -1;

    private int type;
    public int zoneId = -1;
    public int track = TRACK_NONE;
    public int townId = TOWN_NONE;

    public Coord coord;
    
    public List<ScheduleStep> activeConnections;

    public Tile(Coord coord) {
        this.coord = coord;
        this.activeConnections = new ArrayList<>();
    }

    public Tile(Coord coord, int type) {
        this.coord = coord;
        this.setType(type);
    }

    public void setType(int type) {
        this.type = type;
    }

    public int getType() {
        return type;
    }

    public void clear() {
        type = TYPE_GRASS;
    }

    public boolean isValid() {
        return this != NO_TILE;
    }

    public String toString() {
        return this.coord.toString() + " " + this.getType();
    }

    public int getZoneId() {
        return zoneId;
    }

    public void setZoneId(int zoneId) {
        this.zoneId = zoneId;
    }

    public boolean isWater() {
        return type == TYPE_WATER;
    }

    public boolean isMountain() {
        return type == TYPE_MOUNTAIN;
    }

    public boolean canUseTrackOrTown(int playerIdx) {
        return isTown() || canUseTracks(playerIdx);
    }

    public boolean canUseTracks(int playerIdx) {
        return track == playerIdx || track == TRACK_NEUTRAL;
    }

    public boolean isTown() {
        return townId != TOWN_NONE;
    }

    public boolean isTrackOrTown() {
        return isTown() || isTrack();
    }

    public boolean isTrack() {
        return track != TRACK_NONE;
    }

    public boolean isPlains() {
        return type == TYPE_GRASS;
    }
}
