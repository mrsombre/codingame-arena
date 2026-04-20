package com.codingame.game.action;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;

public class Action {
    final ActionType type;
    private Direction direction;
    private Integer birdId;
    private Coord coord;

    private String message;

    public Action(ActionType type) {
        this.type = type;
    }

    public ActionType getType() {
        return type;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    @Override
    public String toString() {
        return "Action [type=" + type + ", direction=" + direction + ", birdId=" + birdId + "]";
    }

    public boolean isMove() {
        return direction != null;
    }
    
    public boolean isMark() {
        return coord != null;
    }

    public Integer getBirdId() {
        return birdId;
    }

    public void setBirdId(Integer birdId) {
        this.birdId = birdId;
    }

    public Direction getDirection() {
        return direction;
    }

    public void setDirection(Direction direction) {
        this.direction = direction;
    }

    public void setCoord(Coord coord) {
        this.coord = coord;
    }
    
    public Coord getCoord() {
        return coord;
    }
}