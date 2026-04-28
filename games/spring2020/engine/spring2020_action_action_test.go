package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoActionSentinel(t *testing.T) {
	assert.True(t, NoAction.IsNoAction())
	assert.Equal(t, ActionWait, NoAction.Type)
}

func TestNewMoveAction(t *testing.T) {
	a := NewMoveAction(Coord{X: 3, Y: 4})
	assert.Equal(t, ActionMove, a.Type)
	assert.Equal(t, Coord{X: 3, Y: 4}, a.Target)
	assert.False(t, a.IsNoAction())
}

func TestNewSpeedAction(t *testing.T) {
	a := NewSpeedAction()
	assert.Equal(t, ActionSpeed, a.Type)
	assert.False(t, a.IsNoAction())
}

func TestNewSwitchAction(t *testing.T) {
	a := NewSwitchAction(TypePaper)
	assert.Equal(t, ActionSwitch, a.Type)
	assert.Equal(t, TypePaper, a.NewType)
	assert.False(t, a.IsNoAction())
}
