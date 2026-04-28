// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java
package engine

import (
	"fmt"
	"math"
	"sort"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java:29-41

public static final int MIN_GRID_HEIGHT = 10;
public static final int MAX_GRID_HEIGHT = 24;
public static final float ASPECT_RATIO = 1.8f;

public GridMaker(Random random, int leagueLevel) {
    this.random = random;
    this.leagueLevel = leagueLevel;
}

public static int SPAWN_HEIGHT = 3;
public static int DESIRED_SPAWNS = 4;
*/

const (
	MIN_GRID_HEIGHT = 10
	MAX_GRID_HEIGHT = 24
	ASPECT_RATIO    = float32(1.8)
	SPAWN_HEIGHT    = 3
	DESIRED_SPAWNS  = 4
)

type Random interface {
	NextDouble() float64
	NextInt(bound int) int
	NextIntRange(origin, bound int) int
}

type GridMaker struct {
	Random      Random
	Grid        *Grid
	LeagueLevel int
}

func NewGridMaker(random Random, leagueLevel int) *GridMaker {
	return &GridMaker{Random: random, LeagueLevel: leagueLevel}
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java:48-254

public Grid make() {
    double skew;
    if (leagueLevel == 1) skew = 2;       // bronze
    else if (leagueLevel == 2) skew = 1;  // silver
    else if (leagueLevel == 3) skew = 0.8;// gold
    else skew = 0.3;                      // legend

    int height = MIN_GRID_HEIGHT + (int) Math.round(Math.pow(rand, skew) * (MAX_GRID_HEIGHT - MIN_GRID_HEIGHT));
    int width = Math.round(height * ASPECT_RATIO);
    if (width % 2 != 0) width += 1;

    // Fill bottom row, random walls, mirror X, remove air pockets,
    // break 3-wall cells, sink lowest island, spawn apples, place spawns.
    checkGrid();
    return grid;
}
*/

func (m *GridMaker) Make() *Grid {
	skew := 0.3
	switch m.LeagueLevel {
	case 1:
		skew = 2
	case 2:
		skew = 1
	case 3:
		skew = 0.8
	}

	rand := m.Random.NextDouble()
	height := MIN_GRID_HEIGHT + javaRound(math.Pow(rand, skew)*float64(MAX_GRID_HEIGHT-MIN_GRID_HEIGHT))
	width := javaRoundFloat32(float32(height) * ASPECT_RATIO)
	if width%2 != 0 {
		width++
	}
	m.Grid = NewGrid(width, height)

	b := 5 + m.Random.NextDouble()*10
	for x := 0; x < width; x++ {
		m.Grid.GetXY(x, height-1).SetType(TileWall)
	}

	for y := height - 2; y >= 0; y-- {
		yNorm := float64(height-1-y) / float64(height-1)
		blockChanceEl := 1 / (yNorm + 0.1) / b
		for x := 0; x < width; x++ {
			if m.Random.NextDouble() < blockChanceEl {
				m.Grid.GetXY(x, y).SetType(TileWall)
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: y}
			opp := m.Grid.Opposite(c)
			m.Grid.Get(opp).SetType(m.Grid.Get(c).Type)
		}
	}

	for _, island := range m.Grid.DetectAirPockets() {
		if len(island) < 10 {
			for c := range island {
				m.Grid.Get(c).SetType(TileWall)
			}
		}
	}

	somethingDestroyed := true
	for somethingDestroyed {
		somethingDestroyed = false
		for _, c := range m.Grid.Coords() {
			if m.Grid.Get(c).IsWall() {
				continue
			}
			neighbourWalls := make([]Coord, 0)
			for _, n := range m.Grid.Neighbours4(c) {
				if m.Grid.Get(n).IsWall() {
					neighbourWalls = append(neighbourWalls, n)
				}
			}
			if len(neighbourWalls) >= 3 {
				destroyable := make([]Coord, 0, len(neighbourWalls))
				for _, n := range neighbourWalls {
					if n.Y <= c.Y {
						destroyable = append(destroyable, n)
					}
				}
				shuffleCoords(destroyable, m.Random)
				m.Grid.Get(destroyable[0]).SetType(TileEmpty)
				m.Grid.Get(m.Grid.Opposite(destroyable[0])).SetType(TileEmpty)
				somethingDestroyed = true
			}
		}
	}

	island := m.Grid.DetectLowestIsland()
	lowerBy := 0
	canLower := true
	for canLower {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: height - 1 - (lowerBy + 1)}
			if !coordSliceContains(island, c) {
				canLower = false
				break
			}
		}
		if canLower {
			lowerBy++
		}
	}
	if lowerBy >= 2 {
		lowerBy = m.Random.NextIntRange(2, lowerBy+1)
	}
	for _, c := range island {
		m.Grid.Get(c).SetType(TileEmpty)
		m.Grid.Get(m.Grid.Opposite(c)).SetType(TileEmpty)
	}
	for _, c := range island {
		lowered := Coord{X: c.X, Y: c.Y + lowerBy}
		if m.Grid.Get(lowered).IsValid() {
			m.Grid.Get(lowered).SetType(TileWall)
			m.Grid.Get(m.Grid.Opposite(lowered)).SetType(TileWall)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			c := Coord{X: x, Y: y}
			if m.Grid.Get(c).IsEmpty() && m.Random.NextDouble() < 0.025 {
				m.Grid.Apples = append(m.Grid.Apples, c, m.Grid.Opposite(c))
			}
		}
	}

	if len(m.Grid.Apples) < 8 {
		m.Grid.Apples = m.Grid.Apples[:0]
		freeTiles := make([]Coord, 0)
		for _, c := range m.Grid.Coords() {
			if m.Grid.Get(c).IsEmpty() {
				freeTiles = append(freeTiles, c)
			}
		}
		shuffleCoords(freeTiles, m.Random)
		minAppleCoords := maxInt(4, int(0.025*float64(len(freeTiles))))
		for len(m.Grid.Apples) < minAppleCoords*2 && len(freeTiles) > 0 {
			c := freeTiles[0]
			freeTiles = freeTiles[1:]
			m.Grid.Apples = append(m.Grid.Apples, c, m.Grid.Opposite(c))
			freeTiles = removeCoord(freeTiles, m.Grid.Opposite(c))
		}
	}

	for _, c := range m.Grid.Coords() {
		if m.Grid.Get(c).IsEmpty() {
			continue
		}
		neighbourWallCount := 0
		for _, n := range m.Grid.Neighbours(c, Adjacency8[:]) {
			if m.Grid.Get(n).IsWall() {
				neighbourWallCount++
			}
		}
		if neighbourWallCount == 0 {
			m.Grid.Get(c).SetType(TileEmpty)
			m.Grid.Get(m.Grid.Opposite(c)).SetType(TileEmpty)
			m.Grid.Apples = append(m.Grid.Apples, c, m.Grid.Opposite(c))
		}
	}

	potentialSpawns := make([]Coord, 0)
	for _, c := range m.Grid.Coords() {
		if !m.Grid.Get(c).IsWall() {
			continue
		}
		aboves := m.FreeAbove(c, SPAWN_HEIGHT)
		if len(aboves) >= SPAWN_HEIGHT {
			potentialSpawns = append(potentialSpawns, c)
		}
	}
	shuffleCoords(potentialSpawns, m.Random)

	desiredSpawns := DESIRED_SPAWNS
	if height <= 15 {
		desiredSpawns--
	}
	if height <= 10 {
		desiredSpawns--
	}
	for desiredSpawns > 0 && len(potentialSpawns) > 0 {
		spawn := potentialSpawns[0]
		potentialSpawns = potentialSpawns[1:]
		spawnLoc := m.FreeAbove(spawn, SPAWN_HEIGHT)

		tooClose := false
		for _, c := range spawnLoc {
			if c.X == width/2-1 || c.X == width/2 {
				tooClose = true
				break
			}
			for _, n := range m.Grid.Neighbours(c, Adjacency8[:]) {
				if coordSliceContains(m.Grid.Spawns, n) || coordSliceContains(m.Grid.Spawns, m.Grid.Opposite(n)) {
					tooClose = true
					break
				}
			}
			if tooClose {
				break
			}
		}
		if tooClose {
			continue
		}

		for _, c := range spawnLoc {
			m.Grid.Spawns = append(m.Grid.Spawns, c)
			m.Grid.Apples = removeCoord(m.Grid.Apples, c)
			m.Grid.Apples = removeCoord(m.Grid.Apples, m.Grid.Opposite(c))
		}
		desiredSpawns--
	}

	if err := m.CheckGrid(); err != nil {
		panic(err)
	}
	return m.Grid
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java:256-265

private void checkGrid() {
    for (Coord c : grid.apples) {
        if (grid.get(c).getType() != Tile.TYPE_EMPTY) {
            throw new RuntimeException("Apple on wall at " + c);
        }
    }
    if (grid.apples.size() != grid.apples.stream().distinct().count()) {
        throw new RuntimeException("Duplicate apples");
    }
}
*/

func (m *GridMaker) CheckGrid() error {
	seen := make(map[Coord]struct{}, len(m.Grid.Apples))
	for _, c := range m.Grid.Apples {
		if !m.Grid.Get(c).IsEmpty() {
			return fmt.Errorf("apple on wall at %s", c)
		}
		if _, ok := seen[c]; ok {
			return fmt.Errorf("duplicate apples")
		}
		seen[c] = struct{}{}
	}
	return nil
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java:267-278

private List<Coord> getFreeAbove(Coord c, int by) {
    List<Coord> result = new ArrayList<>();
    for (int i = 1; i <= by; i++) {
        Coord above = new Coord(c.x, c.y - i);
        if (grid.get(above).isValid() && grid.get(above).getType() == Tile.TYPE_EMPTY) {
            result.add(above);
        } else {
            break;
        }
    }
    return result;
}
*/

func (m *GridMaker) FreeAbove(c Coord, by int) []Coord {
	result := make([]Coord, 0, by)
	for i := 1; i <= by; i++ {
		above := Coord{X: c.X, Y: c.Y - i}
		if m.Grid.Get(above).IsEmpty() {
			result = append(result, above)
		} else {
			break
		}
	}
	return result
}

func SortedCoordsFromSet(set map[Coord]struct{}) []Coord {
	coords := make([]Coord, 0, len(set))
	for c := range set {
		coords = append(coords, c)
	}
	sort.Slice(coords, func(i, j int) bool {
		return coords[i].Less(coords[j])
	})
	return coords
}

func shuffleCoords(coords []Coord, random Random) {
	for i := len(coords); i > 1; i-- {
		j := random.NextInt(i)
		coords[i-1], coords[j] = coords[j], coords[i-1]
	}
}

func coordSliceContains(coords []Coord, target Coord) bool {
	for _, c := range coords {
		if c == target {
			return true
		}
	}
	return false
}

func removeCoord(coords []Coord, target Coord) []Coord {
	for i, c := range coords {
		if c == target {
			return append(coords[:i], coords[i+1:]...)
		}
	}
	return coords
}

func javaRound(v float64) int {
	return int(math.Floor(v + 0.5))
}

func javaRoundFloat32(v float32) int {
	return int(float32(math.Floor(float64(v + 0.5))))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
