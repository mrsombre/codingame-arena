// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Unit.java
package engine

import "strconv"

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:11-22

public class Unit {
    private int id;
    private Player player;
    private Cell cell;
    private int movementSpeed;
    private int carryCapacity;
    private int harvestPower;
    private int chopPower;
    private Inventory inventory = new Inventory();
    public static int idCounter;
*/

type Unit struct {
	ID            int
	Player        *Player
	Cell          *Cell
	MovementSpeed int
	CarryCapacity int
	HarvestPower  int
	ChopPower     int
	Inv           *Inventory
}

// UnitIDCounter mirrors Java's static Unit.idCounter. Reset to 0 at the start
// of every map build (Board.createMap), so ids stay deterministic across
// replays.
var UnitIDCounter int

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:24-30

public static int[] getTrainingCosts(Player player, int[] talents, int league) {
    int baseCost = player.getUnits().size();
    int[] result = new int[Item.values().length];
    for (int i = 0; i <= Item.APPLE.ordinal(); i++) result[i] = baseCost + talents[i] * talents[i];
    if (league >= 3) result[Item.IRON.ordinal()] = baseCost + talents[3] * talents[3];
    return result;
}
*/

// GetTrainingCosts returns the per-item cost array for training a unit with
// the given talents. Talents are [moveSpeed, carryCapacity, harvestPower,
// chopPower]; the chop-power cost (paid in IRON) is only enabled in league 3+.
func GetTrainingCosts(player *Player, talents [4]int, league int) [ItemsCount]int {
	base := len(player.Units)
	var result [ItemsCount]int
	// PLUM(0), LEMON(1), APPLE(2) consume the first three talents.
	for i := 0; i <= int(ItemAPPLE); i++ {
		result[i] = base + talents[i]*talents[i]
	}
	if league >= 3 {
		result[ItemIRON] = base + talents[3]*talents[3]
	}
	return result
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:32-38

public static boolean canTrain(Player player, int[] talents, int league) {
    int[] costs = getTrainingCosts(player, talents, league);
    for (int i = 0; i < costs.length; i++) {
        if (costs[i] > player.getInventory().getItemCount(i)) return false;
    }
    return true;
}
*/

func CanTrain(player *Player, talents [4]int, league int) bool {
	costs := GetTrainingCosts(player, talents, league)
	for i := 0; i < ItemsCount; i++ {
		if costs[i] > player.Inv.GetItemCount(Item(i)) {
			return false
		}
	}
	return true
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:56-69

public Unit(Player player, int[] talents, int league) {
    id = idCounter++;
    int[] costs = getTrainingCosts(player, talents, league);
    for (int i = 0; i < costs.length; i++) {
        player.getInventory().setItem(i, player.getInventory().getItemCount(i) - costs[i]);
    }
    this.player = player;
    player.AddUnit(this);
    this.cell = player.getShack();
    movementSpeed = talents[0];
    carryCapacity = talents[1];
    harvestPower = talents[2];
    chopPower = talents[3];
}
*/

// NewUnit charges the player's inventory for the training cost, registers the
// unit with the player, and seats it on the player's shack. Board.addUnit is
// the caller's responsibility — Java keeps unit registration on the Board too.
func NewUnit(player *Player, talents [4]int, league int) *Unit {
	id := UnitIDCounter
	UnitIDCounter++
	costs := GetTrainingCosts(player, talents, league)
	for i := 0; i < ItemsCount; i++ {
		player.Inv.SetItem(Item(i), player.Inv.GetItemCount(Item(i))-costs[i])
	}
	u := &Unit{
		ID:            id,
		Player:        player,
		Cell:          player.Shack,
		MovementSpeed: talents[0],
		CarryCapacity: talents[1],
		HarvestPower:  talents[2],
		ChopPower:     talents[3],
		Inv:           NewInventory(),
	}
	player.AddUnit(u)
	return u
}

func (u *Unit) GetID() int           { return u.ID }
func (u *Unit) GetPlayer() *Player   { return u.Player }
func (u *Unit) GetCell() *Cell       { return u.Cell }
func (u *Unit) SetCell(c *Cell)      { u.Cell = c }
func (u *Unit) GetChopPower() int    { return u.ChopPower }
func (u *Unit) GetHarvestPower() int { return u.HarvestPower }
func (u *Unit) GetMovementSpeed() int { return u.MovementSpeed }
func (u *Unit) GetCarryCapacity() int { return u.CarryCapacity }
func (u *Unit) GetInventory() *Inventory { return u.Inv }
func (u *Unit) GetFreeCarryCapacity() int { return u.CarryCapacity - u.Inv.GetTotal() }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:90-100

public void harvest(int power) {
    if (power > harvestPower || inventory.getTotal() >= carryCapacity) return;
    inventory.incrementItem(cell.getPlant().getType());
    cell.getPlant().harvest();
}

public void mine() {
    for (int i = 0; i < chopPower && inventory.getTotal() < carryCapacity; i++) {
        inventory.incrementItem(Item.IRON);
    }
}
*/

func (u *Unit) Harvest(power int) {
	if power > u.HarvestPower || u.Inv.GetTotal() >= u.CarryCapacity {
		return
	}
	u.Inv.IncrementItem(u.Cell.Plant.Type)
	u.Cell.Plant.Harvest()
}

func (u *Unit) Mine() {
	for i := 0; i < u.ChopPower && u.Inv.GetTotal() < u.CarryCapacity; i++ {
		u.Inv.IncrementItem(ItemIRON)
	}
}

func (u *Unit) IsNearShack() bool {
	return u.Cell.IsNearShack(u.Player)
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Unit.java:130-133

public String getInputLine(int playerId, int playersCount) {
    int outputId = (playerId + player.getIndex()) % playersCount;
    return id + " " + outputId + " " + cell.getX() + " " + cell.getY() + " " + ... + " " + inventory.getInputLine();
}
*/

// GetInputLine emits the per-unit line for the recipient bot. playerId is the
// recipient's index and playersCount the player count (always 2 here).
func (u *Unit) GetInputLine(playerID, playersCount int) string {
	outputID := (playerID + u.Player.GetIndex()) % playersCount
	return strconv.Itoa(u.ID) + " " +
		strconv.Itoa(outputID) + " " +
		strconv.Itoa(u.Cell.X) + " " +
		strconv.Itoa(u.Cell.Y) + " " +
		strconv.Itoa(u.MovementSpeed) + " " +
		strconv.Itoa(u.CarryCapacity) + " " +
		strconv.Itoa(u.HarvestPower) + " " +
		strconv.Itoa(u.ChopPower) + " " +
		u.Inv.GetInputLine()
}

// itoa is a local helper so engine files can share an int-to-string without
// pulling strconv into every file that builds error/summary text.
func itoa(v int) string { return strconv.Itoa(v) }
