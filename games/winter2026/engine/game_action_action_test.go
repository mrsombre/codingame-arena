package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMoveUp(t *testing.T) {
	a, err := Parse("3 UP")
	assert.NoError(t, err)
	assert.Equal(t, TypeMoveUp, a.Type)
	assert.Equal(t, DirNorth, a.Direction)
	assert.Equal(t, 3, a.BirdID)
	assert.True(t, a.HasBirdID)
	assert.True(t, a.IsMove())
	assert.False(t, a.IsMark())
	assert.False(t, a.HasMessage)
}

func TestParseAllMoveDirections(t *testing.T) {
	cases := map[string]Direction{
		"0 UP":    DirNorth,
		"0 DOWN":  DirSouth,
		"0 LEFT":  DirWest,
		"0 RIGHT": DirEast,
	}
	for raw, want := range cases {
		a, err := Parse(raw)
		assert.NoError(t, err, raw)
		assert.Equal(t, want, a.Direction, raw)
	}
}

func TestParseMoveWithMessage(t *testing.T) {
	a, err := Parse("2 RIGHT hello world")
	assert.NoError(t, err)
	assert.True(t, a.IsMove())
	assert.True(t, a.HasMessage)
	assert.Equal(t, "hello world", a.Message)
}

func TestParseMoveIsCaseInsensitive(t *testing.T) {
	a, err := Parse("1 up")
	assert.NoError(t, err)
	assert.Equal(t, DirNorth, a.Direction)
}

func TestParseMark(t *testing.T) {
	a, err := Parse("MARK 12 3")
	assert.NoError(t, err)
	assert.Equal(t, TypeMark, a.Type)
	assert.True(t, a.IsMark())
	assert.Equal(t, Coord{X: 12, Y: 3}, a.Coord)
}

func TestParseWait(t *testing.T) {
	a, err := Parse("WAIT")
	assert.NoError(t, err)
	assert.Equal(t, TypeWait, a.Type)
	assert.False(t, a.IsMove())
	assert.False(t, a.IsMark())
}

func TestParseInvalidReturnsActionError(t *testing.T) {
	_, err := Parse("DANCE")
	assert.Error(t, err)
	var aerr *ActionError
	assert.ErrorAs(t, err, &aerr)
}

func TestParseRejectsLeadingGarbage(t *testing.T) {
	// Pattern must match the full string.
	_, err := Parse("FOO 0 UP")
	assert.Error(t, err)
}

func TestActionStringIncludesFields(t *testing.T) {
	a := Action{Type: TypeMoveUp, Direction: DirNorth, BirdID: 7}
	s := a.String()
	assert.Contains(t, s, "direction=N")
	assert.Contains(t, s, "birdId=7")
}
