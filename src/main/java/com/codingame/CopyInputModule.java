package com.codingame;

import com.codingame.gameengine.core.AbstractPlayer;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.core.Module;
import com.google.inject.Inject;
import engine.Board;

import java.util.ArrayList;

public class CopyInputModule implements Module {
    private GameManager<AbstractPlayer> gameManager;
    private Board board;

    @Inject
    CopyInputModule(GameManager<AbstractPlayer> gameManager) {
        this.gameManager = gameManager;
        gameManager.registerModule(this);
    }

    /**
     * Called at the beginning of the game
     */
    @Override
    public final void onGameInit() {
        gameManager.setViewGlobalData("inputmodule",  String.join("\n", board.getInitialInputs(0)));
        sendTurnData();
    }

    /**
     * Called at the end of every turn, after the Referee's gameTurn()
     */
    @Override
    public final void onAfterGameTurn() {
        sendTurnData();
    }

    private void sendTurnData() {
        ArrayList<String> inventoryInput = board.getTurnInputs(0);
        while (inventoryInput.size() > 2) inventoryInput.remove(2);
        gameManager.setViewData("inputmodule", String.join("\n", inventoryInput));
    }

    /**
     * Called at the end of the game, after the Referee's onEnd()
     */
    @Override
    public final void onAfterOnEnd() {
    }

    public void setBoard(Board board) {
        this.board = board;
    }
}