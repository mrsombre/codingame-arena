package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeagueRulesLevel1_SinglePacNoAbilitiesNoFog(t *testing.T) {
	r := LeagueRulesFromIndex(1)
	assert.Equal(t, 1, r.MinPacsPerPlayer)
	assert.Equal(t, 1, r.MaxPacsPerPlayer)
	assert.False(t, r.FogOfWar)
	assert.False(t, r.SpeedAbilityAvailable)
	assert.False(t, r.SwitchAbilityAvailable)
	assert.False(t, r.ProvideDeadPacs)
	assert.Equal(t, 4, r.NumberOfCherries)
	assert.True(t, r.BodyBlock)
	assert.True(t, r.FriendlyBodyBlock)
	assert.True(t, r.MapWraps)
}

func TestLeagueRulesLevel2_MultiPacNoAbilitiesNoFog(t *testing.T) {
	r := LeagueRulesFromIndex(2)
	assert.Equal(t, 2, r.MinPacsPerPlayer)
	assert.Equal(t, 5, r.MaxPacsPerPlayer)
	assert.False(t, r.FogOfWar)
	assert.False(t, r.SpeedAbilityAvailable)
	assert.False(t, r.SwitchAbilityAvailable)
	assert.False(t, r.ProvideDeadPacs)
}

func TestLeagueRulesLevel3_FullButNoDeadPacs(t *testing.T) {
	r := LeagueRulesFromIndex(3)
	assert.True(t, r.FogOfWar)
	assert.True(t, r.SpeedAbilityAvailable)
	assert.True(t, r.SwitchAbilityAvailable)
	assert.False(t, r.ProvideDeadPacs)
}

func TestLeagueRulesLevel4_Full(t *testing.T) {
	r := LeagueRulesFromIndex(4)
	assert.True(t, r.FogOfWar)
	assert.True(t, r.SpeedAbilityAvailable)
	assert.True(t, r.SwitchAbilityAvailable)
	assert.True(t, r.ProvideDeadPacs)
	assert.Equal(t, 2, r.MinPacsPerPlayer)
	assert.Equal(t, 5, r.MaxPacsPerPlayer)
	assert.Equal(t, 4, r.NumberOfCherries)
}

func TestNewConfigMirrorsRulesAndJavaConstants(t *testing.T) {
	cfg := NewConfig(LeagueRulesFromIndex(4))
	assert.Equal(t, 28, cfg.MapMinWidth)
	assert.Equal(t, 33, cfg.MapMaxWidth)
	assert.Equal(t, 10, cfg.MapMinHeight)
	assert.Equal(t, 15, cfg.MapMaxHeight)
	assert.Equal(t, 1, cfg.PacmanBase)
	assert.Equal(t, 2, cfg.SpeedBoost)
	assert.Equal(t, 6, cfg.AbilityDur)
	assert.Equal(t, 10, cfg.AbilityCool)
	assert.Equal(t, 4, cfg.NumCherries)
	assert.True(t, cfg.FogOfWar)
	assert.True(t, cfg.MapWraps)
	assert.True(t, cfg.BodyBlock)
	assert.True(t, cfg.FriendlyBlock)
	assert.True(t, cfg.SpeedAvail)
	assert.True(t, cfg.SwitchAvail)
	assert.True(t, cfg.DeadPacs)
}

func TestCherryScoreConstant(t *testing.T) {
	assert.Equal(t, 10, CherryScore)
}

func TestPacTypeIDs(t *testing.T) {
	assert.Equal(t, 0, IDRock)
	assert.Equal(t, 1, IDPaper)
	assert.Equal(t, 2, IDScissors)
	assert.Equal(t, -1, IDNeutral)
}
