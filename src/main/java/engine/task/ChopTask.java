package engine.task;

import com.codingame.game.Player;
import engine.*;

import java.util.ArrayList;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class ChopTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>CHOP)\\s+(?<id>\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    public ChopTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        if (unit.getCell().getPlant() == null)
            addParsingError("troll " + unit.getId() + " is not at a plant", InputError.NO_PLANT, false);
        if (unit.getChopPower() == 0)
            addParsingError("troll " + unit.getId() + " has no chopping power", InputError.NO_CHOP, false);
    }

    @Override
    public int getTaskPriority() {
        return 4;
    }

    @Override
    protected Cell getCell() {
        return unit.getCell();
    }

    @Override
    public int getRequiredLeague() {
        return 3;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (this != concurrentTasks.get(0)) return;
        for (ArrayList<Task> byCell : groupByCell(concurrentTasks)) {
            Plant plant = byCell.get(0).getCell().getPlant();
            for (Task t : byCell) {
                plant.damage(t.unit.getChopPower());
                t.applied = true;
            }
            if (plant.isDead()) {
                int remainingWood = plant.getSize();
                for (int i = 0; i < plant.getSize() && remainingWood > 0; i++) {
                    for (Task t : byCell) {
                        if (t.unit.getFreeCarryCapacity() > 0) {
                            t.unit.getInventory().incrementItem(Item.WOOD);
                            remainingWood--;
                        }
                    }
                }
            }
            for (Task t : byCell) {
                if (t.getDeltaCarry() > 0)
                    t.player.addSummary("troll " + t.unit.getId() + " collected " + t.getDeltaCarry() + " " + Item.WOOD);
                else t.player.addSummary("troll " + t.unit.getId() + " damaged a tree");
            }
        }
    }
}