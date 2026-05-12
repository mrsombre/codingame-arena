package engine;

import java.util.Arrays;
import java.util.Random;

import com.codingame.game.Player;
import view.BoardView;

public class Cell {
    public enum Type {
        GRASS, WATER, ROCK, IRON, SHACK
    }

    private int x, y;
    private Type type;
    private Plant plant;

    private Cell[] neighbors = new Cell[4];

    public Cell(int x, int y) {
        this.x = x;
        this.y = y;
        this.type = Type.GRASS;
    }

    public int getX() {
        return x;
    }

    public int getY() {
        return y;
    }

    public int getId() {
        return x + (y << 16);
    }

    public Type getType() {
        return type;
    }

    public void setType(Type type) {
        this.type = type;
    }

    public Plant getPlant() {
        return plant;
    }

    public void setPlant(Plant plant) {
        this.plant = plant;
    }

    public Cell getNeighbor(int index) {
        return neighbors[index];
    }

    public Cell[] getNeighbors() {
        return neighbors;
    }

    private static final int[] dx = {0, 1, 0, -1};
    private static final int[] dy = {1, 0, -1, 0};

    public void initNeighbors(Board board) {
        for (int dir = 0; dir < 4; dir++) {
            int x_ = x + dx[dir];
            int y_ = y + dy[dir];
            if (x_ < 0 || x_ >= board.getWidth() || y_ < 0 || y_ >= board.getHeight()) continue;
            neighbors[dir] = board.getCell(x_, y_);
        }
    }

    public boolean isWalkable() {
        return type == Type.GRASS;
    }

    public boolean isNearWater() {
        return isNearType(Type.WATER);
    }

    public boolean isNearIron() {
        return isNearType(Type.IRON);
    }

    public boolean isNearShack(Player player) {
        return this == player.getShack() || Arrays.stream(neighbors).anyMatch(c -> c == player.getShack());
    }

    public boolean isNearEdge() {
        return Arrays.stream(neighbors).anyMatch(n -> n == null);
    }

    public int manhattan(Cell cell) {
        return Math.abs(x - cell.x) + Math.abs(y - cell.y);
    }

    private boolean isNearType(Type type) {
        for (Cell neighbor : neighbors) {
            if (neighbor != null && neighbor.type == type) return true;
        }
        return false;
    }

    private static Random random = new Random();

    public int getViewX() {
        return x * BoardView.SPRITE_SIZE;
    }

    public int getViewY() {
        return y * BoardView.SPRITE_SIZE;
    }

    @Override
    public String toString() {
        return "(" + x + "/" + y + ")";
    }
}
