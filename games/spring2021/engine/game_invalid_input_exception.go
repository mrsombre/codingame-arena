// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/InvalidInputException.java
package engine

import "fmt"

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/InvalidInputException.java:1-22

public class InvalidInputException extends Exception {
    private final String expected;
    private final String got;
    public InvalidInputException(String expected, String got) {
        super("Invalid Input: Expected " + expected + " but got '" + got + "'");
        this.expected = expected;
        this.got = got;
    }
}
*/

type InvalidInputError struct {
	Expected string
	Got      string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("Invalid Input: Expected %s but got '%s'", e.Expected, e.Got)
}

func (e *InvalidInputError) GetExpected() string { return e.Expected }
func (e *InvalidInputError) GetGot() string      { return e.Got }
