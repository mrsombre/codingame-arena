// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/GameSummaryManager.java
package engine

import (
	"fmt"
	"strings"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/GameSummaryManager.java:11-25

@Singleton
public class GameSummaryManager {
    private List<String> lines;
    public GameSummaryManager() { this.lines = new ArrayList<>(); }
    public void clear() { this.lines.clear(); }
}
*/

// GameSummaryManager collects per-turn summary text the Java referee dumps
// into the gameSummary panel. The arena keeps the lines so traces can
// reproduce them.
type GameSummaryManager struct {
	Lines []string
	cfg   Config
}

func NewGameSummaryManager(cfg Config) *GameSummaryManager {
	return &GameSummaryManager{cfg: cfg}
}

func (m *GameSummaryManager) String() string {
	return strings.Join(m.Lines, "\n")
}

func (m *GameSummaryManager) Clear() {
	m.Lines = m.Lines[:0]
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/GameSummaryManager.java:26-48

public void addPlayerBadCommand(Player player, InvalidInputException e) {
    lines.add(formatErrorMessage(...));
}
public void addPlayerTimeout(Player player) {
    lines.add(formatErrorMessage(... "%s has not provided an action in time."));
}
*/

func (m *GameSummaryManager) AddPlayerBadCommand(p *Player, err *InvalidInputError) {
	m.Lines = append(m.Lines, fmt.Sprintf(
		"%s provided invalid input. Expected '%s'\nGot '%s'",
		p.NicknameToken(), err.GetExpected(), err.GetGot()))
}

func (m *GameSummaryManager) AddPlayerTimeout(p *Player) {
	m.Lines = append(m.Lines, fmt.Sprintf("%s has not provided an action in time.", p.NicknameToken()))
}

func (m *GameSummaryManager) AddPlayerDisqualified(p *Player) {
	m.Lines = append(m.Lines, fmt.Sprintf("%s was disqualified.", p.NicknameToken()))
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/GameSummaryManager.java:59-89

public void addCutTree(...)    { lines.add(... "%s is ending their tree life on cell %d, scoring %d points"); }
public void addGrowTree(...)   { lines.add(... "%s is growing a tree on cell %d"); }
public void addPlantSeed(...)  { lines.add(... "%s is planting a seed on cell %d from cell %d"); }
*/

func (m *GameSummaryManager) AddCutTree(p *Player, cell *Cell, score int) {
	m.Lines = append(m.Lines, fmt.Sprintf(
		"%s is ending their tree life on cell %d, scoring %d points",
		p.NicknameToken(), cell.GetIndex(), score))
}

func (m *GameSummaryManager) AddGrowTree(p *Player, cell *Cell) {
	m.Lines = append(m.Lines, fmt.Sprintf(
		"%s is growing a tree on cell %d",
		p.NicknameToken(), cell.GetIndex()))
}

func (m *GameSummaryManager) AddPlantSeed(p *Player, target, source *Cell) {
	m.Lines = append(m.Lines, fmt.Sprintf(
		"%s is planting a seed on cell %d from cell %d",
		p.NicknameToken(), target.GetIndex(), source.GetIndex()))
}

func (m *GameSummaryManager) AddWait(p *Player) {
	m.Lines = append(m.Lines, fmt.Sprintf("%s is waiting", p.NicknameToken()))
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/GameSummaryManager.java:100-149

public void addRound(int round)            { ... "Round %d/%d" }
public void addRoundTransition(int round)  { ... "Round %d ends" / "The sun is now pointing towards direction %d" / "Round %d starts" }
public void addGather(Player p, int given) { ... "%s has collected %d sun points" }
public void addSeedConflict(Seed seed)     { ... "Seed conflict on cell %d" }
*/

func (m *GameSummaryManager) AddRound(round int) {
	m.Lines = append(m.Lines, fmt.Sprintf("Round %d/%d", round, m.cfg.MAX_ROUNDS-1))
}

func (m *GameSummaryManager) AddError(err string) {
	m.Lines = append(m.Lines, err)
}

func (m *GameSummaryManager) AddSeedConflict(s Seed) {
	m.Lines = append(m.Lines, fmt.Sprintf("Seed conflict on cell %d", s.TargetCell))
}

func (m *GameSummaryManager) AddRoundTransition(round int) {
	m.Lines = append(m.Lines, fmt.Sprintf("Round %d ends", round))
	if round+1 < m.cfg.MAX_ROUNDS {
		m.Lines = append(m.Lines, fmt.Sprintf("The sun is now pointing towards direction %d", (round+1)%6))
		m.Lines = append(m.Lines, fmt.Sprintf("Round %d starts", round+1))
	}
}

func (m *GameSummaryManager) AddGather(p *Player, given int) {
	m.Lines = append(m.Lines, fmt.Sprintf("%s has collected %d sun points", p.NicknameToken(), given))
}
