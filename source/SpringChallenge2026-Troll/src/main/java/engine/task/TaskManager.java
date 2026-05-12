package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Unit;

import java.util.ArrayList;
import java.util.HashSet;

public class TaskManager {
    private ArrayList<Task> tasks = new ArrayList<>();
    private Board board;

    public void parseTasks(Player player, Board board, String command, int league) {
        player.setMessage("");
        this.board = board;
        HashSet<Unit> usedUnits = new HashSet<>();
        for (String comm : command.split(";")) {
            if (comm.toUpperCase().startsWith("MSG ")) {
                player.setMessage(comm.substring(4).trim());
                continue;
            }
            Task task = Task.parseTask(player, board, comm, league, usedUnits);
            if (task != null && !task.hasFailedParsing()) tasks.add(task);
        }
    }

    public boolean hasTasks() {
        return peekTasks().size() > 0;
    }

    public ArrayList<Task> popTasks() {
        ArrayList<Task> pop = peekTasks();
        for (Task task : pop) tasks.remove(task);
        return pop;
    }

    private ArrayList<Task> peekTasks() {
        ArrayList<Task> result = new ArrayList<>();
        for (Task task : tasks) {
            if (result.size() > 0 && result.get(0).getTaskPriority() > task.getTaskPriority()) result.clear();
            if ((result.size() == 0 || task.getTaskPriority() == result.get(0).getTaskPriority()))
                result.add(task);
        }

        return result;
    }
}
