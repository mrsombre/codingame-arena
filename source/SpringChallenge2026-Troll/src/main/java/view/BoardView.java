package view;

import com.codingame.gameengine.module.entities.*;
import com.codingame.gameengine.module.toggle.ToggleModule;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import engine.Board;
import engine.Cell;
import engine.Constants;

import java.util.Random;

public class BoardView {
    public static final int SPRITE_SIZE = 72;
    private Group mainGroup;
    private BirdView birdView;

    public BoardView(Board board, GraphicEntityModule graphics, TooltipModule tooltips, ToggleModule toggle, Random random) {
        String groundShort = "y";
        String wallShort = "p";
        String decorShort = "x";
        String decorsWaterShort = "t";

        Group explosionContainer = graphics.createGroup().setZIndex(10).setScale(0.25).setX(random.nextInt(1840)).setY(random.nextInt(120, 1000));
        explosionContainer.add(graphics.createSpriteAnimation().setImages("z0..6").setPlaying(false));
        Group[] creatorContainers = {graphics.createGroup().setZIndex(10), graphics.createGroup().setZIndex(10), graphics.createGroup().setZIndex(10), graphics.createGroup().setZIndex(10), graphics.createGroup().setZIndex(10)};
        SpriteAnimation[] creatorAnimations = {graphics.createSpriteAnimation().setPlaying(false), graphics.createSpriteAnimation().setPlaying(false), graphics.createSpriteAnimation(), graphics.createSpriteAnimation(), graphics.createSpriteAnimation().setPlaying(false)};
        for (int i = 0; i < creatorAnimations.length; i++) creatorContainers[i].add(creatorAnimations[i]);

        Rectangle rect1 = graphics.createRectangle().setAlpha(0).setWidth(120).setHeight(120);
        tooltips.setTooltipText(rect1, "click to copy turn input (Firefox only)");
        Rectangle rect2 = graphics.createRectangle().setAlpha(0).setWidth(120).setHeight(120).setX(graphics.getWorld().getWidth() - 120);
        tooltips.setTooltipText(rect2, "click to copy turn input (Firefox only)");

        graphics.createSprite().setImage("wood.jpg");
        graphics.createSprite().setImage("branding.png").setX(graphics.getWorld().getWidth() / 2).setAnchorX(0.5).setScale(50.0 / 343).setY(70);
        BufferedGroup background = graphics.createBufferedGroup();
        Rectangle darkMode = graphics.createRectangle().setY(120).setWidth(graphics.getWorld().getWidth()).setHeight(graphics.getWorld().getHeight()).setFillColor(0).setAlpha(0.5);
        toggle.displayOnToggleState(darkMode, "darkMode", true);
        Group backgroundTooltip = graphics.createGroup();
        mainGroup = graphics.createGroup();
        double scaleX = ((double) graphics.getWorld().getWidth()) / (SPRITE_SIZE * board.getWidth());
        double scaleY = ((double) graphics.getWorld().getHeight() - 120) / (SPRITE_SIZE * board.getHeight());
        double scale = Math.min(scaleX, scaleY);
        mainGroup.setScale(scale).setY(120);
        background.setScale(scale).setY(120);
        backgroundTooltip.setScale(scale).setY(120);

        boolean[] creatorsPlaced = {false, false, false, false, false}; // frog, fish, bird, cat, turtle
        String[][] creatorsTexts = {{"eulerscheZahl\nLooking for delicious flies", "eulerscheZahl\nis a noob", "eulerscheZahl\nenjoying the weather", "eulerscheZahl\ngetting a sunburn"},
                {"Illedan\nswimming in circles", "Illedan\ndrowning", "Illedan\nfreezing in the water", "Illedan\neating mud", "Illedan\nfeeding the piranhas"},
                {"Illedan\nhopping from tree to tree", "Illedan\ndodging the cat", "Illedan\nhiding in plain sight", "Illedan\nlooking for worms"},
                {"aCat\nchasing Illedan", "aCat\nor is that Astrobytes?", "aCat\nscared of trolls"},
                {"MSz\nhiding"},
        };


        for (int x = 0; x < board.getWidth(); x++) {
            for (int y = 0; y < board.getHeight(); y++) {
                Cell cell = board.getCell(x, y);
                Rectangle tip = graphics.createRectangle().setWidth(SPRITE_SIZE).setHeight(SPRITE_SIZE).setX(cell.getViewX()).setY(cell.getViewY()).setAlpha(0);
                tooltips.setTooltipText(tip, "(" + x + ", " + y + ")" + (cell.getType() == Cell.Type.GRASS ? "" : "\n" + cell.getType()));
                backgroundTooltip.add(tip);
                if (cell.getType() != Cell.Type.WATER) {
                    background.add(graphics.createSprite().setImage(groundShort + (24 + random.nextInt(8))).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2));
                }
                if (cell.getType() == Cell.Type.ROCK) {
                    Sprite sprite = graphics.createSprite().setImage(wallShort + random.nextInt(5)).setX(cell.getViewX() + random.nextInt(SPRITE_SIZE - 66)).setY(cell.getViewY() + random.nextInt(SPRITE_SIZE - 66));
                    background.add(sprite);
                }
                if (cell.getType() == Cell.Type.GRASS) {
                    if (!creatorsPlaced[1] && !creatorsPlaced[2] && cell.getPlant() != null && cell.getPlant().getSize() == Constants.PLANT_MAX_SIZE && random.nextDouble() < 0.0005) {
                        birdView = new BirdView(board, cell.getPlant(), creatorContainers, creatorAnimations, creatorsTexts, graphics, tooltips, random, scale);
                        creatorsPlaced[2] = true;
                        creatorsPlaced[3] = true;
                    } else if (random.nextDouble() < 0.3) {
                        Sprite sprite = graphics.createSprite().setImage(decorShort + random.nextInt(16)).setX(cell.getViewX() + random.nextInt(SPRITE_SIZE - 16)).setY(cell.getViewY() + random.nextInt(SPRITE_SIZE - 16));
                        background.add(sprite);
                    }
                }
                if (cell.getType() == Cell.Type.WATER) {
                    int index = 0;
                    if (cell.getNeighbor(0) != null && cell.getNeighbor(0).getType() != Cell.Type.WATER) index += 4;
                    if (cell.getNeighbor(1) != null && cell.getNeighbor(1).getType() != Cell.Type.WATER) index += 1;
                    if (cell.getNeighbor(2) != null && cell.getNeighbor(2).getType() != Cell.Type.WATER) index += 8;
                    if (cell.getNeighbor(3) != null && cell.getNeighbor(3).getType() != Cell.Type.WATER) index += 2;
                    boolean empty = index == 0;
                    if (index == 0) index = new int[]{0, 20, 21, 22, 23}[random.nextInt(5)];
                    Sprite sprite = graphics.createSprite().setImage(groundShort + index).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2);
                    background.add(sprite);

                    if (cell.getNeighbor(1) != null && cell.getNeighbor(0) != null && cell.getNeighbor(1).getType() == Cell.Type.WATER && cell.getNeighbor(0).getType() == Cell.Type.WATER && cell.getNeighbor(0).getNeighbor(1).getType() != Cell.Type.WATER)
                        background.add(graphics.createSprite().setImage(groundShort + 19).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2));
                    if (cell.getNeighbor(1) != null && cell.getNeighbor(2) != null && cell.getNeighbor(1).getType() == Cell.Type.WATER && cell.getNeighbor(2).getType() == Cell.Type.WATER && cell.getNeighbor(2).getNeighbor(1).getType() != Cell.Type.WATER)
                        background.add(graphics.createSprite().setImage(groundShort + 17).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2));
                    if (cell.getNeighbor(3) != null && cell.getNeighbor(2) != null && cell.getNeighbor(3).getType() == Cell.Type.WATER && cell.getNeighbor(2).getType() == Cell.Type.WATER && cell.getNeighbor(2).getNeighbor(3).getType() != Cell.Type.WATER)
                        background.add(graphics.createSprite().setImage(groundShort + 16).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2));
                    if (cell.getNeighbor(3) != null && cell.getNeighbor(0) != null && cell.getNeighbor(3).getType() == Cell.Type.WATER && cell.getNeighbor(0).getType() == Cell.Type.WATER && cell.getNeighbor(0).getNeighbor(3).getType() != Cell.Type.WATER)
                        background.add(graphics.createSprite().setImage(groundShort + 18).setX(cell.getViewX() - 2).setY(cell.getViewY() - 2));

                    if (empty && random.nextDouble() < 0.3) {
                        if (random.nextDouble() < 0.001) {
                            int creatorIndex = random.nextInt(2); // others on land
                            if (creatorIndex == 1 && creatorsPlaced[2] || creatorsPlaced[creatorIndex])
                                creatorIndex = 1 - creatorIndex;
                            if (!creatorsPlaced[creatorIndex] && (creatorIndex == 0 || !creatorsPlaced[2])) {
                                creatorContainers[creatorIndex].setX((int) (cell.getViewX() * scale)).setY((int) (cell.getViewY() * scale + 120));
                                if (creatorIndex == 0) {
                                    creatorAnimations[0].setImages(graphics.createSpriteSheetSplitter()
                                            .setSourceImage("frog.png")
                                            .setImageCount(3 * 8)
                                            .setWidth(48)
                                            .setHeight(48)
                                            .setOrigRow(0)
                                            .setOrigCol(0)
                                            .setImagesPerRow(8)
                                            .setName("s")
                                            .split());
                                    creatorContainers[creatorIndex].setScale(scale * 1.5);
                                }
                                if (creatorIndex == 1) {
                                    creatorContainers[creatorIndex].setScale(scale * 0.4);
                                    creatorAnimations[1].setImages(graphics.createSpriteSheetSplitter()
                                            .setSourceImage("fish.png")
                                            .setImageCount(3 * 4)
                                            .setWidth(200)
                                            .setHeight(200)
                                            .setOrigRow(0)
                                            .setOrigCol(0)
                                            .setImagesPerRow(4)
                                            .setName("r")
                                            .split());
                                }
                                tooltips.setTooltipText(creatorAnimations[creatorIndex], creatorsTexts[creatorIndex][random.nextInt(creatorsTexts[creatorIndex].length)]);
                                creatorsPlaced[creatorIndex] = true;
                            }
                        } else {
                            sprite = graphics.createSprite().setImage(decorsWaterShort + random.nextInt(4)).setX(cell.getViewX()).setY(cell.getViewY()).setScale(SPRITE_SIZE / 375.0);
                            background.add(sprite);
                        }
                    }
                    if (index == 8 && !creatorsPlaced[4] && random.nextDouble() < 0.0005) {
                        creatorAnimations[4].setImages(graphics.createSpriteSheetSplitter()
                                .setSourceImage("turtle.png")
                                .setImageCount(6 * 4 + 1)
                                .setWidth(50)
                                .setHeight(96)
                                .setOrigRow(0)
                                .setOrigCol(0)
                                .setImagesPerRow(6)
                                .setName("q")
                                .split());
                        creatorContainers[4].setX((int) (cell.getViewX() * scale + 10)).setY((int) (cell.getViewY() * scale + 120)).setScale(scale);
                        tooltips.setTooltipText(creatorAnimations[4], creatorsTexts[4][random.nextInt(creatorsTexts[4].length)]);
                        creatorsPlaced[4] = true;
                    }
                }
                if (cell.getType() == Cell.Type.IRON) {
                    Sprite sprite = graphics.createSprite().setImage("iron.png").setX(cell.getViewX()).setY(cell.getViewY());
                    mainGroup.add(sprite);
                }
                if (cell.getType() == Cell.Type.SHACK) {
                    Sprite sprite = graphics.createSprite().setImage("shack.png").setX(cell.getViewX() + 5).setY(cell.getViewY() + 15);
                    sprite.setTint(board.getPlayers().stream().filter(p -> p.getShack() == cell).findFirst().get().getColor());
                    mainGroup.add(sprite);
                }
                if (cell.getType() != Cell.Type.WATER && cell.isNearWater()) {
                    Rectangle nearWater = graphics.createRectangle().setFillColor(0x0000ff).setAlpha(0.05).setX(cell.getViewX()).setY(cell.getViewY()).setWidth(SPRITE_SIZE).setHeight(SPRITE_SIZE);
                    mainGroup.add(nearWater);
                    toggle.displayOnToggleState(nearWater, "debug", true);
                }
            }
        }
        for (int i = 0; i < creatorAnimations.length; i++) {
            if (!creatorsPlaced[i]) creatorAnimations[i].setVisible(false);
        }
        for (int x = 0; x <= board.getWidth(); x++) {
            Line line = graphics.createLine().setY(0).setY2(board.getHeight() * SPRITE_SIZE).setX(x * SPRITE_SIZE).setX2(x * SPRITE_SIZE).setLineColor(0).setLineWidth(2).setAlpha(0.1);
            background.add(line);
        }
        for (int y = 0; y <= board.getHeight(); y++) {
            Line line = graphics.createLine().setX(0).setX2(board.getWidth() * SPRITE_SIZE).setY(y * SPRITE_SIZE).setY2(y * SPRITE_SIZE).setLineColor(0).setLineWidth(2).setAlpha(0.1);
            background.add(line);
        }

        SpritePool.initPool(graphics, mainGroup);
    }

    public void update() {
        if (birdView != null) birdView.update();
    }

    public Group getMainGroup() {
        return mainGroup;
    }
}
