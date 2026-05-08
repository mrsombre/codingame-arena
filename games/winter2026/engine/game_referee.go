// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"encoding/json"

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
		CommandManager: NewCommandManager(&game.summary, game),
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

// DecorateTraceTurn delegates to Game.DecorateTraceTurn for the per-turn
// state payload. The arena runner type-asserts on the Referee, so the
// method must live here even though Game owns the data.
func (r *Referee) DecorateTraceTurn(turn int, players []arena.Player) json.RawMessage {
	return r.Game.DecorateTraceTurn(turn, players)
}

func (r *Referee) TurnTraces(_ int, _ []arena.Player) [2][]arena.TurnTrace {
	var out [2][]arena.TurnTrace
	for i, slot := range r.Game.traces {
		if len(slot) == 0 {
			continue
		}
		out[i] = make([]arena.TurnTrace, len(slot))
		copy(out[i], slot)
	}
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

// EndReason categorizes how the match terminated. Priority: deactivation
// reason (if any side is deactivated) > IsGameOver branch (player wiped out
// → ELIMINATED, otherwise apples consumed → SCORE) > turn cap. Winter 2026
// has no score lock-in shortcut, so SCORE_EARLY is not produced.
//
// "Timeout!" is the literal deactivation message used by both the arena
// runner (on no output) and CommandManager.ParseCommands (on empty input).
// TIMEOUT_START fires when the deactivation lands on the player's first
// prompted turn (loop turn 0 for Winter 2026, which has no leading
// engine-only frame).
func (r *Referee) EndReason(turn int, players []arena.Player, deactivationTurns, firstOutputTurns [2]int) string {
	for i, p := range players {
		if !p.IsDeactivated() {
			continue
		}
		reason := p.DeactivationReason()
		switch {
		case reason == "Timeout!" && deactivationTurns[i] == firstOutputTurns[i]:
			return arena.EndReasonTimeoutStart
		case reason == "Timeout!":
			return arena.EndReasonTimeout
		default:
			return arena.EndReasonInvalid
		}
	}

	if r.Game.IsGameOver() {
		for _, p := range r.Game.Players {
			alive := false
			for _, b := range p.Birds {
				if b.Alive {
					alive = true
					break
				}
			}
			if !alive {
				return arena.EndReasonEliminated
			}
		}
		return arena.EndReasonScore
	}

	return arena.EndReasonTurnsOut
}

func (r *Referee) Metrics() []arena.Metric {
	return []arena.Metric{
		{Label: "apples_remaining", Value: float64(len(r.Game.Grid.Apples))},
		{Label: "losses_p0", Value: float64(r.Game.Losses[0])},
		{Label: "losses_p1", Value: float64(r.Game.Losses[1])},
	}
}
