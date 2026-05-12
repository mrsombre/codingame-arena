package engine;

import com.codingame.CompressionModule;
import com.codingame.game.Player;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.entities.Group;
import com.codingame.gameengine.module.toggle.ToggleModule;
import engine.task.Task;
import view.UnitView;

public class Unit {
    private int id;
    private Player player;
    private Cell cell;
    private int movementSpeed;
    private int carryCapacity;
    private int harvestPower;
    private int chopPower;
    private Inventory inventory = new Inventory();
    private UnitView view;

    public static int idCounter;

    public static int[] getTrainingCosts(Player player, int[] talents, int league) {
        int baseCost = player.getUnits().size();
        int[] result = new int[Item.values().length];
        for (int i = 0; i <= Item.APPLE.ordinal(); i++) result[i] = baseCost + talents[i] * talents[i];
        if (league >= 3) result[Item.IRON.ordinal()] = baseCost + talents[3] * talents[3];
        return result;
    }

    public static boolean canTrain(Player player, int[] talents, int league) {
        int[] costs = getTrainingCosts(player, talents, league);
        for (int i = 0; i < costs.length; i++) {
            if (costs[i] > player.getInventory().getItemCount(i)) return false;
        }
        return true;
    }

    public int getId() {
        return id;
    }

    public Player getPlayer() {
        return player;
    }

    public Cell getCell() {
        return cell;
    }

    public void setCell(Cell cell) {
        this.cell = cell;
    }

    public Unit(Player player, int[] talents, int league) {
        id = idCounter++;
        int[] costs = getTrainingCosts(player, talents, league);
        for (int i = 0; i < costs.length; i++) {
            player.getInventory().setItem(i, player.getInventory().getItemCount(i) - costs[i]);
        }
        this.player = player;
        player.AddUnit(this);
        this.cell = player.getShack();
        movementSpeed = talents[0];
        carryCapacity = talents[1];
        harvestPower = talents[2];
        chopPower = talents[3];
    }

    public Unit(Unit unit) {
        this.id = unit.id;
        this.player = unit.player;
        this.cell = unit.cell;
        this.inventory = new Inventory(unit.inventory);
        this.movementSpeed = unit.movementSpeed;
        this.carryCapacity = unit.carryCapacity;
        this.harvestPower = unit.harvestPower;
        this.chopPower = unit.chopPower;
    }

    public void initView(GraphicEntityModule graphicEntityModule, ToggleModule toggleModule, CompressionModule compressionModule, Group group) {
        if (view == null) view = new UnitView(this, graphicEntityModule, toggleModule, compressionModule, group);
    }

    public void updateView() {
        view.update();
    }

    public void harvest(int power) {
        if (power > harvestPower || inventory.getTotal() >= carryCapacity) return;
        inventory.incrementItem(cell.getPlant().getType());
        cell.getPlant().harvest();
    }

    public void mine() {
        for (int i = 0; i < chopPower && inventory.getTotal() < carryCapacity; i++) {
            inventory.incrementItem(Item.IRON);
        }
    }

    public int getChopPower() {
        return chopPower;
    }

    public int getHarvestPower() {
        return harvestPower;
    }

    public int getMovementSpeed() {
        return movementSpeed;
    }

    public int getCarryCapacity() {
        return carryCapacity;
    }

    public int getFreeCarryCapacity() {
        return carryCapacity - inventory.getTotal();
    }

    public Inventory getInventory() {
        return inventory;
    }

    public boolean isNearShack() {
        return cell.isNearShack(player);
    }

    public String getInputLine(int playerId, int playersCount) {
        int outputId = (playerId + player.getIndex()) % playersCount;
        return id + " " + outputId + " " + cell.getX() + " " + cell.getY() + " " + movementSpeed + " " + carryCapacity + " " + harvestPower + " " + chopPower + " " + inventory.getInputLine();
    }

    @Override
    public String toString() {
        return id + ": " + cell;
    }

    public void setAnimateTask(Task task) {
        view.setAnimateTask(task);
    }

    public String serializeDelta(Unit expected, String alphabet) {
        if (expected == null)
            return " W " +
                    alphabet.charAt(this.id) +
                    alphabet.charAt(this.cell.getX()) +
                    alphabet.charAt(this.cell.getY()) +
                    alphabet.charAt(this.player.getIndex()) +
                    alphabet.charAt(this.movementSpeed) +
                    alphabet.charAt(this.carryCapacity) +
                    alphabet.charAt(this.harvestPower) +
                    alphabet.charAt(this.chopPower);

        String diff = "";
        if (this.cell.getX() != expected.cell.getX()) diff += " x" + alphabet.charAt(this.cell.getX());
        if (this.cell.getY() != expected.cell.getY()) diff += " y" + alphabet.charAt(this.cell.getY());
        for (int i = 0; i < this.inventory.getItemsLength(); i++) {
            if (this.inventory.getItemCount(i) != expected.inventory.getItemCount(i))
                diff += " " + i + alphabet.charAt(this.inventory.getItemCount(i));
        }
        if (diff.length() == 0) return null;
        return diff;
    }
}
