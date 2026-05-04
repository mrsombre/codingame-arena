// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/FrameType.java
package engine

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/FrameType.java:3-8

public enum FrameType {
    GATHERING,
    ACTIONS,
    SUN_MOVE,
    INIT
}
*/

type FrameType int

const (
	FrameGathering FrameType = iota
	FrameActions
	FrameSunMove
	FrameInit
)
