package view;

import com.codingame.gameengine.module.entities.*;

public class UnitObject {
    private Group group;
    private SpriteAnimation sprite;
    private Sprite[] invSprites;
    private Rectangle[] invRects;
    private Rectangle tooltipArea;

    public Rectangle[] getInvRects() {
        return invRects;
    }

    public Group getGroup() {
        return group;
    }

    public SpriteAnimation getSprite() {
        return sprite;
    }

    public Rectangle getTooltipArea() {
        return tooltipArea;
    }

    public Sprite[] getInvSprites() {
        return invSprites;
    }

    public UnitObject(GraphicEntityModule graphics, Group mainGroup, int invSize) {
        group = graphics.createGroup().setScale(0.4).setZIndex(2).setAlpha(0);
        mainGroup.add(group);
        sprite = graphics.createSpriteAnimation()
                .setAnchor(0.5)
                .setLoop(true)
                .setPlaying(true);
        int invRectSize = 28;
        tooltipArea = graphics.createRectangle().setWidth(invRectSize * 3).setHeight(invRectSize * 5).setAlpha(0);
        tooltipArea.setX(-tooltipArea.getWidth() / 2).setY(-tooltipArea.getHeight() / 2);
        group.add(sprite, tooltipArea);
        invSprites = new Sprite[invSize];
        invRects = new Rectangle[invSize];
        for (int i = 0; i < invSize; i++) {
            int x = invRectSize * (i % 3) - invRectSize * 3 / 2;
            int y = 36 - invRectSize * (i / 3);
            invRects[i] = graphics.createRectangle().setLineWidth(2).setX(x).setY(y).setWidth(invRectSize).setHeight(invRectSize).setVisible(false);
            invSprites[i] = graphics.createSprite().setX(x).setY(y).setScale(0.5).setVisible(false);
            group.add(invRects[i], invSprites[i]);
        }
    }
}
