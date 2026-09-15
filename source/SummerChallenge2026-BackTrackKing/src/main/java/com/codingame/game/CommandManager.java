package com.codingame.game;

import java.util.List;
import java.util.Optional;
import java.util.regex.Matcher;

import com.codingame.game.action.Action;
import com.codingame.game.action.ActionType;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class CommandManager {

    @Inject private MultiplayerGameManager<Player> gameManager;

    public void parseCommands(Player player, List<String> lines) {

        String line = lines.get(0);
        String[] commands = line.trim().split(";");
        try {
            Matcher match;
            try {
                for (String _command : commands) {
                    String command = _command.trim();
                    boolean found = false;
                    for (ActionType actionType : ActionType.values()) {

                        match = actionType.getPattern().matcher(command);
                        if (match.matches()) {
                            Action action = new Action(actionType);
                            actionType.getConsumer().accept(match, action);

                            if (action.getType() == ActionType.MESSAGE) {
                                player.setMessage(action.getMessage());
                            } else {
                                player.intents.add(action);
                            }

                            found = true;
                        }
                        if (found) {
                            break;
                        }
                    }
                    if (!found) {
                        throw new InvalidInputException(Game.getExpected(command), command);
                    }
                }
            } catch (Exception e) {
                throw e;
            }

        } catch (Exception e) {
            deactivatePlayer(player, e.getMessage());
            gameManager.addToGameSummary(e.getMessage());
            gameManager.addToGameSummary(GameManager.formatErrorMessage(player.getNicknameToken() + ": disqualified!"));
        }

    }

    private Optional<Integer> getAgentId(String[] tokens) {
        if (tokens.length > 1) {
            try {
                return Optional.of(Integer.parseInt(tokens[0]));
            } catch (NumberFormatException e) {
                return Optional.empty();
            }
        }
        return Optional.empty();
    }

    public void deactivatePlayer(Player player, String message) {
        player.deactivate(escapeHTMLEntities(message));
        player.setScore(-1);
    }

    private String escapeHTMLEntities(String message) {
        return message
            .replace("&lt;", "<")
            .replace("&gt;", ">");
    }
}
