package com.codingame.game;

import java.util.LinkedList;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;

public class Bird {
    int id;
    LinkedList<Coord> body;
    Player owner;
    boolean alive;
    Direction direction;
    public String message;

    public Bird(int id, Player owner) {
        this.id = id;
        this.owner = owner;
        this.alive = true;
        this.message = null;
        body = new LinkedList<>();
    }

    public Coord getHeadPos() {
        return body.get(0);
    }

    public Direction getFacing() {
        if (body.size() < 2) {
            return Direction.UNSET;
        }
        return Direction.fromCoord(
            new Coord(
                body.get(0).getX() - body.get(1).getX(),
                body.get(0).getY() - body.get(1).getY()
            )
        );
    }

    public boolean isAlive() {
        return alive;
    }

    public void setMessage(String message) {
        this.message = message;
        if (message != null && message.length() > 48) {
            this.message = message.substring(0, 46) + "...";
        }
    }

    public boolean hasMessage() {
        return message != null;
    }
}
