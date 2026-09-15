package com.codingame.game.grid.pathfinding;

import java.util.Collections;
import java.util.Deque;
import java.util.LinkedHashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Set;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.Town;

public class TrainBFS {

    Grid grid;
    Town from, to;

    public TrainBFS(Grid grid, Town from, Town to) {
        this.grid = grid;
        this.from = from;
        this.to = to;

    }

    public List<Coord> search() {
        Deque<TrainState> fifo = new LinkedList<>();
        fifo.add(new TrainState(from.coord));
        Set<Coord> visited = new LinkedHashSet<>();
        visited.add(from.coord);

        while (!fifo.isEmpty()) {
            TrainState current = fifo.poll();
            if (isGoal(current)) {
                // reconstruct path
                List<Coord> path = new LinkedList<>();
                TrainState state = current;
                while (state != null) {
                    path.add(0, state.coord);
                    state = state.prev;
                }
                return path;
            }

            List<Coord> sortedNeighs = grid.getNeighbours(current.coord).stream().sorted((a, b) -> {
                return Direction.fromCoord(b.sub(a)).ordinal();
            }).toList();

            for (Coord neighbour : sortedNeighs) {
                if (grid.canTrainPass(neighbour) && !visited.contains(neighbour)) {
                    visited.add(neighbour);
                    fifo.add(new TrainState(neighbour, current));
                }
            }
        }

        return Collections.emptyList();

    }

    private boolean isGoal(TrainState state) {
        return state.coord.equals(to.coord);
    }
}
