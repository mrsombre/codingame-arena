// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Constants.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Constants.java:1-25

public class Constants {

    public static final int RICHNESS_NULL = 0;
    public static final int RICHNESS_POOR = 1;
    public static final int RICHNESS_OK = 2;
    public static final int RICHNESS_LUSH = 3;

    public static final int TREE_SEED = 0;
    public static final int TREE_SMALL = 1;
    public static final int TREE_MEDIUM = 2;
    public static final int TREE_TALL = 3;

    public static final int[] TREE_BASE_COST = new int[] { 0, 1, 3, 7};
    public static final int TREE_COST_SCALE = 1;
    public static final int LIFECYCLE_END_COST = 4;
    public static final int DURATION_ACTION_PHASE = 1000;
    public static final int DURATION_GATHER_PHASE = 2000;
    public static final int DURATION_SUNMOVE_PHASE = 1000;
    public static final int STARTING_TREE_COUNT = 2;
    public static final int RICHNESS_BONUS_OK = 2;
    public static final int RICHNESS_BONUS_LUSH = 4;
}
*/

const (
	RICHNESS_NULL = 0
	RICHNESS_POOR = 1
	RICHNESS_OK   = 2
	RICHNESS_LUSH = 3

	TREE_SEED   = 0
	TREE_SMALL  = 1
	TREE_MEDIUM = 2
	TREE_TALL   = 3

	TREE_COST_SCALE        = 1
	LIFECYCLE_END_COST     = 4
	DURATION_ACTION_PHASE  = 1000
	DURATION_GATHER_PHASE  = 2000
	DURATION_SUNMOVE_PHASE = 1000
	STARTING_TREE_COUNT    = 2
	RICHNESS_BONUS_OK      = 2
	RICHNESS_BONUS_LUSH    = 4
)

var TREE_BASE_COST = [4]int{0, 1, 3, 7}
