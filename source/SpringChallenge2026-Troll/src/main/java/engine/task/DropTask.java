package engine.task;

import com.codingame.game.Player;
import engine.Board;

import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class DropTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>DROP)\\s+(?<id>\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    public DropTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        if (unit.getInventory().getTotal() == 0)
            addParsingError("troll " + unit.getId() + " has nothing to drop", InputError.NO_SEEDS, false);
        if (!unit.isNearShack())
            addParsingError("troll " + unit.getId() + " isn't next to shack", InputError.NO_SHACK, false);
    }

    @Override
    public int getTaskPriority() {
        return 7;
    }

    @Override
    public int getRequiredLeague() {
        return 1;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        for (int i = 0; i < unit.getInventory().getItemsLength(); i++) {
            player.getInventory().setItem(i, player.getInventory().getItemCount(i) + unit.getInventory().getItemCount(i));
            unit.getInventory().setItem(i, 0);
        }
        applied = true;
        String itemText = "item";
        if (getDeltaCarry() < -1) itemText += "s";
        player.addSummary("troll " + unit.getId() + " dropped " + (-getDeltaCarry()) + " " + itemText + " to the shack");
    }
}