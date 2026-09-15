package com.codingame.game.action;

import com.codingame.game.grid.Coord;

public class Action {
    private ActionType type;
    private Coord coord;
    private Integer zoneId;
    private Coord from, to;

    boolean generatedByAutobuild;

    private String message;

    public Action(ActionType type) {
        this(type, false);
    }

    public Action(ActionType type, boolean generatedByAutobuild) {
        this.type = type;
        this.generatedByAutobuild = generatedByAutobuild;
    }

    public ActionType getType() {
        return type;
    }

    public Coord getCoord() {
        return coord;
    }

    public void setCoord(Coord coord) {
        this.coord = coord;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    @Override
    public String toString() {
        return "Action [type=" + type + ", coord=" + coord + ", zone=" + zoneId + "]";
    }

    public boolean isMessage() {
        return type == ActionType.MESSAGE;
    }

    public boolean isPlaceTrack() {
        return type == ActionType.PLACE_TRACK;
    }

    public boolean isDisrupt() {
        return type == ActionType.DISRUPT;
    }

    public int getZoneId() {
        return zoneId;
    }

    public void setZoneId(int zoneId) {
        this.zoneId = zoneId;
    }

    public boolean isAutobuild() {
        return type == ActionType.AUTOPLACE;
    }

    public Coord getFrom() {
        return from;
    }

    public void setFrom(Coord from) {
        this.from = from;
    }

    public Coord getTo() {
        return to;
    }

    public void setTo(Coord to) {
        this.to = to;
    }

    public boolean isGeneratedByAutobuild() {
        return generatedByAutobuild;
    }

    public void setType(ActionType type) {
        this.type = type;
    }
}