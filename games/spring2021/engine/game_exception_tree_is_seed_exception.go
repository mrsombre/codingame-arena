// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeIsSeedException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeIsSeedException.java:3-10

public class TreeIsSeedException extends GameException {
    public TreeIsSeedException(int id) {
        super("The seed on " + id + " cannot produce seeds");
    }
}
*/

func NewTreeIsSeedException(id int) *GameError {
	return &GameError{
		Kind:    GameErrTreeIsSeed,
		Message: fmt.Sprintf("The seed on %d cannot produce seeds", id),
	}
}
