// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/action"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

var (
	playerActionPattern = regexp.MustCompile(`(?i)^(WAIT|MOVE|SWITCH|SPEED|MSG)\s+(\d+).*`)
	playerWaitPattern   = regexp.MustCompile(`(?i)^WAIT\s+(\d+)(?:\s+(.+))?$`)
	playerMovePattern   = regexp.MustCompile(`(?i)^MOVE\s+(\d+)\s+(-?\d+)\s+(-?\d+)(?:\s+(.+))?$`)
	playerSwitchPattern = regexp.MustCompile(`(?i)^SWITCH\s+(\d+)\s+(ROCK|PAPER|SCISSORS)(?:\s+(.+))?$`)
	playerSpeedPattern  = regexp.MustCompile(`(?i)^SPEED\s+(\d+)(?:\s+(.+))?$`)
)

// CommandManager parses player outputs into pacman intents.
type CommandManager struct {
	summary *[]string
	game    *Game
}

func NewCommandManager(summary *[]string, game *Game) *CommandManager {
	return &CommandManager{summary: summary, game: game}
}

func (m *CommandManager) expected() string {
	exp := "MOVE <id> <x> <y>"
	if m.game.Config.SpeedAvail {
		exp += " or SPEED <id>"
	}
	if m.game.Config.SwitchAvail {
		exp += " or SWITCH <id> <type(ROCK|PAPER|SCISSORS)>"
	}
	return exp
}

// ParseCommands walks the commands the player sent this turn, translating them
// into pacman intents. Any parse or validation failure deactivates the player.
func (m *CommandManager) ParseCommands(player *Player, lines []string) {
	if len(lines) == 0 {
		m.deactivate(player, "Timeout!")
		return
	}

	expected := m.expected()
	commands := strings.Split(lines[0], "|")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		pac, handled, err := m.resolvePacman(player, cmd, expected)
		if err != nil {
			m.deactivate(player, err.Error())
			m.addSummary("Bad command: " + err.Error())
			return
		}
		if handled {
			continue
		}

		if match := playerMovePattern.FindStringSubmatch(cmd); match != nil {
			x, _ := strconv.Atoi(match[2])
			y, _ := strconv.Atoi(match[3])
			if x < 0 || x >= m.game.Grid.Width || y < 0 || y >= m.game.Grid.Height {
				msg := fmt.Sprintf(
					"Pac %d cannot reach its target (%d, %d) because it is out of grid!",
					pac.Number, x, y,
				)
				m.deactivate(player, msg)
				m.addSummary("Bad command: " + msg)
				return
			}
			m.handlePacmanCommand(pac, action.NewMoveAction(grid.Coord{X: x, Y: y}), match[4])
			continue
		}
		if match := playerSpeedPattern.FindStringSubmatch(cmd); match != nil {
			m.handlePacmanCommand(pac, action.NewSpeedAction(), match[2])
			continue
		}
		if match := playerSwitchPattern.FindStringSubmatch(cmd); match != nil {
			t, _ := PacmanTypeFromName(strings.ToUpper(match[2]))
			m.handlePacmanCommand(pac, action.NewSwitchAction(action.PacType(t)), match[3])
			continue
		}
		if match := playerWaitPattern.FindStringSubmatch(cmd); match != nil {
			_ = match
			continue
		}

		invalid := &InvalidInputError{Expected: expected, Got: cmd}
		m.deactivate(player, invalid.Error())
		m.addSummary("Bad command: " + invalid.Error())
		return
	}
}

// resolvePacman locates the targeted pacman for cmd and returns it.
// handled is true when the command itself has been fully processed (dead pac
// or twice-commanded skip); the caller should move to the next command.
func (m *CommandManager) resolvePacman(player *Player, cmd, expected string) (*Pacman, bool, error) {
	match := playerActionPattern.FindStringSubmatch(cmd)
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

func (m *CommandManager) handlePacmanCommand(pac *Pacman, intent action.Action, message string) {
	if strings.TrimSpace(message) != "" {
		pac.SetMessage(message)
	}
	pac.Intent = intent

	switch intent.Type {
	case action.ActionSpeed:
		pac.AbilityToUse = AbilitySpeed
		pac.HasAbilityToUse = true
	case action.ActionSwitch:
		target := PacmanType(intent.NewType)
		if pac.Type != target {
			pac.AbilityToUse = AbilityFromSwitchType(target)
			pac.HasAbilityToUse = pac.AbilityToUse != AbilityUnset
		}
	}
}

func (m *CommandManager) deactivate(player *Player, message string) {
	player.Deactivate(escapeHTMLEntities(message))
}

func (m *CommandManager) addSummary(line string) {
	if m.summary != nil {
		*m.summary = append(*m.summary, line)
	}
}

func escapeHTMLEntities(message string) string {
	return strings.NewReplacer("&lt;", "<", "&gt;", ">").Replace(message)
}
