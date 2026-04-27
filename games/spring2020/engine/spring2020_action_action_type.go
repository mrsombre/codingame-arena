// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/ActionType.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/action/ActionType.java:1-5

package com.codingame.spring2020.action;

public enum ActionType {
  WAIT, MOVE, MSG, SPEED, SWITCH
}
*/

// ActionType matches Java's ActionType enum.
type ActionType int

const (
	ActionWait ActionType = iota
	ActionMove
	ActionMsg
	ActionSpeed
	ActionSwitch
)
