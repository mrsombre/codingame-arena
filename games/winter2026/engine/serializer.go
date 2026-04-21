// Package winter2026
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/Serializer.java
package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
)

func serializeGlobalInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0)
	lines = append(lines, strconv.Itoa(player.GetIndex()))
	lines = append(lines, strconv.Itoa(game.grid.Width))
	lines = append(lines, strconv.Itoa(game.grid.Height))
	for y := 0; y < game.grid.Height; y++ {
		var row strings.Builder
		for x := 0; x < game.grid.Width; x++ {
			if game.grid.GetXY(x, y).Type == grid.TileWall {
				row.WriteByte('#')
			} else {
				row.WriteByte('.')
			}
		}
		lines = append(lines, row.String())
	}
	lines = append(lines, strconv.Itoa(len(game.players[0].birds)))
	for _, b := range player.birds {
		lines = append(lines, strconv.Itoa(b.ID))
	}
	for _, b := range game.players[1-player.GetIndex()].birds {
		lines = append(lines, strconv.Itoa(b.ID))
	}
	return lines
}

func serializeFrameInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0)
	lines = append(lines, strconv.Itoa(len(game.grid.Apples)))
	for _, c := range game.grid.Apples {
		lines = append(lines, c.ToIntString())
	}
	liveBirds := game.liveBirds()
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
