package com.codingame.game.grid.pathfinding;

import java.util.List;

import com.codingame.game.grid.Direction;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.Town;

public class TrainAStar extends AbstractAStar<TrainState> {

    Grid grid;
    Town from, to;

    public TrainAStar(Grid grid, Town from, Town to) {
        this.grid = grid;
        this.from = from;
        this.to = to;

    }

    @Override
    protected TrainState getInitialState() {
        return new TrainState(from.coord);
    }

    @Override
    protected double heuristic(TrainState state) {
        return state.coord.manhattanTo(to.coord);
    }

    @Override
    protected double cost(TrainState from, TrainState to) {
        return 1;
    }

    @Override
    protected List<TrainState> getSuccessors(TrainState state) {
        return grid.getNeighbours(state.coord).stream()
            .filter(coord -> grid.canTrainPass(coord))
            .map(TrainState::new)
            .toList();
    }

    @Override
    protected boolean isGoal(TrainState state) {
        return state.coord.equals(to.coord);
    }

    @Override
    protected int tieBreaker(TrainState from, TrainState to) {
        return Direction.fromCoord(to.coord.sub(from.coord)).ordinal();
    }
}
