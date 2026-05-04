package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTreeFatherIndexIsMinusOne(t *testing.T) {
	tree := NewTree()
	assert.Equal(t, -1, tree.FatherIndex)
	assert.Equal(t, 0, tree.Size)
	assert.Nil(t, tree.Owner)
	assert.False(t, tree.Dormant)
}

func TestTreeGrowIncrementsSize(t *testing.T) {
	tree := NewTree()
	tree.Grow()
	assert.Equal(t, 1, tree.Size)
	tree.Grow()
	assert.Equal(t, 2, tree.Size)
}

func TestTreeSetDormantAndReset(t *testing.T) {
	tree := NewTree()
	tree.SetDormant()
	assert.True(t, tree.Dormant)
	tree.Reset()
	assert.False(t, tree.Dormant)
}
