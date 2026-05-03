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

func (r *Referee) ActivePlayers(players []arena.Player) int {
	active := 0
	for _, player := range players {
		if !player.IsDeactivated() {
			active++
		}
	}
	return active
}
