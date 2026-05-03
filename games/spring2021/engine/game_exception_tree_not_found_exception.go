// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeNotFoundException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeNotFoundException.java:3-10

public class TreeNotFoundException extends GameException {
    public TreeNotFoundException(int id) {
        super("There is no tree on cell " + id);
    }
}
*/

func NewTreeNotFoundException(id int) *GameError {
	return &GameError{
		Kind:    GameErrTreeNotFound,
		Message: fmt.Sprintf("There is no tree on cell %d", id),
	}
}
