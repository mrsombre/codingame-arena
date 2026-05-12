// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/Plant.java
package engine

import "strconv"

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Plant.java:8-22

public class Plant {
    private int size;
    private int health;
    private int resources;
    private int cooldown;
    private Cell cell;
    private Item type;

    public Plant(Cell cell, Item type) {
        this.cell = cell;
        this.type = type;
        this.health = Constants.PLANT_FINAL_HEALTH[type.ordinal()] - Constants.PLANT_DELTA_HEALTH[type.ordinal()] * Constants.PLANT_MAX_SIZE;
    }
*/

type Plant struct {
	Size      int
	Health    int
	Resources int
	Cooldown  int
	Cell      *Cell
	Type      Item // PLUM/LEMON/APPLE/BANANA
}

func NewPlant(cell *Cell, kind Item) *Plant {
	return &Plant{
		Cell:   cell,
		Type:   kind,
		Health: PLANT_FINAL_HEALTH[kind] - PLANT_DELTA_HEALTH[kind]*PLANT_MAX_SIZE,
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Plant.java:32-51

public void tick(boolean updateHealth) {
    if (cooldown > 0) cooldown--;
    if (cooldown == 0 && health > 0) {
        if (size < Constants.PLANT_MAX_SIZE) {
            size++;
            if (updateHealth) health += Constants.PLANT_DELTA_HEALTH[type.ordinal()];
            cooldown = getGrowthCooldown();
        } else if (resources < Constants.PLANT_MAX_RESOURCES) {
            resources++;
            cooldown = getGrowthCooldown();
        }
    }
}

public int getGrowthCooldown() {
    int cooldown = Constants.PLANT_COOLDOWN[type.ordinal()];
    if (cell.isNearWater()) cooldown -= Constants.PLANT_WATER_COOLDOWN_BOOST[type.ordinal()];
    return cooldown;
}
*/

func (p *Plant) Tick(updateHealth bool) {
	if p.Cooldown > 0 {
		p.Cooldown--
	}
	if p.Cooldown == 0 && p.Health > 0 {
		if p.Size < PLANT_MAX_SIZE {
			p.Size++
			if updateHealth {
				p.Health += PLANT_DELTA_HEALTH[p.Type]
			}
			p.Cooldown = p.GetGrowthCooldown()
		} else if p.Resources < PLANT_MAX_RESOURCES {
			p.Resources++
			p.Cooldown = p.GetGrowthCooldown()
		}
	}
}

func (p *Plant) GetGrowthCooldown() int {
	c := PLANT_COOLDOWN[p.Type]
	if p.Cell.IsNearWater() {
		c -= PLANT_WATER_COOLDOWN_BOOST[p.Type]
	}
	return c
}

func (p *Plant) GetType() Item   { return p.Type }
func (p *Plant) GetCell() *Cell  { return p.Cell }
func (p *Plant) GetResources() int { return p.Resources }
func (p *Plant) GetHealth() int  { return p.Health }
func (p *Plant) GetSize() int    { return p.Size }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Plant.java:69-76

public void harvest() { if (resources > 0) resources--; }
public void damage(int damage) {
    health = Math.max(health - damage, 0);
    if (isDead()) cell.setPlant(null);
}
*/

func (p *Plant) Harvest() {
	if p.Resources > 0 {
		p.Resources--
	}
}

func (p *Plant) Damage(d int) {
	p.Health -= d
	if p.Health < 0 {
		p.Health = 0
	}
	if p.IsDead() {
		p.Cell.SetPlant(nil)
	}
}

func (p *Plant) IsDead() bool { return p.Health <= 0 }

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/Plant.java:90-92

public String getInputLine() {
    return type + " " + cell.getX() + " " + cell.getY() + " " + size + " " + health + " " + resources + " " + cooldown;
}
*/

func (p *Plant) GetInputLine() string {
	return p.Type.String() + " " +
		strconv.Itoa(p.Cell.X) + " " +
		strconv.Itoa(p.Cell.Y) + " " +
		strconv.Itoa(p.Size) + " " +
		strconv.Itoa(p.Health) + " " +
		strconv.Itoa(p.Resources) + " " +
		strconv.Itoa(p.Cooldown)
}
