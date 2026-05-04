// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotValidException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/CellNotValidException.java:3-10

public class CellNotValidException extends GameException {
    public CellNotValidException(int id) {
        super("You can't plant a seed on cell " + id);
    }
}
*/

func NewCellNotValidException(id int) *GameError {
	return &GameError{
		Kind:    GameErrCellNotValid,
		Message: fmt.Sprintf("You can't plant a seed on cell %d", id),
	}
}
