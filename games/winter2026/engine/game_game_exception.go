// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/GameException.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/GameException.java:1-10

package com.codingame.game;

@SuppressWarnings("serial")
public class GameException extends Exception {

    public GameException(String string) {
        super(string);
    }

}
*/

type GameError struct {
	Message string
}

func (e *GameError) Error() string { return e.Message }
