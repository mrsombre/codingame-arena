// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java
package engine

import (
	"encoding/json"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/game/Referee.java:21-22

private static final int MAX_TURNS = 200;
*/

// MaxMainTurns mirrors Java Referee.MAX_TURNS — the cap on main turns
// (speed sub-turns are extra and do not count toward this limit).
const MaxMainTurns = 200

// Referee drives the Game through the arena runner lifecycle. It tracks the
// Spring 2020 "speed sub-turn" mechanic locally so the arena can treat every
// iteration of its main loop uniformly.
type Referee struct {
	Game           *Game
	CommandManager *CommandManager
	SpeedTurn      bool
	GameOverFrame  bool
	MainTurns      int
}

func NewReferee(game *Game) *Referee {
	r := &Referee{Game: game}
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
	if r.SpeedTurn {
		return
	}
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
*/

func (r *Referee) PerformGameUpdate(turn int) {
	if r.GameOverFrame {
		r.Game.ResetGameTurnData()
		r.Game.PerformGameUpdate()
		r.Game.PerformGameOver()
		r.Game.EndGame()
		return
	}

	if r.SpeedTurn {
		r.Game.PerformGameSpeedUpdate()
	} else {
		r.MainTurns++
		r.Game.PerformGameUpdate()
	}

	if r.Game.IsGameOver() || r.MainTurns >= MaxMainTurns {
		r.GameOverFrame = true
	}
	r.SpeedTurn = r.Game.IsSpeedTurn()
}

func (r *Referee) ResetGameTurnData() {
	if r.SpeedTurn {
		return
	}
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
	return r.SpeedTurn
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

// TraceSnapshot, TracePac, TracePellet, SnapshotTurn, RawScores and Metrics
// have no Java counterpart — they wire optional arena interfaces for replay
// trace, pre-adjustment scores, and per-match metrics.

type TraceSnapshot struct {
	Scores  [2]int        `json:"scores"`
	Pacs    []TracePac    `json:"pacs"`
	Pellets []TracePellet `json:"pellets"`
}

type TracePac struct {
	ID              int    `json:"id"`
	Owner           int    `json:"owner"`
	X               int    `json:"x"`
	Y               int    `json:"y"`
	Type            string `json:"type"`
	AbilityDuration int    `json:"abilityDuration"`
	AbilityCooldown int    `json:"abilityCooldown"`
}

type TracePellet struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Value int `json:"value"`
}

// SnapshotTurn emits the full engine perspective for trace replay god mode.
func (r *Referee) SnapshotTurn(_ int, _ []arena.Player) json.RawMessage {
	snapshot := TraceSnapshot{}
	for _, p := range r.Game.Players {
		idx := p.GetIndex()
		if idx >= 0 && idx < len(snapshot.Scores) {
			snapshot.Scores[idx] = p.Pellets
		}
	}
	for _, pac := range r.Game.Pacmen {
		typeName := pac.Type.Name()
		if pac.Dead {
			typeName = "DEAD"
		}
		owner := 0
		if pac.Owner != nil {
			owner = pac.Owner.GetIndex()
		}
		snapshot.Pacs = append(snapshot.Pacs, TracePac{
			ID:              pac.Number,
			Owner:           owner,
			X:               pac.Position.X,
			Y:               pac.Position.Y,
			Type:            typeName,
			AbilityDuration: pac.AbilityDuration,
			AbilityCooldown: pac.AbilityCooldown,
		})
	}
	for _, p := range r.Game.Grid.AllPellets() {
		snapshot.Pellets = append(snapshot.Pellets, TracePellet{X: p.X, Y: p.Y, Value: 1})
	}
	for _, p := range r.Game.Grid.AllCherries() {
		snapshot.Pellets = append(snapshot.Pellets, TracePellet{X: p.X, Y: p.Y, Value: CHERRY_SCORE})
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil
	}
	return data
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
