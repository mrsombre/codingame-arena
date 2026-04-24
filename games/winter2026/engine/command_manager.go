// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"fmt"
	"strings"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/action"
)

type CommandManager struct {
	summary *[]string
}

func NewCommandManager(summary *[]string) *CommandManager {
	return &CommandManager{summary: summary}
}

func (m *CommandManager) ParseCommands(player *Player, lines []string) {
	if len(lines) == 0 {
		m.deactivatePlayer(player, "Timeout!")
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

		parsed, err := action.Parse(command)
		if err != nil {
			inputErr := &InvalidInputError{Expected: GetExpected(command), Got: command}
			m.deactivatePlayer(player, inputErr.Error())
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

func (m *CommandManager) deactivatePlayer(player *Player, message string) {
	player.Deactivate(escapeHTMLEntities(message))
	player.SetScore(-1)
}

func (m *CommandManager) addSummary(line string) {
	if m.summary != nil {
		*m.summary = append(*m.summary, line)
	}
}

func escapeHTMLEntities(message string) string {
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
