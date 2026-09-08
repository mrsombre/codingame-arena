// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java
package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

// Referee drives Game through the arena runner lifecycle. The gym-mode
// reset/step pair, the debug frame strings and the ByteBuffer observation
// plumbing are RL-harness only and are not ported.
type Referee struct {
	Game           *Game
	CommandManager *CommandManager
}

func NewReferee(game *Game) *Referee {
	return &Referee{Game: game, CommandManager: NewCommandManager(game)}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java:47-64

@Override
public void init() {
    gameManager.setMaxTurns(400);
    gameManager.setFirstTurnMaxTime(1000);
    game.init();
    sendGlobalInfo();
    gameManager.setFrameDuration(1000);
    gameManager.setTurnMaxTime(50);
}
*/

// Init builds the match state. setMaxTurns(400) is the SDK's safety ceiling
// rather than a rule — the real cap is Game.MAX_TURNS — and the timing calls
// are handled by the arena, so neither is modelled. Global info dispatch is
// the runner's job via GlobalInfoFor.
func (r *Referee) Init(players []arena.Player) {
	engine := make([]*Player, len(players))
	for i, p := range players {
		engine[i] = p.(*Player)
	}
	r.Game.Init(engine)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java:71-78,88-94

private void sendGlobalInfo() {
    for (Player player : gameManager.getActivePlayers())
        for (String line : Serializer.serializeGlobalInfoFor(player, game)) player.sendInputLine(line);
}
...
for (String line : Serializer.serializeFrameInfoFor(player, game)) player.sendInputLine(line);
*/

func (r *Referee) GlobalInfoFor(player arena.Player) []string {
	return SerializeGlobalInfoFor(player.(*Player), r.Game)
}

func (r *Referee) FrameInfoFor(player arena.Player) []string {
	return SerializeFrameInfoFor(player.(*Player), r.Game)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java:105-118

private void handlePlayerCommands() {
    for (Player player : gameManager.getActivePlayers()) {
        if (game.shouldSkipPlayerTurn(player)) continue;
        try { commandManager.parseCommands(player, player.getOutputs()); }
        catch (TimeoutException e) { player.deactivate("Timeout!"); ... }
    }
}
*/

// ParsePlayerOutputs mirrors handlePlayerCommands. Timeouts are already
// converted into a deactivation by the arena before this runs, so the
// TimeoutException branch has no counterpart here.
func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	for _, p := range players {
		player := p.(*Player)
		if player.IsDeactivated() || r.Game.ShouldSkipPlayerTurn(player) {
			continue
		}
		r.CommandManager.ParseCommands(player, player.GetOutputs())
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java:82-103

public void gameTurn(int turn) {
    game.resetGameTurnData();
    ... send inputs, execute, handlePlayerCommands ...
    game.performGameUpdate(turn);
    if (gameManager.getActivePlayers().size() < 2) abort();
}

private void abort() { gameManager.endGame(); }
*/

func (r *Referee) PerformGameUpdate(turn int) {
	r.Game.PerformGameUpdate(turn)

	if r.ActivePlayersCount() < 2 {
		r.Game.EndGame()
	}
}

func (r *Referee) ResetGameTurnData() {
	r.Game.ResetGameTurnData()
}

func (r *Referee) Ended() bool { return r.Game.Ended() }

func (r *Referee) EndGame() { r.Game.EndGame() }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java:120-123

@Override
public void onEnd() { game.onEnd(); }
*/

func (r *Referee) OnEnd() { r.Game.OnEnd() }

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

func (r *Referee) ActivePlayersCount() int {
	active := 0
	for _, player := range r.Game.Players {
		if !player.IsDeactivated() {
			active++
		}
	}
	return active
}
