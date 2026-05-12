package engine.task;

import com.codingame.game.Player;
import engine.Board;

import java.util.ArrayList;
import java.util.regex.Pattern;

public class WaitTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>WAIT)\\s*$", Pattern.CASE_INSENSITIVE);

    public WaitTask(Player player, Board board, String command) {
        super(player, board);
    }

    @Override
    public int getTaskPriority() {
        return 10;
    }

    @Override
    public int getRequiredLeague() {
        return 1;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        applied = true;
    }
}