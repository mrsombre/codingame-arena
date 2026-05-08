// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/CommandManager.java
package engine

import (
	"fmt"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
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
	// game is the trace sink for accepted MOVE/WAIT/MARK command traces.
	// Optional — unit tests construct CommandManager standalone and pass nil
	// here, in which case command traces are silently dropped (the tests
	// don't assert on them). The referee always wires a real *Game in.
	game *Game
}

func NewCommandManager(summary *[]string, game *Game) *CommandManager {
	return &CommandManager{summary: summary, game: game}
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
			m.traceMove(player, bird, parsed)
		} else if parsed.IsMark() {
			if !player.AddMark(parsed.Coord) {
				errors = append(errors, formatError(fmt.Sprintf("$%d: Too many MARK actions this turn", player.GetIndex())))
				continue
			}
			m.traceMark(player, parsed.Coord)
		} else if parsed.Type == TypeWait {
			m.traceWait(player)
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

// traceMove emits a MOVE command trace into the player's slot. Direction
// is rendered as the wire-token name (UP/DOWN/LEFT/RIGHT) so analyzers
// don't have to translate the internal NESW alias back to the bot
// vocabulary.
func (m *CommandManager) traceMove(player *Player, bird *Bird, parsed Action) {
	if m.game == nil {
		return
	}
	meta := MoveMeta{
		Bird:      bird.ID,
		Direction: directionToName(parsed.Direction),
	}
	if parsed.HasMessage {
		meta.Debug = bird.Message
	}
	m.game.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceMove, meta))
}

// traceWait emits a bare WAIT command trace (no data). Winter 2026's WAIT
// is a "do nothing" with no bird id and no message group, so the trace
// carries the type alone.
func (m *CommandManager) traceWait(player *Player) {
	if m.game == nil {
		return
	}
	m.game.tracePlayer(player.GetIndex(), arena.TurnTrace{Type: TraceWait})
}

// traceMark emits a MARK command trace per accepted MARK x y. Capped by
// Player.AddMark at four per side per turn; the 5th+ never reaches here.
func (m *CommandManager) traceMark(player *Player, coord Coord) {
	if m.game == nil {
		return
	}
	m.game.tracePlayer(player.GetIndex(), arena.MakeTurnTrace(TraceMark, MarkMeta{
		Coord: [2]int{coord.X, coord.Y},
	}))
}

// directionToName maps the internal Direction to the wire-token name
// the bot used (`UP`/`DOWN`/`LEFT`/`RIGHT`). Unset or out-of-range maps
// to the empty string, but a successfully parsed move always carries one
// of the four cardinals.
func directionToName(d Direction) string {
	switch d {
	case DirNorth:
		return "UP"
	case DirEast:
		return "RIGHT"
	case DirSouth:
		return "DOWN"
	case DirWest:
		return "LEFT"
	}
	return ""
}
