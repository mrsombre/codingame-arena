// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotEmptyException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotEmptyException.java:3-10

public class CellNotEmptyException extends GameException {
    public CellNotEmptyException(int id) {
        super("There is already a tree on cell " + id);
    }
}
*/

func NewCellNotEmptyException(id int) *GameError {
	return &GameError{
		Kind:    GameErrCellNotEmpty,
		Message: fmt.Sprintf("There is already a tree on cell %d", id),
	}
}
