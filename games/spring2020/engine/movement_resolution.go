// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/MovementResolution.java
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/BumpCouple.java
package engine

import "github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"

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

func (m *MovementResolution) AddMoved(pac *Pacman) {
	m.MovedPacmen = append(m.MovedPacmen, pac)
}

func (m *MovementResolution) AddBlocked(pac *Pacman) {
	m.BlockedPacmen = append(m.BlockedPacmen, pac)
}

func (m *MovementResolution) BlockerOf(pac *Pacman) *Pacman {
	return m.BlockedBy[pac]
}

// BumpCouple is a pair of pacmen involved in the same bump.
// Equality is symmetric over (from, fromBlocker).
type BumpCouple struct {
	From        grid.Coord
	FromBlocker grid.Coord
	To          grid.Coord
	Distance    int
}

func (b BumpCouple) matches(other BumpCouple) bool {
	return (b.From == other.From && b.FromBlocker == other.FromBlocker) ||
		(b.From == other.FromBlocker && b.FromBlocker == other.From)
}

// AddUnique appends b to list unless a symmetric couple already exists.
func AddUniqueBumpCouple(list []BumpCouple, b BumpCouple) []BumpCouple {
	for _, existing := range list {
		if existing.matches(b) {
			return list
		}
	}
	return append(list, b)
}
