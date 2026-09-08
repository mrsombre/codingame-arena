package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestEngineRegistersItselfUnderTheGameName(t *testing.T) {
	factory := arena.GetFactory("summer2026")

	require.NotNil(t, factory, "summer2026 is missing from the registry")
	assert.Equal(t, "summer2026", factory.Name())
	assert.Equal(t, MAX_TURNS, factory.MaxTurns())
	assert.Contains(t, arena.Games(), "summer2026")
}
