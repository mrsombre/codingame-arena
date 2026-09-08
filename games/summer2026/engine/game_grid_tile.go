// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Tile.java
package engine

import "fmt"

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Tile.java:9-32

public static final Tile NO_TILE = new Tile(new Coord(-1, -1), -1);

public static final int TYPE_GRASS = 0;
public static final int TYPE_WATER = 1;
public static final int TYPE_MOUNTAIN = 2;
public static final int TYPE_POI = 3;

public static final int TRACK_NONE = -1;
public static final int TRACK_NEUTRAL = 2;
public static final int TOWN_NONE = -1;

private int type;
public int zoneId = -1;
public int track = TRACK_NONE;
public int townId = TOWN_NONE;
public Coord coord;
public List<ScheduleStep> activeConnections;
*/

const (
	TYPE_GRASS    = 0
	TYPE_WATER    = 1
	TYPE_MOUNTAIN = 2
	TYPE_POI      = 3

	TRACK_NONE    = -1
	TRACK_NEUTRAL = 2
	TOWN_NONE     = -1
)

// Tile is one grid cell. Java's shared NO_TILE sentinel has no Go
// counterpart: Grid.Get returns nil out of bounds and every predicate below
// is nil-tolerant so an out-of-bounds read reports what NO_TILE reported.
// The mutators are deliberately not nil-tolerant — an out-of-bounds write is
// a bug and should panic rather than vanish into a shared sentinel.
type Tile struct {
	Type              int
	ZoneID            int
	Track             int
	TownID            int
	Coord             Coord
	ActiveConnections []ScheduleStep
}

func NewTile(coord Coord) *Tile {
	return &Tile{
		Type:   TYPE_GRASS,
		ZoneID: -1,
		Track:  TRACK_NONE,
		TownID: TOWN_NONE,
		Coord:  coord,
	}
}

func (t *Tile) SetType(typ int) { t.Type = typ }

func (t *Tile) GetType() int {
	if t == nil {
		return -1
	}
	return t.Type
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Tile.java:47-49,55-65

public void clear() { type = TYPE_GRASS; }
public String toString() { return this.coord.toString() + " " + this.getType(); }
public int getZoneId() { return zoneId; }
public void setZoneId(int zoneId) { this.zoneId = zoneId; }
*/

func (t *Tile) Clear() { t.Type = TYPE_GRASS }

func (t *Tile) String() string { return fmt.Sprintf("%s %d", t.Coord, t.GetType()) }

func (t *Tile) GetZoneID() int {
	if t == nil {
		return -1
	}
	return t.ZoneID
}

func (t *Tile) SetZoneID(zoneID int) { t.ZoneID = zoneID }

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/grid/Tile.java:51-53,67-97

public boolean isValid() { return this != NO_TILE; }
public boolean isWater() { return type == TYPE_WATER; }
public boolean isMountain() { return type == TYPE_MOUNTAIN; }
public boolean canUseTrackOrTown(int playerIdx) { return isTown() || canUseTracks(playerIdx); }
public boolean canUseTracks(int playerIdx) { return track == playerIdx || track == TRACK_NEUTRAL; }
public boolean isTown() { return townId != TOWN_NONE; }
public boolean isTrackOrTown() { return isTown() || isTrack(); }
public boolean isTrack() { return track != TRACK_NONE; }
public boolean isPlains() { return type == TYPE_GRASS; }
*/

func (t *Tile) IsValid() bool { return t != nil }

func (t *Tile) IsWater() bool { return t != nil && t.Type == TYPE_WATER }

func (t *Tile) IsMountain() bool { return t != nil && t.Type == TYPE_MOUNTAIN }

func (t *Tile) IsPlains() bool { return t != nil && t.Type == TYPE_GRASS }

func (t *Tile) IsTown() bool { return t != nil && t.TownID != TOWN_NONE }

func (t *Tile) IsTrack() bool { return t != nil && t.Track != TRACK_NONE }

func (t *Tile) IsTrackOrTown() bool { return t.IsTown() || t.IsTrack() }

func (t *Tile) CanUseTracks(playerIdx int) bool {
	return t != nil && (t.Track == playerIdx || t.Track == TRACK_NEUTRAL)
}

func (t *Tile) CanUseTrackOrTown(playerIdx int) bool {
	return t.IsTown() || t.CanUseTracks(playerIdx)
}
