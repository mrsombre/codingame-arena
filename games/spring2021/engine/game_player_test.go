package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayerAddAndRemoveSunClampsAtZero(t *testing.T) {
	p := NewPlayer(0)
	p.AddSun(5)
	assert.Equal(t, 5, p.GetSun())
	p.RemoveSun(2)
	assert.Equal(t, 3, p.GetSun())
	p.RemoveSun(10)
	assert.Equal(t, 0, p.GetSun(), "sun cannot go negative")
}

func TestPlayerAddScoreAccumulates(t *testing.T) {
	p := NewPlayer(0)
	p.AddScore(3)
	p.AddScore(7)
	assert.Equal(t, 10, p.GetScore())
}

func TestPlayerSetMessageMarksHasMessage(t *testing.T) {
	p := NewPlayer(0)
	assert.False(t, p.HasMessage)
	p.SetMessage("hello")
	assert.True(t, p.HasMessage)
	assert.Equal(t, "hello", p.Message)
}

func TestPlayerResetClearsMessageAndAction(t *testing.T) {
	p := NewPlayer(0)
	p.SetMessage("hi")
	p.SetAction(NewWaitAction())
	p.Reset()
	assert.Equal(t, "", p.Message)
	assert.False(t, p.HasMessage)
	assert.Equal(t, NoAction, p.GetAction())
}

func TestPlayerNicknameToken(t *testing.T) {
	assert.Equal(t, "$0", NewPlayer(0).NicknameToken())
	assert.Equal(t, "$1", NewPlayer(1).NicknameToken())
}

func TestPlayerIsActiveBeforeAndAfterDeactivate(t *testing.T) {
	p := NewPlayer(0)
	assert.True(t, p.IsActive())
	p.Deactivate("bad input")
	assert.False(t, p.IsActive())
	assert.True(t, p.IsDeactivated())
	assert.Equal(t, "bad input", p.DeactivationReason())
}

func TestPlayerBonusScoreText(t *testing.T) {
	p := NewPlayer(0)
	assert.Equal(t, "", p.GetBonusScoreText(), "no bonus → empty")
	p.SetScore(5)
	p.AddBonusScore(2)
	assert.Equal(t, "3 points and 2 trees", p.GetBonusScoreText())
}
