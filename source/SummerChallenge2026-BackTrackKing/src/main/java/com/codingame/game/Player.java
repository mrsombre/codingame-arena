package com.codingame.game;

import java.util.ArrayList;
import java.util.List;

import com.codingame.game.action.Action;
import com.codingame.gameengine.core.AbstractMultiplayerPlayer;

public class Player extends AbstractMultiplayerPlayer {

    List<Action> intents;
    int dosh;
    public int blotPoints;
    private String message;

    public Player() {
    }

    public void init() {
        intents = new ArrayList<>();
        dosh = Game.STARTING_DOSH;
    }

    public void reset() {
        intents.clear();
        message = null;
    }

    public int getDosh() {
        return dosh;
    }

    @Override
    public int getExpectedOutputLines() {
        return 1;
    }

    public void addScore(int points) {
        setScore(getScore() + points);
    }

    public void pay(int cost) {
        dosh -= cost;

    }

    public void addDosh(int amount) {
        dosh += amount;

    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

}
