// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"fmt"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

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

// TurnTraces implements arena.TurnTraceProvider, draining the turn's event
// buffer. The runner calls it after PerformGameUpdate, so the buffer holds the
// MESSAGE events parsing emitted plus everything the update resolved.
func (r *Referee) TurnTraces(_ int, _ []arena.Player) [2][]arena.TurnTrace {
	return r.Game.TurnTraces()
}

// RawScores implements arena.RawScoresProvider.
func (r *Referee) RawScores() [2]int { return r.Game.RawScores() }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Game.java:752-773

private void writeMetadata() {
    for (Player p : players) {
        gameManager.putMetadata("tracksPlaced_" + p.getIndex(), placedTracks[p.getIndex()]);
        ...
        if (enableSideQuest) gameManager.putMetadata("sideQuestPoints_" + p.getIndex(), sideQuestPoints[p.getIndex()]);
        gameManager.putMetadata(
            "averageTrackOwnershipPercentagePerActiveConnection_" + p.getIndex(),
            trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()] == 0 ? 0f
                : trackOwnershipPercentagePerActiveConnection[p.getIndex()]
                    / (float) trackOwnershipPercentagePerActiveConnectionTotal[p.getIndex()]
        );
        gameManager.putMetadata("executionTimeMs_" + p.getIndex(), executionTimeMs[p.getIndex()]);
    }
}
*/

// Metrics implements arena.MetricsProvider, carrying the counters Java
// publishes as match metadata under the same key names.
//
// Two keys of Java's set are dropped or unguarded: executionTimeMs is the
// SDK's own timing rather than a game counter and the arena measures its own,
// and sideQuestPoints is emitted unconditionally rather than behind
// enableSideQuest, so that a batch aggregates over one stable label set. The
// side quest is inert in this build, so the value is always 0.
func (r *Referee) Metrics() []arena.Metric {
	game := r.Game
	metrics := make([]arena.Metric, 0, 22)
	for idx := range game.PlacedTracks {
		counters := []struct {
			label string
			value float64
		}{
			{"tracksPlaced", float64(game.PlacedTracks[idx])},
			{"tracksPlacedOnPlains", float64(game.TracksPlacedOnPlains[idx])},
			{"tracksPlacedOnRiver", float64(game.TracksPlacedOnRiver[idx])},
			{"tracksPlacedOnMountains", float64(game.TracksPlacedOnMountains[idx])},
			{"zonesInked", float64(game.ZonesInked[idx])},
			{"ownTracksInkedOut", float64(game.OwnTracksInkedOut[idx])},
			{"enemyTracksInkedOut", float64(game.EnemyTracksInkedOut[idx])},
			{"extraTilesInConnection", float64(game.ExtraTilesInConnection[idx])},
			{"autobuildCalled", float64(game.AutobuildCalled[idx])},
			{"sideQuestPoints", float64(game.SideQuestPoints[idx])},
			{"averageTrackOwnershipPercentagePerActiveConnection", averageTrackOwnership(game, idx)},
		}
		for _, counter := range counters {
			metrics = append(metrics, arena.Metric{
				Label: fmt.Sprintf("%s_%d", counter.label, idx),
				Value: counter.value,
			})
		}
	}
	return metrics
}

// averageTrackOwnership divides in float32, as Java does, so the value matches
// the metadata the server publishes bit for bit.
func averageTrackOwnership(game *Game, idx int) float64 {
	total := game.TrackOwnershipPercentagePerActiveConnectionTotal[idx]
	if total == 0 {
		return 0
	}
	return float64(game.TrackOwnershipPercentagePerActiveConnection[idx] / float32(total))
}

// EndReason implements arena.EndReasonProvider. A deactivation decides the
// reason; otherwise the match ended normally, and both normal endings — the
// turn cap and no desired connection having any route left — are SCORE.
func (r *Referee) EndReason(_ int, players []arena.Player, deactivationTurns, firstOutputTurns [2]int) string {
	for i, p := range players {
		if !p.IsDeactivated() {
			continue
		}
		// The arena flags a hard timeout on the player; "Timeout!" is the
		// string Java's referee deactivates with, which a replay carries.
		timedOut := p.IsTimedOut() || p.DeactivationReason() == "Timeout!"
		switch {
		case timedOut && deactivationTurns[i] == firstOutputTurns[i]:
			return arena.EndReasonTimeoutStart
		case timedOut:
			return arena.EndReasonTimeout
		default:
			return arena.EndReasonInvalid
		}
	}
	return arena.EndReasonScore
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
