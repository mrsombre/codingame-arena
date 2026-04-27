// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java:8
public static final Coord[] ADJACENCY = { new Coord(-1, 0), new Coord(1, 0), new Coord(0, -1), new Coord(0, 1) };
*/

// ADJACENCY is the 4-directional neighbour deltas in Java order:
// left, right, up, down. Iteration order matters for parity.
var ADJACENCY = [4]Coord{
	{X: -1, Y: 0},
	{X: 1, Y: 0},
	{X: 0, Y: -1},
	{X: 0, Y: 1},
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java:10-25

public static final int ID_ROCK = 0;
public static final int ID_PAPER = 1;
public static final int ID_SCISSORS = 2;
public static final int ID_NEUTRAL = -1;

public static final int CHERRY_SCORE = 10;

public static int MAP_MIN_WIDTH = 28;
public static int MAP_MAX_WIDTH = 33;
public static int MAP_MIN_HEIGHT = 10;
public static int MAP_MAX_HEIGHT = 15;

public static int PACMAN_BASE_SPEED = 1;
public static int SPEED_BOOST = 2;
public static int ABILITY_DURATION = 6;
public static int ABILITY_COOLDOWN = 10;
*/

// Java Config constants. apply() can override the non-final ones from system
// properties at runtime, but the local engine never calls apply(), so we treat
// them as Go consts.
const (
	CHERRY_SCORE      = 10
	ID_ROCK           = 0
	ID_PAPER          = 1
	ID_SCISSORS       = 2
	ID_NEUTRAL        = -1
	MAP_MIN_WIDTH     = 28
	MAP_MAX_WIDTH     = 33
	MAP_MIN_HEIGHT    = 10
	MAP_MAX_HEIGHT    = 15
	PACMAN_BASE_SPEED = 1
	SPEED_BOOST       = 2
	ABILITY_DURATION  = 6
	ABILITY_COOLDOWN  = 10
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java:27-36

public static int NUMBER_OF_CHERRIES;
public static boolean FOG_OF_WAR;
public static boolean MAP_WRAPS;
public static boolean BODY_BLOCK;
public static boolean FRIENDLY_BODY_BLOCK;
public static boolean SPEED_ABILITY_AVAILABLE;
public static boolean SWITCH_ABILITY_AVAILABLE;
public static int MIN_PACS_PER_PLAYER = 2;
public static int MAX_PACS_PER_PLAYER = 5;
public static boolean PROVIDE_DEAD_PACS;
*/

// Config holds the mutable per-match game configuration.
// Java used static fields; we hold a value-type Config per Game.
type Config struct {
	MAP_MIN_WIDTH            int
	MAP_MAX_WIDTH            int
	MAP_MIN_HEIGHT           int
	MAP_MAX_HEIGHT           int
	PACMAN_BASE_SPEED        int
	SPEED_BOOST              int
	ABILITY_DURATION         int
	ABILITY_COOLDOWN         int
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
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Config.java:38-49

public static void setDefaultValueByLevel(LeagueRules rules) {
    NUMBER_OF_CHERRIES = rules.numberOfCherries;
    FOG_OF_WAR = rules.forOfWar;
    MAP_WRAPS = rules.mapWraps;
    BODY_BLOCK = rules.bodyBlock;
    FRIENDLY_BODY_BLOCK = rules.friendlyBodyBlock;
    SPEED_ABILITY_AVAILABLE = rules.speedAbilityAvailable;
    SWITCH_ABILITY_AVAILABLE = rules.switchAbilityAvailable;
    MIN_PACS_PER_PLAYER = rules.minPacsPerPlayer;
    MAX_PACS_PER_PLAYER = rules.maxPacsPerPlayer;
    PROVIDE_DEAD_PACS = rules.provideDeadPacs;
}
*/

// NewConfig returns the default Config for the given league rules.
func NewConfig(rules LeagueRules) Config {
	return Config{
		MAP_MIN_WIDTH:            MAP_MIN_WIDTH,
		MAP_MAX_WIDTH:            MAP_MAX_WIDTH,
		MAP_MIN_HEIGHT:           MAP_MIN_HEIGHT,
		MAP_MAX_HEIGHT:           MAP_MAX_HEIGHT,
		PACMAN_BASE_SPEED:        PACMAN_BASE_SPEED,
		SPEED_BOOST:              SPEED_BOOST,
		ABILITY_DURATION:         ABILITY_DURATION,
		ABILITY_COOLDOWN:         ABILITY_COOLDOWN,
		NUMBER_OF_CHERRIES:       rules.NUMBER_OF_CHERRIES,
		FOG_OF_WAR:               rules.FOG_OF_WAR,
		MAP_WRAPS:                rules.MAP_WRAPS,
		BODY_BLOCK:               rules.BODY_BLOCK,
		FRIENDLY_BODY_BLOCK:      rules.FRIENDLY_BODY_BLOCK,
		SPEED_ABILITY_AVAILABLE:  rules.SPEED_ABILITY_AVAILABLE,
		SWITCH_ABILITY_AVAILABLE: rules.SWITCH_ABILITY_AVAILABLE,
		MIN_PACS_PER_PLAYER:      rules.MIN_PACS_PER_PLAYER,
		MAX_PACS_PER_PLAYER:      rules.MAX_PACS_PER_PLAYER,
		PROVIDE_DEAD_PACS:        rules.PROVIDE_DEAD_PACS,
	}
}
