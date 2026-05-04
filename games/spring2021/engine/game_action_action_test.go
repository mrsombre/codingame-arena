package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoActionIsActionNone(t *testing.T) {
	assert.Equal(t, ActionNone, NoAction.Kind)
	assert.False(t, NoAction.IsGrow())
	assert.False(t, NoAction.IsSeed())
	assert.False(t, NoAction.IsComplete())
	assert.False(t, NoAction.IsWait())
}

func TestActionConstructors(t *testing.T) {
	g := NewGrowAction(5)
	assert.True(t, g.IsGrow())
	assert.Equal(t, 5, g.GetTargetID())

	c := NewCompleteAction(7)
	assert.True(t, c.IsComplete())
	assert.Equal(t, 7, c.GetTargetID())

	w := NewWaitAction()
	assert.True(t, w.IsWait())

	s := NewSeedAction(2, 8)
	assert.True(t, s.IsSeed())
	assert.Equal(t, 2, s.GetSourceID())
	assert.Equal(t, 8, s.GetTargetID())
}

func TestActionFlagsAreMutuallyExclusive(t *testing.T) {
	cases := []struct {
		name string
		a    Action
	}{
		{"grow", NewGrowAction(0)},
		{"seed", NewSeedAction(0, 1)},
		{"complete", NewCompleteAction(0)},
		{"wait", NewWaitAction()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flags := []bool{tc.a.IsGrow(), tc.a.IsSeed(), tc.a.IsComplete(), tc.a.IsWait()}
			set := 0
			for _, f := range flags {
				if f {
					set++
				}
			}
			assert.Equal(t, 1, set, "exactly one Is* flag is true")
		})
	}
}
