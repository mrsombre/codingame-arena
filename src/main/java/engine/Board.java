package engine;

import com.codingame.CompressionModule;
import com.codingame.game.Player;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.toggle.ToggleModule;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import engine.task.Task;
import engine.task.TaskManager;
import engine.task.TrainTask;
import view.BoardView;

import java.util.*;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.Collectors;
import java.util.stream.Stream;

public class Board {
    private int width, height;
    private Random random;
    private Cell[][] grid;
    private ArrayList<Player> players = new ArrayList<>();
    private ArrayList<Unit> units = new ArrayList<>();
    private ArrayList<Plant> plants = new ArrayList<>();
    private BoardView view;
    private GraphicEntityModule graphicEntityModule;
    private ToggleModule toggleModule;
    private CompressionModule compressionModule;

    private Board(int width, int height, Random random) {
        this.width = width;
        this.height = height;
        this.random = random;
        grid = new Cell[width][height];
        for (int x = 0; x < width; x++) {
            for (int y = 0; y < height; y++) {
                grid[x][y] = new Cell(x, y);
            }
        }
        for (int x = 0; x < width; x++) {
            for (int y = 0; y < height; y++) {
                grid[x][y].initNeighbors(this);
            }
        }
    }

    private void initView(GraphicEntityModule graphicEntityModule, TooltipModule tooltipModule, ToggleModule toggleModule, CompressionModule compressionModule) {
        this.graphicEntityModule = graphicEntityModule;
        this.toggleModule = toggleModule;
        this.compressionModule = compressionModule;
        view = new BoardView(this, graphicEntityModule, tooltipModule, toggleModule, random);
        for (Plant plant : plants) plant.initView(toggleModule, compressionModule);
        for (Unit unit : units) unit.initView(graphicEntityModule, toggleModule, compressionModule, view.getMainGroup());
        for (Player player : players) player.initView(graphicEntityModule, tooltipModule);
    }

    public int getWidth() {
        return width;
    }

    public int getHeight() {
        return height;
    }

    public static Board loadMap(String map) {
        String[] lines = map.split("\n");
        Board board = new Board(lines[0].length(), lines.length, new Random());
        ArrayList<Cell> cells = new ArrayList<>();
        for (int x = 0; x < board.getWidth(); x++) {
            for (int y = 0; y < board.getHeight(); y++) {
                cells.add(board.grid[x][y]);
                switch (lines[y].charAt(x)) {
                    case '~':
                        board.grid[x][y].setType(Cell.Type.WATER);
                        break;
                    case '#':
                        board.grid[x][y].setType(Cell.Type.ROCK);
                        break;
                    case '+':
                        board.grid[x][y].setType(Cell.Type.IRON);
                        break;
                    case '0':
                    case '1':
                        board.grid[x][y].setType(Cell.Type.SHACK);
                        break;
                    default:
                        break;
                }
            }
        }
        return board;
    }

    public static Board createMap(List<Player> players, Random random, int league, GraphicEntityModule graphicEntityModule, TooltipModule tooltipModule, ToggleModule toggleModule, CompressionModule compressionModule) {
        while (true) {
            int height = random.nextInt(Constants.MAP_MAX_HEIGHT - Constants.MAP_MIN_HEIGHT + 1) + Constants.MAP_MIN_HEIGHT;
            if (league <= 2) height = Constants.MAP_MIN_HEIGHT;
            int width = 2 * height;
            Board board = new Board(width, height, random);

            // rivers
            if (league > 2) {
                int maxTotalRiver = width * height - 2 * (Constants.MAP_MAX_ROCK + Constants.PLANT_MAX_SIZE * 4 + Constants.MAP_MAX_IRON + 1) * 4 / 5;
                int riversCount = random.nextInt(Constants.MAP_MAX_RIVER - Constants.MAP_MIN_RIVER + 1) + Constants.MAP_MIN_RIVER;
                for (int i = 0; i < riversCount; i++) {
                    Cell river = board.getRandomCell();
                    for (int j = 0; j<10 && river.isNearEdge(); j++) river = board.getRandomCell();
                    while (river != null && maxTotalRiver > 0) {
                        board.setCellType(river, Cell.Type.WATER);
                        river = river.getNeighbor(random.nextInt(4));
                        maxTotalRiver -= 2;
                    }
                }
            }

            // shacks, inventory
            int[] inventory = new int[1 + Item.IRON.ordinal()];
            if (league > 1) {
                for (int i = 0; i < inventory.length; i++)
                    inventory[i] = Constants.MIN_STARTING_RESOURCE + random.nextInt(Constants.MAX_STARTING_RESOURCE - Constants.MIN_STARTING_RESOURCE + 1);
            }
            if (league < 3) inventory[Item.IRON.ordinal()] = 0;
            Cell shack = board.grid[random.nextInt(width / 2)][random.nextInt(height)];
            while (shack.getType() == Cell.Type.WATER) {
                shack = board.grid[random.nextInt(width / 2)][random.nextInt(height)];
            }
            board.setCellType(shack, Cell.Type.SHACK);

            Unit.idCounter = 0;
            Cell[] shacks = {shack, board.grid[width - 1 - shack.getX()][height - 1 - shack.getY()]};
            for (int i = 0; i < players.size(); i++) {
                Player player = players.get(i);
                player.init(shacks[i], league);
                board.players.add(player);
                board.units.add(player.getUnits().get(0));
                player.setInventory(inventory);
                player.recomputeScore();
            }

            // resources
            if (league > 2) {
                board.placeTerrain(Cell.Type.IRON, Constants.MAP_MIN_IRON, Constants.MAP_MAX_IRON);
                board.placeTerrain(Cell.Type.ROCK, Constants.MAP_MIN_ROCK, Constants.MAP_MAX_ROCK);
            }
            board.placeTree(Item.PLUM, Constants.MAP_MIN_TREE, Constants.MAP_MAX_TREE);
            board.placeTree(Item.LEMON, Constants.MAP_MIN_TREE, Constants.MAP_MAX_TREE);
            board.placeTree(Item.APPLE, Constants.MAP_MIN_TREE, Constants.MAP_MAX_TREE);
            board.placeTree(Item.BANANA, Constants.MAP_MIN_TREE, Constants.MAP_MAX_TREE);

            if (board.isValid(league)) {
                board.initView(graphicEntityModule, tooltipModule, toggleModule, compressionModule);
                return board;
            }
        }
    }

    public Cell getCell(int x, int y) {
        return grid[x][y];
    }

    public void addPlant(Plant plant) {
        plants.add(plant);
    }

    public ArrayList<Plant> getPlants() {
        return plants;
    }

    public ArrayList<Player> getPlayers() {
        return players;
    }

    public void addUnit(Unit unit) {
        units.add(unit);
    }

    public Stream<Unit> getUnitsByPlayerId(int playerId) {
        return units.stream().filter(u -> u.getPlayer().getIndex() == playerId);
    }

    public Stream<Unit> getUnitsByCell(Cell cell) {
        return units.stream().filter(u -> u.getCell() == cell);
    }

    private void placeTree(Item type, int min, int max) {
        int count = random.nextInt(max - min + 1) + min;
        for (int i = 0; i < count; i++) {
            Cell cell = getRandomCell();
            cell.setPlant(new Plant(cell, type));
            int ticks = random.nextInt(1, cell.getPlant().getGrowthCooldown() * (Constants.PLANT_MAX_SIZE + Constants.PLANT_MAX_RESOURCES));
            for (int t = 0; t < ticks; t++) cell.getPlant().tick(true);
            addPlant(cell.getPlant());

            Cell mirror = cell;
            cell = grid[width - 1 - cell.getX()][height - 1 - cell.getY()];
            if (cell == mirror) return;
            cell.setPlant(new Plant(cell, type));
            for (int t = 0; t < ticks; t++) cell.getPlant().tick(true);
            addPlant(cell.getPlant());
        }
    }

    private void placeTerrain(Cell.Type type, int min, int max) {
        int count = random.nextInt(max - min + 1) + min;
        for (int i = 0; i < count; i++) {
            Cell cell = getRandomCell();
            setCellType(cell, type);
        }
    }

    private Cell getRandomCell() {
        while (true) {
            Cell cell = grid[random.nextInt(width)][random.nextInt(height)];
            if (cell.getType() == Cell.Type.GRASS && cell.getPlant() == null) return cell;
        }
    }

    private void setCellType(Cell cell, Cell.Type type) {
        cell.setType(type);
        grid[width - 1 - cell.getX()][height - 1 - cell.getY()].setType(type);
    }

    private boolean isValid(int league) {
        if (players.get(0).getShack().isNearIron()) return false;

        boolean shackReachable = false;
        for (Cell cell : players.get(0).getShack().getNeighbors()) {
            if (cell != null && cell.isWalkable()) shackReachable = true;
        }
        if (!shackReachable) return false;

        ArrayList<Cell> walkables = new ArrayList<>();
        ArrayList<Cell> irons = new ArrayList<>();
        for (int x = 0; x < width; x++) {
            for (int y = 0; y < height; y++) {
                if (grid[x][y].isWalkable()) walkables.add(grid[x][y]);
                if (grid[x][y].getType() == Cell.Type.IRON) irons.add(grid[x][y]);
            }
        }
        boolean canReachIron = false;
        for (Cell iron : irons) {
            for (Cell cell : iron.getNeighbors()) {
                if (cell != null && cell.isWalkable()) canReachIron = true;
            }
        }
        if (!canReachIron && league > 2) return false;

        int[][] dist = getDistances(walkables.get(0));
        for (Cell cell : walkables) {
            if (dist[cell.getX()][cell.getY()] == -1) return false;
        }

        int[][] shackDist = getDistances(players.get(0).getShack());
        int opponentDist = Arrays.stream(players.get(1).getShack().getNeighbors())
                .filter(c -> c != null && c.isWalkable())
                .mapToInt(c -> shackDist[c.getX()][c.getY()] + 1)
                .min()
                .orElse(Integer.MAX_VALUE);
        if (opponentDist > Constants.MAP_MAX_OPP_DIST)
            return false;

        if (league < 3 && plants.stream().noneMatch(p -> p.getResources() > 0)) return false;

        return true;
    }

    public Cell getNextCell(Unit unit, Cell target) {
        return getNextCell(unit.getCell(), target, unit.getMovementSpeed());
    }

    public Cell getNextCell(Cell current, Cell target, int speed) {
        int[][] targetDist = getDistances(target);
        int[][] sourceDist = getDistances(current);
        if (sourceDist[target.getX()][target.getY()] >= 0 && sourceDist[target.getX()][target.getY()] <= speed)
            return target;
        if (sourceDist[target.getX()][target.getY()] == -1) {
            ArrayList<Cell> closest = new ArrayList<>();
            int best = width * height;
            for (int x = 0; x < width; x++) {
                for (int y = 0; y < height; y++) {
                    if (sourceDist[x][y] == -1) continue;
                    int d = target.manhattan(grid[x][y]);
                    if (d < best) {
                        best = d;
                        closest.clear();
                    }
                    if (d == best) closest.add(grid[x][y]);
                }
            }
            targetDist = getDistances(closest);
        }

        ArrayList<Cell> closest = new ArrayList<>();
        int best = width * height;
        for (int x = 0; x < width; x++) {
            for (int y = 0; y < height; y++) {
                if (sourceDist[x][y] > speed || sourceDist[x][y] == -1) continue;
                int d = targetDist[x][y];
                if (d >= 0 && d < best) {
                    best = d;
                    closest.clear();
                }
                if (d == best) closest.add(grid[x][y]);
            }
        }
        return closest.get(random.nextInt(closest.size()));
    }

    public int[][] getDistances(Cell cell) {
        ArrayList<Cell> starts = new ArrayList<>();
        starts.add(cell);
        return getDistances(starts);
    }

    private int[][] getDistances(ArrayList<Cell> starts) {
        int[][] result = new int[width][height];
        for (int x = 0; x < width; x++) {
            for (int y = 0; y < height; y++) {
                result[x][y] = -1;
            }
        }
        Queue<Cell> queue = new ConcurrentLinkedDeque<>();
        for (Cell cell : starts) {
            queue.add(cell);
            result[cell.getX()][cell.getY()] = 0;
        }
        while (queue.size() > 0) {
            Cell c = queue.poll();
            for (Cell n : c.getNeighbors()) {
                if (n != null && n.isWalkable() && result[n.getX()][n.getY()] == -1) {
                    result[n.getX()][n.getY()] = result[c.getX()][c.getY()] + 1;
                    queue.add(n);
                }
            }
        }
        return result;
    }

    public void tick(int turn, TaskManager taskManager, GameManager gameManager) {
        while (taskManager.hasTasks()) {
            ArrayList<Task> tasks = taskManager.popTasks();
            for (Task t : tasks) {
                t.apply(this, tasks);
                if (t.wasApplied() && t instanceof TrainTask)
                    gameManager.addTooltip(t.getPlayer(), t.getPlayer().getNicknameToken() + " trained a unit");
            }
        }

        for (Plant plant : plants) {
            plant.initView(toggleModule, compressionModule);
            plant.tick(true);
        }
        plants = plants.stream().filter(p -> !p.isDead()).collect(Collectors.toCollection(ArrayList::new));
        for (Unit unit : units) {
            unit.initView(graphicEntityModule, toggleModule, compressionModule, view.getMainGroup());
            unit.updateView();
        }
        for (Player player : players) {
            player.recomputeScore();
            player.updateView(turn);
        }
        view.update();
    }

    public ArrayList<String> getInitialInputs(int id) {
        ArrayList<String> result = new ArrayList<>();
        result.add(getWidth() + " " + getHeight());
        for (int y = 0; y < getHeight(); y++) {
            StringBuilder line = new StringBuilder();
            for (int x = 0; x < getWidth(); x++) {
                Cell cell = grid[x][y];
                if (cell.getType() == Cell.Type.GRASS) line.append('.');
                if (cell.getType() == Cell.Type.WATER) line.append('~');
                if (cell.getType() == Cell.Type.IRON) line.append('+');
                if (cell.getType() == Cell.Type.ROCK) line.append('#');
                if (cell.getType() == Cell.Type.SHACK) {
                    Player owner = players.stream().filter(p -> p.getShack() == cell).findFirst().get();
                    line.append((owner.getIndex() - id + players.size()) % players.size());
                }
            }
            result.add(line.toString());
        }
        return result;
    }

    public ArrayList<String> getTurnInputs(int id) {
        ArrayList<String> result = new ArrayList<>();
        for (int i = 0; i < players.size(); i++) {
            Player player = players.get((i + id) % players.size());
            result.add(player.getInventory().getInputLine());
        }
        result.add(String.valueOf(plants.size()));
        for (Plant plant : plants) result.add(plant.getInputLine());
        result.add(String.valueOf(units.size()));
        for (Unit unit : units) result.add(unit.getInputLine(id, players.size()));
        return result;
    }

    public Unit getUnit(int id) {
        for (Unit unit : units) {
            if (unit.getId() == id) return unit;
        }
        return null;
    }

    // the game ends early when either player can enforce a win by doing nothing
    private int noTreeCounter = 0;
    public boolean hasStalled() {
        if (plants.size() > 0) {
            noTreeCounter = 0;
            return false;
        }
        noTreeCounter++;
        if (noTreeCounter == Constants.STALL_LIMIT) return true;
        boolean[] playerStuck = { true, true };
        for (Unit unit : units) {
            if (unit.getInventory().getTotal() > unit.getInventory().getItemCount(Item.IRON)) playerStuck[unit.getPlayer().getIndex()] = false;
        }
        for (Player player : players) {
            for (int i = 0; i <= Item.BANANA.ordinal(); i++)
                if (player.getInventory().getItemCount(i) > 0) playerStuck[player.getIndex()] = false;
        }

        if (playerStuck[0] && playerStuck[1]) return true;
        if (playerStuck[0] && players.get(0).getScore() < players.get(1).getScore()) return true;
        if (playerStuck[1] && players.get(1).getScore() < players.get(0).getScore()) return true;
        return false;
    }
}
