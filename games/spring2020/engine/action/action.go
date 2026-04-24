// Package action
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/Action.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/ActionType.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/MoveAction.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SpeedAction.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/SwitchAction.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/ActionException.java
package action

import "github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"

// ActionType matches Java's ActionType enum.
type ActionType int

const (
	ActionWait ActionType = iota
	ActionMove
	ActionMsg
	ActionSpeed
	ActionSwitch
)

// PacType is the subset of pac types that a SWITCH action can pick.
// Mirrors Java PacmanType values; we duplicate the constant here to avoid
// an import cycle between action and engine.
type PacType int

const (
	PacRock     PacType = 0
	PacPaper    PacType = 1
	PacScissors PacType = 2
	PacNeutral  PacType = -1
)

// Action is a tagged-union port of the Java Action hierarchy.
//   - MoveAction:   Type=ActionMove, Target set
//   - SpeedAction:  Type=ActionSpeed
//   - SwitchAction: Type=ActionSwitch, NewType set
//   - NO_ACTION:    Type=ActionWait (use NoAction to check)
type Action struct {
	Type    ActionType
	Target  grid.Coord
	NewType PacType
}

// NoAction sentinel (equivalent to Java's Action.NO_ACTION).
var NoAction = Action{Type: ActionWait}

// IsNoAction reports whether a is the sentinel no-op action.
func (a Action) IsNoAction() bool {
	return a.Type == ActionWait
}

// NewMoveAction builds a MOVE action targeting coord.
func NewMoveAction(target grid.Coord) Action {
	return Action{Type: ActionMove, Target: target}
}

// NewSpeedAction builds a SPEED action.
func NewSpeedAction() Action {
	return Action{Type: ActionSpeed}
}

// NewSwitchAction builds a SWITCH action to the given pac type.
func NewSwitchAction(t PacType) Action {
	return Action{Type: ActionSwitch, NewType: t}
}
