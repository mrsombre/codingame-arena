package engine;

import com.codingame.CompressionModule;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.toggle.ToggleModule;
import view.PlantView;

public class Plant {
    private int size;
    private int health;
    private int resources;
    private int cooldown;
    private Cell cell;
    private Item type;
    private PlantView view;

    public Plant(Cell cell, Item type) {
        this.cell = cell;
        this.type = type;
        this.health = Constants.PLANT_FINAL_HEALTH[type.ordinal()] - Constants.PLANT_DELTA_HEALTH[type.ordinal()] * Constants.PLANT_MAX_SIZE;
    }

    public Plant(Plant plant) {
        this.cell = plant.cell;
        this.type = plant.type;
        this.size = plant.size;
        this.health = plant.health;
        this.resources = plant.resources;
        this.cooldown = plant.cooldown;
    }

    public void tick(boolean updateHealth) {
        if (cooldown > 0) cooldown--;
        if (cooldown == 0 && health > 0) {
            if (size < Constants.PLANT_MAX_SIZE) {
                size++;
                if (updateHealth) health += Constants.PLANT_DELTA_HEALTH[type.ordinal()];
                cooldown = getGrowthCooldown();
            } else if (resources < Constants.PLANT_MAX_RESOURCES) {
                resources++;
                cooldown = getGrowthCooldown();
            }
        }
        if (view != null) view.update();
    }

    public int getGrowthCooldown() {
        int cooldown = Constants.PLANT_COOLDOWN[type.ordinal()];
        if (cell.isNearWater()) cooldown -= Constants.PLANT_WATER_COOLDOWN_BOOST[type.ordinal()];
        return cooldown;
    }

    public Item getType() {
        return type;
    }

    public Cell getCell() {
        return cell;
    }

    public int getResources() {
        return resources;
    }

    public int getHealth() {
        return health;
    }

    public void harvest() {
        if (resources > 0) resources--;
    }

    public void damage(int damage) {
        health = Math.max(health - damage, 0);
        if (isDead()) cell.setPlant(null);
    }

    public int getSize() {
        return size;
    }

    public boolean isDead() {
        return health <= 0;
    }

    public void initView(ToggleModule toggleModule, CompressionModule compressionModule) {
        if (view == null) view = new PlantView(this, toggleModule, compressionModule);
    }

    public String getInputLine() {
        return type + " " + cell.getX() + " " + cell.getY() + " " + size + " " + health + " " + resources + " " + cooldown;
    }

    public String serializeDelta(Plant expected, String alphabet) {
        if (expected == null)
            return " P " +
                    alphabet.charAt(this.cell.getX()) +
                    alphabet.charAt(this.cell.getY()) +
                    alphabet.charAt(this.type.ordinal()) +
                    alphabet.charAt(this.size + this.resources) +
                    alphabet.charAt(this.cooldown) +
                    alphabet.charAt(this.health) +
                    alphabet.charAt(this.getGrowthCooldown());

        String diff = "";
        if (this.health != expected.health) diff += " h" + alphabet.charAt(this.health);
        if (this.resources != expected.resources) diff += " s" + alphabet.charAt(this.size + this.resources);
        if (this.cooldown != expected.cooldown) diff += " c" + alphabet.charAt(this.cooldown);
        if (diff.length() == 0) return null;
        return diff;
    }
}
