package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/action"
	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

// newParseGame builds a minimal game + player for ParseCommands tests.
func newParseGame(t *testing.T) (*Game, *Player, *CommandManager) {
	t.Helper()
	g := &Game{Config: NewConfig(LeagueRulesFromIndex(4))}
	g.Grid = grid.NewGridFromRows([]string{"#####", "#   #", "#####"}, false)

	player := NewPlayer(0)
	g.Players = []*Player{player, NewPlayer(1)}

	pac0 := NewPacman(0, 0, player, TypeRock)
	pac0.Position = grid.Coord{X: 1, Y: 1}
	pac1 := NewPacman(1, 1, player, TypePaper)
	pac1.Position = grid.Coord{X: 2, Y: 1}
	player.AddPacman(pac0)
	player.AddPacman(pac1)
	g.Pacmen = []*Pacman{pac0, pac1}

	var summary []string
	return g, player, NewCommandManager(&summary, g)
}

func TestParseCommandsMoveSetsIntent(t *testing.T) {
	g, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 0 3 1"})
	assert.False(t, player.IsDeactivated())
	pac := player.Pacmen[0]
	assert.Equal(t, action.ActionMove, pac.Intent.Type)
	assert.Equal(t, grid.Coord{X: 3, Y: 1}, pac.Intent.Target)
	_ = g
}

func TestParseCommandsSpeedSetsAbility(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"SPEED 0"})
	pac := player.Pacmen[0]
	assert.Equal(t, action.ActionSpeed, pac.Intent.Type)
	assert.True(t, pac.HasAbilityToUse)
	assert.Equal(t, AbilitySpeed, pac.AbilityToUse)
}

func TestParseCommandsSwitchSetsAbility(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"SWITCH 0 PAPER"})
	pac := player.Pacmen[0]
	assert.Equal(t, action.ActionSwitch, pac.Intent.Type)
	assert.True(t, pac.HasAbilityToUse)
	assert.Equal(t, AbilitySetPaper, pac.AbilityToUse)
}

func TestParseCommandsSwitchToSameTypeIsNoop(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"SWITCH 0 ROCK"})
	pac := player.Pacmen[0]
	assert.Equal(t, action.ActionSwitch, pac.Intent.Type)
	// Same type → no ability queued.
	assert.False(t, pac.HasAbilityToUse)
	assert.Equal(t, AbilityUnset, pac.AbilityToUse)
}

func TestParseCommandsWaitLeavesIntentNoop(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"WAIT 0"})
	pac := player.Pacmen[0]
	assert.True(t, pac.Intent.IsNoAction())
	assert.False(t, player.IsDeactivated())
}

func TestParseCommandsMultipleCommandsPipeSeparated(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 0 3 1 | SPEED 1"})
	assert.False(t, player.IsDeactivated())
	assert.Equal(t, action.ActionMove, player.Pacmen[0].Intent.Type)
	assert.Equal(t, action.ActionSpeed, player.Pacmen[1].Intent.Type)
	assert.True(t, player.Pacmen[1].HasAbilityToUse)
}

func TestParseCommandsMessageAttached(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 0 3 1 hello there"})
	pac := player.Pacmen[0]
	assert.Equal(t, "hello there", pac.Message)
	assert.True(t, pac.HasMsg)
}

func TestParseCommandsEmptyOutputDeactivatesAsTimeout(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, nil)
	assert.True(t, player.IsDeactivated())
	assert.Equal(t, "Timeout!", player.DeactivationReason())
}

func TestParseCommandsUnknownCommandDeactivates(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"DANCE 0"})
	assert.True(t, player.IsDeactivated())
}

func TestParseCommandsMoveOutOfGridDeactivates(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 0 -1 1"})
	assert.True(t, player.IsDeactivated())

	// Also y out of range, fresh player.
	_, player2, cm2 := newParseGame(t)
	cm2.ParseCommands(player2, []string{"MOVE 0 1 9"})
	assert.True(t, player2.IsDeactivated())
}

func TestParseCommandsUnknownPacDeactivates(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 99 1 1"})
	assert.True(t, player.IsDeactivated())
	assert.Contains(t, player.DeactivationReason(), "doesn't exist")
}

func TestParseCommandsDeadPacIsSilentlySkipped(t *testing.T) {
	_, player, cm := newParseGame(t)
	player.Pacmen[0].Dead = true
	cm.ParseCommands(player, []string{"MOVE 0 3 1"})
	assert.False(t, player.IsDeactivated())
	assert.True(t, player.Pacmen[0].Intent.IsNoAction())
}

func TestParseCommandsTwiceCommandedDeactivates(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"MOVE 0 3 1 | MOVE 0 2 1"})
	assert.True(t, player.IsDeactivated())
	assert.Contains(t, player.DeactivationReason(), "cannot be commanded twice")
}

func TestParseCommandsIsCaseInsensitive(t *testing.T) {
	_, player, cm := newParseGame(t)
	cm.ParseCommands(player, []string{"move 0 3 1"})
	assert.False(t, player.IsDeactivated())
	assert.Equal(t, action.ActionMove, player.Pacmen[0].Intent.Type)
}

func TestExpectedMessageReflectsLeague(t *testing.T) {
	// League 2: no SPEED or SWITCH.
	g := &Game{Config: NewConfig(LeagueRulesFromIndex(2))}
	g.Grid = grid.NewGridFromRows([]string{"#####", "#   #", "#####"}, false)
	cm := NewCommandManager(nil, g)
	assert.Equal(t, "MOVE <id> <x> <y>", cm.expected())

	// League 4: full.
	g4 := &Game{Config: NewConfig(LeagueRulesFromIndex(4))}
	g4.Grid = g.Grid
	cm4 := NewCommandManager(nil, g4)
	assert.Contains(t, cm4.expected(), "SPEED")
	assert.Contains(t, cm4.expected(), "SWITCH")
}
