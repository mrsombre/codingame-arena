// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotFoundException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotFoundException.java:3-10

public class CellNotFoundException extends GameException {
    public CellNotFoundException(int id) {
        super("Cell " + id + " not found");
    }
}
*/

func NewCellNotFoundException(id int) *GameError {
	return &GameError{
		Kind:    GameErrCellNotFound,
		Message: fmt.Sprintf("Cell %d not found", id),
	}
}
