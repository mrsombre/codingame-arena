// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Ability.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/PacmanType.java
package engine

import (
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/action"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

// PacmanType matches Java PacmanType enum.
type PacmanType int

const (
	TypeRock     PacmanType = PacmanType(IDRock)
	TypePaper    PacmanType = PacmanType(IDPaper)
	TypeScissors PacmanType = PacmanType(IDScissors)
	TypeNeutral  PacmanType = PacmanType(IDNeutral)
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
// Returns (0, false) on unknown input.
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

// Ability types.
type AbilityType int

const (
	AbilityUnset AbilityType = iota
	AbilitySpeed
	AbilitySetRock
	AbilitySetPaper
	AbilitySetScissors
)

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

// pacTypeFromAbility returns the PacmanType an ability sets the pac to,
// or TypeNeutral for non-switch abilities.
func pacTypeFromAbility(a AbilityType) PacmanType {
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

// Pacman is one playable unit on the grid.
type Pacman struct {
	Owner  *Player
	ID     int
	Number int

	Position grid.Coord
	Type     PacmanType
	Speed    int

	AbilityDuration int
	AbilityCooldown int

	Message string
	HasMsg  bool

	Intent          action.Action
	AbilityToUse    AbilityType
	HasAbilityToUse bool
	IntendedPath    []grid.Coord

	CurrentPathStep  int
	PreviousPathStep int
	Blocked          bool
	Dead             bool
	EndOfSpeed       bool
	WarnPathMsg      string
	HasWarnPathMsg   bool
}

// NewPacman creates a Pacman with the given id/number/type for an owner.
func NewPacman(id, number int, owner *Player, t PacmanType) *Pacman {
	return &Pacman{
		Owner:        owner,
		ID:           id,
		Number:       number,
		Type:         t,
		Speed:        1,
		Intent:       action.NoAction,
		IntendedPath: make([]grid.Coord, 0),
	}
}

// SetMessage truncates to 48 chars like Java Pacman.setMessage.
func (p *Pacman) SetMessage(msg string) {
	if len(msg) > 48 {
		msg = msg[:46] + "..."
	}
	p.Message = msg
	p.HasMsg = msg != ""
}

func (p *Pacman) TickAbilityDuration() {
	if p.AbilityDuration > 0 {
		p.AbilityDuration--
		p.EndOfSpeed = p.AbilityDuration == 0
	} else {
		p.EndOfSpeed = false
	}
}

func (p *Pacman) TickAbilityCooldown() {
	if p.AbilityCooldown > 0 {
		p.AbilityCooldown--
	}
}

// TurnReset clears per-turn state. Called at start of each main turn.
func (p *Pacman) TurnReset() {
	p.Message = ""
	p.HasMsg = false
	if !p.Dead {
		p.TickAbilityDuration()
		p.TickAbilityCooldown()
	}
	p.AbilityToUse = AbilityUnset
	p.HasAbilityToUse = false
	p.CurrentPathStep = 0
	p.PreviousPathStep = 0
	p.Blocked = false
	p.Intent = action.NoAction
}

func (p *Pacman) SetCurrentPathStep(step int) {
	p.PreviousPathStep = p.CurrentPathStep
	p.CurrentPathStep = step
}

func (p *Pacman) MoveFinished() bool {
	return p.CurrentPathStep == len(p.IntendedPath)-1
}

func (p *Pacman) FastEnoughToMoveAt(step int) bool {
	return p.Speed > step
}

func (p *Pacman) IsSpeeding(cfg *Config) bool {
	return p.Speed == cfg.SpeedBoost
}
