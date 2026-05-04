// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/GameException.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/GameException.java:1-10

public class GameException extends Exception {
    public GameException(String string) { super(string); }
}
*/

// GameError is the Go counterpart of Java GameException — base type for the
// per-action validation failures (CellNotEmpty, NotEnoughSun, etc.). Subclasses
// in Java only differ by message string, so we keep the message and a Kind tag
// for callers that need to switch on the failure reason.
type GameError struct {
	Kind    GameErrorKind
	Message string
}

func (e *GameError) Error() string { return e.Message }

type GameErrorKind int

const (
	GameErrUnknown GameErrorKind = iota
	GameErrAlreadyActivatedTree
	GameErrCellNotEmpty
	GameErrCellNotFound
	GameErrCellNotValid
	GameErrNotEnoughSun
	GameErrNotOwnerOfTree
	GameErrTreeAlreadyTall
	GameErrTreeIsSeed
	GameErrTreeNotFound
	GameErrTreeNotTall
	GameErrTreeTooFar
)
