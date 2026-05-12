package engine.task;

import com.codingame.game.Player;
import engine.*;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class PlantTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>PLANT)\\s+(?<id>\\d+)\\s+(?<type>\\w+)\\s*$", Pattern.CASE_INSENSITIVE);

    private int type;

    public PlantTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        String typeText = matcher.group("type");
        if (typeText.equalsIgnoreCase("PLUM")) type = Item.PLUM.ordinal();
        else if (typeText.equalsIgnoreCase("LEMON")) type = Item.LEMON.ordinal();
        else if (typeText.equalsIgnoreCase("APPLE")) type = Item.APPLE.ordinal();
        else if (typeText.equalsIgnoreCase("BANANA")) type = Item.BANANA.ordinal();
        else if (typeText.equalsIgnoreCase("IRON")) type = Item.IRON.ordinal();
        else if (typeText.equalsIgnoreCase("WOOD")) type = Item.WOOD.ordinal();
        else type = Integer.parseInt(typeText);
        if (type < 0 || type >= Item.values().length || !Item.values()[type].isPlant())
            addParsingError(type + " is not a plant", InputError.INVALID_PLANT, false);
        if (unit.getCell().getType() != Cell.Type.GRASS)
            addParsingError("troll " + unit.getId() + " is not on grass. How did you even get there?", InputError.NO_GRASS, false);
        if (unit.getCell().getPlant() != null)
            addParsingError("troll " + unit.getId() + " can't plant on top of existing plant", InputError.EXISTING_PLANT, false);
        if (unit.getInventory().getItemCount(type) == 0)
            addParsingError("troll " + unit.getId() + " has no " + Item.values()[type] + " seed to plant", InputError.NO_SEEDS, false);
    }

    @Override
    public int getTaskPriority() {
        return 3;
    }

    @Override
    public int getRequiredLeague() {
        return 2;
    }

    @Override
    protected Cell getCell() {
        return unit.getCell();
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (this != concurrentTasks.get(0)) return;
        for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
            HashSet<Integer> types = new HashSet<>();
            for (Task t : byCell) {
                types.add(((PlantTask) t).type);
            }
            if (types.size() == 1) { // only proceed if all tasks plant the same type
                for (Task t : byCell) {
                    int type = ((PlantTask) t).type;
                    t.unit.getInventory().decrementItem(type);
                    if (t.getCell().getPlant() == null) {
                        Plant newPlant = new Plant(t.unit.getCell(), Item.values()[type]);
                        t.unit.getCell().setPlant(newPlant);
                        board.addPlant(newPlant);
                    }
                    t.applied = true;
                    t.player.addSummary("troll " + t.unit.getId() + " planted a " + Item.values()[type]);
                }
            } else { // contradicting planting, no one gets to plant
                for (Task t : byCell) {
                    addParsingError("troll " + t.unit.getId() + " can't plant " + Item.values()[type] + ", contradicting opponent planting", InputError.OPPONENT_BLOCKING, false);
                }
            }
        }
    }
}