// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java:11-16

@Singleton
public class Referee extends AbstractReferee {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private CommandManager commandManager;
    @Inject private Game game;
*/

type Referee struct {
	Game           *Game
	CommandManager *CommandManager
}

func NewReferee(game *Game) *Referee {
	return &Referee{
		Game:           game,
		CommandManager: NewCommandManager(&game.summary),
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java:18-34

@Override
public void init() {
    gameManager.setMaxTurns(200);
    gameManager.setFirstTurnMaxTime(1000);
    game.init();
    sendGlobalInfo();
    gameManager.setFrameDuration(1000);
    gameManager.setTurnMaxTime(50);
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
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java:74-87

private void handlePlayerCommands() {
    for (Player player : gameManager.getActivePlayers()) {
        if (game.shouldSkipPlayerTurn(player)) continue;
        try {
            commandManager.parseCommands(player, player.getOutputs());
        } catch (TimeoutException e) {
            player.deactivate("Timeout!");
        }
    }
}
*/

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	for _, player := range players {
		p := player.(*Player)
		if p.IsDeactivated() || r.Game.ShouldSkipPlayerTurn(p) {
			continue
		}
		r.CommandManager.ParseCommands(p, p.GetOutputs())
	}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java:51-72

@Override
public void gameTurn(int turn) {
    game.resetGameTurnData();
    for (Player player : gameManager.getActivePlayers()) {
        if (game.shouldSkipPlayerTurn(player)) continue;
        // sendInputLine for each frame info line
        player.execute();
    }
    handlePlayerCommands();
    game.performGameUpdate(turn);
}
*/

func (r *Referee) PerformGameUpdate(turn int) {
	r.Game.PerformGameUpdate(turn)
}

func (r *Referee) ResetGameTurnData() {
	r.Game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.Game.ended
}

func (r *Referee) EndGame() {
	r.Game.EndGame()
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java:89-92

@Override
public void onEnd() {
    game.onEnd();
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

func (r *Referee) TurnTraces(_ int, _ []arena.Player) []arena.TurnTrace {
	if len(r.Game.traces) == 0 {
		return nil
	}
	out := make([]arena.TurnTrace, len(r.Game.traces))
	copy(out, r.Game.traces)
	return out
}

// RawScores returns each player's score as the sum of alive bird segments.
// This is the intrinsic game metric before OnEnd applies its tie-breaking
// adjustment (subtracting losses when raw scores tie) that can produce
// negative values in Player.GetScore.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.Game.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		for _, b := range p.Birds {
			if b.Alive {
				scores[idx] += len(b.Body)
			}
		}
	}
	return scores
}

func (r *Referee) Metrics() []arena.Metric {
	return []arena.Metric{
		{Label: "apples_remaining", Value: float64(len(r.Game.Grid.Apples))},
		{Label: "losses_p0", Value: float64(r.Game.Losses[0])},
		{Label: "losses_p1", Value: float64(r.Game.Losses[1])},
	}
}
