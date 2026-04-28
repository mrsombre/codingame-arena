// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/PacmanType.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/PacmanType.java:5-16

public enum PacmanType {
    ROCK(Config.ID_ROCK), PAPER(Config.ID_PAPER), SCISSORS(Config.ID_SCISSORS), NEUTRAL(Config.ID_NEUTRAL);
    int id;
    private PacmanType(int id) { this.id = id; }
    public int getId() { return id; }
}
*/

// PacmanType matches Java PacmanType enum.
type PacmanType int

const (
	TypeRock     PacmanType = PacmanType(ID_ROCK)
	TypePaper    PacmanType = PacmanType(ID_PAPER)
	TypeScissors PacmanType = PacmanType(ID_SCISSORS)
	TypeNeutral  PacmanType = PacmanType(ID_NEUTRAL)
)

func (t PacmanType) Name() string {
	switch t {
	case TypeRock:
		return "ROCK"
	case TypePaper:
		return "PAPER"
	case TypeScissors:
		return "SCISSORS"
	case TypeNeutral:
		return "NEUTRAL"
	}
	return "NEUTRAL"
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/PacmanType.java:25-42

public static PacmanType fromCharacter(char c) {
    switch (c) {
    case 'r': case 'R': return ROCK;
    case 'p': case 'P': return PAPER;
    case 's': case 'S': return SCISSORS;
    case 'n': case 'N': return NEUTRAL;
    }
    throw new RuntimeException(c + " is not a valid pac type");
}
*/

// PacmanTypeFromCharacter maps a state-string char to its PacmanType.
// Lowercase: player 0. Uppercase: player 1.
func PacmanTypeFromCharacter(c byte) PacmanType {
	switch c {
	case 'r', 'R':
		return TypeRock
	case 'p', 'P':
		return TypePaper
	case 's', 'S':
		return TypeScissors
	case 'n', 'N':
		return TypeNeutral
	}
	panic(string(c) + " is not a valid pac type")
}

// PacmanTypeFromName maps a name like "ROCK"/"PAPER"/"SCISSORS" to PacmanType.
// Returns (0, false) on unknown input. Go-only inverse of Name().
func PacmanTypeFromName(s string) (PacmanType, bool) {
	switch s {
	case "ROCK":
		return TypeRock, true
	case "PAPER":
		return TypePaper, true
	case "SCISSORS":
		return TypeScissors, true
	case "NEUTRAL":
		return TypeNeutral, true
	}
	return 0, false
}
