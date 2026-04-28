// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/LeagueRules.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/LeagueRules.java:3-13

public class LeagueRules {
    public int numberOfCherries = 4;
    public boolean forOfWar = true;
    public boolean mapWraps = true;
    public boolean bodyBlock = true;
    public boolean friendlyBodyBlock = true;
    public boolean speedAbilityAvailable = true;
    public boolean switchAbilityAvailable = true;
    public int minPacsPerPlayer = 2;
    public int maxPacsPerPlayer = 5;
    public boolean provideDeadPacs = true;
}
*/

// LeagueRules describes per-league toggles.
type LeagueRules struct {
	NUMBER_OF_CHERRIES       int
	FOG_OF_WAR               bool
	MAP_WRAPS                bool
	BODY_BLOCK               bool
	FRIENDLY_BODY_BLOCK      bool
	SPEED_ABILITY_AVAILABLE  bool
	SWITCH_ABILITY_AVAILABLE bool
	MIN_PACS_PER_PLAYER      int
	MAX_PACS_PER_PLAYER      int
	PROVIDE_DEAD_PACS        bool
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/LeagueRules.java:15-32

public static LeagueRules fromIndex(int index) {
    LeagueRules rules = new LeagueRules();
    if (index == 1) {
        rules.minPacsPerPlayer = 1;
        rules.maxPacsPerPlayer = 1;
    }
    if (index <= 2) {
        rules.forOfWar = false;
        rules.speedAbilityAvailable = false;
        rules.switchAbilityAvailable = false;
    }
    if (index <= 3) {
        rules.provideDeadPacs = false;
    }
    return rules;
}
*/

// LeagueRulesFromIndex mirrors LeagueRules.fromIndex.
// Index 1: single pac, no fog, no abilities.
// Index 2: two pacs, no fog, no abilities.
// Index 3: full game sans dead-pac reporting.
// Index >=4: full rules.
func LeagueRulesFromIndex(index int) LeagueRules {
	r := LeagueRules{
		NUMBER_OF_CHERRIES:       4,
		FOG_OF_WAR:               true,
		MAP_WRAPS:                true,
		BODY_BLOCK:               true,
		FRIENDLY_BODY_BLOCK:      true,
		SPEED_ABILITY_AVAILABLE:  true,
		SWITCH_ABILITY_AVAILABLE: true,
		MIN_PACS_PER_PLAYER:      2,
		MAX_PACS_PER_PLAYER:      5,
		PROVIDE_DEAD_PACS:        true,
	}
	if index == 1 {
		r.MIN_PACS_PER_PLAYER = 1
		r.MAX_PACS_PER_PLAYER = 1
	}
	if index <= 2 {
		r.FOG_OF_WAR = false
		r.SPEED_ABILITY_AVAILABLE = false
		r.SWITCH_ABILITY_AVAILABLE = false
	}
	if index <= 3 {
		r.PROVIDE_DEAD_PACS = false
	}
	return r
}
