// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Referee.java:16-28

@Singleton
public class Referee extends AbstractReferee {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private CommandManager commandManager;
    @Inject private Game game;
    @Inject private GameSummaryManager gameSummaryManager;
    long seed;
}
*/

type Referee struct {
	Game           *Game
	CommandManager *CommandManager
}

func NewReferee(game *Game) *Referee {
	return &Referee{
		Game:           game,
		CommandManager: NewCommandManager(game, game.Summary),
	}
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Referee.java:29-49

@Override
public void init() {
    this.seed = gameManager.getSeed();
    Config.load(gameManager.getGameParameters());
    gameManager.setFirstTurnMaxTime(1000);
    gameManager.setTurnMaxTime(100);
    game.init(seed);
    sendGlobalInfo();
}
*/

func (r *Referee) Init(players []arena.Player) {
	typed := make([]*Player, 0, len(players))
	for _, player := range players {
		typed = append(typed, player.(*Player))
	}
	r.Game.Init(typed)
}

func (r *Referee) GlobalInfoFor(player arena.Player) []string {
	return SerializeGlobalInfoFor(player.(*Player), r.Game)
}

func (r *Referee) FrameInfoFor(player arena.Player) []string {
	return SerializeFrameInfoFor(player.(*Player), r.Game)
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Referee.java:64-83

@Override
public void gameTurn(int turn) {
    game.resetGameTurnData();
    if (game.getCurrentFrameType() == FrameType.ACTIONS) {
        for (Player player : gameManager.getActivePlayers()) {
            if (!player.isWaiting()) {
                // sendInputLine ...
                player.execute();
            }
        }
        handlePlayerCommands();
    }
    game.performGameUpdate();
}
*/

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	if r.Game.CurrentFrameType != FrameActions {
		return
	}
	for _, player := range players {
		p := player.(*Player)
		if p.IsDeactivated() || r.Game.ShouldSkipPlayerTurn(p) {
			continue
		}
		r.CommandManager.ParseCommands(p, p.GetOutputs())
	}
}

func (r *Referee) PerformGameUpdate(turn int) {
	r.Game.PerformGameUpdate(turn)
}

func (r *Referee) ResetGameTurnData() {
	r.Game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.Game.Ended()
}

func (r *Referee) EndGame() {
	r.Game.EndGame()
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Referee.java:100-115

@Override
public void onEnd() {
    game.onEnd();
    int[] scores = gameManager.getPlayers().stream().mapToInt(Player::getScore).toArray();
    String[] displayedText = gameManager.getPlayers().stream().map(Player::getBonusScore).toArray(String[]::new);
    endScreenModule.setScores(scores, displayedText);
}
*/

func (r *Referee) OnEnd() {
	r.Game.OnEnd()
}

func (r *Referee) ShouldSkipPlayerTurn(player arena.Player) bool {
	return r.Game.ShouldSkipPlayerTurn(player.(*Player))
}

// TurnTraces returns a copy of this turn's accumulated game traces.
// Drained after each PerformGameUpdate by the runner; the underlying
// slice is reset on the next call.
func (r *Referee) TurnTraces(_ int, _ []arena.Player) []arena.TurnTrace {
	if len(r.Game.traces) == 0 {
		return nil
	}
	out := make([]arena.TurnTrace, len(r.Game.traces))
	copy(out, r.Game.traces)
	return out
}

func (r *Referee) ActivePlayers(players []arena.Player) int {
	active := 0
	for _, player := range players {
		if !player.IsDeactivated() {
			active++
		}
	}
	return active
}

// RawScores returns each player's pre-OnEnd accumulated score (sun-points
// converted via COMPLETE actions). OnEnd then layers on the floor(sun/3)
// bonus and the all-equal tree-count tiebreaker, so the post-OnEnd value
// reported via Player.GetScore is not the raw in-match score.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.Game.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		scores[idx] = p.GetScore()
	}
	return scores
}

// EndReason categorizes how the match terminated. Priority: deactivation
// reason (if any side is deactivated) > round cap reached (SCORE) > turn
// cap (TURNS_OUT). Spring 2021 has no ELIMINATED equivalent — there is no
// "all units dead" condition, only round exhaustion or deactivation.
//
// "Timeout!" is the literal deactivation message used by both the arena
// runner (on no output) and CommandManager.ParseCommands (on empty input).
func (r *Referee) EndReason(_ int, players []arena.Player, deactivationTurns [2]int) string {
	for i, p := range players {
		if !p.IsDeactivated() {
			continue
		}
		reason := p.DeactivationReason()
		switch {
		case reason == "Timeout!" && deactivationTurns[i] == 0:
			return arena.EndReasonTimeoutStart
		case reason == "Timeout!":
			return arena.EndReasonTimeout
		default:
			return arena.EndReasonInvalid
		}
	}

	if r.Game.Round >= r.Game.MAX_ROUNDS {
		return arena.EndReasonScore
	}
	return arena.EndReasonTurnsOut
}
