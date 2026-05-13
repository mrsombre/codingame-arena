// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Constants.java
package engine

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Constants.java:3-32

public class Constants {
    public static final int TIME_PER_TURN = 50;
    public static final int GAME_TURNS = 300;
    public static final int GAME_TURNS_LOW_LEAGUE = 100;
    public static final int STALL_LIMIT = 10;

    public static final int PLANT_MAX_SIZE = 4;
    public static final int PLANT_MAX_RESOURCES = 3;
    public static final int[] PLANT_COOLDOWN = {8, 8, 9, 6};
    public static final int[] PLANT_WATER_COOLDOWN_BOOST = {5, 5, 7, 2};
    public static final int[] PLANT_FINAL_HEALTH = {12, 12, 20, 6};
    public static final int[] PLANT_DELTA_HEALTH = {2, 2, 3, 1};

    public static final int MAP_MIN_HEIGHT = 8;
    public static final int MAP_MAX_HEIGHT = 11;
    public static final int MAP_MIN_RIVER = 2;
    public static final int MAP_MAX_RIVER = 3;
    public static final int MAP_MIN_IRON = 1;
    public static final int MAP_MAX_IRON = 2;
    public static final int MAP_MIN_ROCK = 1;
    public static final int MAP_MAX_ROCK = 10;
    public static final int MAP_MIN_TREE = 1;
    public static final int MAP_MAX_TREE = 3;
    public static final int MAP_MAX_OPP_DIST = 16;

    public static final int MIN_STARTING_RESOURCE = 2;
    public static final int MAX_STARTING_RESOURCE = 10;

    public static final int WOOD_POINTS = 4;
}
*/

const (
	TIME_PER_TURN         = 50
	GAME_TURNS            = 300
	GAME_TURNS_LOW_LEAGUE = 100
	STALL_LIMIT           = 10

	PLANT_MAX_SIZE      = 4
	PLANT_MAX_RESOURCES = 3

	MAP_MIN_HEIGHT   = 8
	MAP_MAX_HEIGHT   = 11
	MAP_MIN_RIVER    = 2
	MAP_MAX_RIVER    = 3
	MAP_MIN_IRON     = 1
	MAP_MAX_IRON     = 2
	MAP_MIN_ROCK     = 1
	MAP_MAX_ROCK     = 10
	MAP_MIN_TREE     = 1
	MAP_MAX_TREE     = 3
	MAP_MAX_OPP_DIST = 16

	MIN_STARTING_RESOURCE = 2
	MAX_STARTING_RESOURCE = 10

	WOOD_POINTS = 4
)

// Per-tree-type tables. Indexed by Item ordinal for PLUM/LEMON/APPLE/BANANA.
var (
	PLANT_COOLDOWN              = [4]int{8, 8, 9, 6}
	PLANT_WATER_COOLDOWN_BOOST  = [4]int{5, 5, 7, 2}
	PLANT_FINAL_HEALTH          = [4]int{12, 12, 20, 6}
	PLANT_DELTA_HEALTH          = [4]int{2, 2, 3, 1}
)
