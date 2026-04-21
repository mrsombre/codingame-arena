// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/GameException.java
// Source: SpringChallenge2020/src/main/java/com/codingame/game/InvalidInputException.java
package engine

import "fmt"

type GameError struct {
	Message string
}

func (e *GameError) Error() string { return e.Message }

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
