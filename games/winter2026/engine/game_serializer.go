// Package engine
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Serializer.java
package engine

import (
	"strconv"
	"strings"
)

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Serializer.java:100-123

public static List<String> SerializeGlobalInfoFor(Player player, Game game) {
    lines.add(player.getIndex());
    lines.add(game.grid.width);
    lines.add(game.grid.height);
    for (int y = 0; y < game.grid.height; ++y) {
        // row of '#' for wall, '.' for empty
    }
    lines.add(game.players.get(0).birds.size());
    for (Bird b : player.birds) lines.add(b.id);
    for (Bird b : game.players.get(1 - player.getIndex()).birds) lines.add(b.id);
    return lines.stream().map(String::valueOf).collect(Collectors.toList());
}
*/

func SerializeGlobalInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0)
	lines = append(lines, strconv.Itoa(player.GetIndex()))
	lines = append(lines, strconv.Itoa(game.Grid.Width))
	lines = append(lines, strconv.Itoa(game.Grid.Height))
	for y := 0; y < game.Grid.Height; y++ {
		var row strings.Builder
		for x := 0; x < game.Grid.Width; x++ {
			if game.Grid.GetXY(x, y).IsWall() {
				row.WriteByte('#')
			} else {
				row.WriteByte('.')
			}
		}
		lines = append(lines, row.String())
	}
	lines = append(lines, strconv.Itoa(len(game.Players[0].Birds)))
	for _, b := range player.Birds {
		lines = append(lines, strconv.Itoa(b.ID))
	}
	for _, b := range game.Players[1-player.GetIndex()].Birds {
		lines = append(lines, strconv.Itoa(b.ID))
	}
	return lines
}

/*
Java: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Serializer.java:126-140

public static List<String> SerializeFrameInfoFor(Player player, Game game) {
    lines.add(game.grid.apples.size());
    for (Coord c : game.grid.apples) lines.add(c.toIntString());
    List<Bird> liveBirds = game.getLiveBirds();
    lines.add(liveBirds.size());
    for (Bird b : liveBirds) {
        String body = b.body.stream().map(c -> c.getX() + "," + c.getY()).collect(Collectors.joining(":"));
        lines.add(join(b.id, body));
    }
    return lines.stream().map(String::valueOf).collect(Collectors.toList());
}
*/

func SerializeFrameInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0)
	lines = append(lines, strconv.Itoa(len(game.Grid.Apples)))
	for _, c := range game.Grid.Apples {
		lines = append(lines, c.ToIntString())
	}
	liveBirds := game.LiveBirds()
	lines = append(lines, strconv.Itoa(len(liveBirds)))
	for _, b := range liveBirds {
		bodyParts := make([]string, 0, len(b.Body))
		for _, c := range b.Body {
			bodyParts = append(bodyParts, strconv.Itoa(c.X)+","+strconv.Itoa(c.Y))
		}
		lines = append(lines, strings.Join([]string{strconv.Itoa(b.ID), strings.Join(bodyParts, ":")}, " "))
	}
	return lines
}
