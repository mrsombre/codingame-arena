package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Item;

import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class PickTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>PICK)\\s+(?<id>\\d+)\\s+(?<type>\\w+)\\s*$", Pattern.CASE_INSENSITIVE);

    private int type;

    public PickTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        if (unit.getFreeCarryCapacity() == 0)
            addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
        String typeText = matcher.group("type");
        if (typeText.equalsIgnoreCase("PLUM")) type = Item.PLUM.ordinal();
        else if (typeText.equalsIgnoreCase("LEMON")) type = Item.LEMON.ordinal();
        else if (typeText.equalsIgnoreCase("APPLE")) type = Item.APPLE.ordinal();
        else if (typeText.equalsIgnoreCase("BANANA")) type = Item.BANANA.ordinal();
        else if (typeText.equalsIgnoreCase("IRON")) type = Item.IRON.ordinal();
        else if (typeText.equalsIgnoreCase("WOOD")) type = Item.WOOD.ordinal();
        else type = Integer.parseInt(typeText);
        if (type < 0 || type >= Item.values().length || !Item.values()[type].isPlant())
            addParsingError(typeText + " is not a plant", InputError.INVALID_PLANT, false);
        else if (player.getInventory().getItemCount(type) == 0)
            addParsingError(Item.values()[type] + " is out of stock", InputError.OUT_OF_STOCK, false);
        if (!unit.isNearShack())
            addParsingError("troll " + unit.getId() + " isn't next to shack", InputError.NO_SHACK, false);
    }

    @Override
    public int getTaskPriority() {
        return 5;
    }

    @Override
    public int getRequiredLeague() {
        return 2;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (player.getInventory().getItemCount(type) > 0) {
            player.getInventory().decrementItem(type);
            unit.getInventory().incrementItem(type);
            applied = true;
            player.addSummary("troll " + unit.getId() + " picked 1 " + Item.values()[type]);
        } else
            player.addError(new InputError("troll " + unit.getId() + " can't pick " + Item.values()[type] + ", out of stock", InputError.CANT_AFFORD, false));
    }
}