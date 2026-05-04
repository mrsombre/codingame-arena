// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Config.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Config.java:6-20

public class Config {

    public static int STARTING_SUN = 0;
    public static int MAP_RING_COUNT = 3;
    public static int STARTING_NUTRIENTS = 20;
    public static int MAX_ROUNDS = 24;
    public static int MAX_EMPTY_CELLS = 10;
}
*/

// Config holds per-match tunables. The Java version uses static fields and a
// load(Properties) hook; we keep the same defaults on a per-Game value so
// repeated simulations don't leak state between games.
type Config struct {
	STARTING_SUN       int
	MAP_RING_COUNT     int
	STARTING_NUTRIENTS int
	MAX_ROUNDS         int
	MAX_EMPTY_CELLS    int
}

func NewConfig() Config {
	return Config{
		STARTING_SUN:       0,
		MAP_RING_COUNT:     3,
		STARTING_NUTRIENTS: 20,
		MAX_ROUNDS:         24,
		MAX_EMPTY_CELLS:    10,
	}
}
