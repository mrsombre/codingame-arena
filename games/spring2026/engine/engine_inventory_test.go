package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventoryIncrementAndTotal(t *testing.T) {
	inv := NewInventory()
	inv.IncrementItem(ItemAPPLE)
	inv.IncrementItem(ItemAPPLE)
	inv.IncrementItem(ItemWOOD)
	assert.Equal(t, 2, inv.GetItemCount(ItemAPPLE))
	assert.Equal(t, 1, inv.GetItemCount(ItemWOOD))
	assert.Equal(t, 3, inv.GetTotal())
}

func TestInventoryDecrementBoundary(t *testing.T) {
	// Java rule: decrementItem returns false (no-op) when count is already 0.
	// Used by Pick / Plant to validate the player can still pay.
	inv := NewInventory()
	assert.False(t, inv.DecrementItem(ItemPLUM), "no plums to take")

	inv.SetItem(ItemPLUM, 1)
	assert.True(t, inv.DecrementItem(ItemPLUM))
	assert.Equal(t, 0, inv.GetItemCount(ItemPLUM))
	assert.False(t, inv.DecrementItem(ItemPLUM), "drained")
}

func TestInventoryCloneIsIndependent(t *testing.T) {
	// Plant.Plant(Plant) and Unit.Unit(Unit) rely on Inventory copies that
	// don't share storage with the source.
	a := NewInventory()
	a.SetItem(ItemAPPLE, 5)
	b := a.Clone()
	b.SetItem(ItemAPPLE, 0)
	assert.Equal(t, 5, a.GetItemCount(ItemAPPLE), "source untouched")
	assert.Equal(t, 0, b.GetItemCount(ItemAPPLE))
}

func TestInventoryGetInputLineEmits6FieldsInItemOrder(t *testing.T) {
	// Java rule: getInputLine joins counts in PLUM/LEMON/APPLE/BANANA/IRON/
	// WOOD order, space-separated. Bot stdin contract depends on this shape.
	inv := NewInventory()
	inv.SetItem(ItemPLUM, 1)
	inv.SetItem(ItemAPPLE, 3)
	inv.SetItem(ItemWOOD, 4)
	assert.Equal(t, "1 0 3 0 0 4", inv.GetInputLine())
}
