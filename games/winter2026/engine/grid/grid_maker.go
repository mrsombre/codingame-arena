// Package grid
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/grid/GridMaker.java
package grid

import (
	"fmt"
	"math"
	"sort"
)

const (
	MinGridHeight = 10
	MaxGridHeight = 24
	AspectRatio   = float32(1.8)
	SpawnHeight   = 3
	DesiredSpawns = 4
)

type Random interface {
	NextDouble() float64
	NextInt(bound int) int
	NextIntRange(origin, bound int) int
}

type GridMaker struct {
	random      Random
	grid        *Grid
	leagueLevel int
}

func NewGridMaker(random Random, leagueLevel int) *GridMaker {
	return &GridMaker{random: random, leagueLevel: leagueLevel}
}

func (m *GridMaker) Make() *Grid {
	skew := 0.3
	switch m.leagueLevel {
	case 1:
		skew = 2
	case 2:
		skew = 1
	case 3:
		skew = 0.8
	}

	rand := m.random.NextDouble()
	height := MinGridHeight + javaRound(math.Pow(rand, skew)*float64(MaxGridHeight-MinGridHeight))
	width := javaRoundFloat32(float32(height) * AspectRatio)
	if width%2 != 0 {
		width++
	}
	m.grid = NewGrid(width, height)

	b := 5 + m.random.NextDouble()*10
	for x := 0; x < width; x++ {
		m.grid.GetXY(x, height-1).SetType(TileWall)
	}

	for y := height - 2; y >= 0; y-- {
		yNorm := float64(height-1-y) / float64(height-1)
		blockChanceEl := 1 / (yNorm + 0.1) / b
		for x := 0; x < width; x++ {
			if m.random.NextDouble() < blockChanceEl {
				m.grid.GetXY(x, y).SetType(TileWall)
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: y}
			opp := m.grid.Opposite(c)
			m.grid.Get(opp).SetType(m.grid.Get(c).Type)
		}
	}

	for _, island := range m.grid.DetectAirPockets() {
		if len(island) < 10 {
			for c := range island {
				m.grid.Get(c).SetType(TileWall)
			}
		}
	}

	somethingDestroyed := true
	for somethingDestroyed {
		somethingDestroyed = false
		for _, c := range m.grid.Coords() {
			if m.grid.Get(c).Type == TileWall {
				continue
			}
			neighbourWalls := make([]Coord, 0)
			for _, n := range m.grid.Neighbours4(c) {
				if m.grid.Get(n).Type == TileWall {
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
				shuffleCoords(destroyable, m.random)
				m.grid.Get(destroyable[0]).SetType(TileEmpty)
				m.grid.Get(m.grid.Opposite(destroyable[0])).SetType(TileEmpty)
				somethingDestroyed = true
			}
		}
	}

	island := m.grid.DetectLowestIsland()
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
		lowerBy = m.random.NextIntRange(2, lowerBy+1)
	}
	for _, c := range island {
		m.grid.Get(c).SetType(TileEmpty)
		m.grid.Get(m.grid.Opposite(c)).SetType(TileEmpty)
	}
	for _, c := range island {
		lowered := Coord{X: c.X, Y: c.Y + lowerBy}
		if m.grid.Get(lowered).IsValid() {
			m.grid.Get(lowered).SetType(TileWall)
			m.grid.Get(m.grid.Opposite(lowered)).SetType(TileWall)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			c := Coord{X: x, Y: y}
			if m.grid.Get(c).Type == TileEmpty && m.random.NextDouble() < 0.025 {
				m.grid.Apples = append(m.grid.Apples, c, m.grid.Opposite(c))
			}
		}
	}

	if len(m.grid.Apples) < 8 {
		m.grid.Apples = m.grid.Apples[:0]
		freeTiles := make([]Coord, 0)
		for _, c := range m.grid.Coords() {
			if m.grid.Get(c).Type == TileEmpty {
				freeTiles = append(freeTiles, c)
			}
		}
		shuffleCoords(freeTiles, m.random)
		minAppleCoords := maxInt(4, int(0.025*float64(len(freeTiles))))
		for len(m.grid.Apples) < minAppleCoords*2 && len(freeTiles) > 0 {
			c := freeTiles[0]
			freeTiles = freeTiles[1:]
			m.grid.Apples = append(m.grid.Apples, c, m.grid.Opposite(c))
			freeTiles = removeCoord(freeTiles, m.grid.Opposite(c))
		}
	}

	for _, c := range m.grid.Coords() {
		if m.grid.Get(c).Type == TileEmpty {
			continue
		}
		neighbourWallCount := 0
		for _, n := range m.grid.Neighbours(c, Adjacency8[:]) {
			if m.grid.Get(n).Type == TileWall {
				neighbourWallCount++
			}
		}
		if neighbourWallCount == 0 {
			m.grid.Get(c).SetType(TileEmpty)
			m.grid.Get(m.grid.Opposite(c)).SetType(TileEmpty)
			m.grid.Apples = append(m.grid.Apples, c, m.grid.Opposite(c))
		}
	}

	potentialSpawns := make([]Coord, 0)
	for _, c := range m.grid.Coords() {
		if m.grid.Get(c).Type != TileWall {
			continue
		}
		aboves := m.freeAbove(c, SpawnHeight)
		if len(aboves) >= SpawnHeight {
			potentialSpawns = append(potentialSpawns, c)
		}
	}
	shuffleCoords(potentialSpawns, m.random)

	desiredSpawns := DesiredSpawns
	if height <= 15 {
		desiredSpawns--
	}
	if height <= 10 {
		desiredSpawns--
	}
	for desiredSpawns > 0 && len(potentialSpawns) > 0 {
		spawn := potentialSpawns[0]
		potentialSpawns = potentialSpawns[1:]
		spawnLoc := m.freeAbove(spawn, SpawnHeight)

		tooClose := false
		for _, c := range spawnLoc {
			if c.X == width/2-1 || c.X == width/2 {
				tooClose = true
				break
			}
			for _, n := range m.grid.Neighbours(c, Adjacency8[:]) {
				if coordSliceContains(m.grid.Spawns, n) || coordSliceContains(m.grid.Spawns, m.grid.Opposite(n)) {
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
			m.grid.Spawns = append(m.grid.Spawns, c)
			m.grid.Apples = removeCoord(m.grid.Apples, c)
			m.grid.Apples = removeCoord(m.grid.Apples, m.grid.Opposite(c))
		}
		desiredSpawns--
	}

	if err := m.checkGrid(); err != nil {
		panic(err)
	}
	return m.grid
}

func (m *GridMaker) checkGrid() error {
	seen := make(map[Coord]struct{}, len(m.grid.Apples))
	for _, c := range m.grid.Apples {
		if m.grid.Get(c).Type != TileEmpty {
			return fmt.Errorf("apple on wall at %s", c)
		}
		if _, ok := seen[c]; ok {
			return fmt.Errorf("duplicate apples")
		}
		seen[c] = struct{}{}
	}
	return nil
}

func (m *GridMaker) freeAbove(c Coord, by int) []Coord {
	result := make([]Coord, 0, by)
	for i := 1; i <= by; i++ {
		above := Coord{X: c.X, Y: c.Y - i}
		if m.grid.Get(above).IsValid() && m.grid.Get(above).Type == TileEmpty {
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
