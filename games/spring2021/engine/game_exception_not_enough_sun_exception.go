// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/exception/NotEnoughSunException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/exception/NotEnoughSunException.java:3-10

public class NotEnoughSunException extends GameException {
    public NotEnoughSunException(int cost, int sun) {
        super(String.format("Not enough sun. You need %d but have %d", cost, sun));
    }
}
*/

func NewNotEnoughSunException(cost, sun int) *GameError {
	return &GameError{
		Kind:    GameErrNotEnoughSun,
		Message: fmt.Sprintf("Not enough sun. You need %d but have %d", cost, sun),
	}
}
