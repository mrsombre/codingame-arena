package com.codingame.game.grid.pathfinding;

import java.util.*;

public abstract class AbstractAStar<T> {

    protected abstract T getInitialState();

    protected abstract boolean isGoal(T state);

    protected abstract double heuristic(T state);

    protected abstract List<T> getSuccessors(T state);

    protected abstract double cost(T from, T to);

    protected int tieBreaker(T from, T to) {
        return 0;
    }

    public Optional<List<T>> search() {
        PriorityQueue<Node> open = new PriorityQueue<>(
            Comparator.comparingDouble((Node n) -> n.f)
                .thenComparingInt(n -> n.tieBreaker)
        );
        Map<T, Double> gScores = new LinkedHashMap<>();
        Map<T, T> cameFrom = new LinkedHashMap<>();

        T start = getInitialState();
        gScores.put(start, 0.0);
        open.add(new Node(start, heuristic(start), 0));

        while (!open.isEmpty()) {
            Node current = open.poll();
            if (isGoal(current.state)) {
                return Optional.of(reconstructPath(cameFrom, current.state));
            }

            for (T neighbor : getSuccessors(current.state)) {
                double tentativeG = gScores.get(current.state) + cost(current.state, neighbor);
                if (tentativeG < gScores.getOrDefault(neighbor, Double.POSITIVE_INFINITY)) {
                    cameFrom.put(neighbor, current.state);
                    gScores.put(neighbor, tentativeG);
                    int tieBreakerValue = tieBreaker(current.state, neighbor);
                    open.add(new Node(neighbor, tentativeG + heuristic(neighbor), tieBreakerValue));
                }
            }
        }

        return Optional.empty();
    }

    private List<T> reconstructPath(Map<T, T> cameFrom, T current) {
        List<T> path = new ArrayList<>();
        while (current != null) {
            path.add(current);
            current = cameFrom.get(current);
        }
        Collections.reverse(path);
        return path;
    }

    private class Node {
        T state;
        double f;
        int tieBreaker;

        Node(T state, double f, int tieBreaker) {
            this.state = state;
            this.f = f;
            this.tieBreaker = tieBreaker;
        }
    }
}
