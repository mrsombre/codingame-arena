package com.codingame;

import com.codingame.gameengine.core.AbstractPlayer;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.core.Module;
import com.google.inject.Inject;
import engine.Plant;
import engine.Unit;

import java.util.ArrayList;
import java.util.HashMap;

public class CompressionModule implements Module {
    private GameManager<AbstractPlayer> gameManager;
    private HashMap<Integer, Unit> units = new HashMap<>();
    private HashMap<Integer, Plant> plants = new HashMap<>();
    private HashMap<Integer, Unit> prevUnits = new HashMap<>();
    private HashMap<Integer, Plant> prevPlants = new HashMap<>();
    private static String alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";

    @Inject
    CompressionModule(GameManager<AbstractPlayer> gameManager) {
        this.gameManager = gameManager;
        gameManager.registerModule(this);
    }

    /**
     * Called at the beginning of the game
     */
    @Override
    public final void onGameInit() {
        sendData();
    }

    /**
     * Called at the end of every turn, after the Referee's gameTurn()
     */
    @Override
    public final void onAfterGameTurn() {
        sendData();
    }

    private void sendData() {
        ArrayList<String> delta = new ArrayList<>();
        for (int id : units.keySet()) {
            String diff = units.get(id).serializeDelta(prevUnits.get(id), alphabet);
            if (diff != null) delta.add(id + diff);
            prevUnits.put(id, units.get(id));
        }
        for (int id : plants.keySet()) {
            Plant plant = plants.get(id);
            String diff = plant.serializeDelta(prevPlants.get(id), alphabet);
            if (diff != null) delta.add(id + diff);
            plant.tick(false);
            prevPlants.put(id, plant);
        }
        gameManager.setViewData("diff", String.join(";", delta));
        //System.err.println(String.join(";", delta));
        units.clear();
        plants.clear();
    }

    /**
     * Called at the end of the game, after the Referee's onEnd()
     */
    @Override
    public final void onAfterOnEnd() {
    }

    public void addPlant(Plant plant, int id) {
        plants.put(id, new Plant(plant));
    }

    public void addUnit(Unit unit, int id) {
        units.put(id, new Unit(unit));
    }
}