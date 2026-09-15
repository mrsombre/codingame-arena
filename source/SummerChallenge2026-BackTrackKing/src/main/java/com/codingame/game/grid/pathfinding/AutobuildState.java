package com.codingame.game.grid.pathfinding;

import java.util.Objects;
import java.util.Set;

import com.codingame.game.action.Action;
import com.codingame.game.grid.Coord;

public class AutobuildState {

    public enum CursorState {
        UNBOUGHT,
        BUILDABLE,
        BUILT,
        // For invalid coords
        BROKEN;
    }

    // For history
    public Action action;

    // For pathfinding
    Set<Integer> ownedZones;
    Set<Coord> tracks;
    int moneySpent;

    Coord cursor;
    CursorState cursorState;

    @Override
    public String toString() {
        return "State{" +
            cursorState +
            ", action='" + action + '\'' +
            ", ownedZones=" + ownedZones +
            ", tracks=" + tracks +
            ", moneySpent=" + moneySpent +
            ", cursor=" + cursor +
            '}';
    }

    @Override
    public int hashCode() {
        return Objects.hash(cursor, cursorState);
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (obj == null) return false;
        if (getClass() != obj.getClass()) return false;
        AutobuildState other = (AutobuildState) obj;
        return Objects.equals(cursor, other.cursor)
            && Objects.equals(cursorState, other.cursorState);
    }

}
