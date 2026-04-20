package com.codingame.game;

import com.codingame.gameengine.core.AbstractPlayer.TimeoutException;
import com.codingame.gameengine.core.AbstractReferee;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.codingame.view.ViewModule;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class Referee extends AbstractReferee {

    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private CommandManager commandManager;
    @Inject private Game game;
    @Inject private ViewModule viewModule;

    @Override
    public void init() {
        try {
            gameManager.setMaxTurns(200);
            gameManager.setFirstTurnMaxTime(1000);

            game.init();
            sendGlobalInfo();

            gameManager.setFrameDuration(1000);
            gameManager.setTurnMaxTime(50);
        } catch (Exception e) {
            e.printStackTrace();
            System.err.println("Referee failed to initialize");
            abort();
        }
    }

    private void abort() {
        gameManager.endGame();

    }

    private void sendGlobalInfo() {
        // Give input to players
        for (Player player : gameManager.getActivePlayers()) {
            for (String line : Serializer.serializeGlobalInfoFor(player, game)) {
                player.sendInputLine(line);
            }
        }
    }

    @Override
    public void gameTurn(int turn) {
        game.resetGameTurnData();

        // Give input to players
        for (Player player : gameManager.getActivePlayers()) {
            if (game.shouldSkipPlayerTurn(player)) {
                continue;
            }
            for (String line : Serializer.serializeFrameInfoFor(player, game)) {
                player.sendInputLine(line);
            }
            player.execute();
        }
        // Get output from players
        handlePlayerCommands();

        game.performGameUpdate(turn);

        if (gameManager.getActivePlayers().size() < 2) {
            abort();
        }
    }

    private void handlePlayerCommands() {
        for (Player player : gameManager.getActivePlayers()) {
            if (game.shouldSkipPlayerTurn(player)) {
                continue;
            }
            try {
                commandManager.parseCommands(player, player.getOutputs());
            } catch (TimeoutException e) {
                player.deactivate("Timeout!");
                gameManager.addToGameSummary(player.getNicknameToken() + " has not provided " + player.getExpectedOutputLines() + " lines in time");
            }
        }

    }

    @Override
    public void onEnd() {
        game.onEnd();
    }
}
