package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Item;

import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class MineTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>MINE)\\s+(?<id>\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    public MineTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        if (!unit.getCell().isNearIron())
            addParsingError("troll " + unit.getId() + " is not next to iron", InputError.NO_IRON, false);
        if (unit.getFreeCarryCapacity() == 0)
            addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
        if (unit.getChopPower() == 0)
            addParsingError("troll " + unit.getId() + " has no chopping power", InputError.NO_CHOP, false);
    }

    @Override
    public int getTaskPriority() {
        return 8;
    }

    @Override
    public int getRequiredLeague() {
        return 3;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        unit.mine();
        applied = true;
        player.addSummary("troll " + unit.getId() + " collected " + getDeltaCarry() + " " + Item.IRON);
    }
}