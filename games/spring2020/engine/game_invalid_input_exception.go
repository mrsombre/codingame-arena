// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/InvalidInputException.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/InvalidInputException.java:1-14

package com.codingame.game;

@SuppressWarnings("serial")
public class InvalidInputException extends Exception {

    public InvalidInputException(String expected, String got) {
        super("Invalid Input: Expected " + expected + " but got '" + got + "'");
    }

    public InvalidInputException(String error, String expected, String got) {
        super(error + ": Expected " + expected + " but got '" + got + "'");
    }

}
*/

import "fmt"

type InvalidInputError struct {
	Expected string
	Got      string
	Prefix   string
}

func (e *InvalidInputError) Error() string {
	if e.Prefix != "" {
		return fmt.Sprintf("%s: Expected %s but got '%s'", e.Prefix, e.Expected, e.Got)
	}
	return fmt.Sprintf("Invalid Input: Expected %s but got '%s'", e.Expected, e.Got)
}
