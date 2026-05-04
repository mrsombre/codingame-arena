// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/action/Action.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/action/Action.java:3-34

public abstract class Action {
    public Integer sourceId;
    public Integer targetId;

    public static final Action NO_ACTION = new Action() {};

    public boolean isGrow()     { return false; }
    public boolean isComplete() { return false; }
    public boolean isSeed()     { return false; }
    public boolean isWait()     { return false; }
}
*/

// ActionKind tags Java's Action subclass hierarchy.
type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionWait
	ActionSeed
	ActionGrow
	ActionComplete
)

// Action holds a parsed player command — Go-flat replacement for the Java
// abstract Action + four subclasses. SourceId/TargetId mirror the Java
// Integer fields (presence is implied by Kind).
type Action struct {
	Kind     ActionKind
	SourceID int
	TargetID int
	Message  string
}

// NoAction is the sentinel equivalent of Java Action.NO_ACTION (the abstract
// instance returned by getActionType() before any command is parsed).
var NoAction = Action{Kind: ActionNone}

func (a Action) IsGrow() bool     { return a.Kind == ActionGrow }
func (a Action) IsComplete() bool { return a.Kind == ActionComplete }
func (a Action) IsSeed() bool     { return a.Kind == ActionSeed }
func (a Action) IsWait() bool     { return a.Kind == ActionWait }
func (a Action) GetSourceID() int { return a.SourceID }
func (a Action) GetTargetID() int { return a.TargetID }
