package view;

import com.codingame.game.Player;
import com.codingame.gameengine.module.entities.*;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import engine.Item;

import java.util.ArrayList;

public class PlayerView {
    private Player player;
    private Text messageText;
    private Text scoreText;
    private Sprite avatar;
    private GraphicEntityModule graphics;
    private TooltipModule tooltips;
    public static final String[] inventorySprites = {"u0", "u1", "u2", "u3", "u4", "u5"};
    private ArrayList<Text> inventoryTexts = new ArrayList<>();

    public PlayerView(Player player, GraphicEntityModule graphics, TooltipModule tooltips) {
        this.graphics = graphics;
        this.tooltips = tooltips;
        this.player = player;
        int avatarX = player.getIndex() == 0 ? 10 : graphics.getWorld().getWidth() - 100 - 10;
        graphics.createRoundedRectangle().setRadius(5).setLineColor(0x505050).setLineWidth(8).setX(avatarX - 4).setY(6).setHeight(108).setWidth(108).setFillColor(player.getDarkColor());
        String image = player.getAvatarToken();
        avatar = graphics.createSprite().setImage(image).setBaseHeight(100).setBaseWidth(100).setY(10).setX(avatarX);
        graphics.createRoundedRectangle().setRadius(5).setLineColor(0x505050).setLineWidth(8).setX(avatarX - 4).setY(6).setHeight(108).setWidth(108).setFillAlpha(0);

        int nickX = player.getIndex() == 0 ? 130 : graphics.getWorld().getWidth() - 430 - 130;
        graphics.createRoundedRectangle().setRadius(3).setLineColor(0x505050).setLineWidth(5).setX(nickX).setY(5).setHeight(60).setWidth(430).setFillColor(player.getDarkColor());
        Text nickname = graphics.createText().setText(player.getNicknameToken()).setX(nickX + 5).setY(10).setFillColor(player.getLightColor()).setFontSize(40).setStrokeThickness(2).setMaxWidth(420);
        graphics.createRoundedRectangle().setRadius(3).setLineColor(0x505050).setLineWidth(5).setX(nickX).setY(75).setHeight(40).setWidth(430).setFillColor(player.getDarkColor());
        messageText = graphics.createText("").setX(nickX + 5).setY(80).setFillColor(player.getLightColor()).setFontSize(25).setStrokeThickness(2).setMaxWidth(420);
        for (int i = 0; i < inventorySprites.length; i++) {
            int baseX = 570 + 100 * (i % 3);
            if (player.getIndex() == 1) baseX = graphics.getWorld().getWidth() - 860 + 100 * (i % 3);
            int baseY = 60 * (i / 3) + 5;
            graphics.createRoundedRectangle().setRadius(3).setLineColor(0x505050).setLineWidth(5).setX(baseX).setY(baseY).setHeight(50).setWidth(90).setFillColor(player.getDarkColor());
            Sprite invSprite = graphics.createSprite().setImage(inventorySprites[i]).setScale(5.0 / 6).setX(baseX).setY(baseY);
            Rectangle rectangle = graphics.createRectangle().setX(baseX).setY(baseY).setWidth(50).setHeight(50).setAlpha(0);
            tooltips.setTooltipText(rectangle, Item.values()[i].toString());
            Text invText = graphics.createText().setX(baseX + 50).setY(baseY + 10).setFillColor(player.getLightColor()).setFontSize(25).setStrokeThickness(2);
            inventoryTexts.add(invText);
        }
        int scoreX = player.getIndex() == 0 ? 870 : graphics.getWorld().getWidth() - 870 - 80;
        graphics.createRoundedRectangle().setRadius(3).setLineColor(0x505050).setLineWidth(5).setX(scoreX).setY(5).setHeight(50).setWidth(80).setFillColor(player.getDarkColor());
        scoreText = graphics.createText().setX(scoreX + 40).setY(12).setFontSize(30).setStrokeThickness(2).setFillColor(player.getLightColor()).setAnchorX(0.5);
        update(0);
    }

    private int lastUnitCount = 1;
    private String lastMessage = "";

    public void update(int turn) {
        messageText.setText(player.getMessage());
        if (!player.getMessage().equals(lastMessage)) graphics.commitEntityState(1e-3, messageText);
        lastMessage = player.getMessage();
        scoreText.setText(String.valueOf(player.getScore()));
        for (int i = 0; i < inventoryTexts.size(); i++) {
            inventoryTexts.get(i).setText(String.valueOf(player.getInventory().getItemCount(i)));
            if (lastUnitCount != player.getUnits().size())
                graphics.commitEntityState(1e-3, inventoryTexts.get(i), scoreText);
        }
        lastUnitCount = player.getUnits().size();
        if (player.hasTimelimitExceededLastTurn()) {
            int x = avatar.getX() + 50 * (player.getTimelimitsExceeded() - 1);
            Sprite stopwatch = graphics.createSprite().setImage("stopwatch.png").setX(x).setY(60).setScale(0.2);
            if (player.getTimelimitsExceeded() == 3) stopwatch.setX(avatar.getX()).setY(avatar.getY()).setScale(0.4);
            tooltips.setTooltipText(stopwatch, "Time limit exceeded in turn " + turn + ": " + player.getLastExectionTimeMs() + " ms");
        }
    }
}

