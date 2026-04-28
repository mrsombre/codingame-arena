// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Ability.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Ability.java:4-5

public enum Type {
    SPEED, SET_ROCK, SET_PAPER, SET_SCISSORS;
}
*/

// Ability types.
type AbilityType int

const (
	AbilityUnset AbilityType = iota
	AbilitySpeed
	AbilitySetRock
	AbilitySetPaper
	AbilitySetScissors
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Ability.java:7-18

public static Type fromType(PacmanType switchType) {
    switch (switchType) {
    case PAPER:    return SET_PAPER;
    case ROCK:     return SET_ROCK;
    case SCISSORS: return SET_SCISSORS;
    default:       return null;
    }
}
*/

// AbilityFromSwitchType returns the SET_* ability matching a PacmanType switch.
// Returns AbilityUnset for TypeNeutral (Java returned null).
func AbilityFromSwitchType(t PacmanType) AbilityType {
	switch t {
	case TypeRock:
		return AbilitySetRock
	case TypePaper:
		return AbilitySetPaper
	case TypeScissors:
		return AbilitySetScissors
	}
	return AbilityUnset
}

// PacTypeFromAbility returns the PacmanType an ability sets the pac to,
// or TypeNeutral for non-switch abilities. Go-only inverse helper.
func PacTypeFromAbility(a AbilityType) PacmanType {
	switch a {
	case AbilitySetRock:
		return TypeRock
	case AbilitySetPaper:
		return TypePaper
	case AbilitySetScissors:
		return TypeScissors
	}
	return TypeNeutral
}
