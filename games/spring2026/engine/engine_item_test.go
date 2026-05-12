package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemStringMatchesOrdinal(t *testing.T) {
	// Item.String must align with Java enum ordinals: PLUM=0, LEMON=1,
	// APPLE=2, BANANA=3, IRON=4, WOOD=5. Bot input uses these names verbatim.
	assert.Equal(t, "PLUM", ItemPLUM.String())
	assert.Equal(t, "LEMON", ItemLEMON.String())
	assert.Equal(t, "APPLE", ItemAPPLE.String())
	assert.Equal(t, "BANANA", ItemBANANA.String())
	assert.Equal(t, "IRON", ItemIRON.String())
	assert.Equal(t, "WOOD", ItemWOOD.String())
}

func TestItemIsPlantOnlyForFruits(t *testing.T) {
	// Java rule: isPlant() == (ordinal() <= BANANA.ordinal()). IRON/WOOD are
	// items but not seeds; PLANT and PICK reject them with INVALID_PLANT.
	for _, fruit := range []Item{ItemPLUM, ItemLEMON, ItemAPPLE, ItemBANANA} {
		assert.Truef(t, fruit.IsPlant(), "%s is a plant", fruit)
	}
	assert.False(t, ItemIRON.IsPlant())
	assert.False(t, ItemWOOD.IsPlant())
}

func TestItemFromNameMatchesAllKnownAndRejectsRest(t *testing.T) {
	cases := map[string]Item{
		"PLUM": ItemPLUM, "LEMON": ItemLEMON, "APPLE": ItemAPPLE,
		"BANANA": ItemBANANA, "IRON": ItemIRON, "WOOD": ItemWOOD,
	}
	for name, want := range cases {
		assert.Equalf(t, want, ItemFromName(name), "round-trip for %s", name)
	}
	assert.Equal(t, Item(-1), ItemFromName("STONE"), "unknown returns sentinel")
	assert.Equal(t, Item(-1), ItemFromName(""), "empty returns sentinel")
}
