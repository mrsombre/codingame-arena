// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/GameException.java
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/InvalidInputException.java
package engine

import "fmt"

type GameError struct {
	Message string
}

func (e *GameError) Error() string { return e.Message }

type InvalidInputError struct {
	Expected string
	Got      string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("Invalid Input: Expected %s but got '%s'", e.Expected, e.Got)
}
