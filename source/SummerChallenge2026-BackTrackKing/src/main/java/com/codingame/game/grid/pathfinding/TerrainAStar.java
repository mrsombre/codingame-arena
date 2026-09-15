package com.codingame.game.grid.pathfinding;

import java.util.List;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.Town;

public class TerrainAStar extends AbstractAStar<Coord> {

    Grid grid;
    Town from, to;

    public TerrainAStar(Grid grid, Town from, Town to) {
        this.grid = grid;
        this.from = from;
        this.to = to;

    }

    @Override
    protected Coord getInitialState() {
        return from.coord;
    }

    @Override
    protected double heuristic(Coord state) {
        return state.manhattanTo(to.coord);
    }

    @Override
    protected double cost(Coord from, Coord to) {
        return 1;
    }

    @Override
    protected List<Coord> getSuccessors(Coord state) {
        List<Coord> neighs = grid.getNeighbours(state).stream()
            .filter(coord -> !grid.zones.get(grid.get(coord).zoneId).inked)
            .toList();
        return neighs;
    }

    @Override
    protected boolean isGoal(Coord state) {
        return state.equals(to.coord);
    }
}
