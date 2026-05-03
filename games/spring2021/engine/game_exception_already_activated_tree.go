// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/AlreadyActivatedTree.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/AlreadyActivatedTree.java:3-10

public class AlreadyActivatedTree extends GameException {
    public AlreadyActivatedTree(int id) {
        super("Tree on cell " + id + " is dormant (has already been used this round)");
    }
}
*/

func NewAlreadyActivatedTree(id int) *GameError {
	return &GameError{
		Kind:    GameErrAlreadyActivatedTree,
		Message: fmt.Sprintf("Tree on cell %d is dormant (has already been used this round)", id),
	}
}
