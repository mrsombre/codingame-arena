// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"fmt"
	"strings"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java:17-21

@Singleton
public class CommandManager {
    @Inject private MultiplayerGameManager<Player> gameManager;

    public void parseCommands(Player player, List<String> lines) {
*/

type CommandManager struct {
	summary *[]string
}

func NewCommandManager(summary *[]string) *CommandManager {
	return &CommandManager{summary: summary}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java:21-110

public void parseCommands(Player player, List<String> lines) {
    String[] commands = lines.get(0).split(";");
    int reasonableLimitForActions = 30;
    for (String command : commands) {
        if (reasonableLimitForActions-- <= 0) return;
        try {
            boolean found = false;
            try {
                for (ActionType actionType : ActionType.values()) {
                    match = actionType.getPattern().matcher(command);
                    if (match.matches()) {
                        // build action, apply move/mark, track errors
                        found = true; break;
                    }
                }
            } catch (GameException e) { errors.add(...); continue; }
            if (!found) throw new InvalidInputException(Game.getExpected(command), command);
        } catch (InvalidInputException e) {
            deactivatePlayer(player, e.getMessage());
            gameManager.addToGameSummary(...);
            break;
        }
    }
    // emit up to 4 non-fatal errors to summary
}
*/

func (m *CommandManager) ParseCommands(player *Player, lines []string) {
	if len(lines) == 0 {
		m.DeactivatePlayer(player, "Timeout!")
		player.SetTimedOut(true)
		return
	}

	errors := make([]string, 0)
	commands := splitCommands(lines[0])
	reasonableLimitForActions := 30
	for _, command := range commands {
		if reasonableLimitForActions <= 0 {
			return
		}
		reasonableLimitForActions--

		parsed, err := Parse(command)
		if err != nil {
			inputErr := &InvalidInputError{Expected: GetExpected(command), Got: command}
			m.DeactivatePlayer(player, inputErr.Error())
			m.addSummary(inputErr.Error())
			m.addSummary(fmt.Sprintf("$%d: disqualified!", player.GetIndex()))
			break
		}

		if parsed.IsMove() {
			birdID := parsed.BirdID
			bird := player.BirdByID(birdID)
			switch {
			case bird == nil:
				errors = append(errors, formatError(fmt.Sprintf("$%d: Bird not found for id %d", player.GetIndex(), birdID)))
				continue
			case !bird.Alive:
				errors = append(errors, formatError(fmt.Sprintf("$%d: Bird with id %d is dead", player.GetIndex(), birdID)))
				continue
			case bird.HasMove:
				errors = append(errors, formatError(fmt.Sprintf("$%d: Bird id %d has already been given a move", player.GetIndex(), birdID)))
				continue
			case bird.Facing().Opposite() == parsed.Direction:
				errors = append(errors, formatError(fmt.Sprintf("$%d: Bird id %d cannot move backwards", player.GetIndex(), birdID)))
				continue
			}
			bird.Direction = parsed.Direction
			bird.HasMove = true
			if parsed.HasMessage {
				bird.SetMessage(parsed.Message)
			}
		} else if parsed.IsMark() {
			if !player.AddMark(parsed.Coord) {
				errors = append(errors, formatError(fmt.Sprintf("$%d: Too many MARK actions this turn", player.GetIndex())))
			}
		}
	}

	maxErrs := 4
	if len(errors) <= maxErrs+1 {
		for _, err := range errors {
			m.addSummary(err)
		}
	} else {
		for _, err := range errors[:maxErrs] {
			m.addSummary(err)
		}
		m.addSummary(formatError(fmt.Sprintf("...and %d more errors.", len(errors)-maxErrs)))
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java:123-126

public void deactivatePlayer(Player player, String message) {
    player.deactivate(EscapeHTMLEntities(message));
    player.setScore(-1);
}
*/

func (m *CommandManager) DeactivatePlayer(player *Player, message string) {
	player.Deactivate(EscapeHTMLEntities(message))
	player.SetScore(-1)
}

func (m *CommandManager) addSummary(line string) {
	if m.summary != nil {
		*m.summary = append(*m.summary, line)
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java:128-132

private String EscapeHTMLEntities(String message) {
    return message
        .replace("&lt;", "<")
        .replace("&gt;", ">");
}
*/

func EscapeHTMLEntities(message string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">").Replace(message)
}

func formatError(message string) string {
	return message
}

func splitCommands(line string) []string {
	commands := strings.Split(line, ";")
	for len(commands) > 1 && commands[len(commands)-1] == "" {
		commands = commands[:len(commands)-1]
	}
	return commands
}
