package com.codingame.game.grid.pathfinding;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.LinkedHashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Set;

import com.codingame.game.Game;
import com.codingame.game.Player;
import com.codingame.game.action.Action;
import com.codingame.game.action.ActionType;
import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;
import com.codingame.game.grid.Grid;
import com.codingame.game.grid.Tile;
import com.codingame.game.grid.Zone;
import com.codingame.game.grid.pathfinding.AutobuildState.CursorState;

public class AutobuildAStar extends AbstractAStar<AutobuildState> {

    Grid grid;
    Player player;
    Coord from, to;
    List<Zone> zones;

    public AutobuildAStar(Grid grid, Player player, Coord from, Coord to) {
        this.grid = grid;
        this.player = player;
        this.from = from;
        this.to = to;
        this.zones = grid.zones;

    }

    @Override
    protected AutobuildState getInitialState() {
        AutobuildState s = new AutobuildState();
        s.action = null;

        s.ownedZones = new HashSet<>();
        s.moneySpent = 0;
        s.tracks = new HashSet<>();
        s.cursor = from;
        Tile t = grid.get(from);

        if (!t.isValid()) {
            s.cursorState = CursorState.BROKEN;
            return s;
        }

        Zone zone = zones.get(t.zoneId);

        boolean usableTrack = t.isTrackOrTown();

        s.cursorState = usableTrack
            ? AutobuildState.CursorState.BUILT 
            : AutobuildState.CursorState.BUILDABLE;

        return s;
    }

    @Override
    protected int tieBreaker(AutobuildState from, AutobuildState to) {
        return Direction.fromCoord(to.cursor.sub(from.cursor)).ordinal();
    }

    @Override
    protected boolean isGoal(AutobuildState state) {
        if (state.cursor.equals(to)) {
            return true;
        }

        Tile t = grid.get(state.cursor);
        if (t.isTrackOrTown()) {
            return isPartOfRailBlock(state.cursor, to);
        }

        return false;

    }

    private boolean isPartOfRailBlock(Coord coord, Coord goal) {
        LinkedList<Coord> fifo = new LinkedList<>();
        Set<Coord> visited = new HashSet<>();
        fifo.add(coord);
        visited.add(coord);
        while (!fifo.isEmpty()) {
            Coord current = fifo.poll();
            if (current.equals(goal)) {
                return true;
            }
            List<Coord> neighbours = grid.getNeighbours(current);
            for (Coord n : neighbours) {
                Tile t = grid.get(n);
                if (!visited.contains(n) && t.isTrackOrTown()) {
                    fifo.add(n);
                    visited.add(n);
                }
            }
        }
        return false;
    }

    @Override
    protected double heuristic(AutobuildState state) {
        return state.cursor.manhattanTo(state.cursor) * Game.BASE_RAIL_COST * Game.GRASS_COST_MULTIPLIER;
    }

    @Override
    protected double cost(AutobuildState from, AutobuildState to) {
        return (to.moneySpent - from.moneySpent);
    }

    @Override
    protected List<AutobuildState> getSuccessors(AutobuildState state) {

        // Invalid start point
        if (state.cursorState == CursorState.BROKEN) {
            return List.of();
        }

        // Invalid end point
        if (!grid.get(to).isValid()) {
            return List.of();
        }

        Set<AutobuildState> result = new LinkedHashSet<>();

        List<Coord> scope = new ArrayList<>(4);

        Tile currentTile = grid.get(state.cursor);
        boolean currentIsTown = currentTile.townId != Tile.TOWN_NONE;
        boolean usableTrack = state.tracks.contains(state.cursor)
            || currentTile.isTrackOrTown();

        if (usableTrack) {
            if (state.tracks.contains(state.cursor)) {
                // Potential track created during path finding
                scope.addAll(grid.getNeighbours(state.cursor));
            } else {
                // Preexisting track
                scope.addAll(getAllNeighboursOfRailBlock(state.cursor));
            }
        } else {
            scope.add(state.cursor);
        }

        for (Coord neigh : scope) {
            Tile t = grid.get(neigh);
            Zone zone = zones.get(t.zoneId);
            boolean isTown = t.townId != Tile.TOWN_NONE;
            boolean buildableZone = !zone.inked;
            if (state.ownedZones.contains(t.zoneId) || buildableZone || isTown) {
                // No point in looping back
                if (state.tracks.contains(neigh)) {
                    continue; // this never happens
                }
                if (t.isTrackOrTown()) {
                    if (neigh.equals(state.cursor)) {
                        continue; // skip self, should never happen either
                    }
                    // Move cursor onto town
                    AutobuildState next = new AutobuildState();
                    next.action = null;
                    next.cursor = neigh;
                    next.moneySpent = state.moneySpent;
                    next.ownedZones = state.ownedZones;
                    next.tracks = state.tracks;
                    next.cursorState = AutobuildState.CursorState.BUILT;
                    result.add(next);
                } else {
                    int railCost = Game.getRailCost(t);
                    AutobuildState next = new AutobuildState();
                    next.action = new Action(ActionType.PLACE_TRACK, true);
                    next.action.setCoord(neigh);
                    next.cursor = neigh;
                    next.moneySpent = state.moneySpent + railCost;
                    next.ownedZones = state.ownedZones;
                    next.tracks = new HashSet<>(state.tracks);
                    next.tracks.add(neigh);
                    next.cursorState = AutobuildState.CursorState.BUILT;
                    result.add(next);
                }
            }
        }

        return result.stream().toList();
    }

    private List<Coord> getAllNeighboursOfRailBlock(Coord cursor) {
        List<Coord> neighs = new ArrayList<>();
        LinkedList<Coord> fifo = new LinkedList<>();
        fifo.add(cursor);
        Set<Coord> visited = new HashSet<>();
        visited.add(cursor);
        while (!fifo.isEmpty()) {
            Coord current = fifo.poll();
            boolean usableTrack = grid.get(current).isTrackOrTown();

            if (!usableTrack) {
                neighs.add(current);
                continue; // not part of rail block
            }
            List<Coord> neighbours = grid.getNeighbours(current);
            for (Coord n : neighbours) {
                if (!visited.contains(n)) {
                    fifo.add(n);
                    visited.add(n);
                }
            }
        }
        return neighs;
    }

}
