// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/MovementResolution.java
package engine

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/MovementResolution.java:8-13

public class MovementResolution {
    List<Pacman> pacmenToKill = new ArrayList<>();
    List<Pacman> movedPacmen = new ArrayList<>();
    List<Pacman> blockedPacmen = new ArrayList<>();
    Map<Pacman, Pacman> blockedBy = new HashMap<>();
}
*/

// MovementResolution accumulates per-step movement outcomes.
type MovementResolution struct {
	MovedPacmen   []*Pacman
	BlockedPacmen []*Pacman
	BlockedBy     map[*Pacman]*Pacman
}

func NewMovementResolution() *MovementResolution {
	return &MovementResolution{
		BlockedBy: make(map[*Pacman]*Pacman),
	}
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/MovementResolution.java:14-17,27-34

public void addMovedPacman(Pacman pac)   { movedPacmen.add(pac); }
public void addBlockedPacmen(Pacman pac) { blockedPacmen.add(pac); }
public Pacman getBlockerOf(Pacman pac)   { return blockedBy.get(pac); }
*/

func (m *MovementResolution) AddMoved(pac *Pacman) {
	m.MovedPacmen = append(m.MovedPacmen, pac)
}

func (m *MovementResolution) AddBlocked(pac *Pacman) {
	m.BlockedPacmen = append(m.BlockedPacmen, pac)
}

func (m *MovementResolution) BlockerOf(pac *Pacman) *Pacman {
	return m.BlockedBy[pac]
}
