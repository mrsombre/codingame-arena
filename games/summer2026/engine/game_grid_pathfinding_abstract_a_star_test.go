package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lineProblem is a minimal AStarProblem: walk the integers from 0 to Goal one
// step at a time, with a step backwards always available.
type lineProblem struct {
	Goal int
	Max  int
}

func (p lineProblem) InitialState() int           { return 0 }
func (p lineProblem) IsGoal(state int) bool       { return state == p.Goal }
func (p lineProblem) Heuristic(state int) float64 { return float64(abs(p.Goal - state)) }
func (p lineProblem) Cost(_, _ int) float64       { return 1 }
func (p lineProblem) TieBreaker(_, _ int) int     { return 0 }
func (p lineProblem) StateKey(state int) int      { return state }

func (p lineProblem) Successors(state int) []int {
	var out []int
	if state > 0 {
		out = append(out, state-1)
	}
	if state < p.Max {
		out = append(out, state+1)
	}
	return out
}

func TestAStarSearchReturnsThePathStartFirst(t *testing.T) {
	path, ok := AStarSearch[int, int](lineProblem{Goal: 4, Max: 10})

	require.True(t, ok)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, path)
}

func TestAStarSearchReturnsTheStartAloneWhenItIsAlreadyTheGoal(t *testing.T) {
	path, ok := AStarSearch[int, int](lineProblem{Goal: 0, Max: 10})

	require.True(t, ok)
	assert.Equal(t, []int{0}, path)
}

func TestAStarSearchReportsFailureWhenTheGoalIsUnreachable(t *testing.T) {
	path, ok := AStarSearch[int, int](lineProblem{Goal: 20, Max: 5})

	assert.False(t, ok)
	assert.Nil(t, path)
}
