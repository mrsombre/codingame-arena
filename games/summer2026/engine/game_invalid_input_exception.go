// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/InvalidInputException.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/InvalidInputException.java:4-12

public class InvalidInputException extends Exception {
    public InvalidInputException(String expected, String got) {
        super("Invalid Input: Expected " + expected + " but got '" + got + "'");
    }

    public InvalidInputException(String error, String expected, String got) {
        super(error + ": Expected " + expected + " but got '" + got + "'");
    }
}
*/

// InvalidInputError is the disqualification message a player sees when a
// command does not parse. The exact wording reaches the bot author through
// the match summary, so it is part of the observable contract. Java's second
// constructor swaps the "Invalid Input" prefix for a caller-supplied one,
// which Prefix carries.
type InvalidInputError struct {
	Prefix   string
	Expected string
	Got      string
}

func NewInvalidInputError(expected, got string) *InvalidInputError {
	return &InvalidInputError{Prefix: "Invalid Input", Expected: expected, Got: got}
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("%s: Expected %s but got '%s'", e.Prefix, e.Expected, e.Got)
}
