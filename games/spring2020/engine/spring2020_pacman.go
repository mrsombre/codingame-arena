// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:9-40

public class Pacman {
    private Player owner;
    private int id;
    private int number;
    private String message;

    private Coord position;
    private int speed = 1;
    private int abilityDuration = 0;
    private int abilityCooldown = 0;
    private boolean endOfSpeed = false;

    private Action intent;
    private Ability.Type abilityToUse;
    private List<Coord> intendedPath = new ArrayList<Coord>();
    private boolean pathResolved = false;
    private String warningPathMessage;

    private PacmanType type;
    private int currentPathStep;
    private int previousPathStep;
    private boolean blocked;
    private boolean dead = false;

    public Pacman(int id, int number, Player owner, PacmanType type) {
        this.owner = owner;
        this.id = id;
        this.number = number;
        this.setType(type);
    }
}
*/

// Pacman is one playable unit on the grid.
type Pacman struct {
	Owner  *Player
	ID     int
	Number int

	Position Coord
	Type     PacmanType
	Speed    int

	AbilityDuration int
	AbilityCooldown int

	Message string
	HasMsg  bool

	Intent          Action
	AbilityToUse    AbilityType
	HasAbilityToUse bool
	IntendedPath    []Coord

	CurrentPathStep  int
	PreviousPathStep int
	Blocked          bool
	Dead             bool
	EndOfSpeed       bool
	WarnPathMsg      string
	HasWarnPathMsg   bool
}

// NewPacman creates a Pacman with the given id/number/type for an owner.
func NewPacman(id, number int, owner *Player, t PacmanType) *Pacman {
	return &Pacman{
		Owner:        owner,
		ID:           id,
		Number:       number,
		Type:         t,
		Speed:        1,
		Intent:       NoAction,
		IntendedPath: make([]Coord, 0),
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:41-46

public void setMessage(String message) {
    this.message = message;
    if (message != null && message.length() > 48) {
        this.message = message.substring(0, 46) + "...";
    }
}
*/

// SetMessage truncates to 48 chars like Java Pacman.setMessage.
func (p *Pacman) SetMessage(msg string) {
	if len(msg) > 48 {
		msg = msg[:46] + "..."
	}
	p.Message = msg
	p.HasMsg = msg != ""
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:72-79,89-93

public void tickAbilityDuration() {
    if (abilityDuration > 0) {
        abilityDuration--;
        endOfSpeed = abilityDuration == 0;
    } else {
        endOfSpeed = false;
    }
}

public void tickAbilityCooldown() {
    if (abilityCooldown > 0) {
        abilityCooldown--;
    }
}
*/

func (p *Pacman) TickAbilityDuration() {
	if p.AbilityDuration > 0 {
		p.AbilityDuration--
		p.EndOfSpeed = p.AbilityDuration == 0
	} else {
		p.EndOfSpeed = false
	}
}

func (p *Pacman) TickAbilityCooldown() {
	if p.AbilityCooldown > 0 {
		p.AbilityCooldown--
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:95-105

public void turnReset() {
    message = null;
    if (!isDead()) {
        tickAbilityDuration();
        tickAbilityCooldown();
    }
    setAbilityToUse(null);
    setCurrentPathStep(0);
    blocked = false;
    this.intent = Action.NO_ACTION;
}
*/

// TurnReset clears per-turn state. Called at start of each main turn.
func (p *Pacman) TurnReset() {
	p.Message = ""
	p.HasMsg = false
	if !p.Dead {
		p.TickAbilityDuration()
		p.TickAbilityCooldown()
	}
	p.AbilityToUse = AbilityUnset
	p.HasAbilityToUse = false
	p.CurrentPathStep = 0
	p.PreviousPathStep = 0
	p.Blocked = false
	p.Intent = NoAction
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:175-183

public void setCurrentPathStep(int step) {
    setPreviousPathStep(currentPathStep);
    currentPathStep = step;
}

public boolean moveFinished() {
    return getCurrentPathStep() == intendedPath.size() - 1;
}
*/

func (p *Pacman) SetCurrentPathStep(step int) {
	p.PreviousPathStep = p.CurrentPathStep
	p.CurrentPathStep = step
}

func (p *Pacman) MoveFinished() bool {
	return p.CurrentPathStep == len(p.IntendedPath)-1
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Pacman.java:197-203

public boolean fastEnoughToMoveAt(int step) {
    return speed > step;
}

public boolean isSpeeding() {
    return speed == Config.SPEED_BOOST;
}
*/

func (p *Pacman) FastEnoughToMoveAt(step int) bool {
	return p.Speed > step
}

func (p *Pacman) IsSpeeding(cfg *Config) bool {
	return p.Speed == cfg.SPEED_BOOST
}
