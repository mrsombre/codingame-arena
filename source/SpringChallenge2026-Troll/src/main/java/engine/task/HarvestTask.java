package engine.task;

import com.codingame.game.Player;
import engine.*;

import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class HarvestTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>HARVEST)\\s+(?<id>\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    public HarvestTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        Plant plant = unit.getCell().getPlant();
        if (plant == null) addParsingError("troll " + unit.getId() + " is not at a plant", InputError.NO_PLANT, false);
        else if (plant.getResources() == 0)
            addParsingError("troll " + unit.getId() + " has no fruits to harvest", InputError.NO_FRUIT, false);
        if (unit.getFreeCarryCapacity() == 0)
            addParsingError("troll " + unit.getId() + " has no capacity", InputError.NO_CAPACITY, false);
        if (unit.getHarvestPower() == 0)
            addParsingError("troll " + unit.getId() + " has no harvest power", InputError.NO_HARVEST, false);
    }

    @Override
    public int getTaskPriority() {
        return 2;
    }

    @Override
    public int getRequiredLeague() {
        return 1;
    }

    @Override
    protected Cell getCell() {
        return unit.getCell();
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (this != concurrentTasks.get(0)) return;
        for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
            Plant plant = byCell.get(0).getCell().getPlant();
            for (int i = 1; i <= Constants.PLANT_MAX_RESOURCES; i++) {
                if (plant.getResources() == 0) break;
                for (Task t : byCell) {
                    t.unit.harvest(i);
                    t.applied = true;
                }
            }
            for (Task t : byCell) {
                String itemText = plant.getType().toString();
                if (t.getDeltaCarry() > 1) itemText += "s";
                t.player.addSummary("troll " + t.unit.getId() + " harvested " + t.getDeltaCarry() + " " + itemText);
            }
        }
    }
}