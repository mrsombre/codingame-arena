// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/CommandManager.java
package engine

import "strings"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/CommandManager.java:19-62

public void parseCommands(Player player, List<String> lines) {
    String line = lines.get(0);
    String[] commands = line.trim().split(";");
    try {
        for (String _command : commands) {
            String command = _command.trim();
            boolean found = false;
            for (ActionType actionType : ActionType.values()) {
                match = actionType.getPattern().matcher(command);
                if (match.matches()) {
                    Action action = new Action(actionType);
                    actionType.getConsumer().accept(match, action);
                    if (action.getType() == ActionType.MESSAGE) player.setMessage(action.getMessage());
                    else player.intents.add(action);
                    found = true;
                }
                if (found) break;
            }
            if (!found) throw new InvalidInputException(Game.getExpected(command), command);
        }
    } catch (InvalidInputException e) {
        deactivatePlayer(player, e.getMessage());
        gameManager.addToGameSummary(e.getMessage());
        gameManager.addToGameSummary(GameManager.formatErrorMessage(player.getNicknameToken() + ": disqualified!"));
    }
}
*/

// CommandManager turns one output line into player intents. An unparseable
// command disqualifies; a parseable-but-illegal one is left for Game to
// report and skip.
type CommandManager struct {
	game *Game
}

func NewCommandManager(game *Game) *CommandManager {
	return &CommandManager{game: game}
}

func (cm *CommandManager) ParseCommands(player *Player, lines []string) {
	// Java reads lines.get(0) unconditionally. The arena hands us an empty
	// slice when a bot wrote nothing; an empty line matches no ActionType, so
	// the player is disqualified either way — the game requires at least one
	// action per turn.
	var line string
	if len(lines) > 0 {
		line = lines[0]
	}

	for _, rawCommand := range javaSplit(strings.TrimSpace(line), ";") {
		command := strings.TrimSpace(rawCommand)
		found := false
		for _, actionType := range ActionTypes {
			match := actionType.Pattern().FindStringSubmatch(command)
			if match == nil {
				continue
			}
			action := NewAction(actionType)
			if err := actionType.Apply(match, action); err != nil {
				// Java would throw NumberFormatException out of the referee
				// here; treating an unrepresentable integer as invalid input
				// keeps the failure local to the offending player.
				cm.disqualify(player, NewInvalidInputError(GetExpected(command), command))
				return
			}
			if action.Type == ACTION_MESSAGE {
				player.SetMessage(action.Message)
			} else {
				player.Intents = append(player.Intents, action)
			}
			found = true
			break
		}
		if !found {
			cm.disqualify(player, NewInvalidInputError(GetExpected(command), command))
			return
		}
	}
}

func (cm *CommandManager) disqualify(player *Player, err *InvalidInputError) {
	cm.DeactivatePlayer(player, err.Error())
	cm.game.AddToGameSummary(err.Error())
	cm.game.AddToGameSummary(formatErrorMessage(playerNickname(player) + ": disqualified!"))
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/CommandManager.java:75-84

public void deactivatePlayer(Player player, String message) {
    player.deactivate(escapeHTMLEntities(message));
    player.setScore(-1);
}

private String escapeHTMLEntities(String message) {
    return message.replace("&lt;", "<").replace("&gt;", ">");
}
*/

func (cm *CommandManager) DeactivatePlayer(player *Player, message string) {
	player.Deactivate(escapeHTMLEntities(message))
	player.SetScore(-1)
}

func escapeHTMLEntities(message string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">").Replace(message)
}

// javaSplit reproduces String.split(literal): a string with no separator
// comes back whole, and trailing empty fields are dropped — so "WAIT;" is one
// command and ";;" is none at all.
func javaSplit(s, sep string) []string {
	parts := strings.Split(s, sep)
	if len(parts) == 1 {
		return parts
	}
	n := len(parts)
	for n > 0 && parts[n-1] == "" {
		n--
	}
	return parts[:n]
}
