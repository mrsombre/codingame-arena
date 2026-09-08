// Package engine
// Source: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Serializer.java
package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Only the two *InfoFor functions are ported. serializeGlobalData and
// serializeFrameData feed the CodinGame viewer's frame protocol and have no
// bot-visible effect, so they are out of the simulation port.

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Serializer.java:107-139

public static List<String> serializeGlobalInfoFor(Player player, Game game) {
    lines.add(player.getIndex());
    lines.add(game.grid.width);
    lines.add(game.grid.height);
    for (int y = 0; y < game.grid.height; ++y)
        for (int x = 0; x < game.grid.width; ++x)
            lines.add(join(tile.getZoneId(), tile.getType()));
    lines.add(game.grid.towns.size());
    for (Town t : game.grid.towns)
        lines.add(join(t.id, t.coord.getX(), t.coord.getY(), serializeTowns(t.desiredConnections)));
    return ...;
}
*/

// SerializeGlobalInfoFor produces the one-off initialization block: player
// id, grid dimensions, one `regionId type` line per cell in row-major order,
// then the town roster.
func SerializeGlobalInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0, 4+game.Grid.Width*game.Grid.Height+len(game.Grid.Towns))
	lines = append(lines,
		strconv.Itoa(player.GetIndex()),
		strconv.Itoa(game.Grid.Width),
		strconv.Itoa(game.Grid.Height),
	)

	for y := 0; y < game.Grid.Height; y++ {
		for x := 0; x < game.Grid.Width; x++ {
			tile := game.Grid.GetXY(x, y)
			lines = append(lines, fmt.Sprintf("%d %d", tile.GetZoneID(), tile.GetType()))
		}
	}

	lines = append(lines, strconv.Itoa(len(game.Grid.Towns)))
	for _, t := range game.Grid.Towns {
		lines = append(lines, fmt.Sprintf(
			"%d %d %d %s",
			t.ID, t.Coord.X, t.Coord.Y, serializeTowns(t.DesiredConnections),
		))
	}

	return lines
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Serializer.java:75-80

private static String serializeTowns(List<Town> towns) {
    return towns.isEmpty() ? "x"
        : towns.stream().map((Town tc) -> Integer.toString(tc.id)).collect(Collectors.joining(","));
}
*/

func serializeTowns(towns []*Town) string {
	if len(towns) == 0 {
		return "x"
	}
	ids := make([]string, len(towns))
	for i, t := range towns {
		ids[i] = strconv.Itoa(t.ID)
	}
	return strings.Join(ids, ",")
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Serializer.java:141-170

public static List<String> serializeFrameInfoFor(Player player, Game game) {
    Stream.of(player, game.players.get(1 - player.getIndex()))
        .forEach(p -> lines.add(join(p.getScore())));
    for (int y = 0; y < game.grid.height; ++y)
        for (int x = 0; x < game.grid.width; ++x)
            lines.add(join(tile.track,
                           game.grid.zones.get(tile.zoneId).instability,
                           game.grid.zones.get(tile.zoneId).inked ? 1 : 0,
                           serializeLiveConnections(tile.activeConnections)));
    return ...;
}
*/

// SerializeFrameInfoFor produces the per-turn block. The two score lines are
// always "mine then theirs", so player 1 sees the pair swapped relative to
// player 0.
func SerializeFrameInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0, 2+game.Grid.Width*game.Grid.Height)
	for _, p := range []*Player{player, game.Players[1-player.GetIndex()]} {
		lines = append(lines, strconv.Itoa(p.GetScore()))
	}

	for y := 0; y < game.Grid.Height; y++ {
		for x := 0; x < game.Grid.Width; x++ {
			tile := game.Grid.GetXY(x, y)
			zone := game.Grid.Zones[tile.ZoneID]
			inked := 0
			if zone.Inked {
				inked = 1
			}
			lines = append(lines, fmt.Sprintf(
				"%d %d %d %s",
				tile.Track, zone.Instability, inked, serializeLiveConnections(tile.ActiveConnections),
			))
		}
	}

	return lines
}

/*
Java: SummerChallenge2026-BackTrackKing/src/main/java/com/codingame/game/Serializer.java:172-179

private static String serializeLiveConnections(List<ScheduleStep> connection) {
    return connection.isEmpty() ? "x"
        : connection.stream().map(ss -> ss.fromTownId() + "-" + ss.toTownId()).sorted().collect(Collectors.joining(","));
}
*/

// serializeLiveConnections sorts the pairs as strings, not numerically —
// "10-2" therefore precedes "2-3", matching Java's natural String ordering.
func serializeLiveConnections(connection []ScheduleStep) string {
	if len(connection) == 0 {
		return "x"
	}
	pairs := make([]string, len(connection))
	for i, ss := range connection {
		pairs[i] = fmt.Sprintf("%d-%d", ss.FromTownID, ss.ToTownID)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}
