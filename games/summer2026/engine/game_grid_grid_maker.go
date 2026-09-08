// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java
package engine

import (
	"math"

	"github.com/mrsombre/codingame-arena/internal/util/javahash"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// javaRoundFloat is Math.round(float): floor(x + 0.5), rounding halves
// towards positive infinity. Go's math.Round rounds halves away from zero,
// so -2.5 comes out -3 there and -2 here.
func javaRoundFloat(v float32) int {
	return int(math.Floor(float64(v) + 0.5))
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:20-38

private class River {
    public Coord current; // equal to last item in history
    public List<Coord> history;
    public Direction preferredDirection;

    public River(Coord coord, List<Coord> history, Direction preferredDirection) {
        this.current = coord;
        this.history = new ArrayList<>(history);
        this.history.add(coord);
        this.preferredDirection = preferredDirection;
    }

    boolean isStart() { return history.size() == 1; }
}
*/

type river struct {
	current            Coord
	history            []Coord
	preferredDirection Direction
}

func newRiver(coord Coord, history []Coord, preferredDirection Direction) *river {
	h := make([]Coord, len(history), len(history)+1)
	copy(h, history)
	h = append(h, coord)
	return &river{current: coord, history: h, preferredDirection: preferredDirection}
}

func (r *river) isStart() bool { return len(r.history) == 1 }

// GridMaker builds a map from an RNG stream. Every ordering decision here is
// load-bearing for seed parity: LinkedList poll/remove semantics, the
// insertion order of the river weight map, and — through
// getAvailableNeighbours — java.util.HashSet bucket order.
type GridMaker struct {
	random          *sha1prng.Random
	h, w            int
	enableSideQuest bool
	freeBorders     []Coord
	freeCoords      []Coord
	grid            *Grid
}

func NewGridMaker() *GridMaker { return &GridMaker{} }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:47-58

private boolean isCorner(int x, int y) {
    return (x == 0 && y == 0) || (x == 0 && y == h - 1) || (x == w - 1 && y == 0) || (x == w - 1 && y == h - 1);
}

private boolean isEdge(int x, int y) { return x == 0 || y == 0 || x == w - 1 || y == h - 1; }
*/

func (m *GridMaker) isCorner(x, y int) bool {
	return (x == 0 && y == 0) ||
		(x == 0 && y == m.h-1) ||
		(x == m.w-1 && y == 0) ||
		(x == m.w-1 && y == m.h-1)
}

func (m *GridMaker) isEdge(x, y int) bool {
	return x == 0 || y == 0 || x == m.w-1 || y == m.h-1
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:60-72

private boolean hasWaterNearby(Grid grid, Coord tile, List<Coord> riverHistory) {
    List<Coord> ignoreCoords = riverHistory.subList(Math.max(riverHistory.size() - 2, 0), riverHistory.size());
    var neighs = grid.getNeighbours(tile, Grid.ADJACENCY_8);
    return neighs.stream().filter(n -> !ignoreCoords.contains(n)).anyMatch(n -> grid.get(n).isWater());
}

private boolean isAtNFromSides(Coord coord, int n) {
    return coord.x - n < 0 || coord.y - n < 0 || coord.x + n > w - 1 || coord.y + n > h - 1;
}
*/

// hasWaterNearby ignores the last two cells of the river's own history, so a
// river does not see the water it just laid down behind itself.
func (m *GridMaker) hasWaterNearby(grid *Grid, tile Coord, riverHistory []Coord) bool {
	ignore := riverHistory[max(len(riverHistory)-2, 0):]
	for _, n := range grid.NeighboursWith(tile, ADJACENCY_8) {
		if contains(ignore, n) {
			continue
		}
		if grid.Get(n).IsWater() {
			return true
		}
	}
	return false
}

func (m *GridMaker) isAtNFromSides(coord Coord, n int) bool {
	return coord.X-n < 0 || coord.Y-n < 0 || coord.X+n > m.w-1 || coord.Y+n > m.h-1
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:74-90

private List<Coord> getAvailableNeighbours(Grid grid, List<Coord> coords, Predicate<Tile> checkAccessible) {
    Set<Coord> availables = new HashSet<>();
    for (Coord coord : coords) {
        List<Coord> neighs = grid.getNeighbours(coord);
        for (Coord neigh : neighs) {
            if (checkAccessible.test(grid.get(neigh))) availables.add(neigh);
        }
    }
    return availables.stream().toList();
}

private List<Coord> getMountainAvailableNeighbours(Grid grid, List<Coord> mountains) {
    return getAvailableNeighbours(grid, mountains, this::isAccessible);
}
*/

// getAvailableNeighbours returns candidates in java.util.HashSet iteration
// order, because the callers index into the result with nextInt(size). Any
// other order still produces a valid map, but not the seed's map.
func (m *GridMaker) getAvailableNeighbours(grid *Grid, coords []Coord, checkAccessible func(*Tile) bool) []Coord {
	availables := javahash.NewSet[Coord]()
	for _, coord := range coords {
		for _, neigh := range grid.Neighbours(coord) {
			if checkAccessible(grid.Get(neigh)) {
				availables.Add(neigh)
			}
		}
	}
	return availables.Values()
}

func (m *GridMaker) getMountainAvailableNeighbours(grid *Grid, mountains []Coord) []Coord {
	return m.getAvailableNeighbours(grid, mountains, m.isAccessible)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:92-97

public void init(Random random, boolean enableSideQuest) {
    this.random = random;
    this.enableSideQuest = enableSideQuest;
    this.h = random.nextInt(Game.MIN_GRID_HEIGHT, Game.MAX_GRID_HEIGHT + 1);
    this.w = Math.round(h * Game.ASPECT_RATIO);
}
*/

// Init draws the grid dimensions. h * ASPECT_RATIO is a float32 expression in
// Java; widening it to float64 changes the rounding for some heights.
func (m *GridMaker) Init(random *sha1prng.Random, enableSideQuest bool) {
	m.random = random
	m.enableSideQuest = enableSideQuest
	m.h = random.NextIntRange(MIN_GRID_HEIGHT, MAX_GRID_HEIGHT+1)
	m.w = javaRoundFloat(float32(m.h) * ASPECT_RATIO)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:99-121

private void initializeGrid() {
    LinkedList<Coord> freeBorders = new LinkedList<>();
    this.grid = new Grid(w, h, true);
    for (int y = 0; y < h; ++y) for (int x = 0; x < w; ++x) {
        grid.get(x, y).setType(Tile.TYPE_GRASS);
        if (isEdge(x, y) && !isCorner(x, y)) freeBorders.add(grid.get(x, y).coord);
    }
    LinkedList<Coord> freeCoords = new LinkedList<>(grid.cells.values().stream().map(t -> t.coord).toList());
    Collections.shuffle(freeBorders, random);
    Collections.shuffle(freeCoords, random);
    this.freeBorders = freeBorders;
    this.freeCoords = freeCoords;
}
*/

func (m *GridMaker) initializeGrid() {
	freeBorders := make([]Coord, 0, 2*(m.w+m.h))
	m.grid = NewGridSym(m.w, m.h, true)
	for y := 0; y < m.h; y++ {
		for x := 0; x < m.w; x++ {
			tile := m.grid.GetXY(x, y)
			tile.SetType(TYPE_GRASS)
			if m.isEdge(x, y) && !m.isCorner(x, y) {
				freeBorders = append(freeBorders, tile.Coord)
			}
		}
	}

	freeCoords := m.grid.Coords()

	sha1prng.Shuffle(m.random, freeBorders)
	sha1prng.Shuffle(m.random, freeCoords)

	m.freeBorders = freeBorders
	m.freeCoords = freeCoords
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:123-147

public Grid make() {
    this.makingOf = new ArrayList<>();
    initializeGrid();
    makeMountains();
    makeRivers();
    int averageTilesPerZone = h / Game.AVERAGE_TILES_PER_ZONE_COEFF_TO_GRID_HEIGHT;
    int nZones = Math.max(1, (h * w) / averageTilesPerZone);
    List<Zone> zones = makeZones(nZones);
    int nTowns = Math.max(4, (h * w) / Game.AVERAGE_TILES_PER_TOWN);
    List<Town> towns = makeTowns(zones, nTowns, averageTilesPerZone);
    makeTownConnections(towns);
    makePOIs(towns);
    grid.towns = towns;
    grid.zones = zones;
    return grid;
}
*/

// Make builds the whole map. `makingOf`, a per-step clone list feeding the
// viewer's generation animation, is not ported: it consumes no randomness.
func (m *GridMaker) Make() *Grid {
	m.initializeGrid()

	m.makeMountains()
	m.makeRivers()

	averageTilesPerZone := m.h / AVERAGE_TILES_PER_ZONE_COEFF_TO_GRID_HEIGHT
	nZones := max(1, (m.h*m.w)/averageTilesPerZone)
	zones := m.makeZones(nZones)

	nTowns := max(4, (m.h*m.w)/AVERAGE_TILES_PER_TOWN)
	towns := m.makeTowns(zones, nTowns, averageTilesPerZone)

	m.makeTownConnections(towns)
	m.makePOIs(towns)

	m.grid.Towns = towns
	m.grid.Zones = zones

	return m.grid
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:149-176

private void makePOIs(List<Town> towns) {
    if (!this.enableSideQuest) return;
    int nPois = random.nextInt(3) + 2;
    LinkedList<Coord> sideCoords = new LinkedList<>(
        grid.cells.values().stream().map(t -> t.coord)
            .filter(c -> isAtNFromSides(c, 3))
            .filter(c -> towns.stream().allMatch(t -> t.coord.manhattanTo(c) >= Game.MIN_TOWN_DISTANCE))
            .filter(c -> !grid.get(c).isWater()).toList());
    Collections.shuffle(sideCoords, random);
    for (int i = 0; i < nPois; ++i) {
        if (!sideCoords.isEmpty()) {
            Coord coord = sideCoords.poll();
            grid.get(coord).setType(Tile.TYPE_POI);
            grid.setPoi(coord);
        }
    }
}
*/

// The isAtNFromSides filter keeps coords within 3 of a side, despite its name.
func (m *GridMaker) makePOIs(towns []*Town) {
	if !m.enableSideQuest {
		return
	}
	nPois := m.random.NextInt(3) + 2

	sideCoords := make([]Coord, 0, len(m.grid.Cells))
	for _, tile := range m.grid.Cells {
		c := tile.Coord
		if !m.isAtNFromSides(c, 3) {
			continue
		}
		farEnough := true
		for _, t := range towns {
			if t.Coord.ManhattanTo(c) < MIN_TOWN_DISTANCE {
				farEnough = false
				break
			}
		}
		if !farEnough || m.grid.Get(c).IsWater() {
			continue
		}
		sideCoords = append(sideCoords, c)
	}

	sha1prng.Shuffle(m.random, sideCoords)
	for i := 0; i < nPois; i++ {
		if len(sideCoords) == 0 {
			continue
		}
		var coord Coord
		coord, sideCoords = sideCoords[0], sideCoords[1:]
		m.grid.Get(coord).SetType(TYPE_POI)
		m.grid.SetPOI(coord)
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:178-193

private void makeTownConnections(List<Town> towns) {
    for (Town t : towns) {
        List<Town> otherTowns = new ArrayList<>(towns.stream().filter(ot -> !ot.equals(t)).toList());
        Collections.shuffle(otherTowns, random);
        int atLeast = Math.min(3, otherTowns.size());
        int atMost = Math.max(atLeast, otherTowns.size() - 4);
        t.desiredConnections = otherTowns.subList(0, random.nextInt(atLeast, atMost + 1));
        t.desiredConnections.sort(Comparator.comparingInt(Town::getId));
    }

    // Remove reciprocal connections
    for (Town t : towns) {
        t.desiredConnections = t.desiredConnections.stream().filter(ot -> !ot.desiredConnections.contains(t)).toList();
    }
}
*/

// makeTownConnections gives every town a shuffled subset of the others, then
// strips any connection whose target already asks for this town. The second
// loop mutates in place as it goes, so a town is filtered against the
// already-pruned lists of the towns before it.
func (m *GridMaker) makeTownConnections(towns []*Town) {
	for _, t := range towns {
		otherTowns := make([]*Town, 0, len(towns))
		for _, ot := range towns {
			if ot != t {
				otherTowns = append(otherTowns, ot)
			}
		}
		sha1prng.Shuffle(m.random, otherTowns)

		atLeast := min(3, len(otherTowns))
		atMost := max(atLeast, len(otherTowns)-4)
		n := m.random.NextIntRange(atLeast, atMost+1)

		t.DesiredConnections = otherTowns[:n]
		javaListSort(t.DesiredConnections, func(x, y *Town) int { return compareInt(x.ID, y.ID) })
	}

	for _, t := range towns {
		kept := make([]*Town, 0, len(t.DesiredConnections))
		for _, ot := range t.DesiredConnections {
			if !contains(ot.DesiredConnections, t) {
				kept = append(kept, ot)
			}
		}
		t.DesiredConnections = kept
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:195-269

private List<Town> makeTowns(List<Zone> zones, int nTowns, int averageTilesPerZone) {
    LinkedList<Zone> availableZones = new LinkedList<>(zones);
    List<Zone> blacklist = new ArrayList<>(zones.size());
    List<Town> towns = new ArrayList<>();
    int step = Math.max(1, Game.AVERAGE_TILES_PER_TOWN / averageTilesPerZone);
    int retries = 100;
    if (availableZones.isEmpty()) return towns;
    for (int i = 0; i < nTowns; i++) {
        int index = random.nextInt(i * step, (i + 1) * step);
        Zone zone = availableZones.get(index % availableZones.size());
        boolean found = false;
        if (!blacklist.contains(zone)) {
            LinkedList<Coord> townCoords = new LinkedList<>(zone.getCoords());
            Collections.shuffle(townCoords, random);
            while (!townCoords.isEmpty()) {
                Coord townCoord = townCoords.poll();
                if (isAccessible(grid.get(townCoord)) && !isEdge(townCoord.x, townCoord.y)
                        && towns.stream().allMatch(town -> townCoord.manhattanTo(town.coord) >= Game.MIN_TOWN_DISTANCE)) {
                    towns.add(new Town(i, townCoord));
                    found = true;
                    blacklist.add(zone);
                    zone.getNeighbours().stream().map(zid -> zones.get(zid)).forEach(blacklist::add);
                    break;
                }
            }
        }
        if (!found && retries > 0) { i--; retries--; }
    }

    int townsLeftToPlace = nTowns - towns.size();
    availableZones.removeAll(blacklist);
    Collections.shuffle(availableZones, random);
    for (int i = 0; i < townsLeftToPlace; ++i) {
        if (availableZones.isEmpty()) break;
        Zone z = availableZones.poll();
        if (blacklist.contains(z)) continue;
        LinkedList<Coord> townCoords = new LinkedList<>(z.getCoords());
        Collections.shuffle(townCoords, random);
        while (!townCoords.isEmpty()) {
            Coord townCoord = townCoords.poll();
            if (towns.stream().allMatch(town -> townCoord.manhattanTo(town.coord) >= Game.MIN_TOWN_DISTANCE)) {
                towns.add(new Town(i, townCoord));
                z.getNeighbours().stream().map(zid -> zones.get(zid)).forEach(blacklist::add);
                break;
            }
        }
    }

    int idx = 0;
    for (Town town : towns) { town.id = idx++; }

    for (Town town : towns) {
        Tile tile = grid.get(town.coord);
        Zone zone = zones.get(tile.getZoneId());
        zone.addTown(town);
        tile.townId = town.id;
        tile.setType(Tile.TYPE_GRASS);
    }
    return towns;
}
*/

// makeTowns places one town per zone, blacklisting the chosen zone and its
// neighbours so towns spread out.
//
// UPSTREAM QUIRK — the second loop. It runs when the primary loop burns
// through its 100 retries without placing nTowns towns, and it is materially
// laxer than the first: it drops the isAccessible check (so a town can land
// on river or mountain), drops the not-on-edge check, and constructs
// Town(i, ...) from its own loop counter, which restarts at 0 and can collide
// with an id the primary loop already used. The final renumbering pass
// papers over the duplicate ids, but the terrain and edge placements survive.
// This is faithful to upstream and must not be "fixed" — doing so diverges
// every generated map from the server's.
func (m *GridMaker) makeTowns(zones []*Zone, nTowns, averageTilesPerZone int) []*Town {
	availableZones := make([]*Zone, len(zones))
	copy(availableZones, zones)
	blacklist := make([]*Zone, 0, len(zones))
	towns := make([]*Town, 0, nTowns)

	step := max(1, AVERAGE_TILES_PER_TOWN/averageTilesPerZone)
	retries := 100
	if len(availableZones) == 0 {
		return towns
	}

	for i := 0; i < nTowns; i++ {
		index := m.random.NextIntRange(i*step, (i+1)*step)
		zone := availableZones[index%len(availableZones)]
		found := false

		if !contains(blacklist, zone) {
			townCoords := make([]Coord, len(zone.Coords))
			copy(townCoords, zone.Coords)
			sha1prng.Shuffle(m.random, townCoords)

			for len(townCoords) > 0 {
				var townCoord Coord
				townCoord, townCoords = townCoords[0], townCoords[1:]
				if m.isAccessible(m.grid.Get(townCoord)) &&
					!m.isEdge(townCoord.X, townCoord.Y) &&
					m.farFromEveryTown(towns, townCoord) {
					towns = append(towns, NewTown(i, townCoord))
					found = true
					blacklist = append(blacklist, zone)
					for _, zid := range zone.Neighbours {
						blacklist = append(blacklist, zones[zid])
					}
					break
				}
			}
		}

		if !found && retries > 0 {
			i--
			retries--
		}
	}

	townsLeftToPlace := nTowns - len(towns)
	availableZones = removeAll(availableZones, blacklist)
	sha1prng.Shuffle(m.random, availableZones)

	for i := 0; i < townsLeftToPlace; i++ {
		if len(availableZones) == 0 {
			break
		}
		var z *Zone
		z, availableZones = availableZones[0], availableZones[1:]
		if contains(blacklist, z) {
			continue
		}

		townCoords := make([]Coord, len(z.Coords))
		copy(townCoords, z.Coords)
		sha1prng.Shuffle(m.random, townCoords)

		for len(townCoords) > 0 {
			var townCoord Coord
			townCoord, townCoords = townCoords[0], townCoords[1:]
			if m.farFromEveryTown(towns, townCoord) {
				towns = append(towns, NewTown(i, townCoord))
				for _, zid := range z.Neighbours {
					blacklist = append(blacklist, zones[zid])
				}
				break
			}
		}
	}

	for idx, town := range towns {
		town.ID = idx
	}

	for _, town := range towns {
		tile := m.grid.Get(town.Coord)
		zone := zones[tile.GetZoneID()]
		zone.AddTown(town)
		tile.TownID = town.ID
		tile.SetType(TYPE_GRASS)
	}

	return towns
}

func (m *GridMaker) farFromEveryTown(towns []*Town, coord Coord) bool {
	for _, town := range towns {
		if coord.ManhattanTo(town.Coord) < MIN_TOWN_DISTANCE {
			return false
		}
	}
	return true
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:271-340

private List<Zone> makeZones(int nZones) {
    int cols = (int) Math.ceil(Math.sqrt(nZones));
    int rows = (int) Math.ceil((double) nZones / cols);
    double cellH = (double) h / rows;
    List<Zone> zones = new ArrayList<>();
    for (int i = 0; i < nZones; i++) {
        int row = i / cols, col = i % cols;
        double x, y;
        if (row == rows - 1 && nZones % cols != 0) {
            int lastRowPoints = nZones % cols;
            x = (col + 0.5) * ((double) w / lastRowPoints);
        } else {
            x = (col + 0.5) * ((double) w / cols);
        }
        y = (row + 0.5) * cellH;
        Coord zoneCenter = new Coord((int) x, (int) y);
        Zone zoneSeed = new Zone(i, new ArrayList<>(List.of(zoneCenter)));
        grid.get(zoneCenter).setZoneId(i);
        zones.add(zoneSeed);
    }

    int zoneId = 0;
    Set<Integer> blockedZones = new HashSet<>();
    while (true) {
        if (!blockedZones.contains(zoneId)) {
            if (zones.size() <= zoneId) return zones;
            Zone zone = zones.get(zoneId);
            List<Coord> neighs = getAvailableNeighbours(grid, zone.getCoords(), t -> t.getZoneId() == -1);
            if (neighs.isEmpty()) {
                blockedZones.add(zoneId);
                if (blockedZones.size() == nZones) break;
            } else {
                Coord neighToAdd = neighs.get(random.nextInt(neighs.size()));
                grid.get(neighToAdd).setZoneId(zoneId);
                zone.getCoords().add(neighToAdd);
            }
        }
        zoneId = (zoneId + 1) % nZones;
    }

    for (Zone zone : zones) {
        Set<Integer> allNeighs = new HashSet<>();
        for (Coord coord : zone.getCoords()) {
            for (Coord neigh : grid.getNeighbours(coord)) {
                int otherId = grid.get(neigh).getZoneId();
                if (otherId != zone.id) allNeighs.add(otherId);
            }
        }
        zone.setNeighbours(allNeighs.stream().toList());
    }
    return zones;
}
*/

// makeZones seeds one zone per grid subdivision and grows them round-robin,
// one cell each, until nothing unclaimed remains adjacent to any of them.
//
// Two seeds can land on the same cell when the subdivision is degenerate; the
// later zone wins the cell's zoneId while both keep it in their coord list,
// as in Java.
func (m *GridMaker) makeZones(nZones int) []*Zone {
	cols := int(math.Ceil(math.Sqrt(float64(nZones))))
	rows := int(math.Ceil(float64(nZones) / float64(cols)))
	cellH := float64(m.h) / float64(rows)

	zones := make([]*Zone, 0, nZones)
	for i := 0; i < nZones; i++ {
		row, col := i/cols, i%cols
		var x float64
		if row == rows-1 && nZones%cols != 0 {
			lastRowPoints := nZones % cols
			x = (float64(col) + 0.5) * (float64(m.w) / float64(lastRowPoints))
		} else {
			x = (float64(col) + 0.5) * (float64(m.w) / float64(cols))
		}
		y := (float64(row) + 0.5) * cellH

		zoneCenter := Coord{int(x), int(y)}
		zones = append(zones, NewZone(i, []Coord{zoneCenter}))
		m.grid.Get(zoneCenter).SetZoneID(i)
	}

	zoneID := 0
	blockedZones := map[int]bool{}
	for {
		if !blockedZones[zoneID] {
			if len(zones) <= zoneID {
				return zones
			}
			zone := zones[zoneID]
			neighs := m.getAvailableNeighbours(m.grid, zone.Coords, func(t *Tile) bool {
				return t.GetZoneID() == -1
			})
			if len(neighs) == 0 {
				blockedZones[zoneID] = true
				if len(blockedZones) == nZones {
					break
				}
			} else {
				neighToAdd := neighs[m.random.NextInt(len(neighs))]
				m.grid.Get(neighToAdd).SetZoneID(zoneID)
				zone.Coords = append(zone.Coords, neighToAdd)
			}
		}
		zoneID = (zoneID + 1) % nZones
	}

	for _, zone := range zones {
		allNeighs := javahash.NewSet[javaInt]()
		for _, coord := range zone.Coords {
			for _, neigh := range m.grid.Neighbours(coord) {
				otherID := m.grid.Get(neigh).GetZoneID()
				if otherID != zone.ID {
					allNeighs.Add(javaInt(otherID))
				}
			}
		}
		ids := make([]int, 0, allNeighs.Len())
		for _, v := range allNeighs.Values() {
			ids = append(ids, int(v))
		}
		zone.Neighbours = ids
	}

	return zones
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:342-431

private void makeRivers() {
    int nRiverCells = Math.round(w * h * Game.RIVER_TO_LAND_MIN_RATIO);
    LinkedList<Coord> availableRiverSources = new LinkedList<>(freeBorders);
    if (random.nextBoolean()) availableRiverSources.addFirst(new Coord(w / 2, h / 2));
    else availableRiverSources.addFirst(new Coord(w / 2, h / 2 + 1));

    boolean initialRiver = true;
    List<River> generatedRivers = new ArrayList<>();

    while (nRiverCells > 0 && !availableRiverSources.isEmpty()) {
        Coord riverStart = availableRiverSources.poll();
        freeBorders.remove(riverStart);
        Direction direction = getDirectionFromRiverStart(riverStart);
        if (hasWaterNearby(grid, riverStart, List.of())) continue;
        LinkedList<River> riversToExpand = new LinkedList<>();
        createWater(riversToExpand, new River(riverStart, new ArrayList<>(), direction));
        nRiverCells--;

        while (!riversToExpand.isEmpty()) {
            makingOf.add(grid.clone());
            River river = riversToExpand.poll();
            Coord current = river.current;
            if (isEdge(current.x, current.y) && !river.isStart()) { generatedRivers.add(river); continue; }
            List<Coord> neighs = getAvailableNeighboursForRiverToFlow(river);
            if (neighs.isEmpty()) { generatedRivers.add(river); continue; }
            boolean goingToSplit = nRiverCells > 0 && neighs.size() >= 2 && random.nextFloat() <= Game.RIVER_SPLIT_PROBA;
            goingToSplit |= initialRiver;
            initialRiver = false;
            LinkedHashMap<Coord, Float> weights = createRiverFlowWeights(river, neighs);
            Coord nextCoord = getRandomCoord(weights);
            River nextRiver = new River(nextCoord, river.history,
                goingToSplit ? Direction.fromCoord(nextCoord.sub(river.current)) : river.preferredDirection);
            createWater(riversToExpand, nextRiver);
            nRiverCells--;
            if (!goingToSplit) continue;
            List<Coord> remainingNeighs = new ArrayList<>(neighs.stream().filter(c -> !c.equals(nextCoord)).toList());
            Collections.shuffle(remainingNeighs, random);
            if (remainingNeighs.isEmpty()) continue;
            Coord splitCoord = remainingNeighs.get(0);
            nextRiver = new River(splitCoord, river.history, Direction.fromCoord(splitCoord.sub(river.current)));
            createWater(riversToExpand, nextRiver);
            nRiverCells--;
        }
    }
    deleteShortRivers(generatedRivers);
}
*/

// makeRivers floods water outward from border sources. The first source is
// always the grid centre (one of two cells, chosen by a coin flip), and the
// very first expansion always splits regardless of the split probability.
func (m *GridMaker) makeRivers() {
	nRiverCells := javaRoundFloat(float32(m.w*m.h) * RIVER_TO_LAND_MIN_RATIO)

	availableRiverSources := make([]Coord, 0, len(m.freeBorders)+1)
	if m.random.NextBoolean() {
		availableRiverSources = append(availableRiverSources, Coord{m.w / 2, m.h / 2})
	} else {
		availableRiverSources = append(availableRiverSources, Coord{m.w / 2, m.h/2 + 1})
	}
	availableRiverSources = append(availableRiverSources, m.freeBorders...)

	initialRiver := true
	var generatedRivers []*river

	for nRiverCells > 0 && len(availableRiverSources) > 0 {
		var riverStart Coord
		riverStart, availableRiverSources = availableRiverSources[0], availableRiverSources[1:]
		m.freeBorders = removeFirst(m.freeBorders, riverStart)

		direction := m.getDirectionFromRiverStart(riverStart)

		if m.hasWaterNearby(m.grid, riverStart, nil) {
			continue
		}

		riversToExpand := []*river{}
		riversToExpand = m.createWater(riversToExpand, newRiver(riverStart, nil, direction))
		nRiverCells--

		for len(riversToExpand) > 0 {
			var r *river
			r, riversToExpand = riversToExpand[0], riversToExpand[1:]
			current := r.current

			if m.isEdge(current.X, current.Y) && !r.isStart() {
				generatedRivers = append(generatedRivers, r)
				continue
			}

			neighs := m.getAvailableNeighboursForRiverToFlow(r)
			if len(neighs) == 0 {
				generatedRivers = append(generatedRivers, r)
				continue
			}

			goingToSplit := nRiverCells > 0 && len(neighs) >= 2 && m.random.NextFloat() <= RIVER_SPLIT_PROBA
			goingToSplit = goingToSplit || initialRiver
			initialRiver = false

			weights := m.createRiverFlowWeights(r, neighs)
			nextCoord := m.getRandomCoord(weights)

			nextDirection := r.preferredDirection
			if goingToSplit {
				nextDirection = DirectionFromCoord(nextCoord.Sub(r.current))
			}
			riversToExpand = m.createWater(riversToExpand, newRiver(nextCoord, r.history, nextDirection))
			nRiverCells--

			if !goingToSplit {
				continue
			}

			remainingNeighs := make([]Coord, 0, len(neighs))
			for _, c := range neighs {
				if c != nextCoord {
					remainingNeighs = append(remainingNeighs, c)
				}
			}
			sha1prng.Shuffle(m.random, remainingNeighs)
			if len(remainingNeighs) == 0 {
				continue
			}
			splitCoord := remainingNeighs[0]

			riversToExpand = m.createWater(riversToExpand, newRiver(
				splitCoord, r.history, DirectionFromCoord(splitCoord.Sub(r.current)),
			))
			nRiverCells--
		}
	}

	m.deleteShortRivers(generatedRivers)
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:433-444

private void deleteShortRivers(List<River> generatedRivers) {
    for (River river : generatedRivers) {
        if (river.history.size() < Game.MIN_RIVER_LENGTH) {
            for (Coord coord : river.history) {
                grid.get(coord).setType(Tile.TYPE_GRASS);
                if (!freeCoords.contains(coord)) freeCoords.add(coord);
            }
        }
    }
}
*/

// deleteShortRivers reverts stubs to grass. A river's history includes every
// cell back to the source, so erasing a stub can also erase the trunk cells
// it branched from — upstream behaviour, not an oversight here.
func (m *GridMaker) deleteShortRivers(generatedRivers []*river) {
	for _, r := range generatedRivers {
		if len(r.history) >= MIN_RIVER_LENGTH {
			continue
		}
		for _, coord := range r.history {
			m.grid.Get(coord).SetType(TYPE_GRASS)
			if !contains(m.freeCoords, coord) {
				m.freeCoords = append(m.freeCoords, coord)
			}
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:446-452

private LinkedHashMap<Coord, Float> createRiverFlowWeights(River river, List<Coord> neighs) {
    LinkedHashMap<Coord, Float> weights = new LinkedHashMap<>();
    neighs.forEach(neig -> weights.put(neig, getWeight(neig, river.current, river.preferredDirection)));
    return weights;
}
*/

type coordWeight struct {
	coord  Coord
	weight float32
}

// createRiverFlowWeights keeps neighbour order, standing in for the
// LinkedHashMap the weighted pick walks.
func (m *GridMaker) createRiverFlowWeights(r *river, neighs []Coord) []coordWeight {
	weights := make([]coordWeight, 0, len(neighs))
	for _, neig := range neighs {
		weights = append(weights, coordWeight{neig, m.getWeight(neig, r.current, r.preferredDirection)})
	}
	return weights
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:454-456,458-467,469-475

private boolean isAccessible(Tile t) { return !t.isWater() && t.getType() != Tile.TYPE_MOUNTAIN; }

private List<Coord> getAvailableNeighboursForRiverToFlow(River river) {
    return grid.getNeighbours(river.current).stream().filter(n ->
        !hasWaterNearby(grid, n, river.history) && !grid.get(n).isWater()
            && (!river.isStart() || !isEdge(n.x, n.y)) && isAccessible(grid.get(n))).toList();
}

private Direction getDirectionFromRiverStart(Coord riverStart) {
    return riverStart.x == 0 ? Direction.EAST
        : (riverStart.y == 0 ? Direction.SOUTH
            : (riverStart.x == w - 1 ? Direction.WEST
                : (riverStart.y == h - 1 ? Direction.NORTH : Direction.UNSET)));
}
*/

func (m *GridMaker) isAccessible(t *Tile) bool {
	return !t.IsWater() && t.GetType() != TYPE_MOUNTAIN
}

func (m *GridMaker) getAvailableNeighboursForRiverToFlow(r *river) []Coord {
	all := m.grid.Neighbours(r.current)
	out := make([]Coord, 0, len(all))
	for _, n := range all {
		if !m.hasWaterNearby(m.grid, n, r.history) &&
			!m.grid.Get(n).IsWater() &&
			(!r.isStart() || !m.isEdge(n.X, n.Y)) &&
			m.isAccessible(m.grid.Get(n)) {
			out = append(out, n)
		}
	}
	return out
}

// getDirectionFromRiverStart is UNSET for the centre source, which is not on
// any edge — the initial river then has no preferred direction.
func (m *GridMaker) getDirectionFromRiverStart(riverStart Coord) Direction {
	switch {
	case riverStart.X == 0:
		return EAST
	case riverStart.Y == 0:
		return SOUTH
	case riverStart.X == m.w-1:
		return WEST
	case riverStart.Y == m.h-1:
		return NORTH
	}
	return UNSET
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:477-503

private void makeMountains() {
    int nMountains = Math.max(Game.MIN_MOUNTAINS, Math.round(random.nextFloat(w * h * Game.MOUNTAIN_TO_CELL_RATIO)));
    if (w * h < 10) nMountains = 0;
    for (int i = 0; i < nMountains; ++i) {
        int mountainSize = random.nextInt(2, 8);
        Coord mountainBase = freeCoords.poll();
        if (mountainBase == null) return;
        List<Coord> mountainCoords = new ArrayList<>(mountainSize);
        grid.get(mountainBase).setType(Tile.TYPE_MOUNTAIN);
        mountainCoords.add(mountainBase);
        freeBorders.remove(mountainBase);
        for (int j = 0; j < mountainSize; j++) {
            List<Coord> neighs = getMountainAvailableNeighbours(grid, mountainCoords);
            if (neighs.isEmpty()) break;
            Coord newMountain = neighs.get(random.nextInt(neighs.size()));
            grid.get(newMountain).setType(Tile.TYPE_MOUNTAIN);
            mountainCoords.add(newMountain);
            freeBorders.remove(newMountain);
            freeCoords.remove(newMountain);
        }
    }
}
*/

// makeMountains grows each range from a free cell by mountainSize accretions,
// each picked out of the HashSet-ordered candidate list.
func (m *GridMaker) makeMountains() {
	nMountains := max(MIN_MOUNTAINS, javaRoundFloat(
		m.random.NextFloatBound(float32(m.w*m.h)*MOUNTAIN_TO_CELL_RATIO),
	))
	if m.w*m.h < 10 {
		nMountains = 0
	}

	for i := 0; i < nMountains; i++ {
		mountainSize := m.random.NextIntRange(2, 8)
		if len(m.freeCoords) == 0 {
			return
		}
		var mountainBase Coord
		mountainBase, m.freeCoords = m.freeCoords[0], m.freeCoords[1:]

		mountainCoords := make([]Coord, 0, mountainSize)
		m.grid.Get(mountainBase).SetType(TYPE_MOUNTAIN)
		mountainCoords = append(mountainCoords, mountainBase)
		m.freeBorders = removeFirst(m.freeBorders, mountainBase)

		for j := 0; j < mountainSize; j++ {
			neighs := m.getMountainAvailableNeighbours(m.grid, mountainCoords)
			if len(neighs) == 0 {
				break
			}
			newMountain := neighs[m.random.NextInt(len(neighs))]
			m.grid.Get(newMountain).SetType(TYPE_MOUNTAIN)
			mountainCoords = append(mountainCoords, newMountain)
			m.freeBorders = removeFirst(m.freeBorders, newMountain)
			m.freeCoords = removeFirst(m.freeCoords, newMountain)
		}
	}
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:505-509,511-521

private void createWater(LinkedList<River> riversToExpand, River riverToAdd) {
    riversToExpand.add(riverToAdd);
    grid.get(riverToAdd.current).setType(Tile.TYPE_WATER);
    freeCoords.remove(riverToAdd.current);
}

private float getWeight(Coord neig, Coord current, Direction preferredDirection) {
    Coord direction = neig.sub(current);
    if (direction.equals(preferredDirection.coord)) return 1.75f;
    else if (direction.equals(preferredDirection.opposite().coord)) return 0.25f;
    return 1.0f;
}
*/

func (m *GridMaker) createWater(riversToExpand []*river, riverToAdd *river) []*river {
	riversToExpand = append(riversToExpand, riverToAdd)
	m.grid.Get(riverToAdd.current).SetType(TYPE_WATER)
	m.freeCoords = removeFirst(m.freeCoords, riverToAdd.current)
	return riversToExpand
}

// getWeight favours flowing on and disfavours doubling back. With an UNSET
// preferred direction both deltas are (0,0) and no neighbour matches, so
// every candidate weighs 1.
func (m *GridMaker) getWeight(neig, current Coord, preferredDirection Direction) float32 {
	direction := neig.Sub(current)
	switch direction {
	case preferredDirection.Coord():
		return 1.75
	case preferredDirection.Opposite().Coord():
		return 0.25
	}
	return 1.0
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/GridMaker.java:523-541

private Coord getRandomCoord(LinkedHashMap<Coord, Float> weights) {
    float totalWeight = 0f;
    for (float weight : weights.values()) totalWeight += weight;
    float rand = (float) (random.nextFloat() * totalWeight);
    float cumulative = 0f;
    for (Map.Entry<Coord, Float> entry : weights.entrySet()) {
        cumulative += entry.getValue();
        if (rand < cumulative) { Coord key = entry.getKey(); weights.remove(key); return key; }
    }
    return weights.keySet().iterator().next();
}
*/

// getRandomCoord is a weighted pick over the insertion-ordered map. The sums
// stay float32: accumulating in float64 shifts the cut points and picks a
// different neighbour on some draws.
func (m *GridMaker) getRandomCoord(weights []coordWeight) Coord {
	var totalWeight float32
	for _, w := range weights {
		totalWeight += w.weight
	}

	rand := m.random.NextFloat() * totalWeight
	var cumulative float32
	for _, w := range weights {
		cumulative += w.weight
		if rand < cumulative {
			return w.coord
		}
	}

	return weights[0].coord
}

// contains is List.contains. Town and Zone override neither equals nor
// hashCode in Java, so on those the Go pointer comparison is exactly the
// reference identity the source compares by.
func contains[T comparable](s []T, v T) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

// removeFirst is LinkedList.remove(Object): drop the first occurrence only,
// keeping the order of everything else.
func removeFirst[T comparable](s []T, v T) []T {
	for i, e := range s {
		if e == v {
			return append(s[:i:i], s[i+1:]...)
		}
	}
	return s
}

// removeAll is List.removeAll: drop every element present in remove.
func removeAll[T comparable](s, remove []T) []T {
	out := make([]T, 0, len(s))
	for _, e := range s {
		if !contains(remove, e) {
			out = append(out, e)
		}
	}
	return out
}
