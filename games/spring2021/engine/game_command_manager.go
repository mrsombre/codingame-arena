// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"regexp"
	"strconv"
	"strings"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CommandManager.java:14-32

@Singleton
public class CommandManager {
    @Inject private GameSummaryManager gameSummaryManager;

    static final Pattern PLAYER_WAIT_PATTERN     = Pattern.compile("^WAIT(?:\\s+(?<message>.*))?");
    static final Pattern PLAYER_SEED_PATTERN     = Pattern.compile("^SEED (?<sourceId>\\d+) (?<targetId>\\d+)(?:\\s+(?<message>.*))?");
    static final Pattern PLAYER_GROW_PATTERN     = Pattern.compile("^GROW (?<targetId>\\d+)(?:\\s+(?<message>.*))?");
    static final Pattern PLAYER_COMPLETE_PATTERN = Pattern.compile("^COMPLETE (?<targetId>\\d+)(?:\\s+(?<message>.*))?");
}
*/

var (
	playerWaitPattern     = regexp.MustCompile(`^WAIT(?:\s+(.*))?`)
	playerSeedPattern     = regexp.MustCompile(`^SEED (\d+) (\d+)(?:\s+(.*))?`)
	playerGrowPattern     = regexp.MustCompile(`^GROW (\d+)(?:\s+(.*))?`)
	playerCompletePattern = regexp.MustCompile(`^COMPLETE (\d+)(?:\s+(.*))?`)
)

// CommandManager parses player outputs into actions on the Player struct, and
// deactivates a player on invalid or absent output (matching Java).
type CommandManager struct {
	Game    *Game
	Summary *GameSummaryManager
}

func NewCommandManager(game *Game, summary *GameSummaryManager) *CommandManager {
	return &CommandManager{Game: game, Summary: summary}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CommandManager.java:33-91

public void parseCommands(Player player, List<String> lines, Game game) {
    String command = lines.get(0);
    if (player.isWaiting()) return;
    try {
        // try WAIT, then GROW (if ENABLE_GROW), then COMPLETE, then SEED (if ENABLE_SEED)
        // throw new InvalidInputException(Game.getExpected(), command);
    } catch (InvalidInputException e) {
        deactivatePlayer(player, e.getMessage());
        gameSummaryManager.addPlayerBadCommand(...);
        gameSummaryManager.addPlayerDisqualified(...);
    }
}
*/

func (m *CommandManager) ParseCommands(player *Player, lines []string) {
	if player.IsWaiting() {
		return
	}
	if len(lines) == 0 {
		// Match the arena convention used elsewhere — empty stdout is the
		// arena runner's stand-in for a TimeoutException.
		m.DeactivatePlayer(player, "Timeout!")
		player.SetTimedOut(true)
		m.Summary.AddPlayerTimeout(player)
		m.Summary.AddPlayerDisqualified(player)
		return
	}

	command := lines[0]

	if match := playerWaitPattern.FindStringSubmatch(command); match != nil {
		player.SetAction(NewWaitAction())
		matchMessage(player, match[1])
		return
	}

	if m.Game.ENABLE_GROW {
		if match := playerGrowPattern.FindStringSubmatch(command); match != nil {
			targetID, err := strconv.Atoi(match[1])
			if err != nil {
				m.handleInvalid(player, command, err.Error())
				return
			}
			player.SetAction(NewGrowAction(targetID))
			matchMessage(player, match[2])
			return
		}
	}

	if match := playerCompletePattern.FindStringSubmatch(command); match != nil {
		targetID, err := strconv.Atoi(match[1])
		if err != nil {
			m.handleInvalid(player, command, err.Error())
			return
		}
		player.SetAction(NewCompleteAction(targetID))
		matchMessage(player, match[2])
		return
	}

	if m.Game.ENABLE_SEED {
		if match := playerSeedPattern.FindStringSubmatch(command); match != nil {
			sourceID, err := strconv.Atoi(match[1])
			if err != nil {
				m.handleInvalid(player, command, err.Error())
				return
			}
			targetID, err := strconv.Atoi(match[2])
			if err != nil {
				m.handleInvalid(player, command, err.Error())
				return
			}
			player.SetAction(NewSeedAction(sourceID, targetID))
			matchMessage(player, match[3])
			return
		}
	}

	m.handleInvalid(player, command, command)
}

func (m *CommandManager) handleInvalid(player *Player, command, got string) {
	err := &InvalidInputError{Expected: m.Game.GetExpected(), Got: got}
	m.DeactivatePlayer(player, err.Error())
	m.Summary.AddPlayerBadCommand(player, err)
	m.Summary.AddPlayerDisqualified(player)
	_ = command
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CommandManager.java:93-102

public void deactivatePlayer(Player player, String message) {
    player.deactivate(escapeHTMLEntities(message));
    player.setScore(-1);
}
private String escapeHTMLEntities(String message) {
    return message.replace("&lt;", "<").replace("&gt;", ">");
}
*/

func (m *CommandManager) DeactivatePlayer(player *Player, message string) {
	player.Deactivate(escapeHTMLEntities(message))
	player.SetScore(-1)
}

func escapeHTMLEntities(message string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">").Replace(message)
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/CommandManager.java:104-113

private void matchMessage(Player player, Matcher match) {
    String message = match.group("message");
    if (message != null) {
        String trimmed = message.trim();
        if (trimmed.length() > 48) trimmed = trimmed.substring(0, 46) + "...";
        player.setMessage(trimmed);
    }
}
*/

func matchMessage(player *Player, message string) {
	if message == "" {
		return
	}
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return
	}
	if len(trimmed) > 48 {
		trimmed = trimmed[:46] + "..."
	}
	player.SetMessage(trimmed)
}
