// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Item.java
package engine

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Item.java:3-14

public enum Item {
    PLUM,
    LEMON,
    APPLE,
    BANANA,
    IRON,
    WOOD;

    public boolean isPlant() {
        return this.ordinal() <= BANANA.ordinal();
    }
}
*/

type Item int

const (
	ItemPLUM Item = iota
	ItemLEMON
	ItemAPPLE
	ItemBANANA
	ItemIRON
	ItemWOOD
)

// ItemsCount is the total number of items, matching Java Item.values().length.
const ItemsCount = 6

var itemNames = [ItemsCount]string{"PLUM", "LEMON", "APPLE", "BANANA", "IRON", "WOOD"}

func (i Item) String() string { return itemNames[i] }

// IsPlant returns true for PLUM/LEMON/APPLE/BANANA, false for IRON/WOOD.
func (i Item) IsPlant() bool { return i <= ItemBANANA }

// ItemFromName maps an uppercase name (PLUM, LEMON, ...) to its Item value.
// Returns -1 when unknown — caller is expected to validate.
func ItemFromName(name string) Item {
	for i, n := range itemNames {
		if n == name {
			return Item(i)
		}
	}
	return -1
}
