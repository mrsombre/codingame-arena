package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
)

// newCmdPlayer returns a player owning a single bird facing UP, ready to
// receive commands.
func newCmdPlayer(birdID int) (*Player, *Bird) {
	p := NewPlayer(0)
	p.Init()
	b := NewBird(birdID, p)
	// Body with head NORTH of neck → Facing() = DirNorth.
	b.Body = []grid.Coord{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 7}}
	p.birds = []*Bird{b}
	return p, b
}

func TestParseCommandsMoveSetsDirection(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT"})

	assert.False(t, p.IsDeactivated())
	assert.Equal(t, grid.DirEast, b.Direction)
	assert.True(t, b.HasMove)
}

func TestParseCommandsMoveWithMessage(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 LEFT hey"})

	assert.Equal(t, grid.DirWest, b.Direction)
	assert.Equal(t, "hey", b.Message)
	assert.True(t, b.HasMessage())
}

func TestParseCommandsSemicolonSeparated(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT;MARK 3 4"})

	assert.False(t, p.IsDeactivated())
	assert.Equal(t, grid.DirEast, b.Direction)
	assert.Len(t, p.marks, 1)
	assert.Equal(t, grid.Coord{X: 3, Y: 4}, p.marks[0])
}

func TestParseCommandsInvalidSyntaxDeactivates(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"DANCE"})

	assert.True(t, p.IsDeactivated())
	assert.Equal(t, -1, p.GetScore())
}

func TestParseCommandsEmptyOutputIsTimeout(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, nil)

	assert.True(t, p.IsDeactivated())
	assert.Equal(t, "Timeout!", p.DeactivationReason())
	assert.True(t, p.IsTimedOut())
}

func TestParseCommandsBackwardsMoveRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)
	// Bird faces North; DOWN is its opposite → should be rejected.

	cm.ParseCommands(p, []string{"0 DOWN"})

	assert.False(t, p.IsDeactivated(), "backwards is a soft error, not a kick")
	assert.False(t, b.HasMove, "bird is not given a new move")
	assert.Equal(t, grid.DirUnset, b.Direction)
	assert.NotEmpty(t, summary, "summary records the error")
}

func TestParseCommandsDoubleMoveOnSameBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT;0 LEFT"})

	// First command succeeds — second is rejected (bird already has a move).
	assert.Equal(t, grid.DirEast, b.Direction)
	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary)
}

func TestParseCommandsUnknownBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"99 UP"})

	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary)
}

func TestParseCommandsDeadBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, b := newCmdPlayer(0)
	b.Alive = false

	cm.ParseCommands(p, []string{"0 RIGHT"})

	assert.False(t, b.HasMove)
	assert.NotEmpty(t, summary)
}

func TestParseCommandsMarkOverflowReportsError(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"MARK 0 0;MARK 1 0;MARK 2 0;MARK 3 0;MARK 4 0"})

	assert.Len(t, p.marks, 4, "only first four accepted")
	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary, "overflow mark recorded as an error")
}

func TestSplitCommandsDropsTrailingEmpty(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, splitCommands("a;b;"))
	assert.Equal(t, []string{"a"}, splitCommands("a;;"))
	assert.Equal(t, []string{""}, splitCommands(""))
}

func TestEscapeHTMLEntities(t *testing.T) {
	assert.Equal(t, "<b>x</b>", escapeHTMLEntities("&lt;b&gt;x&lt;/b&gt;"))
}
