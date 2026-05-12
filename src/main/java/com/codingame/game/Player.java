package com.codingame.game;
import com.codingame.gameengine.core.AbstractMultiplayerPlayer;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import engine.*;
import engine.task.InputError;
import view.PlayerView;

import java.util.ArrayList;
import java.util.stream.Collectors;

public class Player extends AbstractMultiplayerPlayer {
    private ArrayList<Unit> units;
    private Inventory inventory;
    private Cell shack;
    private String message = "";
    private ArrayList<InputError> errors = new ArrayList<>();
    private ArrayList<String> summaries = new ArrayList<>();
    private PlayerView view;

    public void init(Cell shack, int league) {
        this.units = new ArrayList<>();
        this.inventory = new Inventory();
        this.shack = shack;
        Unit unit = new Unit(this, new int[]{1, 1, 1, league >= 3 ? 1 : 0}, league);
    }

    public void setInventory(int[] inventory) {
        for (int i = 0; i < inventory.length; i++) this.inventory.setItem(i, inventory[i]);
    }

    public ArrayList<Unit> getUnits() {
        return units;
    }

    public void AddUnit(Unit unit) {
        units.add(unit);
    }

    public Cell getShack() {
        return shack;
    }

    public Inventory getInventory() {
        return inventory;
    }

    public void recomputeScore() {
        if (isActive())
            setScore(inventory.getItemCount(Item.PLUM) +
                    inventory.getItemCount(Item.LEMON) +
                    inventory.getItemCount(Item.APPLE) +
                    inventory.getItemCount(Item.BANANA) +
                    Constants.WOOD_POINTS * inventory.getItemCount(Item.WOOD));
    }

    public int getColor() {
        return getIndex() == 0 ? 0xff8080 : 0x8080ff;
    }

    public int getLightColor() {
        return getIndex() == 0 ? 0xffa0a0 : 0xa0a0ff;
    }

    public int getDarkColor() {
        return getIndex() == 0 ? 0x603030 : 0x303060;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
        if (this.message.length() > 50) this.message = this.message.substring(0, 50);
    }

    @Override
    public int getExpectedOutputLines() {
        return 1;
    }

    public ArrayList<InputError> popErrors() {
        ArrayList<InputError> result = new ArrayList<>();
        while (errors.size() > 0) {
            InputError error = errors.get(0);
            ArrayList<InputError> errorsOfType = errors.stream().filter(e -> e.getErrorCode() == error.getErrorCode()).collect(Collectors.toCollection(ArrayList::new));
            errors.removeAll(errorsOfType);
            if (errorsOfType.size() <= 3) result.addAll(errorsOfType);
            else {
                result.add(errorsOfType.get(0));
                result.add(errorsOfType.get(1));
                result.add(new InputError((errorsOfType.size() - 2) + " more errors of that type", error.getErrorCode(), error.isCritical()));
            }
        }
        return result;
    }

    public ArrayList<String> popSummaries() {
        ArrayList<String> result = summaries;
        summaries = new ArrayList<>();
        return result;
    }

    public void addError(InputError error) {
        errors.add(error);
    }

    public void addSummary(String summary) {
        summaries.add(summary);
    }

    public void initView(GraphicEntityModule graphicEntityModule, TooltipModule tooltipModule) {
        view = new PlayerView(this, graphicEntityModule, tooltipModule);
    }

    public void updateView(int turn) {
        view.update(turn);
    }
}
