package engine;

import java.util.Arrays;

public class Inventory {
    private int[] items;

    public Inventory() {
        items = new int[Item.values().length];
    }

    public Inventory(Inventory inventory) {
        items = new int[Item.values().length];
        System.arraycopy(inventory.items, 0, items, 0, items.length);
    }

    public void incrementItem(Item item) {
        incrementItem(item.ordinal());
    }

    public void incrementItem(int item) {
        items[item]++;
    }

    public void setItem(int index, int value) {
        items[index] = value;
    }

    public boolean decrementItem(Item item) {
        return decrementItem(item.ordinal());
    }

    public boolean decrementItem(int item) {
        if (items[item] <= 0) return false;
        items[item]--;
        return true;
    }

    public int getTotal() {
        return Arrays.stream(items).sum();
    }

    public int getItemsLength() {
        return items.length;
    }

    public int getItemCount(Item item) {
        return getItemCount(item.ordinal());
    }

    public int getItemCount(int item) {
        return items[item];
    }

    public String getInputLine() {
        String output = "";
        for (int i : items) output += " " + i;
        return output.trim();
    }
}
