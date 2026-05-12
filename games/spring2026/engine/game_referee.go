// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java:19-28

public class Referee extends AbstractReferee {
    @Inject private MultiplayerGameManager<Player> gameManager;
    @Inject private GraphicEntityModule graphicEntityModule;
    ...
    private Board board;
*/

// Referee drives the Board through the arena runner lifecycle.
type Referee struct {
	Board    *Board
	MaxTurns int
	pending  *pendingTasks
}

func NewReferee(board *Board) *Referee {
	return &Referee{
		Board:    board,
		MaxTurns: MainTurnsForLeague(board.League),
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java:30-36

@Override
public void init() {
    gameManager.setTurnMaxTime(Constants.TIME_PER_TURN);
    gameManager.setMaxTurns(...);
    board = Board.createMap(...);
}

The Board is built inside the factory NewGame; Init only re-attaches players
to the existing Board.
*/

func (r *Referee) Init(players []arena.Player) {
	// Map gen has already populated Board.Players (with the same *Player
	// pointers the factory created), so there is nothing left to wire here.
	// Validate the array shape only.
	_ = players
}

func (r *Referee) GlobalInfoFor(player arena.Player) []string {
	return r.Board.GetInitialInputs(player.GetIndex())
}

func (r *Referee) FrameInfoFor(player arena.Player) []string {
	return r.Board.GetTurnInputs(player.GetIndex())
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java:38-78

@Override
public void gameTurn(int turn) {
    for (Player player : gameManager.getActivePlayers()) {
        if (turn == 1) sendInitialInputs;
        sendTurnInputs;
        player.execute();
    }
    TaskManager taskManager = new TaskManager();
    for (Player player : gameManager.getActivePlayers()) {
        try { taskManager.parseTasks(player, board, outputs.get(0), league); }
        catch (TimeoutException e) { killPlayer(player, "timeout"); }
    }
    board.tick(turn, taskManager, gameManager);
    // surface errors per player; critical errors deactivate
    if (board.hasStalled()) gameManager.endGame();
}
*/

// pendingTasks is the TaskManager built up across Init→Parse→Update calls.
// The arena lifecycle splits Java's monolithic gameTurn across three methods,
// so we hold parsed commands here until PerformGameUpdate consumes them.
type pendingTasks struct {
	manager *TaskManager
}

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	mgr := NewTaskManager()
	for _, p := range players {
		player := p.(*Player)
		if player.IsDeactivated() {
			continue
		}
		outputs := player.GetOutputs()
		if len(outputs) == 0 {
			// Arena runner already deactivates on output timeout; this is a
			// defensive path for engines invoked directly (e.g. tests).
			r.killPlayer(player, "timeout")
			continue
		}
		mgr.ParseTasks(player, r.Board, outputs[0], r.Board.League)
	}
	r.pending = &pendingTasks{manager: mgr}
}

func (r *Referee) PerformGameUpdate(turn int) {
	r.Board.Turn = turn
	mgr := NewTaskManager()
	if r.pending != nil {
		mgr = r.pending.manager
	}
	r.Board.Tick(turn, mgr)

	// Surface error / summary tape per player, mirroring the Java referee.
	for _, p := range r.Board.Players {
		for _, err := range p.PopErrors() {
			if err.IsCritical() {
				r.killPlayer(p, err.GetMessage())
			} else {
				r.Board.Summary = append(r.Board.Summary, prefixWith(p, ": [failed] "+err.GetMessage()))
			}
		}
		for _, s := range p.PopSummaries() {
			r.Board.Summary = append(r.Board.Summary, prefixWith(p, ": "+s))
		}
	}

	if r.Board.HasStalled() {
		r.Board.stalled = true
		r.Board.ended = true
	}
	if turn >= r.MaxTurns {
		r.Board.ended = true
	}
}

func prefixWith(p *Player, msg string) string {
	return "P" + itoa(p.GetIndex()) + msg
}

func (r *Referee) killPlayer(player *Player, message string) {
	if player.IsDeactivated() {
		return
	}
	player.Deactivate(strings.TrimSpace(message))
	player.SetScore(-2)
	r.Board.ended = true
}

func (r *Referee) ResetGameTurnData() {
	r.pending = nil
}

func (r *Referee) Ended() bool { return r.Board.ended }

func (r *Referee) EndGame() {
	r.Board.ended = true
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Referee.java:88-99

@Override
public void onEnd() {
    int[] scores = ...
    endScreenModule.setScores(scores, texts);
}
*/

func (r *Referee) OnEnd() {
	for _, p := range r.Board.Players {
		if p.IsDeactivated() {
			p.SetScore(-1)
		}
	}
}

func (r *Referee) ShouldSkipPlayerTurn(player arena.Player) bool {
	return false
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

// pending is set by ParsePlayerOutputs and consumed by PerformGameUpdate.
// Stored on the Referee so the arena runner doesn't see internal task state.
var _ = (*pendingTasks)(nil)

// Implementing arena.RawScoresProvider — engine-side score reflects in-game
// inventory (PLUM/LEMON/APPLE/BANANA + 4*WOOD). OnEnd only flips disqualified
// scores to -1, so raw == final when both sides survive.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.Board.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		// Use the recomputed engine score (post-RecomputeScore), regardless of
		// any OnEnd adjustment for deactivation.
		score := p.Inv.GetItemCount(ItemPLUM) +
			p.Inv.GetItemCount(ItemLEMON) +
			p.Inv.GetItemCount(ItemAPPLE) +
			p.Inv.GetItemCount(ItemBANANA) +
			WOOD_POINTS*p.Inv.GetItemCount(ItemWOOD)
		scores[idx] = score
	}
	return scores
}

// EndReason categorises how the match terminated. Priority: deactivation
// reason > stall-driven end > turn cap.
func (r *Referee) EndReason(turn int, players []arena.Player, deactivationTurns, firstOutputTurns [2]int) string {
	for i, p := range players {
		if !p.IsDeactivated() {
			continue
		}
		reason := p.DeactivationReason()
		switch {
		case strings.Contains(strings.ToLower(reason), "timeout") && deactivationTurns[i] == firstOutputTurns[i]:
			return arena.EndReasonTimeoutStart
		case strings.Contains(strings.ToLower(reason), "timeout"):
			return arena.EndReasonTimeout
		default:
			return arena.EndReasonInvalid
		}
	}
	if r.Board.stalled {
		return arena.EndReasonScoreEarly
	}
	if turn >= r.MaxTurns {
		return arena.EndReasonScore
	}
	return arena.EndReasonTurnsOut
}
