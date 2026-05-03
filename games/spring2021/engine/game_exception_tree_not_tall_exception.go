// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeNotTallException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeNotTallException.java:3-10

public class TreeNotTallException extends GameException {
    public TreeNotTallException(int id) {
        super("The tree on cell " + id + " is not large enough");
    }
}
*/

func NewTreeNotTallException(id int) *GameError {
	return &GameError{
		Kind:    GameErrTreeNotTall,
		Message: fmt.Sprintf("The tree on cell %d is not large enough", id),
	}
}
