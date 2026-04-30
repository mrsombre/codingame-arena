// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:21-22

private static final int MAX_TURNS = 200;
*/

// MaxMainTurns mirrors Java Referee.MAX_TURNS — the cap on main turns. We
// fold speed sub-steps into the same PerformGameUpdate call, so a "turn"
// here corresponds 1:1 to a Java main turn.
const MaxMainTurns = 200

// Referee drives the Game through the arena runner lifecycle.
type Referee struct {
	Game           *Game
	CommandManager *CommandManager
	GameOverFrame  bool
	MainTurns      int
}

func NewReferee(game *Game) *Referee {
	r := &Referee{
		Game: game,
	}
	r.CommandManager = NewCommandManager(&game.Summary, game)
	return r
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:32-61

@Override
public void init() {
    gameOverFrame = false;
    this.seed = gameManager.getSeed();
    Config.setDefaultValueByLevel(LeagueRules.fromIndex(gameManager.getLeagueLevel()));
    maxFrames = MAX_TURNS;
    gameManager.setMaxTurns(MAX_TURNS);
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
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:114-125

private void handlePlayerCommands() {
    for (Player player : gameManager.getActivePlayers()) {
        try {
            commandManager.parseCommands(player, player.getOutputs(), game);
        } catch (TimeoutException e) {
            player.deactivate("Timeout!");
            player.setTimedOut(true);
        }
    }
}
*/

func (r *Referee) ParsePlayerOutputs(players []arena.Player) {
	for _, player := range players {
		p := player.(*Player)
		if p.IsDeactivated() {
			continue
		}
		r.CommandManager.ParseCommands(p, p.GetOutputs())
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:80-112

@Override
public void gameTurn(int turn) {
    if (!gameOverFrame) {
        if (game.isSpeedTurn()) {
            game.performGameSpeedUpdate();
            maxFrames++;
            gameManager.setMaxTurns(maxFrames);
        } else {
            game.resetGameTurnData();
            // ... handlePlayerCommands();
            game.performGameUpdate();
        }
        if (game.isGameOver()) {
            gameOverFrame = true;
        }
    } else {
        game.resetGameTurnData();
        game.performGameUpdate();
        game.performGameOver();
        gameManager.endGame();
    }
}

We deviate from the Java frame split: PerformGameUpdate folds the speed
sub-turn into a single call (see Game.PerformGameUpdate). Each arena turn
maps 1:1 to a Java main turn, no skip-input bookkeeping needed.
*/

func (r *Referee) PerformGameUpdate(turn int) {
	if r.GameOverFrame {
		r.Game.ResetGameTurnData()
		r.Game.PerformGameUpdate()
		r.Game.PerformGameOver()
		r.Game.EndGame()
		return
	}

	r.MainTurns++
	r.Game.PerformGameUpdate()

	if r.Game.IsGameOver() || r.MainTurns >= MaxMainTurns {
		r.GameOverFrame = true
	}
}

func (r *Referee) ResetGameTurnData() {
	r.Game.ResetGameTurnData()
}

func (r *Referee) Ended() bool {
	return r.Game.Ended()
}

func (r *Referee) EndGame() {
	if r.Game.IsGameOver() {
		r.Game.PerformGameOver()
	}
	r.Game.EndGame()
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:131-153

@Override
public void onEnd() {
    gameManager.getPlayers().forEach(player -> player.setScore(player.pellets));
    printWinner();
}
*/

func (r *Referee) OnEnd() {
	r.Game.OnEnd()
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

// TurnTraces, RawScores, and Metrics have no Java counterpart — they wire
// optional arena interfaces for per-turn opaque traces, pre-adjustment scores,
// and per-match metrics.

// TurnTraces returns a copy of this turn's accumulated game traces.
func (r *Referee) TurnTraces(_ int, _ []arena.Player) []arena.TurnTrace {
	if len(r.Game.traces) == 0 {
		return nil
	}
	out := make([]arena.TurnTrace, len(r.Game.traces))
	copy(out, r.Game.traces)
	return out
}

// RawScores returns per-player pellet counts — the raw in-match score.
func (r *Referee) RawScores() [2]int {
	var scores [2]int
	for _, p := range r.Game.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= len(scores) {
			continue
		}
		scores[idx] = p.Pellets
	}
	return scores
}

// EndReason categorizes how the match terminated. Priority: deactivation
// reason (if any side is deactivated) > IsGameOver via score lock-in > turn
// cap. The "all pacmen dead" deactivation maps to ELIMINATED rather than a
// disqualification — it's a normal in-game elimination, not a fault.
//
// "Timeout!" is the literal deactivation message used by both the arena
// runner (on no output) and CommandManager.ParseCommands (on empty input).
func (r *Referee) EndReason(turn int, players []arena.Player, deactivationTurns [2]int) string {
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
		case reason == "all pacmen dead":
			return arena.EndReasonEliminated
		default:
			return arena.EndReasonInvalid
		}
	}

	// Both sides still active. IsGameOver here means canImproveRanking lock-in.
	if r.Game.IsGameOver() {
		if r.Game.RemainingPellets() == 0 {
			return arena.EndReasonScore
		}
		return arena.EndReasonScoreEarly
	}

	return arena.EndReasonTurnsOut
}

// Metrics emits Spring 2020-specific per-match metrics.
func (r *Referee) Metrics() []arena.Metric {
	remainingPellets := 0
	remainingCherries := 0
	for _, cell := range r.Game.Grid.Cells {
		if cell.HasPellet {
			remainingPellets++
		}
		if cell.HasCherry {
			remainingCherries++
		}
	}
	alive0, alive1 := 0, 0
	for _, pac := range r.Game.Pacmen {
		if pac.Owner.GetIndex() == 0 {
			alive0++
		} else {
			alive1++
		}
	}
	return []arena.Metric{
		{Label: "pellets_remaining", Value: float64(remainingPellets)},
		{Label: "cherries_remaining", Value: float64(remainingCherries)},
		{Label: "pacs_alive_p0", Value: float64(alive0)},
		{Label: "pacs_alive_p1", Value: float64(alive1)},
	}
}
