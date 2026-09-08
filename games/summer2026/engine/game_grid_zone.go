// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Zone.java
package engine

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Zone.java:6-23

public class Zone {
    private List<Coord> coords;
    private List<Integer> neighbours;
    private List<Town> containedTowns;
    public final int id;

    public static final int FOR_SALE = -1;
    public static final int NEUTRAL = 2;

    public int instability = 0;
    public boolean inked;

    public Zone(int id, List<Coord> coords) {
        this.id = id;
        this.setCoords(coords);
        this.containedTowns = new ArrayList<>(1);
        this.inked = false;
    }
*/

const (
	ZONE_FOR_SALE = -1
	ZONE_NEUTRAL  = 2
)

// Zone is a contiguous region of the grid — the unit disruption and inking
// operate on. Bots see it as `regionId` on every cell.
type Zone struct {
	ID             int
	Coords         []Coord
	Neighbours     []int
	ContainedTowns []*Town
	Instability    int
	Inked          bool
}

func NewZone(id int, coords []Coord) *Zone {
	return &Zone{
		ID:             id,
		Coords:         coords,
		ContainedTowns: []*Town{},
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Zone.java:29-31,54-56

public int getCost() { return getCoords().size() * 2; }
public void addTown(Town t) { containedTowns.add(t); }
*/

func (z *Zone) Cost() int { return len(z.Coords) * 2 }

func (z *Zone) AddTown(t *Town) { z.ContainedTowns = append(z.ContainedTowns, t) }

func (z *Zone) GetContainedTowns() []*Town { return z.ContainedTowns }
