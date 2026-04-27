// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:28-59

static final Pattern PLAYER_ACTION_PATTERN = Pattern.compile(
    "^(WAIT|MOVE|SWITCH|SPEED|MSG)\\s+(?<id>\\d+).*",
    Pattern.CASE_INSENSITIVE
);
static final Pattern PLAYER_WAIT_PATTERN = Pattern.compile(
    "^WAIT\\s+(?<id>\\d+)(?:\\s+(?<message>.+))?$",
    Pattern.CASE_INSENSITIVE
);
static final Pattern PLAYER_MOVE_PATTERN = Pattern.compile(
    "^MOVE\\s+(?<id>\\d+)\\s+(?<x>-?\\d+)\\s+(?<y>-?\\d+)(?:\\s+(?<message>.+))?$",
    Pattern.CASE_INSENSITIVE
);
static final Pattern PLAYER_SWITCH_TYPE_PATTERN = Pattern.compile(
    "^SWITCH\\s+(?<id>\\d+)\\s+(?<switchType>(ROCK|PAPER|SCISSORS))(?:\\s+(?<message>.+))?$",
    Pattern.CASE_INSENSITIVE
);
static final Pattern PLAYER_SPEED_TYPE_PATTERN = Pattern.compile(
    "^SPEED\\s+(?<id>\\d+)(?:\\s+(?<message>.+))?$",
    Pattern.CASE_INSENSITIVE
);
*/

var (
	PlayerActionPattern = regexp.MustCompile(`(?i)^(WAIT|MOVE|SWITCH|SPEED|MSG)\s+(\d+).*`)
	PlayerWaitPattern   = regexp.MustCompile(`(?i)^WAIT\s+(\d+)(?:\s+(.+))?$`)
	PlayerMovePattern   = regexp.MustCompile(`(?i)^MOVE\s+(\d+)\s+(-?\d+)\s+(-?\d+)(?:\s+(.+))?$`)
	PlayerSwitchPattern = regexp.MustCompile(`(?i)^SWITCH\s+(\d+)\s+(ROCK|PAPER|SCISSORS)(?:\s+(.+))?$`)
	PlayerSpeedPattern  = regexp.MustCompile(`(?i)^SPEED\s+(\d+)(?:\s+(.+))?$`)
)

// CommandManager parses player outputs into pacman intents.
type CommandManager struct {
	Summary *[]string
	Game    *Game
}

func NewCommandManager(summary *[]string, game *Game) *CommandManager {
	return &CommandManager{Summary: summary, Game: game}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:72-76

if (EXPECTED == null) {
    EXPECTED = MOVE_EXPECTED;
    if (Config.SPEED_ABILITY_AVAILABLE)  EXPECTED += " or " + SPEED_EXPECTED;
    if (Config.SWITCH_ABILITY_AVAILABLE) EXPECTED += " or " + SWITCH_EXPECTED;
}
*/

func (m *CommandManager) Expected() string {
	exp := "MOVE <id> <x> <y>"
	if m.Game.Config.SPEED_ABILITY_AVAILABLE {
		exp += " or SPEED <id>"
	}
	if m.Game.Config.SWITCH_ABILITY_AVAILABLE {
		exp += " or SWITCH <id> <type(ROCK|PAPER|SCISSORS)>"
	}
	return exp
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:78-178

public void parseCommands(Player player, List<String> lines, Game game) {
    String[] commands = lines.get(0).split("\\|");
    for (String command : commands) {
        String cStr = command.trim();
        if (cStr.isEmpty()) continue;
        try {
            // (resolves pac via PLAYER_ACTION_PATTERN; throws GameException on dead/missing/twice)
            // (matches MOVE → bounds-check x,y → MoveAction)
            // (matches SPEED → SpeedAction)
            // (matches SWITCH → SwitchAction with PacmanType)
            // (matches WAIT → no-op)
            // else throw new InvalidInputException(EXPECTED, cStr);
        } catch (InvalidInputException | GameException e) {
            deactivatePlayer(player, e.getMessage());
            return;
        }
    }
    player.getAlivePacmen()
        .filter(pac -> pac.getIntent() == Action.NO_ACTION)
        .forEach(pac -> pac.addToGameSummary("Pac " + pac.getNumber() + " received no command."));
}
*/

// ParseCommands walks the commands the player sent this turn, translating them
// into pacman intents. Any parse or validation failure deactivates the player.
func (m *CommandManager) ParseCommands(player *Player, lines []string) {
	if len(lines) == 0 {
		m.DeactivatePlayer(player, "Timeout!")
		player.SetTimedOut(true)
		return
	}

	expected := m.Expected()
	commands := strings.Split(lines[0], "|")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		pac, handled, err := m.ResolvePacman(player, cmd, expected)
		if err != nil {
			m.DeactivatePlayer(player, err.Error())
			m.AddSummary("Bad command: " + err.Error())
			return
		}
		if handled {
			continue
		}

		if match := PlayerMovePattern.FindStringSubmatch(cmd); match != nil {
			x, _ := strconv.Atoi(match[2])
			y, _ := strconv.Atoi(match[3])
			if x < 0 || x >= m.Game.Grid.Width || y < 0 || y >= m.Game.Grid.Height {
				msg := fmt.Sprintf(
					"Pac %d (p%d) cannot reach its target (%d, %d) because it is out of grid!",
					pac.Number, player.GetIndex(), x, y,
				)
				m.DeactivatePlayer(player, msg)
				m.AddSummary("Bad command: " + msg)
				return
			}
			m.HandlePacmanCommand(pac, NewMoveAction(Coord{X: x, Y: y}), match[4])
			continue
		}
		if match := PlayerSpeedPattern.FindStringSubmatch(cmd); match != nil {
			m.HandlePacmanCommand(pac, NewSpeedAction(), match[2])
			continue
		}
		if match := PlayerSwitchPattern.FindStringSubmatch(cmd); match != nil {
			t, _ := PacmanTypeFromName(strings.ToUpper(match[2]))
			m.HandlePacmanCommand(pac, NewSwitchAction(t), match[3])
			continue
		}
		if match := PlayerWaitPattern.FindStringSubmatch(cmd); match != nil {
			_ = match
			continue
		}

		invalid := &InvalidInputError{Expected: expected, Got: cmd}
		m.DeactivatePlayer(player, invalid.Error())
		m.AddSummary("Bad command: " + invalid.Error())
		return
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:89-105

int pacNumber = Integer.parseInt(match.group("id"));
try {
    pac = player.getPacmen().stream()
        .filter(value -> value.getNumber() == pacNumber)
        .findFirst().get();
    if (pac.isDead()) {
        pac.addToGameSummary(String.format("Pac %d is dead! It cannot be commanded anymore!", pacNumber));
        continue;
    }
    if (pac.getIntent() != Action.NO_ACTION) {
        throw new GameException(String.format("Pac %d cannot be commanded twice!", pacNumber));
    }
} catch (NoSuchElementException e) {
    throw new GameException(String.format("Pac %d doesn't exist", pacNumber));
}
*/

// ResolvePacman locates the targeted pacman for cmd and returns it.
// handled is true when the command itself has been fully processed (dead pac
// or twice-commanded skip); the caller should move to the next command.
func (m *CommandManager) ResolvePacman(player *Player, cmd, expected string) (*Pacman, bool, error) {
	match := PlayerActionPattern.FindStringSubmatch(cmd)
	if match == nil {
		return nil, false, &InvalidInputError{Expected: expected, Got: cmd}
	}
	pacNumber, _ := strconv.Atoi(match[2])
	var pac *Pacman
	for _, candidate := range player.Pacmen {
		if candidate.Number == pacNumber {
			pac = candidate
			break
		}
	}
	if pac == nil {
		return nil, false, &GameError{Message: fmt.Sprintf("Pac %d doesn't exist", pacNumber)}
	}
	if pac.Dead {
		return pac, true, nil
	}
	if !pac.Intent.IsNoAction() {
		return nil, false, &GameError{Message: fmt.Sprintf("Pac %d cannot be commanded twice!", pacNumber)}
	}
	return pac, false, nil
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:180-199

private void handlePacmanCommand(Pacman pac, Action intent, String message) {
    if (message != null && !message.trim().isEmpty()) {
        pac.setMessage(message);
    }
    pac.setIntent(intent);
    if (intent.getActionType() == ActionType.SPEED) {
        pac.setAbilityToUse(Ability.Type.SPEED);
    } else if (intent.getActionType() == ActionType.SWITCH) {
        PacmanType switchType = intent.getType();
        if (pac.getType() != switchType) {
            pac.setAbilityToUse(Ability.Type.fromType(switchType));
        }
    }
}
*/

func (m *CommandManager) HandlePacmanCommand(pac *Pacman, intent Action, message string) {
	if strings.TrimSpace(message) != "" {
		pac.SetMessage(message)
	}
	pac.Intent = intent

	switch intent.Type {
	case ActionSpeed:
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	case ActionSwitch:
		if pac.Type != intent.NewType {
			pac.AbilityToUse = AbilityFromSwitchType(intent.NewType)
			pac.HasAbilityToUse = pac.AbilityToUse != AbilityUnset
		}
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java:201-209

private void deactivatePlayer(Player player, String message) {
    player.deactivate(escapeHTMLEntities(message));
}
private String escapeHTMLEntities(String message) {
    return message.replace("&lt;", "<").replace("&gt;", ">");
}
*/

func (m *CommandManager) DeactivatePlayer(player *Player, message string) {
	player.Deactivate(EscapeHTMLEntities(message))
}

func (m *CommandManager) AddSummary(line string) {
	if m.Summary != nil {
		*m.Summary = append(*m.Summary, line)
	}
}

func EscapeHTMLEntities(message string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">").Replace(message)
}
