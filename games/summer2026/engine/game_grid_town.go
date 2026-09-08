// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Town.java
package engine

import "sort"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Town.java:8-23

public class Town {
    public int id;
    public Coord coord;
    public List<Town> desiredConnections;
    public List<Town> activeConnections;
    public Map<Integer, List<Coord>> paths;

    public Town(int id, Coord coord) {
        this.id = id;
        this.coord = coord;
        this.activeConnections = new ArrayList<>();
        this.paths = new TreeMap<>();
    }
*/

// Town is a scoring endpoint. DesiredConnections is unilateral: a town in
// this list does not necessarily list this town back.
type Town struct {
	ID                 int
	Coord              Coord
	DesiredConnections []*Town
	ActiveConnections  []*Town
	// Paths maps a connected town id to the current shortest path. Java uses
	// a TreeMap, so callers that iterate must go through SortedPathTownIDs.
	Paths map[int][]Coord
}

func NewTown(id int, coord Coord) *Town {
	return &Town{
		ID:                id,
		Coord:             coord,
		ActiveConnections: []*Town{},
		Paths:             map[int][]Coord{},
	}
}

// SortedPathTownIDs returns the Paths keys in ascending order, reproducing
// TreeMap<Integer, ...> iteration.
func (t *Town) SortedPathTownIDs() []int {
	ids := make([]int, 0, len(t.Paths))
	for id := range t.Paths {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
