package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeagueRulesLevel1_SinglePacNoAbilitiesNoFog(t *testing.T) {
	r := LeagueRulesFromIndex(1)
	assert.Equal(t, 1, r.MIN_PACS_PER_PLAYER)
	assert.Equal(t, 1, r.MAX_PACS_PER_PLAYER)
	assert.False(t, r.FOG_OF_WAR)
	assert.False(t, r.SPEED_ABILITY_AVAILABLE)
	assert.False(t, r.SWITCH_ABILITY_AVAILABLE)
	assert.False(t, r.PROVIDE_DEAD_PACS)
	assert.Equal(t, 4, r.NUMBER_OF_CHERRIES)
	assert.True(t, r.BODY_BLOCK)
	assert.True(t, r.FRIENDLY_BODY_BLOCK)
	assert.True(t, r.MAP_WRAPS)
}

func TestLeagueRulesLevel2_MultiPacNoAbilitiesNoFog(t *testing.T) {
	r := LeagueRulesFromIndex(2)
	assert.Equal(t, 2, r.MIN_PACS_PER_PLAYER)
	assert.Equal(t, 5, r.MAX_PACS_PER_PLAYER)
	assert.False(t, r.FOG_OF_WAR)
	assert.False(t, r.SPEED_ABILITY_AVAILABLE)
	assert.False(t, r.SWITCH_ABILITY_AVAILABLE)
	assert.False(t, r.PROVIDE_DEAD_PACS)
}

func TestLeagueRulesLevel3_FullButNoDeadPacs(t *testing.T) {
	r := LeagueRulesFromIndex(3)
	assert.True(t, r.FOG_OF_WAR)
	assert.True(t, r.SPEED_ABILITY_AVAILABLE)
	assert.True(t, r.SWITCH_ABILITY_AVAILABLE)
	assert.False(t, r.PROVIDE_DEAD_PACS)
}

func TestLeagueRulesLevel4_Full(t *testing.T) {
	r := LeagueRulesFromIndex(4)
	assert.True(t, r.FOG_OF_WAR)
	assert.True(t, r.SPEED_ABILITY_AVAILABLE)
	assert.True(t, r.SWITCH_ABILITY_AVAILABLE)
	assert.True(t, r.PROVIDE_DEAD_PACS)
	assert.Equal(t, 2, r.MIN_PACS_PER_PLAYER)
	assert.Equal(t, 5, r.MAX_PACS_PER_PLAYER)
	assert.Equal(t, 4, r.NUMBER_OF_CHERRIES)
}

func TestNewConfigMirrorsRulesAndJavaConstants(t *testing.T) {
	cfg := NewConfig(LeagueRulesFromIndex(4))
	assert.Equal(t, 28, cfg.MAP_MIN_WIDTH)
	assert.Equal(t, 33, cfg.MAP_MAX_WIDTH)
	assert.Equal(t, 10, cfg.MAP_MIN_HEIGHT)
	assert.Equal(t, 15, cfg.MAP_MAX_HEIGHT)
	assert.Equal(t, 1, cfg.PACMAN_BASE_SPEED)
	assert.Equal(t, 2, cfg.SPEED_BOOST)
	assert.Equal(t, 6, cfg.ABILITY_DURATION)
	assert.Equal(t, 10, cfg.ABILITY_COOLDOWN)
	assert.Equal(t, 4, cfg.NUMBER_OF_CHERRIES)
	assert.True(t, cfg.FOG_OF_WAR)
	assert.True(t, cfg.MAP_WRAPS)
	assert.True(t, cfg.BODY_BLOCK)
	assert.True(t, cfg.FRIENDLY_BODY_BLOCK)
	assert.True(t, cfg.SPEED_ABILITY_AVAILABLE)
	assert.True(t, cfg.SWITCH_ABILITY_AVAILABLE)
	assert.True(t, cfg.PROVIDE_DEAD_PACS)
}

func TestCHERRY_SCOREConstant(t *testing.T) {
	assert.Equal(t, 10, CHERRY_SCORE)
}

func TestPacmanTypeIDs(t *testing.T) {
	assert.Equal(t, 0, ID_ROCK)
	assert.Equal(t, 1, ID_PAPER)
	assert.Equal(t, 2, ID_SCISSORS)
	assert.Equal(t, -1, ID_NEUTRAL)
}
