package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilCellMatchesJavaNoCellSentinel(t *testing.T) {
	var c *Cell // nil — equivalent to Java's NO_CELL
	assert.False(t, c.IsValid())
	assert.Equal(t, -1, c.GetIndex())
	assert.Equal(t, 0, c.GetRichness())
}

func TestCellGettersAndSetter(t *testing.T) {
	c := NewCell(7)
	assert.True(t, c.IsValid())
	assert.Equal(t, 7, c.GetIndex())
	assert.Equal(t, 0, c.GetRichness(), "default richness is 0 until SetRichness")
	c.SetRichness(RICHNESS_LUSH)
	assert.Equal(t, RICHNESS_LUSH, c.GetRichness())
}
