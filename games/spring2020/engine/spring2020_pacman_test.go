package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacmanTypeName(t *testing.T) {
	assert.Equal(t, "ROCK", TypeRock.Name())
	assert.Equal(t, "PAPER", TypePaper.Name())
	assert.Equal(t, "SCISSORS", TypeScissors.Name())
	assert.Equal(t, "NEUTRAL", TypeNeutral.Name())
	assert.Equal(t, "NEUTRAL", PacmanType(99).Name())
}

func TestPacmanTypeFromCharacter(t *testing.T) {
	pairs := map[byte]PacmanType{
		'r': TypeRock, 'R': TypeRock,
		'p': TypePaper, 'P': TypePaper,
		's': TypeScissors, 'S': TypeScissors,
		'n': TypeNeutral, 'N': TypeNeutral,
	}
	for c, want := range pairs {
		assert.Equal(t, want, PacmanTypeFromCharacter(c), string(c))
	}
	assert.Panics(t, func() { PacmanTypeFromCharacter('Z') })
}

func TestPacmanTypeFromName(t *testing.T) {
	cases := map[string]PacmanType{
		"ROCK":     TypeRock,
		"PAPER":    TypePaper,
		"SCISSORS": TypeScissors,
		"NEUTRAL":  TypeNeutral,
	}
	for s, want := range cases {
		got, ok := PacmanTypeFromName(s)
		assert.True(t, ok, s)
		assert.Equal(t, want, got, s)
	}

	_, ok := PacmanTypeFromName("rock")
	assert.False(t, ok, "case sensitive")

	_, ok = PacmanTypeFromName("UNKNOWN")
	assert.False(t, ok)
}

func TestAbilityFromSwitchType(t *testing.T) {
	assert.Equal(t, AbilitySetRock, AbilityFromSwitchType(TypeRock))
	assert.Equal(t, AbilitySetPaper, AbilityFromSwitchType(TypePaper))
	assert.Equal(t, AbilitySetScissors, AbilityFromSwitchType(TypeScissors))
	assert.Equal(t, AbilityUnset, AbilityFromSwitchType(TypeNeutral))
}

func TestPacmanSetMessageTruncates(t *testing.T) {
	p := &Pacman{}
	p.SetMessage("")
	assert.False(t, p.HasMsg)
	assert.Equal(t, "", p.Message)

	p.SetMessage("short")
	assert.True(t, p.HasMsg)
	assert.Equal(t, "short", p.Message)

	long := strings.Repeat("x", 60)
	p.SetMessage(long)
	// Java truncates to 46 + "...".
	assert.Equal(t, 49, len(p.Message))
	assert.True(t, strings.HasSuffix(p.Message, "..."))
}

func TestPacmanTickAbilityDurationSetsEndOfSpeed(t *testing.T) {
	p := NewPacman(0, 0, nil, TypeRock)
	p.AbilityDuration = 2

	p.TickAbilityDuration()
	assert.Equal(t, 1, p.AbilityDuration)
	assert.False(t, p.EndOfSpeed)

	p.TickAbilityDuration()
	assert.Equal(t, 0, p.AbilityDuration)
	assert.True(t, p.EndOfSpeed)

	// Next tick should clear EndOfSpeed.
	p.TickAbilityDuration()
	assert.Equal(t, 0, p.AbilityDuration)
	assert.False(t, p.EndOfSpeed)
}

func TestPacmanTickAbilityCooldownFloorsAtZero(t *testing.T) {
	p := NewPacman(0, 0, nil, TypeRock)
	p.AbilityCooldown = 1
	p.TickAbilityCooldown()
	assert.Equal(t, 0, p.AbilityCooldown)
	p.TickAbilityCooldown()
	assert.Equal(t, 0, p.AbilityCooldown)
}

func TestPacmanTurnResetSkipsTicksWhenDead(t *testing.T) {
	p := NewPacman(0, 0, nil, TypeRock)
	p.AbilityCooldown = 5
	p.AbilityDuration = 3
	p.Dead = true

	p.TurnReset()
	assert.Equal(t, 5, p.AbilityCooldown)
	assert.Equal(t, 3, p.AbilityDuration)
}

func TestPacmanTurnResetClearsIntentAndPerTurnFlags(t *testing.T) {
	p := NewPacman(0, 0, nil, TypeRock)
	p.Intent = NewSpeedAction()
	p.AbilityToUse = AbilitySpeed
	p.HasAbilityToUse = true
	p.Blocked = true
	p.CurrentPathStep = 5
	p.PreviousPathStep = 4
	p.HasMsg = true
	p.Message = "hello"

	p.TurnReset()

	assert.True(t, p.Intent.IsNoAction())
	assert.Equal(t, AbilityUnset, p.AbilityToUse)
	assert.False(t, p.HasAbilityToUse)
	assert.False(t, p.Blocked)
	assert.Equal(t, 0, p.CurrentPathStep)
	assert.Equal(t, 0, p.PreviousPathStep)
	assert.False(t, p.HasMsg)
	assert.Equal(t, "", p.Message)
}

func TestPacmanMoveFinished(t *testing.T) {
	p := NewPacman(0, 0, nil, TypeRock)
	p.IntendedPath = []Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}

	p.CurrentPathStep = 0
	assert.False(t, p.MoveFinished())

	p.CurrentPathStep = 1
	assert.False(t, p.MoveFinished())

	p.CurrentPathStep = 2
	assert.True(t, p.MoveFinished())
}

func TestPacmanFastEnoughAndIsSpeeding(t *testing.T) {
	cfg := NewConfig(LeagueRulesFromIndex(4))
	p := NewPacman(0, 0, nil, TypeRock)

	// Base speed=1 → can only move at step 0.
	assert.True(t, p.FastEnoughToMoveAt(0))
	assert.False(t, p.FastEnoughToMoveAt(1))
	assert.False(t, p.IsSpeeding(&cfg))

	p.Speed = cfg.SPEED_BOOST
	assert.True(t, p.FastEnoughToMoveAt(0))
	assert.True(t, p.FastEnoughToMoveAt(1))
	assert.False(t, p.FastEnoughToMoveAt(2))
	assert.True(t, p.IsSpeeding(&cfg))
}
