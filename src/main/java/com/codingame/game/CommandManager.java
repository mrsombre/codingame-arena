package com.codingame.game;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.regex.Matcher;

import com.codingame.game.action.Action;
import com.codingame.game.action.ActionType;
import com.codingame.game.grid.Coord;
import com.codingame.gameengine.core.GameManager;
import com.codingame.gameengine.core.MultiplayerGameManager;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class CommandManager {

    @Inject private MultiplayerGameManager<Player> gameManager;

    public void parseCommands(Player player, List<String> lines) {

        List<String> errors = new ArrayList<>();

        String line = lines.get(0);
        String[] commands = line.split(";");

        int reasonableLimitForActions = 30;
        for (String command : commands) {
            if (reasonableLimitForActions-- <= 0) {
                return;
            }

            try {
                Matcher match;

                boolean found = false;
                try {
                    for (ActionType actionType : ActionType.values()) {
                        match = actionType.getPattern().matcher(command);
                        if (match.matches()) {
                            Action action = new Action(actionType);
                            actionType.getConsumer().accept(match, action);

                            if (action.isMove()) {
                                Integer birdId = action.getBirdId();
                                Bird bird = player.getBirdById(birdId);
                                if (bird == null) {
                                    throw new GameException("Bird not found for id " + birdId);
                                }
                                if (!bird.alive) {
                                    throw new GameException("Bird with id " + birdId + " is dead");
                                }
                                if (bird.direction != null) {
                                    throw new GameException("Bird id " + birdId + " has already been given a move");
                                }

                                if (bird.getFacing().opposite() == action.getDirection()) {
                                    throw new GameException("Bird id " + birdId + " cannot move backwards");
                                }
                                bird.direction = action.getDirection();

                                // parse message
                                if (action.getMessage() != null) {
                                    bird.setMessage(action.getMessage());
                                }
                            } else if (action.isMark()) {
                                if (!player.addMark(action.getCoord())) {
                                    throw new GameException("Too many MARK actions this turn");
                                }
                            }
                            found = true;
                            break;
                        }
                    }
                } catch (GameException e) {
                    errors.add(GameManager.formatErrorMessage(player.getNicknameToken() + ": " + e.getMessage()));
                    continue;
                } catch (Exception e) {
                    throw e;
                }

                if (!found) {
                    throw new InvalidInputException(Game.getExpected(command), command);
                }

            } catch (InvalidInputException e) {
                deactivatePlayer(player, e.getMessage());
                gameManager.addToGameSummary(e.getMessage());
                gameManager.addToGameSummary(GameManager.formatErrorMessage(player.getNicknameToken() + ": disqualified!"));
                break;
            }
        }

        int maxErrs = 4;

        if (errors.size() <= maxErrs + 1) {
            for (String err : errors) {
                gameManager.addToGameSummary(err);
            }
        } else {
            for (String err : errors.stream().limit(maxErrs).toList()) {
                gameManager.addToGameSummary(err);
            }
            gameManager.addToGameSummary(
                GameManager.formatErrorMessage("...and " + (errors.size() - maxErrs) + " more errors.")
            );
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
