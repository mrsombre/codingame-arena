// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/InvalidInputException.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/InvalidInputException.java:1-14

package com.codingame.game;

@SuppressWarnings("serial")
public class InvalidInputException extends Exception {

    public InvalidInputException(String expected, String got) {
        super("Invalid Input: Expected " + expected + " but got '" + got + "'");
    }

}
*/

import "fmt"

type InvalidInputError struct {
	Expected string
	Got      string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("Invalid Input: Expected %s but got '%s'", e.Expected, e.Got)
}
