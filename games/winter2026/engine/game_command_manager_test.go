package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// newCmdPlayer returns a player owning a single bird facing UP, ready to
// receive commands.
func newCmdPlayer(birdID int) (*Player, *Bird) {
	p := NewPlayer(0)
	p.Init()
	b := NewBird(birdID, p)
	// Body with head NORTH of neck → Facing() = DirNorth.
	b.Body = []Coord{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 7}}
	p.Birds = []*Bird{b}
	return p, b
}

// newCmdGame returns a Game wrapping the player+bird from newCmdPlayer so
// command-trace assertions can read g.traces. Side 1 gets a placeholder
// player so tracePlayer's slot bounds check passes.
func newCmdGame(birdID int) (*Game, *Player, *Bird) {
	p, b := newCmdPlayer(birdID)
	g := &Game{}
	g.Players = []*Player{p, NewPlayer(1)}
	return g, p, b
}

// firstTrace returns the first trace of the given type emitted into the
// player slot at side, or fails the test if none was found.
func firstTrace(t *testing.T, g *Game, side int, typ string) arena.TurnTrace {
	t.Helper()
	for _, tr := range g.traces[side] {
		if tr.Type == typ {
			return tr
		}
	}
	t.Fatalf("no %q trace in slot %d", typ, side)
	return arena.TurnTrace{}
}

func TestParseCommandsMoveSetsDirection(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT"})

	assert.False(t, p.IsDeactivated())
	assert.Equal(t, DirEast, b.Direction)
	assert.True(t, b.HasMove)
}

func TestParseCommandsMoveWithMessage(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 LEFT hey"})

	assert.Equal(t, DirWest, b.Direction)
	assert.Equal(t, "hey", b.Message)
	assert.True(t, b.HasMessage())
}

func TestParseCommandsSemicolonSeparated(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT;MARK 3 4"})

	assert.False(t, p.IsDeactivated())
	assert.Equal(t, DirEast, b.Direction)
	assert.Len(t, p.Marks, 1)
	assert.Equal(t, Coord{X: 3, Y: 4}, p.Marks[0])
}

func TestParseCommandsInvalidSyntaxDeactivates(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"DANCE"})

	assert.True(t, p.IsDeactivated())
	assert.Equal(t, -1, p.GetScore())
}

func TestParseCommandsEmptyOutputIsTimeout(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, nil)

	assert.True(t, p.IsDeactivated())
	assert.Equal(t, "Timeout!", p.DeactivationReason())
	assert.True(t, p.IsTimedOut())
}

func TestParseCommandsBackwardsMoveRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)
	// Bird faces North; DOWN is its opposite → should be rejected.

	cm.ParseCommands(p, []string{"0 DOWN"})

	assert.False(t, p.IsDeactivated(), "backwards is a soft error, not a kick")
	assert.False(t, b.HasMove, "bird is not given a new move")
	assert.Equal(t, DirUnset, b.Direction)
	assert.NotEmpty(t, summary, "summary records the error")
}

func TestParseCommandsDoubleMoveOnSameBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"0 RIGHT;0 LEFT"})

	// First command succeeds — second is rejected (bird already has a move).
	assert.Equal(t, DirEast, b.Direction)
	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary)
}

func TestParseCommandsUnknownBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"99 UP"})

	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary)
}

func TestParseCommandsDeadBirdRejected(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, b := newCmdPlayer(0)
	b.Alive = false

	cm.ParseCommands(p, []string{"0 RIGHT"})

	assert.False(t, b.HasMove)
	assert.NotEmpty(t, summary)
}

func TestParseCommandsMarkOverflowReportsError(t *testing.T) {
	var summary []string
	cm := NewCommandManager(&summary, nil)
	p, _ := newCmdPlayer(0)

	cm.ParseCommands(p, []string{"MARK 0 0;MARK 1 0;MARK 2 0;MARK 3 0;MARK 4 0"})

	assert.Len(t, p.Marks, 4, "only first four accepted")
	assert.False(t, p.IsDeactivated())
	assert.NotEmpty(t, summary, "overflow mark recorded as an error")
}

func TestSplitCommandsDropsTrailingEmpty(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, splitCommands("a;b;"))
	assert.Equal(t, []string{"a"}, splitCommands("a;;"))
	assert.Equal(t, []string{""}, splitCommands(""))
}

func TestEscapeHTMLEntities(t *testing.T) {
	assert.Equal(t, "<b>x</b>", EscapeHTMLEntities("&lt;b&gt;x&lt;/b&gt;"))
}

func TestParseCommandsEmitsMoveTrace(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)

	cm.ParseCommands(p, []string{"0 RIGHT"})

	tr := firstTrace(t, g, 0, TraceMove)
	var meta MoveMeta
	require := assert.New(t)
	require.NoError(json.Unmarshal(tr.Data, &meta))
	require.Equal(0, meta.Bird)
	require.Equal("RIGHT", meta.Direction)
	require.Empty(meta.Debug, "bare move command carries no debug field")
	assert.Empty(t, g.traces[1], "side 0's command must not leak to side 1")
}

func TestParseCommandsEmitsMoveTraceWithDebug(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)

	cm.ParseCommands(p, []string{"0 LEFT hello world"})

	tr := firstTrace(t, g, 0, TraceMove)
	var meta MoveMeta
	assert.NoError(t, json.Unmarshal(tr.Data, &meta))
	assert.Equal(t, "LEFT", meta.Direction)
	assert.Equal(t, "hello world", meta.Debug)
}

func TestParseCommandsEmitsWaitTrace(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)

	cm.ParseCommands(p, []string{"WAIT"})

	tr := firstTrace(t, g, 0, TraceWait)
	assert.Empty(t, tr.Data, "WAIT trace carries no data field")
}

func TestParseCommandsEmitsMarkTrace(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)

	cm.ParseCommands(p, []string{"MARK 3 7;MARK 12 4"})

	var marks []MarkMeta
	for _, tr := range g.traces[0] {
		if tr.Type != TraceMark {
			continue
		}
		var m MarkMeta
		assert.NoError(t, json.Unmarshal(tr.Data, &m))
		marks = append(marks, m)
	}
	assert.Equal(t, []MarkMeta{
		{Coord: [2]int{3, 7}}, {Coord: [2]int{12, 4}},
	}, marks)
}

func TestParseCommandsRejectedMoveEmitsNoTrace(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)
	// Bird faces NORTH; DOWN is its opposite — a soft-rejected move
	// must produce no MOVE trace (parity with spring2020's "errored
	// commands drop the message entirely" rule).

	cm.ParseCommands(p, []string{"0 DOWN"})

	for _, tr := range g.traces[0] {
		assert.NotEqual(t, TraceMove, tr.Type, "rejected backwards move emitted a MOVE trace")
	}
}

func TestParseCommandsRejectedMarkEmitsNoTrace(t *testing.T) {
	g, p, _ := newCmdGame(0)
	var summary []string
	cm := NewCommandManager(&summary, g)
	// 5 MARKs in one turn — engine accepts the first four, the 5th is
	// soft-rejected and must emit no MARK trace.

	cm.ParseCommands(p, []string{"MARK 0 0;MARK 1 0;MARK 2 0;MARK 3 0;MARK 4 0"})

	count := 0
	for _, tr := range g.traces[0] {
		if tr.Type == TraceMark {
			count++
		}
	}
	assert.Equal(t, 4, count, "only the four accepted marks emit traces")
}

func TestParseCommandsCommandTraceOrderMatchesInput(t *testing.T) {
	// Multi-bird side, mirroring the bot output sample from replay
	// 886499214: `0 LEFT (0);MARK 0 10;1 LEFT (1)` should produce traces
	// in command-issue order, not bird-id order.
	g := &Game{}
	p := NewPlayer(0)
	p.Init()
	for id := 0; id < 2; id++ {
		b := NewBird(id, p)
		b.Body = []Coord{{X: 5, Y: 5 + id*4}, {X: 5, Y: 6 + id*4}, {X: 5, Y: 7 + id*4}}
		p.Birds = append(p.Birds, b)
	}
	g.Players = []*Player{p, NewPlayer(1)}
	var summary []string
	cm := NewCommandManager(&summary, g)

	cm.ParseCommands(p, []string{"0 LEFT (0);MARK 0 10;1 LEFT (1)"})

	types := make([]string, 0, len(g.traces[0]))
	for _, tr := range g.traces[0] {
		types = append(types, tr.Type)
	}
	assert.Equal(t, []string{TraceMove, TraceMark, TraceMove}, types)
}
