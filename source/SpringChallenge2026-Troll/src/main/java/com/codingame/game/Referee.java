package com.codingame.game;
import java.util.List;

import com.codingame.CompressionModule;
import com.codingame.CopyInputModule;
import com.codingame.gameengine.core.AbstractPlayer.TimeoutException;
import com.codingame.gameengine.core.AbstractReferee;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.codingame.gameengine.module.endscreen.EndScreenModule;
import com.codingame.gameengine.module.entities.GraphicEntityModule;
import com.codingame.gameengine.module.toggle.ToggleModule;
import com.codingame.gameengine.module.tooltip.TooltipModule;
import com.google.inject.Inject;
import engine.Board;
import engine.Constants;
import engine.task.InputError;
import engine.task.TaskManager;

public class Referee extends AbstractReferee {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private GraphicEntityModule graphicEntityModule;
    @Inject private TooltipModule tooltipModule;
    @Inject private ToggleModule toggleModule;
    @Inject private EndScreenModule endScreenModule;
    @Inject private CopyInputModule copyInputModule;
    @Inject private CompressionModule compressionModule;

    private Board board;

    @Override
    public void init() {
        gameManager.setTurnMaxTime(Constants.TIME_PER_TURN);
        gameManager.setMaxTurns(gameManager.getLeagueLevel() > 2 ? Constants.GAME_TURNS : Constants.GAME_TURNS_LOW_LEAGUE);
        board = Board.createMap(gameManager.getActivePlayers(), gameManager.getRandom(), gameManager.getLeagueLevel(), graphicEntityModule, tooltipModule, toggleModule, compressionModule);
        copyInputModule.setBoard(board);
    }

    @Override
    public void gameTurn(int turn) {
        for (Player player : gameManager.getActivePlayers()) {
            if (turn == 1) {
                for (String line : board.getInitialInputs(player.getIndex())) player.sendInputLine(line);
            }
            for (String line : board.getTurnInputs(player.getIndex())) player.sendInputLine(line);
            player.execute();
        }

        TaskManager taskManager = new TaskManager();
        for (Player player : gameManager.getActivePlayers()) {
            try {
                List<String> outputs = player.getOutputs();
                taskManager.parseTasks(player, board, outputs.get(0), gameManager.getLeagueLevel());
            } catch (TimeoutException e) {
                killPlayer(player, "timeout");
            }
        }

        board.tick(turn, taskManager, gameManager);
        for (Player player : gameManager.getPlayers()) {
            for (InputError error : player.popErrors()) {
                if (error.isCritical()) {
                    killPlayer(player, error.getMessage());
                } else {
                    gameManager.addToGameSummary(player.getNicknameToken() + ": [failed] " + error.getMessage());
                }
            }
            for (String summary : player.popSummaries())
                gameManager.addToGameSummary(player.getNicknameToken() + ": " + summary);
            if (player.hasTimelimitExceededLastTurn()) {
                String strike = player.getTimelimitsExceeded() + "th";
                if (player.getTimelimitsExceeded() == 1) strike = "1st";
                if (player.getTimelimitsExceeded() == 2) strike = "2nd";
                if (player.getTimelimitsExceeded() == 3) strike = "3rd";
                gameManager.addToGameSummary(player.getNicknameToken() + " exceeded the time limit: " + player.getLastExectionTimeMs() + " ms, " + strike + " strike");
            }
        }
        if (board.hasStalled()) gameManager.endGame();
    }

    private void killPlayer(Player player, String message) {
        if (!player.isActive()) return;
        gameManager.addToGameSummary(player.getNicknameToken() + ": " + message);
        player.deactivate(String.format("$%d %s!", player.getIndex(), message));
        player.setScore(-2);
        gameManager.endGame();
    }

    @Override
    public void onEnd() {
        int[] scores = gameManager.getPlayers().stream().mapToInt(p -> p.getScore()).toArray();
        String[] texts = new String[scores.length];
        for (int i = 0; i < scores.length; i++)
        {
            texts[i] = scores[i] + " points";
            int timeLimits = gameManager.getPlayers().get(i).getTimelimitsExceeded();
            if (timeLimits > 0) texts[i] += " (" + timeLimits + " x time limit exceeded)";
        }
        endScreenModule.setScores(scores, texts);
    }
}
