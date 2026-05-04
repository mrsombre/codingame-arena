// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeTooFarException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/TreeTooFarException.java:3-10

public class TreeTooFarException extends GameException {
    public TreeTooFarException(int from, int to) {
        super(String.format("The tree on cell %d is too far from cell %d to plant a seed there", from, to));
    }
}
*/

func NewTreeTooFarException(from, to int) *GameError {
	return &GameError{
		Kind:    GameErrTreeTooFar,
		Message: fmt.Sprintf("The tree on cell %d is too far from cell %d to plant a seed there", from, to),
	}
}
