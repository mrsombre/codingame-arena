// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/LeagueRules.java
package engine

// Config holds the mutable per-match game configuration.
// Java used static fields; we hold a value-type Config per Game.
type Config struct {
	MapMinWidth   int
	MapMaxWidth   int
	MapMinHeight  int
	MapMaxHeight  int
	PacmanBase    int
	SpeedBoost    int
	AbilityDur    int
	AbilityCool   int
	NumCherries   int
	FogOfWar      bool
	MapWraps      bool
	BodyBlock     bool
	FriendlyBlock bool
	SpeedAvail    bool
	SwitchAvail   bool
	MinPacs       int
	MaxPacs       int
	DeadPacs      bool
}

// CherryScore is a constant in the Java engine.
const CherryScore = 10

// IDs for pacman types (match Java Config.ID_*).
const (
	IDRock     = 0
	IDPaper    = 1
	IDScissors = 2
	IDNeutral  = -1
)

// NewConfig returns the default Config for the given league rules.
func NewConfig(rules LeagueRules) Config {
	return Config{
		MapMinWidth:   28,
		MapMaxWidth:   33,
		MapMinHeight:  10,
		MapMaxHeight:  15,
		PacmanBase:    1,
		SpeedBoost:    2,
		AbilityDur:    6,
		AbilityCool:   10,
		NumCherries:   rules.NumberOfCherries,
		FogOfWar:      rules.FogOfWar,
		MapWraps:      rules.MapWraps,
		BodyBlock:     rules.BodyBlock,
		FriendlyBlock: rules.FriendlyBodyBlock,
		SpeedAvail:    rules.SpeedAbilityAvailable,
		SwitchAvail:   rules.SwitchAbilityAvailable,
		MinPacs:       rules.MinPacsPerPlayer,
		MaxPacs:       rules.MaxPacsPerPlayer,
		DeadPacs:      rules.ProvideDeadPacs,
	}
}

// LeagueRules describes per-league toggles.
type LeagueRules struct {
	NumberOfCherries       int
	FogOfWar               bool
	MapWraps               bool
	BodyBlock              bool
	FriendlyBodyBlock      bool
	SpeedAbilityAvailable  bool
	SwitchAbilityAvailable bool
	MinPacsPerPlayer       int
	MaxPacsPerPlayer       int
	ProvideDeadPacs        bool
}

// LeagueRulesFromIndex mirrors LeagueRules.fromIndex.
// Index 1: single pac, no fog, no abilities.
// Index 2: two pacs, no fog, no abilities.
// Index 3: full game sans dead-pac reporting.
// Index >=4: full rules.
func LeagueRulesFromIndex(index int) LeagueRules {
	r := LeagueRules{
		NumberOfCherries:       4,
		FogOfWar:               true,
		MapWraps:               true,
		BodyBlock:              true,
		FriendlyBodyBlock:      true,
		SpeedAbilityAvailable:  true,
		SwitchAbilityAvailable: true,
		MinPacsPerPlayer:       2,
		MaxPacsPerPlayer:       5,
		ProvideDeadPacs:        true,
	}
	if index == 1 {
		r.MinPacsPerPlayer = 1
		r.MaxPacsPerPlayer = 1
	}
	if index <= 2 {
		r.FogOfWar = false
		r.SpeedAbilityAvailable = false
		r.SwitchAbilityAvailable = false
	}
	if index <= 3 {
		r.ProvideDeadPacs = false
	}
	return r
}
