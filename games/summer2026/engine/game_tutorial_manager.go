// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/TutorialManager.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/TutorialManager.java:10-30

@Singleton
public class TutorialManager {
    @Inject private Game game;
    int leagueLevel;
    List<Player> players;
    private boolean playerOneHasInkedEnemyTrack;

    public boolean initTutorial() {
        this.leagueLevel = game.leagueLevel;
        this.players = gameManager.getPlayers();
        playerOneHasInkedEnemyTrack = false;
        if (leagueLevel == 1) return true;
        else if (leagueLevel == 2) return true;
        return false;
    }
}
*/

// TutorialManager replaces the win condition in the two tutorial leagues.
// League 1 asks player 0 to score a single point; league 2 asks it to ink a
// region holding an opposing track. Either way the match stops the moment the
// objective is met, and the final scores are a verdict rather than a count.
//
// Leagues 3 and up leave it inert: InitTutorial reports false and Game never
// consults it again.
type TutorialManager struct {
	LeagueLevel int
	Players     []*Player
	// PlayerOneHasInkedEnemyTrack latches the league 2 objective and never
	// clears once set.
	PlayerOneHasInkedEnemyTrack bool
}

func NewTutorialManager() *TutorialManager { return &TutorialManager{} }

// InitTutorial reports whether this league is a tutorial, and is what sets
// Game.InTutorial.
func (m *TutorialManager) InitTutorial(game *Game) bool {
	m.LeagueLevel = game.LeagueLevel
	m.Players = game.Players
	m.PlayerOneHasInkedEnemyTrack = false

	return m.LeagueLevel == 1 || m.LeagueLevel == 2
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/TutorialManager.java:32-41

public void handleEnd(String[] scoreTexts) {
    if (leagueLevel <= 2) {
        boolean win = objectiveComplete();
        players.get(0).setScore(win ? 0 : -1);
        players.get(1).setScore(win ? -1 : 0);
        scoreTexts[0] = win ? "objective complete" : "objective failed";
        scoreTexts[1] = "-";
    }
}
*/

// HandleEnd overwrites both scores with the verdict: the winner gets 0 and
// the loser -1, so the points earned during the match do not survive into the
// result. The score texts are viewer-only and are not ported.
func (m *TutorialManager) HandleEnd() {
	if m.LeagueLevel > 2 {
		return
	}
	win := m.ObjectiveComplete()
	if win {
		m.Players[0].SetScore(0)
		m.Players[1].SetScore(-1)
		return
	}
	m.Players[0].SetScore(-1)
	m.Players[1].SetScore(0)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/TutorialManager.java:43-54

public boolean objectiveComplete() {
    if (leagueLevel == 1) return game.players.get(0).getScore() >= 1;
    else if (leagueLevel == 2) return playerOneHasInkedEnemyTrack;
    return false;
}

public void setPlayerOneHasInkedEnemyTrack() { playerOneHasInkedEnemyTrack = true; }
*/

// ObjectiveComplete is polled every turn by Game.IsGameOver and again by
// HandleEnd. League 1 reads the live score, which HandleEnd then overwrites —
// so the order of those two calls matters and matches Java's.
func (m *TutorialManager) ObjectiveComplete() bool {
	switch m.LeagueLevel {
	case 1:
		return m.Players[0].GetScore() >= 1
	case 2:
		return m.PlayerOneHasInkedEnemyTrack
	}
	return false
}

// SetPlayerOneHasInkedEnemyTrack is called when an inking credited to player 0
// destroyed at least one track owned by player 1.
func (m *TutorialManager) SetPlayerOneHasInkedEnemyTrack() {
	m.PlayerOneHasInkedEnemyTrack = true
}
