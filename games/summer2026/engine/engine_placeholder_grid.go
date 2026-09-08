// Package engine
package engine

// Placeholder map generation. GridMaker — the real generator, and the only
// one that reproduces a CodinGame seed — is not ported yet, so a match runs
// on the fixed grid below instead. It exists to keep the arena end to end
// runnable; nothing about it is faithful to upstream, and it disappears the
// moment GridMaker lands.

const (
	placeholderWidth  = 21
	placeholderHeight = 14
)

// MakePlaceholderGrid returns an all-plains grid with two towns that want to
// be connected to each other. Each town sits in its own single-cell region so
// the region layout is at least shaped like a generated one; everything else
// is one large region.
func MakePlaceholderGrid() *Grid {
	grid := NewGrid(placeholderWidth, placeholderHeight)

	townCoords := []Coord{{4, 7}, {16, 7}}
	for id, coord := range townCoords {
		town := NewTown(id, coord)
		grid.Towns = append(grid.Towns, town)
		grid.Get(coord).TownID = id
	}
	grid.Towns[0].DesiredConnections = []*Town{grid.Towns[1]}
	grid.Towns[1].DesiredConnections = []*Town{grid.Towns[0]}

	// Region 0 and 1 are the two town cells; region 2 is the rest.
	rest := make([]Coord, 0, len(grid.Cells))
	for _, tile := range grid.Cells {
		if tile.IsTown() {
			tile.ZoneID = tile.TownID
			continue
		}
		tile.ZoneID = len(townCoords)
		rest = append(rest, tile.Coord)
	}
	for id, coord := range townCoords {
		zone := NewZone(id, []Coord{coord})
		zone.AddTown(grid.Towns[id])
		grid.Zones = append(grid.Zones, zone)
	}
	grid.Zones = append(grid.Zones, NewZone(len(townCoords), rest))

	return grid
}
