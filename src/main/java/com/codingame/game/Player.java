package com.codingame.game;

import java.util.ArrayList;
import java.util.List;

import com.codingame.game.grid.Coord;
import com.codingame.gameengine.core.AbstractMultiplayerPlayer;
import com.google.common.base.Objects;

public class Player extends AbstractMultiplayerPlayer {

    int points;
    List<Bird> birds;
    List<Coord> marks;

    public Player() {

    }

    public void reset() {
        birds.stream().forEach(bird -> {
            bird.direction = null;
            bird.message = null;
        });
        marks.clear();
    }

    public void init() {
        birds = new ArrayList<>();
        marks = new ArrayList<>();
    }

    @Override
    public int getExpectedOutputLines() {
        return 1;
    }

    public void addScore(int points) {
        this.points += points;
        setScore(this.points);
    }

    public Bird getBirdById(int id) {
        for (Bird bird : birds) {
            if (bird.id == id) {
                return bird;
            }
        }
        return null;
    }

    public boolean addMark(Coord coord) {
        if (marks.size() < 4) {
            marks.add(coord);
            return true;
        }
        return false;
    }

}
