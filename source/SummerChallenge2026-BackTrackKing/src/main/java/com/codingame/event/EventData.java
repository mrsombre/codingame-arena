package com.codingame.event;

import com.codingame.game.grid.Coord;

public class EventData {
    public static final int BUILD = 1;
    public static final int DISRUPT = 2;
    public static final int EARN_POINTS = 3;
    public static final int INK = 5;
    public static final int CONNECTION_LOST = 6;
    public static final int CONNECTION_GAINED = 7;
    public static final int BUMP = 8;
    public static final int POI_BUMP = 9;

    public int type;
    public AnimationData animData;

    public Coord coord, target;
    public int[] params;

    public EventData() {

    }

}
