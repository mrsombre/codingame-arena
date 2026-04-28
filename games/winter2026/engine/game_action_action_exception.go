// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionException.java
package engine

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionException.java:1-10

package com.codingame.game.action;

@SuppressWarnings("serial")
public class ActionException extends Exception {

    public ActionException(String message) {
        super(message);
    }

}
*/

// Intentionally empty: ActionException.java is represented by Go error returns in callers.
// Parse returns an *ActionError on invalid input; CommandManager converts that to
// an *InvalidInputError for player deactivation.
