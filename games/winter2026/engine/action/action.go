// Package action
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/Action.java
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionException.java
package action

import (
	"fmt"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
)

// Action holds a parsed player command.
type Action struct {
	Type       ActionType
	Direction  grid.Direction
	BirdID     int
	HasBirdID  bool
	Coord      grid.Coord
	HasCoord   bool
	Message    string
	HasMessage bool
}

func (a Action) IsMove() bool { return a.HasBirdID && a.Direction != grid.DirUnset }
func (a Action) IsMark() bool { return a.HasCoord }

func (a Action) String() string {
	return fmt.Sprintf("Action [type=%d, direction=%s, birdId=%d]", a.Type, a.Direction, a.BirdID)
}

// ActionError is returned when an action string cannot be parsed.
type ActionError struct {
	Message string
}

func (e *ActionError) Error() string { return e.Message }

// Parse tries each pattern against a single action string.
// Returns an ActionError if no pattern matches.
func Parse(raw string) (Action, error) {
	s := raw
	for _, p := range patterns {
		match := p.re.FindStringSubmatch(s)
		if match == nil || match[0] != s {
			continue
		}
		a := Action{Type: p.actionType}
		p.populate(match, subexpMap(p.re), &a)
		return a, nil
	}
	return Action{}, &ActionError{Message: fmt.Sprintf("invalid action: %s", raw)}
}
