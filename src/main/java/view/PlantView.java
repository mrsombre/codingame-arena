package view;

import com.codingame.CompressionModule;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.entities.Rectangle;
import com.codingame.gameengine.module.entities.Sprite;
import com.codingame.gameengine.module.toggle.ToggleModule;
import engine.Constants;
import engine.Plant;

public class PlantView {
    private Plant plant;
    private Sprite sprite;
    private Rectangle hpBar;
    private Rectangle hpBackground;
    private CompressionModule compressionModule;

    public PlantView(Plant plant, ToggleModule toggleModule, CompressionModule compressionModule) {
        this.compressionModule = compressionModule;
        this.plant = plant;
        sprite = SpritePool.getSprite(1).setX(plant.getCell().getViewX() + BoardView.SPRITE_SIZE / 2).setY(plant.getCell().getViewY() + BoardView.SPRITE_SIZE - 12).setAnchorX(0.5).setAnchorY(0.85);
        int barX = plant.getCell().getViewX() + 10;
        int barY = plant.getCell().getViewY() + 10;
        int barWidth = BoardView.SPRITE_SIZE - 20;
        int barHeight = 6;
        hpBackground = SpritePool.getHpRectangle().setX(barX).setY(barY).setWidth(barWidth).setHeight(barHeight).setLineWidth(2).setFillColor(0x111111).setLineColor(0x111111);
        hpBar = SpritePool.getHpRectangle().setX(barX).setY(barY).setWidth(barWidth).setHeight(barHeight).setFillColor(0x80ff80);
        toggleModule.displayOnToggleState(hpBar, "debug", true);
        toggleModule.displayOnToggleState(hpBackground, "debug", true);
        update();
    }

    public void update() {
        sprite.setImage("v" + Math.max(0, 7 * plant.getType().ordinal() + plant.getSize() + plant.getResources() - 1));
        int fullHp = Constants.PLANT_FINAL_HEALTH[plant.getType().ordinal()] - Constants.PLANT_DELTA_HEALTH[plant.getType().ordinal()] * (Constants.PLANT_MAX_SIZE - plant.getSize());
        hpBar.setWidth((BoardView.SPRITE_SIZE - 20) * plant.getHealth() / fullHp);
        hpBar.setVisible(plant.getHealth() < fullHp && !plant.isDead());
        hpBackground.setVisible(plant.getHealth() < fullHp && !plant.isDead());
        compressionModule.addPlant(plant, sprite.getId());
        if (plant.isDead()) sprite.setAlpha(0).setRotation(45);
    }
}
