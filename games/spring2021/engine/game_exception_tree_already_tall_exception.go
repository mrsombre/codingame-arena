// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeAlreadyTallException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeAlreadyTallException.java:3-10

public class TreeAlreadyTallException extends GameException {
    public TreeAlreadyTallException(int id) {
        super("Tree on cell " + id + " cannot grow more (max size is 3).");
    }
}
*/

func NewTreeAlreadyTallException(id int) *GameError {
	return &GameError{
		Kind:    GameErrTreeAlreadyTall,
		Message: fmt.Sprintf("Tree on cell %d cannot grow more (max size is 3).", id),
	}
}
