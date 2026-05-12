package view;

import com.codingame.CompressionModule;
import com.codingame.gameengine.module.entities.*;
import com.codingame.gameengine.module.entities.Rectangle;
import com.codingame.gameengine.module.toggle.ToggleModule;
import engine.Cell;
import engine.Inventory;
import engine.Unit;
import engine.task.*;

import java.awt.*;
import java.util.*;
import java.util.stream.Collectors;

public class UnitView {
    private Unit unit;

    private SpriteAnimation sprite;
    private Rectangle tooltipArea;
    private GraphicEntityModule graphics;
    private CompressionModule compressionModule;
    private Task task;
    private String lastDir = "Front_";
    private Cell lastPos;
    private Inventory lastInventory = new Inventory();
    private ArrayList<Sprite> fruits = new ArrayList<>();
    private static Hashtable<String, String[]> spriteSheets;
    private Group mainGroup, trollGroup;
    private Sprite[] inventorySprites;
    private boolean didWalk = false;

    public UnitView(Unit unit, GraphicEntityModule graphics, ToggleModule toggleModule, CompressionModule compressionModule, Group group) {
        this.graphics = graphics;
        this.compressionModule = compressionModule;
        this.mainGroup = group;
        if (spriteSheets == null) {
            spriteSheets = new Hashtable<>();
            String[] colors = {"blue", "red"};
            String[] directions = {"Back", "Front", "Left", "Right"};
            for (String color : colors) {
                for (String dir : directions) {
                    String key = color + "_" + dir + "_";
                    parseSheet(key + "Attacking", 10);
                    parseSheet(key + "Drop", 16);
                    parseSheet(key + "Hurt", 10);
                    parseSheet(key + "Idle", 16);
                    parseSheet(key + "Walking", 20);
                }
            }
            parseSheet("blue_Left_Harvest", 16);
            parseSheet("red_Right_Harvest", 16);
        }
        this.unit = unit;

        UnitObject unitObject = SpritePool.getUnit(unit.getCarryCapacity());
        trollGroup = unitObject.getGroup();
        sprite = unitObject.getSprite();
        tooltipArea = unitObject.getTooltipArea();
        inventorySprites = unitObject.getInvSprites();
        for (int i = 0; i < unit.getCarryCapacity(); i++) {
            Rectangle rect = unitObject.getInvRects()[i].setFillColor(unit.getPlayer().getDarkColor()).setLineColor(unit.getPlayer().getLightColor()).setVisible(true);
            rect.setX(rect.getX() + (unit.getPlayer().getIndex() == 0 ? 5 : -5));
            Sprite invSprite = inventorySprites[i].setX(rect.getX());
            graphics.commitEntityState(0, rect);
            toggleModule.displayOnToggleState(rect, "debug", true);
            toggleModule.displayOnToggleState(invSprite, "debug", true);
        }
        update();
        graphics.commitEntityState(0, trollGroup);
        trollGroup.setAlpha(1);
    }

    private int spriteOffset = 0;
    private void parseSheet(String key, int count) {
        String[] images = new String[count];
        for (int i = 1; i <= count; i++) images[i - 1] = "w" + (spriteOffset + i);
        spriteSheets.put(key, images);
        spriteOffset += 20;
    }

    public void setAnimateTask(Task task) {
        this.task = task;
    }

    private String getDirectionString(Cell from, Cell to) {
        if (to == null) return lastDir;
        if (to.getX() > from.getX()) return "Right_";
        if (to.getX() < from.getX()) return "Left_";
        if (to.getY() < from.getY()) return "Back_";
        if (to.getY() > from.getY()) return "Front_";
        return lastDir;
    }

    private void animateMoveFruit(Point from, Point to, int iteration, int totalDelta) {
        Sprite fruit = null;
        if (fruits.size() <= iteration) {
            fruit = SpritePool.getSprite(3).setScale(2);
            fruits.add(fruit);
            graphics.commitEntityState(1e-3, mainGroup);
        }

        int item = -1;
        int inventoryIndex = -1;
        boolean gain = unit.getInventory().getTotal() > lastInventory.getTotal();
        if (gain) { // harvest/chop/mine/pick
            for (int i = unit.getInventory().getItemsLength() - 1; i >= 0; i--) {
                if (unit.getInventory().getItemCount(i) != lastInventory.getItemCount(i)) item = i;
            }
            for (int i = inventorySprites.length - 1; i >= 0; i--) {
                if (!inventorySprites[i].isVisible()) inventoryIndex = i;
            }
        } else { // drop/plant
            for (int i = inventorySprites.length - 1; i >= 0; i--) {
                if (!inventorySprites[i].isVisible()) continue;
                item = Arrays.stream(PlayerView.inventorySprites).collect(Collectors.toCollection(ArrayList::new)).indexOf(inventorySprites[i].getImage());
                if (unit.getInventory().getItemCount(item) != lastInventory.getItemCount(item)) {
                    inventoryIndex = i;
                    break;
                }
            }
        }

        fruit = fruits.get(iteration);
        fruit.setImage(PlayerView.inventorySprites[item]).setX((int)from.getX()).setY((int)from.getY()).setAlpha(0).setScale(0.33).setAnchor(0.5);
        double offset = 0.5 * iteration / totalDelta;
        if (!gain) offset = 0.5 * (totalDelta - 1 - iteration) / totalDelta;
        graphics.commitEntityState(0.5 - offset, fruit);
        fruit.setAlpha(1);
        graphics.commitEntityState(0.6 - offset, fruit);
        fruit.setX((int)to.getX()).setY((int)to.getY());
        graphics.commitEntityState(0.9 - offset, fruit);
        fruit.setAlpha(0);
        graphics.commitEntityState(1 - offset, fruit);

        if (gain) {
            inventorySprites[inventoryIndex].setImage(PlayerView.inventorySprites[item]).setVisible(true);
            graphics.commitEntityState(1 - offset,  inventorySprites[inventoryIndex]);
            lastInventory.setItem(item, lastInventory.getItemCount(item) + 1);
        } else {
            inventorySprites[inventoryIndex].setVisible(false);
            graphics.commitEntityState(0.5 - offset,  inventorySprites[inventoryIndex]);
            lastInventory.setItem(item, lastInventory.getItemCount(item) - 1);
        }
    }

    private Point getPlayerCenter() {
        return new Point(
                unit.getCell().getViewX() + BoardView.SPRITE_SIZE / 2 + (unit.getPlayer().getIndex() == 0 ? -20 : 20),
                unit.getCell().getViewY() + BoardView.SPRITE_SIZE * 3 / 5);
    }

    private Point getCellCenter(Cell cell) {
        return new Point(
                cell.getViewX() + BoardView.SPRITE_SIZE / 2,
                cell.getViewY() + BoardView.SPRITE_SIZE / 2);
    }

    private Point getTreeTop() {
        return new Point(
                unit.getCell().getViewX() + BoardView.SPRITE_SIZE / 2,
                unit.getCell().getViewY() + BoardView.SPRITE_SIZE * 1 / 3);
    }

    private Point getTreeBottom() {
        return new Point(
                unit.getCell().getViewX() + BoardView.SPRITE_SIZE / 2,
                unit.getCell().getViewY() + BoardView.SPRITE_SIZE * 3 / 4);
    }

    private void harvestFruit(int iteration, int totalDelta) {
        animateMoveFruit(getTreeTop(), getPlayerCenter(), iteration, totalDelta);
    }

    private void mineIron(int iteration, int totalDelta) {
        Cell iron = Arrays.stream(unit.getCell().getNeighbors()).filter(c -> c != null && c.getType() == Cell.Type.IRON).findFirst().get();
        animateMoveFruit(getCellCenter(iron), getPlayerCenter(), iteration, totalDelta);
    }

    private void pickItem(int iteration, int totalDelta) {
        Cell shack = unit.getPlayer().getShack();
        animateMoveFruit(getCellCenter(shack), getPlayerCenter(), iteration, totalDelta);
    }

    private void dropItem(int iteration, int totalDelta) {
        Cell shack = unit.getPlayer().getShack();
        animateMoveFruit(getPlayerCenter(), getCellCenter(shack), iteration, totalDelta);
    }

    private void plantTree(int iteration, int totalDelta) {
        animateMoveFruit(getPlayerCenter(), getTreeBottom(), iteration, totalDelta);
    }

    private void chopTree(int iteration, int totalDelta) {
        animateMoveFruit(getTreeBottom(), getPlayerCenter(), iteration, totalDelta);
    }

    public void update() {
        String key = unit.getPlayer().getIndex() == 0 ? "red_" : "blue_";

        if (task != null && (task instanceof ChopTask || task instanceof HarvestTask || task instanceof PlantTask))
            lastDir = unit.getPlayer().getIndex() == 0 ? "Right_" : "Left_";
        if (task != null && task instanceof MineTask) {
            Optional<Cell> iron = Arrays.stream(unit.getCell().getNeighbors()).filter(c -> c != null && c.getType() == Cell.Type.IRON).findFirst();
            if (iron.isPresent()) lastDir = getDirectionString(unit.getCell(), iron.get());
        }
        if (task != null && (task instanceof PickTask || task instanceof DropTask)) {
            lastDir = getDirectionString(unit.getCell(), unit.getPlayer().getShack());
        }
        if (task != null && task instanceof MoveTask) {
            if (unit.getCell() == lastPos) lastDir = getDirectionString(unit.getCell(), ((MoveTask) task).getTarget());
            else lastDir = getDirectionString(lastPos, unit.getCell());
        }
        key += lastDir;

        if (task != null && !task.wasApplied()) key += "Hurt";
        else if (task != null && (task instanceof ChopTask || task instanceof MineTask)) key += "Attacking";
        else if (task != null && (task instanceof DropTask || task instanceof PlantTask || task instanceof PickTask))
            key += "Drop";
        else if (task != null && task instanceof HarvestTask) key += "Harvest";
        else if (task != null && task instanceof MoveTask && unit.getCell() != lastPos) key += "Walking";
        else key += "Idle";

        boolean walking = key.contains("Walk");
        int duration = walking ? 1000 / 3 : 1000;
        boolean changed = sprite.getImages() != spriteSheets.get(key) || sprite.getDuration() != duration;
        sprite.setImages(spriteSheets.get(key));
        sprite.setDuration(duration);
        if (didWalk) sprite.reset(); // walk duration is not a divisor of frame duration
        if (didWalk || changed) graphics.commitEntityState(1e-3, sprite);
        didWalk = walking;
        trollGroup.setX(40 * unit.getPlayer().getIndex() - 20 + unit.getCell().getViewX() + BoardView.SPRITE_SIZE / 2)
                .setY(unit.getCell().getViewY() + BoardView.SPRITE_SIZE / 2 + 10);
        compressionModule.addUnit(unit, tooltipArea.getId());

        int delta = Math.abs(unit.getInventory().getTotal() - lastInventory.getTotal());
        for (int m = 0; m < delta; m++) {
            if (task instanceof HarvestTask) harvestFruit(m, delta);
            if (task instanceof MineTask) mineIron(m, delta);
            if (task instanceof DropTask) dropItem(m, delta);
            if (task instanceof PickTask) pickItem(m, delta);
            if (task instanceof PlantTask) plantTree(m, delta);
            if (task instanceof ChopTask) chopTree(m, delta);
        }
        task = null;
        lastPos = unit.getCell();
        lastInventory = new Inventory(unit.getInventory());
    }
}
