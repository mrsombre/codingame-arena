package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Constants;
import engine.Item;
import engine.Unit;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class TrainTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>TRAIN)\\s+(?<movementSpeed>\\d+)\\s+(?<carryCapacity>\\d+)\\s+(?<harvestPower>\\d+)\\s+(?<chopPower>\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    private int[] talents;
    private int league;

    public TrainTask(Player player, Board board, String command, int league) {
        super(player, board);
        this.league = league;
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        int movementSpeed = Integer.parseInt(matcher.group("movementSpeed"));
        int carryCapacity = Integer.parseInt(matcher.group("carryCapacity"));
        int harvestPower = Integer.parseInt(matcher.group("harvestPower"));
        int chopPower = Integer.parseInt(matcher.group("chopPower"));
        talents = new int[]{movementSpeed, carryCapacity, harvestPower, chopPower};
        if (movementSpeed < 1 || movementSpeed > board.getWidth() * board.getHeight())
            addParsingError("invalid movement speed: " + movementSpeed, InputError.INVALID_SKILL, false);
        if (carryCapacity < 0 || carryCapacity > 1000)
            addParsingError("invalid carry capacity: " + carryCapacity, InputError.INVALID_SKILL, false);
        if (harvestPower < 0 || harvestPower > Constants.PLANT_MAX_RESOURCES)
            addParsingError("invalid harvest power: " + harvestPower, InputError.INVALID_SKILL, false);
        if (league < 3 && chopPower > 0)
            addParsingError("chop power is not available in this league", InputError.NOT_AVAILABLE, false);
        if (chopPower < 0 || chopPower > Arrays.stream(Constants.PLANT_FINAL_HEALTH).max().getAsInt())
            addParsingError("invalid chop power: " + chopPower, InputError.INVALID_SKILL, false);
        if (!Unit.canTrain(player, talents, league))
            addCantAfford();
    }

    private void addCantAfford() {
        int[] costs = Unit.getTrainingCosts(player, talents, league);
        addParsingError("can't afford unit training, costs: " + costs[0] + " " + Item.PLUM + ", " + costs[1] + " " + Item.LEMON + ", " + costs[2] + " " + Item.APPLE + ", " + costs[Item.IRON.ordinal()] + " " + Item.IRON, InputError.CANT_AFFORD, false);
    }

    @Override
    public int getTaskPriority() {
        return 6;
    }

    @Override
    public int getRequiredLeague() {
        return 2;
    }

    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (!Unit.canTrain(player, talents, league)) {
            addCantAfford();
            return;
        }
        if (board.getUnitsByCell(player.getShack()).count() > 0) {
            addParsingError("can't train unit, cell blocked", InputError.MOVE_BLOCKED, false);
            return;
        }
        Unit unit = new Unit(player, talents, league);
        board.addUnit(unit);
        applied = true;
        player.addSummary("trained a troll");
    }
}