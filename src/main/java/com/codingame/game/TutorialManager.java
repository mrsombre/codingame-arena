package com.codingame.game;

import java.util.List;

import com.codingame.gameengine.core.MultiplayerGameManager;
import com.google.inject.Inject;
import com.google.inject.Singleton;

@Singleton
public class TutorialManager {
    @Inject private Game game;
    @Inject private MultiplayerGameManager<Player> gameManager;

    int leagueLevel;
    List<Player> players;
    private boolean playerOneHasInkedEnemyTrack;

    public boolean initTutorial() {
        this.leagueLevel = game.leagueLevel;
        this.players = gameManager.getPlayers();

        playerOneHasInkedEnemyTrack = false;

        if (leagueLevel == 1) {
            return true;
        } else if (leagueLevel == 2) {
            return true;
        }
        return false;
    }

    public void handleEnd(String[] scoreTexts) {
        if (leagueLevel <= 2) {
            boolean win = objectiveComplete();
            players.get(0).setScore(win ? 0 : -1);
            players.get(1).setScore(win ? -1 : 0);
            scoreTexts[0] = win ? "objective complete" : "objective failed";
            scoreTexts[1] = "-";
        }
    }

    public boolean objectiveComplete() {
        if (leagueLevel == 1) {
            return game.players.get(0).getScore() >= 1;
        } else if (leagueLevel == 2) {
            return playerOneHasInkedEnemyTrack;
        }
        return false;
    }

    public void setPlayerOneHasInkedEnemyTrack() {
        playerOneHasInkedEnemyTrack = true;

    }

}
