// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/action/Action.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/action/Action.java:5-22

public class Action {
    private ActionType type;
    private Coord coord;
    private Integer zoneId;
    private Coord from, to;
    boolean generatedByAutobuild;
    private String message;

    public Action(ActionType type) { this(type, false); }
    public Action(ActionType type, boolean generatedByAutobuild) { ... }
*/

// Action is one parsed player intent. Java's nullable Coord / Integer fields
// become value fields; nothing in the engine reads a field its ActionType
// does not populate.
type Action struct {
	Type   ActionType
	Coord  Coord
	ZoneID int
	From   Coord
	To     Coord

	GeneratedByAutobuild bool

	Message string
}

func NewAction(actionType ActionType) *Action {
	return NewAutobuildAction(actionType, false)
}

func NewAutobuildAction(actionType ActionType, generatedByAutobuild bool) *Action {
	return &Action{Type: actionType, GeneratedByAutobuild: generatedByAutobuild}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/action/Action.java:44-71,89-95

@Override public String toString() { return "Action [type=" + type + ", coord=" + coord + ", zone=" + zoneId + "]"; }
public boolean isMessage() { return type == ActionType.MESSAGE; }
public boolean isPlaceTrack() { return type == ActionType.PLACE_TRACK; }
public boolean isDisrupt() { return type == ActionType.DISRUPT; }
public boolean isAutobuild() { return type == ActionType.AUTOPLACE; }
public boolean isGeneratedByAutobuild() { return generatedByAutobuild; }
*/

func (a *Action) String() string {
	return fmt.Sprintf("Action [type=%s, coord=%s, zone=%d]", a.Type, a.Coord, a.ZoneID)
}

func (a *Action) IsMessage() bool { return a.Type == ACTION_MESSAGE }

func (a *Action) IsPlaceTrack() bool { return a.Type == ACTION_PLACE_TRACK }

func (a *Action) IsDisrupt() bool { return a.Type == ACTION_DISRUPT }

func (a *Action) IsAutobuild() bool { return a.Type == ACTION_AUTOPLACE }
