// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/NotOwnerOfTreeException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/NotOwnerOfTreeException.java:5-12

public class NotOwnerOfTreeException extends GameException {
    public NotOwnerOfTreeException(int id, Player player) {
        super("The tree on cell " + id + " is owned by opponent");
    }
}
*/

func NewNotOwnerOfTreeException(id int, _ *Player) *GameError {
	return &GameError{
		Kind:    GameErrNotOwnerOfTree,
		Message: fmt.Sprintf("The tree on cell %d is owned by opponent", id),
	}
}
