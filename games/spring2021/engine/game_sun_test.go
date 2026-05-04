package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSunSetOrientationModSix(t *testing.T) {
	s := Sun{}
	s.SetOrientation(7)
	assert.Equal(t, 1, s.Orientation)
	s.SetOrientation(12)
	assert.Equal(t, 0, s.Orientation)
}

func TestSunMoveIncrementsAndWraps(t *testing.T) {
	s := Sun{Orientation: 5}
	s.Move()
	assert.Equal(t, 0, s.Orientation, "wraps after 5")
	for i := 0; i < 6; i++ {
		s.Move()
	}
	assert.Equal(t, 0, s.Orientation, "6 moves = full revolution")
}
