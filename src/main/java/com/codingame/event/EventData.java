package com.codingame.event;

public class EventData {
    public static final int MOVE = 0;
    public static final int FALL = 1;
    public static final int EAT = 2;
    public static final int DEATH = 3;
    public static final int BEHEAD = 4;

    public int type;
    public AnimationData animData;

    public int[] params;

    public EventData() {

    }

}
