package action

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

func TestNoActionSentinel(t *testing.T) {
	assert.True(t, NoAction.IsNoAction())
	assert.Equal(t, ActionWait, NoAction.Type)
}

func TestNewMoveAction(t *testing.T) {
	a := NewMoveAction(grid.Coord{X: 3, Y: 4})
	assert.Equal(t, ActionMove, a.Type)
	assert.Equal(t, grid.Coord{X: 3, Y: 4}, a.Target)
	assert.False(t, a.IsNoAction())
}

func TestNewSpeedAction(t *testing.T) {
	a := NewSpeedAction()
	assert.Equal(t, ActionSpeed, a.Type)
	assert.False(t, a.IsNoAction())
}

func TestNewSwitchAction(t *testing.T) {
	a := NewSwitchAction(PacPaper)
	assert.Equal(t, ActionSwitch, a.Type)
	assert.Equal(t, PacPaper, a.NewType)
	assert.False(t, a.IsNoAction())
}
