package view;

import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.entities.Group;
import com.codingame.gameengine.module.entities.SpriteAnimation;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import engine.*;

import java.util.*;
import java.util.stream.Collectors;

public class BirdView {
    private Group birdGroup, catGroup;
    private SpriteAnimation birdAnimation, catAnimation;
    private Board board;
    private Plant tree;
    private Cell cell;
    private Hashtable<String, String[]> birdSprites = new Hashtable<>();
    private Hashtable<String, String[]> catSprites = new Hashtable<>();
    private ArrayList<Unit> lastUnits = new ArrayList<>();
    private double scale;
    private Random random;
    private GraphicEntityModule graphics;
    private boolean isFlying;
    private boolean birdDied;
    private boolean isStanding;

    public BirdView(Board board, Plant plant, Group[] creatorContainers, SpriteAnimation[] creatorAnimations, String[][] creatorsTexts, GraphicEntityModule graphics, TooltipModule tooltips, Random random, double scale) {
        this.board = board;
        this.tree = plant;
        this.cell = plant.getCell();
        this.scale = scale;
        this.random = random;
        this.graphics = graphics;
        this.birdGroup = creatorContainers[2];
        this.catGroup = creatorContainers[3];
        this.birdAnimation = creatorAnimations[2].setScale(scale * 0.5).setLoop(true).setPlaying(true);
        this.catAnimation = creatorAnimations[3].setScale(scale * 1.5).setLoop(true).setPlaying(true);
        tooltips.setTooltipText(birdAnimation, creatorsTexts[2][random.nextInt(creatorsTexts[2].length)]);
        tooltips.setTooltipText(catAnimation, creatorsTexts[3][random.nextInt(creatorsTexts[3].length)]);

        String[] birdSplit = graphics.createSpriteSheetSplitter()
                .setSourceImage("bird.png")
                .setImageCount(11 * 11)
                .setWidth(50)
                .setHeight(34)
                .setOrigRow(0)
                .setOrigCol(0)
                .setImagesPerRow(11)
                .setName("b")
                .split();
        birdSprites.put("death", getSequence(birdSplit, 11 * 0, 9));
        birdSprites.put("fly_left", getSequence(birdSplit, 11 * 1, 9));
        birdSprites.put("fly_right", getSequence(birdSplit, 11 * 2, 9));
        birdSprites.put("idle1", getSequence(birdSplit, 11 * 3, 9));
        birdSprites.put("idle2", getSequence(birdSplit, 11 * 4, 8));
        birdSprites.put("idle3", getSequence(birdSplit, 11 * 5, 8));
        birdSprites.put("idle4", getSequence(birdSplit, 11 * 6, 6));
        birdSprites.put("land_left", getSequence(birdSplit, 11 * 7, 7));
        birdSprites.put("land_right", getSequence(birdSplit, 11 * 8, 7));
        birdSprites.put("takeoff_left", getSequence(birdSplit, 11 * 9, 11));
        birdSprites.put("takeoff_right", getSequence(birdSplit, 11 * 10, 11));

        String[] catSplit = graphics.createSpriteSheetSplitter()
                .setSourceImage("cat.png")
                .setImageCount(10 * 12)
                .setWidth(34)
                .setHeight(34)
                .setOrigRow(0)
                .setOrigCol(0)
                .setImagesPerRow(10)
                .setName("c")
                .split();
        catSprites.put("danger", getSequence(catSplit, 10 * 0, 8));
        catSprites.put("death", getSequence(catSplit, 10 * 1, 9));
        catSprites.put("fight", getSequence(catSplit, 10 * 2, 4));
        catSprites.put("idle1", getSequence(catSplit, 10 * 3, 8));
        catSprites.put("idle2", getSequence(catSplit, 10 * 4, 8));
        catSprites.put("jump", getSequence(catSplit, 10 * 5, 7));
        catSprites.put("sit_left", getSequence(catSplit, 10 * 6, 4));
        catSprites.put("sit_right", getSequence(catSplit, 10 * 7, 4));
        catSprites.put("stand_left", getSequence(catSplit, 10 * 8, 4));
        catSprites.put("stand_right", getSequence(catSplit, 10 * 9, 4));
        catSprites.put("walk_left", getSequence(catSplit, 10 * 10, 8));
        catSprites.put("walk_right", getSequence(catSplit, 10 * 11, 8));

        update();
    }

    private String[] getSequence(String[] images, int start, int count) {
        String[] result = new String[count];
        for (int i = 0; i < count; i++) result[i] = images[start + i];
        return result;
    }

    public void update() {
        if (birdDied) return;
        if (tree.isDead()) {
            int[][] dist = board.getDistances(cell);
            int closest = board.getWidth() * board.getHeight();
            for (Plant plant : board.getPlants()) {
                int tmp = dist[plant.getCell().getX()][plant.getCell().getY()];
                if (tmp < closest && plant.getSize() == Constants.PLANT_MAX_SIZE) {
                    closest = tmp;
                    tree = plant;
                }
            }
        }

        if (cell == tree.getCell()) {
            birdAnimation.setImages(birdSprites.get("idle" + random.nextInt(1, 5)));
            catAnimation.setImages(catSprites.get("idle" + random.nextInt(1, 3)));
            ArrayList<Unit> currentUnits =  board.getUnitsByCell(cell).collect(Collectors.toCollection(ArrayList::new));
            if (currentUnits.stream().anyMatch(u -> u.getPlayer().getIndex() == 0 && !lastUnits.contains(u))) {
                catAnimation.setImages(catSprites.get("danger"));
            } else if (currentUnits.stream().anyMatch(u -> u.getPlayer().getIndex() == 1 && !lastUnits.contains(u))) {
                catAnimation.setImages(catSprites.get("fight"));
            }
            lastUnits = currentUnits;
        } else {
            Cell next = board.getNextCell(cell, tree.getCell(), 1);
            String direction = next.getX() > cell.getX() ? "right" : "left";
            if (!isFlying) {
                isFlying = true;
                birdAnimation.setImages(birdSprites.get("takeoff_" + direction));
            } else if (next == tree.getCell()) {
                isFlying = false;
                birdAnimation.setImages(birdSprites.get("land_" + direction));
            } else birdAnimation.setImages(birdSprites.get("fly_" + direction));

            if (!isStanding) {
                isStanding = true;
                catAnimation.setImages(catSprites.get("stand_" + direction));
            } else if (next == tree.getCell()) {
                isStanding = false;
                catAnimation.setImages(catSprites.get("sit_" + direction));
            } else catAnimation.setImages(catSprites.get("walk_" + direction));
            cell = next;
        }
        if (tree.isDead()) {
            birdAnimation.setImages(birdSprites.get("death")).setLoop(false).reset();
            catAnimation.setImages(catSprites.get("death")[0]).setLoop(false).reset();
            birdDied = true;
        }
        graphics.commitEntityState(1e-3, birdAnimation);
        graphics.commitEntityState(1e-3, catAnimation);
        birdGroup.setX((int) ((cell.getViewX() + 40) * scale)).setY((int) ((cell.getViewY() * scale + 20) + 120));
        catGroup.setX((int) (cell.getViewX() * scale)).setY((int) ((cell.getViewY() * scale + 30) + 120));
    }
}
