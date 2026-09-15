package com.codingame.game.grid.pathfinding;

import java.util.Objects;

import com.codingame.game.grid.Coord;

public class TrainState {

    public Coord coord;
    public TrainState prev;

    public TrainState(Coord coord, TrainState prev) {
        this.coord = coord;
        this.prev = prev;
    }
    
    public TrainState(Coord coord) {
        this(coord, null);
    }

    @Override
    public String toString() {
        return "State{" +
            ", coord=" + coord +
            '}';
    }

    @Override
    public int hashCode() {
        return Objects.hash(coord);
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (obj == null) return false;
        if (getClass() != obj.getClass()) return false;
        TrainState other = (TrainState) obj;
        return Objects.equals(coord, other.coord);
    }

}
