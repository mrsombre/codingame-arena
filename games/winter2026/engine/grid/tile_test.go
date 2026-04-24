package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTileDefaults(t *testing.T) {
	tile := NewTile(Coord{X: 2, Y: 3})
	assert.True(t, tile.IsValid())
	assert.Equal(t, TileEmpty, tile.Type)
	assert.Equal(t, Coord{X: 2, Y: 3}, tile.Coord)
}

func TestNewTileWithType(t *testing.T) {
	tile := NewTileWithType(Coord{X: 1, Y: 1}, TileWall)
	assert.True(t, tile.IsValid())
	assert.Equal(t, TileWall, tile.Type)
}

func TestTileSetTypeAndClear(t *testing.T) {
	tile := NewTile(Coord{})
	tile.SetType(TileWall)
	assert.Equal(t, TileWall, tile.Type)
	tile.Clear()
	assert.Equal(t, TileEmpty, tile.Type)
}

func TestNoTileIsInvalid(t *testing.T) {
	assert.False(t, NoTile.IsValid())
	assert.Equal(t, -1, NoTile.Type)
}
