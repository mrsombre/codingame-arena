package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTrainingCostsFormula(t *testing.T) {
	// Java rule: cost[i] = existingTrolls + talents[i]² for each of PLUM,
	// LEMON, APPLE (indices 0..2). League 3+ adds the IRON cost using
	// talents[3]² (chopPower).
	p := NewPlayer(0)
	// pretend they already own 2 trolls
	p.Units = []*Unit{{}, {}}

	got := GetTrainingCosts(p, [4]int{3, 2, 1, 0}, 3)
	assert.Equal(t, 2+3*3, got[ItemPLUM], "PLUM = 2 + 9")
	assert.Equal(t, 2+2*2, got[ItemLEMON], "LEMON = 2 + 4")
	assert.Equal(t, 2+1*1, got[ItemAPPLE], "APPLE = 2 + 1")
	assert.Equal(t, 2+0*0, got[ItemIRON], "IRON = 2 + 0 in league 3")
	assert.Equal(t, 0, got[ItemBANANA], "BANANA never costs anything")
}

func TestGetTrainingCostsLeague2HasNoIronCharge(t *testing.T) {
	// In league 2, chopPower is rejected (TRAIN rejects chop>0) and IRON
	// cost is intentionally left at zero — players can't pay iron yet.
	p := NewPlayer(0)
	got := GetTrainingCosts(p, [4]int{1, 1, 1, 0}, 2)
	assert.Equal(t, 0, got[ItemIRON], "IRON cost suppressed in league 2")
}

func TestCanTrainGuardsInsufficientItem(t *testing.T) {
	p := NewPlayer(0)
	p.Inv.SetItem(ItemPLUM, 1)
	p.Inv.SetItem(ItemLEMON, 1)
	p.Inv.SetItem(ItemAPPLE, 1)
	// costs at zero existing trolls + talents [1,1,1,0] = [1,1,1,0]
	assert.True(t, CanTrain(p, [4]int{1, 1, 1, 0}, 2))

	p.Inv.SetItem(ItemPLUM, 0)
	assert.False(t, CanTrain(p, [4]int{1, 1, 1, 0}, 2), "missing one PLUM blocks the train")
}

func TestUnitHarvestRespectsPowerAndCapacity(t *testing.T) {
	// Java Unit.harvest checks power<=harvestPower and total<carryCapacity.
	// The test exercises one tick of fruit transfer.
	cell := NewCell(0, 0)
	plant := NewPlant(cell, ItemAPPLE)
	plant.Resources = 2
	cell.Plant = plant

	u := &Unit{
		Cell:          cell,
		MovementSpeed: 1,
		CarryCapacity: 3,
		HarvestPower:  2,
		Inv:           NewInventory(),
	}

	u.Harvest(1) // within power, free capacity
	assert.Equal(t, 1, u.Inv.GetItemCount(ItemAPPLE))
	assert.Equal(t, 1, plant.Resources)

	u.Harvest(3) // power > harvestPower → no-op
	assert.Equal(t, 1, u.Inv.GetItemCount(ItemAPPLE))
	assert.Equal(t, 1, plant.Resources)

	u.Inv.SetItem(ItemAPPLE, 3) // at carryCapacity → no-op
	u.Harvest(1)
	assert.Equal(t, 3, u.Inv.GetItemCount(ItemAPPLE))
	assert.Equal(t, 1, plant.Resources, "tree untouched when troll is full")
}

func TestUnitMineCappedByCapacity(t *testing.T) {
	u := &Unit{
		CarryCapacity: 2,
		ChopPower:     5,
		Inv:           NewInventory(),
	}
	u.Mine()
	assert.Equal(t, 2, u.Inv.GetItemCount(ItemIRON),
		"capacity caps the iron pull even when chopPower is higher")
}

func TestUnitGetInputLineSwapsOwnership(t *testing.T) {
	// Java rule: troll line emits ownerId = (recipient + ownerIndex) % 2 so
	// recipient sees their own trolls as player=0 regardless of engine side.
	player1 := NewPlayer(1)
	u := &Unit{
		ID:            7,
		Player:        player1,
		Cell:          NewCell(2, 4),
		MovementSpeed: 1,
		CarryCapacity: 1,
		HarvestPower:  1,
		ChopPower:     1,
		Inv:           NewInventory(),
	}
	// Recipient is player 1 viewing themselves: outputId = (1 + 1) % 2 = 0.
	got := u.GetInputLine(1, 2)
	assert.Contains(t, got, "7 0 2 4 ", "own troll seen as player=0")

	// Recipient is player 0 viewing player 1's troll: outputId = (0 + 1) % 2 = 1.
	got = u.GetInputLine(0, 2)
	assert.Contains(t, got, "7 1 2 4 ", "opponent troll seen as player=1")
}
