// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Inventory.java
package engine

import (
	"strconv"
	"strings"
)

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Inventory.java:5-60

public class Inventory {
    private int[] items;

    public Inventory() { items = new int[Item.values().length]; }
    public Inventory(Inventory inventory) { ... }
    public void incrementItem(Item item) { items[item.ordinal()]++; }
    public boolean decrementItem(int item) { ... }
    public int getTotal() { return Arrays.stream(items).sum(); }
    public int getItemCount(int item) { return items[item]; }
    public String getInputLine() { ... }
}
*/

// Inventory holds counts indexed by Item ordinal (0..ItemsCount-1).
type Inventory struct {
	Items [ItemsCount]int
}

func NewInventory() *Inventory { return &Inventory{} }

func (inv *Inventory) Clone() *Inventory {
	c := *inv
	return &c
}

func (inv *Inventory) IncrementItem(item Item) { inv.Items[item]++ }

// DecrementItem returns true when it actually decremented (count was > 0).
func (inv *Inventory) DecrementItem(item Item) bool {
	if inv.Items[item] <= 0 {
		return false
	}
	inv.Items[item]--
	return true
}

func (inv *Inventory) SetItem(item Item, value int) { inv.Items[item] = value }

func (inv *Inventory) GetItemCount(item Item) int { return inv.Items[item] }

func (inv *Inventory) GetTotal() int {
	sum := 0
	for _, v := range inv.Items {
		sum += v
	}
	return sum
}

func (inv *Inventory) GetItemsLength() int { return ItemsCount }

// GetInputLine returns the space-separated item counts as expected by the bot
// (matches Java Inventory.getInputLine).
func (inv *Inventory) GetInputLine() string {
	parts := make([]string, 0, ItemsCount)
	for _, v := range inv.Items {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, " ")
}
