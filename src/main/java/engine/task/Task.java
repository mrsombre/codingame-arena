package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Cell;
import engine.Unit;

import java.util.*;
import java.util.regex.Matcher;
import java.util.stream.Collectors;

public abstract class Task {
    protected Player player;
    protected Board board;
    protected Unit unit;
    private int unitInitialCarry;
    protected boolean failedParsing = false;
    protected boolean applied;

    protected Task(Player player, Board board) {
        this.player = player;
        this.board = board;
    }

    public Player getPlayer() {
        return player;
    }

    public static Task parseTask(Player player, Board board, String command, int league, HashSet<Unit> usedUnits) {
        if (command.trim().equals("")) return null;
        if (command.trim().toUpperCase().equals("WAIT")) return null;
        Task task = null;
        try {
            if (MoveTask.pattern.matcher(command).matches())
                task = new MoveTask(player, board, command);
            if (HarvestTask.pattern.matcher(command).matches())
                task = new HarvestTask(player, board, command);
            if (PlantTask.pattern.matcher(command).matches())
                task = new PlantTask(player, board, command);
            if (ChopTask.pattern.matcher(command).matches())
                task = new ChopTask(player, board, command);
            if (PickTask.pattern.matcher(command).matches())
                task = new PickTask(player, board, command);
            if (TrainTask.pattern.matcher(command).matches())
                task = new TrainTask(player, board, command, league);
            if (DropTask.pattern.matcher(command).matches())
                task = new DropTask(player, board, command);
            if (MineTask.pattern.matcher(command).matches())
                task = new MineTask(player, board, command);
            if (WaitTask.pattern.matcher(command).matches())
                task = new WaitTask(player, board, command);
        } catch (Exception ex) {
            player.addError(new InputError("Unknown command: " + command, InputError.UNKNOWN_COMMAND, true));
            return null;
        }
        if (task == null) {
            player.addError(new InputError("Unknown command: " + command, InputError.UNKNOWN_COMMAND, true));
            return null;
        }
        if (task.unit != null && !usedUnits.add(task.unit)) {
            task.addParsingError("Troll " + task.unit.getId() + " already used", InputError.ALREADY_USED, false);
            return null;
        }
        if (task.getRequiredLeague() > league) {
            task.addParsingError("Command not available in your league: " + command, InputError.NOT_AVAILABLE, false);
            return null;
        }
        return task;
    }

    protected void parseUnit(Matcher matcher) {
        int id = Integer.parseInt(matcher.group("id"));
        unit = board.getUnit(id);
        if (unit == null) addParsingError("Troll " + id + " does not exist", InputError.UNIT_NOT_FOUND, false);
        else {
            unitInitialCarry = unit.getInventory().getTotal();
            if (unit.getPlayer() == player) unit.setAnimateTask(this);
            else addParsingError("You don't own troll " + id, InputError.UNIT_NOT_OWNED, false);
        }
    }

    public Unit getUnit() {
        return unit;
    }

    public int getDeltaCarry() {
        return unit.getInventory().getTotal() - unitInitialCarry;
    }

    public boolean hasFailedParsing() {
        return failedParsing;
    }

    public boolean wasApplied() { return applied; }

    protected void addParsingError(String message, int errorCode, boolean critical) {
        if (failedParsing) return;
        failedParsing = true;
        player.addError(new InputError(message, errorCode, critical));
    }

    public abstract int getTaskPriority();

    public abstract int getRequiredLeague();

    protected Cell getCell() {
        return null;
    }

    public abstract void apply(Board board, ArrayList<Task> concurrentTasks);

    protected ArrayList<ArrayList<Task>> groupByCell(ArrayList<Task> tasks) {
        Hashtable<Cell, ArrayList<Task>> byCell = new Hashtable<>();
        for (Task task : tasks) {
            if (!byCell.containsKey(task.getCell())) byCell.put(task.getCell(), new ArrayList<>());
            byCell.get(task.getCell()).add(task);
        }
        ArrayList<ArrayList<Task>> result = new ArrayList<>();
        for (ArrayList<Task> list : byCell.values().stream().sorted(Comparator.comparing(task -> task.get(0).getCell().getId())).collect(Collectors.toList()))
            result.add(list);
        return result;
    }

    protected ArrayList<ArrayList<Task>> groupByPlayer(ArrayList<Task> tasks) {
        Hashtable<Player, ArrayList<Task>> byPlayer = new Hashtable<>();
        for (Task task : tasks) {
            if (!byPlayer.containsKey(task.player)) byPlayer.put(task.player, new ArrayList<>());
            byPlayer.get(task.player).add(task);
        }
        ArrayList<ArrayList<Task>> result = new ArrayList<>();
        for (ArrayList<Task> list : byPlayer.values()) result.add(list);
        return result;
    }
}