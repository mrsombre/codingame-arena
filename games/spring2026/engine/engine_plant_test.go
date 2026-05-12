package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPlantInitialSaplingHealth(t *testing.T) {
	// Java rule: starting health = FINAL_HEALTH - DELTA_HEALTH * PLANT_MAX_SIZE,
	// i.e. the residual once all per-growth bumps are subtracted out.
	// PLUM: 12 - 2*4 = 4; APPLE: 20 - 3*4 = 8; BANANA: 6 - 1*4 = 2.
	cell := NewCell(0, 0)
	assert.Equal(t, 12-2*PLANT_MAX_SIZE, NewPlant(cell, ItemPLUM).Health)
	assert.Equal(t, 20-3*PLANT_MAX_SIZE, NewPlant(cell, ItemAPPLE).Health)
	assert.Equal(t, 6-1*PLANT_MAX_SIZE, NewPlant(cell, ItemBANANA).Health)
}

func TestPlantTickGrowsWhenCooldownHitsZero(t *testing.T) {
	// Rules: when cooldown reaches 0 and size < max, the plant grows by 1 and
	// gains DELTA_HEALTH (updateHealth=true). Cooldown resets to growth cooldown.
	cell := NewCell(0, 0)
	p := NewPlant(cell, ItemBANANA) // base cooldown 6, no water
	p.Size = 0
	p.Cooldown = 1

	p.Tick(true)
	assert.Equal(t, 1, p.Size, "size advanced")
	assert.Equal(t, PLANT_COOLDOWN[ItemBANANA], p.Cooldown, "cooldown reset")
}

func TestPlantTickProducesFruitOnceFullyGrown(t *testing.T) {
	cell := NewCell(0, 0)
	p := NewPlant(cell, ItemAPPLE)
	p.Size = PLANT_MAX_SIZE
	p.Cooldown = 1
	p.Resources = 0

	p.Tick(true)
	assert.Equal(t, 1, p.Resources, "first fruit produced")

	// Cap at PLANT_MAX_RESOURCES (=3). At 3 fruits, the next tick is a no-op.
	p.Resources = PLANT_MAX_RESOURCES
	p.Cooldown = 1
	p.Tick(true)
	assert.Equal(t, PLANT_MAX_RESOURCES, p.Resources, "no overshoot")
}

func TestPlantGrowthCooldownReducedByWaterAdjacency(t *testing.T) {
	// A tree's growth cooldown drops by PLANT_WATER_COOLDOWN_BOOST[type]
	// when any 4-neighbour is WATER. Need a row long enough that one cell is
	// genuinely two steps from the water tile.
	grid := cellGrid(5, 1)
	grid[2][0].Type = CellWATER
	dry := NewPlant(grid[0][0], ItemAPPLE) // two cells from water, neighbours are grass + edge
	wet := NewPlant(grid[1][0], ItemAPPLE) // directly adjacent to water

	assert.Equal(t, PLANT_COOLDOWN[ItemAPPLE], dry.GetGrowthCooldown())
	assert.Equal(t,
		PLANT_COOLDOWN[ItemAPPLE]-PLANT_WATER_COOLDOWN_BOOST[ItemAPPLE],
		wet.GetGrowthCooldown())
}

func TestPlantDamageClearsCellWhenKilled(t *testing.T) {
	// Java rule: when health reaches 0, the plant detaches itself from its
	// cell (cell.setPlant(null)). The Board.Tick sweep then removes it from
	// the live list.
	cell := NewCell(0, 0)
	p := NewPlant(cell, ItemBANANA)
	cell.SetPlant(p)
	p.Health = 3
	p.Damage(2)
	assert.Equal(t, 1, p.Health)
	assert.Equal(t, p, cell.Plant, "still attached")
	p.Damage(10)
	assert.Equal(t, 0, p.Health, "clamped to zero")
	assert.True(t, p.IsDead())
	assert.Nil(t, cell.Plant, "detached on death")
}

func TestPlantHarvestNoopWhenEmpty(t *testing.T) {
	cell := NewCell(0, 0)
	p := NewPlant(cell, ItemAPPLE)
	p.Resources = 1
	p.Harvest()
	assert.Equal(t, 0, p.Resources)
	p.Harvest()
	assert.Equal(t, 0, p.Resources, "stays at zero, no underflow")
}
