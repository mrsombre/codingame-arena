package view;

import com.codingame.gameengine.module.entities.*;

import java.util.ArrayList;
import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedDeque;

// creating sprites later causes re-render, breaking custom animations
public class SpritePool {
    private static Group container;
    private static GraphicEntityModule graphics;
    private static Queue<Sprite> sprites1 = new ConcurrentLinkedDeque<>(); // trees
    private static Queue<Sprite> sprites3 = new ConcurrentLinkedDeque<>(); // moving items to/from trolls
    private static Queue<Rectangle> rectangles1 = new ConcurrentLinkedDeque<>(); // tree HP bars
    private static ArrayList<UnitObject> unitObjects = new ArrayList<>(); // trolls

    private static void initSprites(int size, int zIndex) {
        for (int i = 0; i < size; i++) {
            Sprite sprite = graphics.createSprite().setZIndex(zIndex);
            container.add(sprite);
            if (zIndex == 1) {
                sprites1.add(sprite); // tree
                for (int j = 0; j < 2; j++) { // health bar
                    Rectangle rect = graphics.createRectangle().setVisible(false).setZIndex(zIndex);
                    container.add(rect);
                    rectangles1.add(rect);
                }
            }
            else sprites3.add(sprite);
        }
    }

    public static void initPool(GraphicEntityModule graphics, Group container) {
        SpritePool.container = container;
        SpritePool.graphics = graphics;
        initSprites(150, 1);
        initSprites(50, 3);
        initUnits(new int[]{1, 2, 3, 4, 5, 6, 7}, new int[]{4, 8, 4, 4, 4, 2, 2});
    }

    private static void initUnits(int[] sizes, int[] counts) {
        for (int i = 0; i < sizes.length; i++) {
            for (int j = 0; j < counts[i]; j++) unitObjects.add(new UnitObject(graphics, container, sizes[i]));
        }
    }

    public static Sprite getSprite(int zIndex) {
        if (zIndex == 1) {
            if (sprites1.size() == 0) initSprites(20, 1);
            return sprites1.poll();
        }
        if (zIndex == 3) {
            if (sprites3.size() == 0) initSprites(20, 3);
            return sprites3.poll();
        }
        return null;
    }

    public static Rectangle getHpRectangle() {
        if (rectangles1.size() == 0) initSprites(20, 1);
        return rectangles1.poll();
    }

    public static UnitObject getUnit(int invSize) {
        for (UnitObject obj : unitObjects) {
            if (obj.getInvRects().length >= invSize) {
                unitObjects.remove(obj);
                return obj;
            }
        }
        return new UnitObject(graphics, container, invSize);
    }
}
