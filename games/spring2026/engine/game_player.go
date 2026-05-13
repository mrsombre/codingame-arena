// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java
package engine

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:12-19

public class Player extends AbstractMultiplayerPlayer {
    private ArrayList<Unit> units;
    private Inventory inventory;
    private Cell shack;
    private String message = "";
    private ArrayList<InputError> errors = new ArrayList<>();
    private ArrayList<String> summaries = new ArrayList<>();
*/

// Player is the engine-side handle for one agent. It satisfies arena.Player.
// Game state mirrors the upstream Java class — Units, Inventory, Shack — plus
// the message/errors/summaries channels that the referee consumes each turn.
type Player struct {
	index   int
	score   int
	Units   []*Unit
	Inv     *Inventory
	Shack   *Cell
	Message string

	errors    []*InputError
	summaries []string

	deactivated        bool
	deactivationReason string
	timedOut           bool
	inputLines         []string
	outputs            []string
	outputError        error
	executeFunc        func() error
}

func NewPlayer(index int) *Player {
	return &Player{index: index, Inv: NewInventory()}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:21-26

public void init(Cell shack, int league) {
    this.units = new ArrayList<>();
    this.inventory = new Inventory();
    this.shack = shack;
    Unit unit = new Unit(this, new int[]{1, 1, 1, league >= 3 ? 1 : 0}, league);
}
*/

func (p *Player) InitForGame(shack *Cell) {
	p.Units = nil
	p.Inv = NewInventory()
	p.Shack = shack
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:28-30

public void setInventory(int[] inventory) {
    for (int i = 0; i < inventory.length; i++) this.inventory.setItem(i, inventory[i]);
}
*/

func (p *Player) SetInventory(inventory []int) {
	for i, v := range inventory {
		p.Inv.SetItem(Item(i), v)
	}
}

func (p *Player) GetUnits() []*Unit        { return p.Units }
func (p *Player) AddUnit(u *Unit)          { p.Units = append(p.Units, u) }
func (p *Player) GetShack() *Cell          { return p.Shack }
func (p *Player) GetInventory() *Inventory { return p.Inv }

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:48-55

public void recomputeScore() {
    if (isActive())
        setScore(inventory.getItemCount(Item.PLUM) + ... + Constants.WOOD_POINTS * inventory.getItemCount(Item.WOOD));
}
*/

func (p *Player) RecomputeScore() {
	if p.IsDeactivated() {
		return
	}
	s := p.Inv.GetItemCount(ItemPLUM) +
		p.Inv.GetItemCount(ItemLEMON) +
		p.Inv.GetItemCount(ItemAPPLE) +
		p.Inv.GetItemCount(ItemBANANA) +
		WOOD_POINTS*p.Inv.GetItemCount(ItemWOOD)
	p.SetScore(s)
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:69-76

public String getMessage() { return message; }
public void setMessage(String message) {
    this.message = message;
    if (this.message.length() > 50) this.message = this.message.substring(0, 50);
}
*/

func (p *Player) GetMessage() string { return p.Message }
func (p *Player) SetMessage(m string) {
	if len(m) > 50 {
		m = m[:50]
	}
	p.Message = m
}

/*
Java: SpringChallenge2026-Troll/src/main/java/com/codingame/game/Player.java:83-97

public ArrayList<InputError> popErrors() {
    // groups by error code, caps to <=3 per code; "N more errors of that type"
}
*/

// PopErrors drains pending errors. For each error code, at most three are
// returned; longer runs collapse into a "(N more errors of that type)"
// summary, mirroring Java's groupBy-code behaviour.
func (p *Player) PopErrors() []*InputError {
	result := make([]*InputError, 0)
	for len(p.errors) > 0 {
		first := p.errors[0]
		group := make([]*InputError, 0)
		remaining := p.errors[:0]
		for _, e := range p.errors {
			if e.ErrorCode == first.ErrorCode {
				group = append(group, e)
			} else {
				remaining = append(remaining, e)
			}
		}
		p.errors = remaining
		if len(group) <= 3 {
			result = append(result, group...)
		} else {
			result = append(result, group[0])
			result = append(result, group[1])
			result = append(result, NewInputError(
				itoa(len(group)-2)+" more errors of that type",
				first.ErrorCode,
				first.Critical,
			))
		}
	}
	return result
}

func (p *Player) PopSummaries() []string {
	out := p.summaries
	p.summaries = nil
	return out
}

func (p *Player) AddError(e *InputError) { p.errors = append(p.errors, e) }
func (p *Player) AddSummary(s string)    { p.summaries = append(p.summaries, s) }

// --- arena.Player implementation ---

func (p *Player) GetIndex() int { return p.index }
func (p *Player) GetScore() int { return p.score }
func (p *Player) SetScore(score int) {
	p.score = score
}
func (p *Player) IsDeactivated() bool { return p.deactivated }
func (p *Player) Deactivate(reason string) {
	p.deactivated = true
	p.deactivationReason = reason
}
func (p *Player) DeactivationReason() string { return p.deactivationReason }
func (p *Player) IsTimedOut() bool           { return p.timedOut }
func (p *Player) SetTimedOut(timedOut bool)  { p.timedOut = timedOut }

// GetExpectedOutputLines: troll bots emit a single semicolon-joined line per
// turn, matching the upstream protocol.
func (p *Player) GetExpectedOutputLines() int { return 1 }

func (p *Player) SendInputLine(line string) {
	p.inputLines = append(p.inputLines, line)
}
func (p *Player) ConsumeInputLines() []string {
	lines := append([]string(nil), p.inputLines...)
	p.inputLines = p.inputLines[:0]
	return lines
}
func (p *Player) GetOutputs() []string { return p.outputs }
func (p *Player) SetOutputs(outputs []string) {
	p.outputs = outputs
	p.outputError = nil
}
func (p *Player) GetOutputError() error { return p.outputError }
func (p *Player) SetExecuteFunc(fn func() error) {
	p.executeFunc = fn
}
func (p *Player) Execute() error {
	if p.executeFunc == nil {
		return nil
	}
	err := p.executeFunc()
	p.outputError = err
	return err
}
