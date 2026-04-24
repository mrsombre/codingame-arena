// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java
package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/games/spring2020/engine/grid"
)

// serializeGlobalInfoFor builds the initial input lines for a player.
// Format: "width height" followed by height rows of '#' or ' '.
func serializeGlobalInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0, game.Grid.Height+1)
	lines = append(lines, strconv.Itoa(game.Grid.Width)+" "+strconv.Itoa(game.Grid.Height))
	for y := 0; y < game.Grid.Height; y++ {
		var row strings.Builder
		for x := 0; x < game.Grid.Width; x++ {
			if game.Grid.GetXY(x, y).IsWall() {
				row.WriteByte('#')
			} else {
				row.WriteByte(' ')
			}
		}
		lines = append(lines, row.String())
	}
	return lines
}

// serializeFrameInfoFor builds the per-turn input lines for a player.
func serializeFrameInfoFor(player *Player, game *Game) []string {
	opponent := opponentOf(player, game.Players)
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("%d %d", player.Pellets, opponent.Pellets))

	visible := visiblePacmen(player, game)
	if game.Config.DeadPacs {
		visible = append(visible, player.DeadPacmen()...)
		visible = append(visible, opponent.DeadPacmen()...)
	}

	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].ID < visible[j].ID
	})

	lines = append(lines, strconv.Itoa(len(visible)))
	for _, pac := range visible {
		lines = append(lines, pacmanLine(player, pac))
	}

	pellets := visiblePellets(player, game)
	cherries := game.Grid.AllCherries()
	lines = append(lines, strconv.Itoa(len(pellets)+len(cherries)))
	for _, p := range pellets {
		lines = append(lines, fmt.Sprintf("%d %d %d", p.X, p.Y, 1))
	}
	for _, c := range cherries {
		lines = append(lines, fmt.Sprintf("%d %d %d", c.X, c.Y, CherryScore))
	}
	return lines
}

func pacmanLine(player *Player, pac *Pacman) string {
	owned := 0
	if pac.Owner == player {
		owned = 1
	}
	typeName := pac.Type.Name()
	if pac.Dead {
		typeName = "DEAD"
	}
	return fmt.Sprintf("%d %d %d %d %s %d %d",
		pac.Number,
		owned,
		pac.Position.X, pac.Position.Y,
		typeName,
		pac.AbilityDuration,
		pac.AbilityCooldown,
	)
}

func opponentOf(player *Player, players []*Player) *Player {
	for _, p := range players {
		if p != player {
			return p
		}
	}
	return player
}

func visiblePacmen(player *Player, game *Game) []*Pacman {
	if !game.Config.FogOfWar {
		return append([]*Pacman(nil), game.Pacmen...)
	}
	visibleCoords := findVisibleCoords(player, game, func(c grid.Coord) bool {
		for _, pac := range game.Pacmen {
			if pac.Position == c {
				return true
			}
		}
		return false
	})
	coordSet := make(map[grid.Coord]struct{}, len(visibleCoords))
	for _, c := range visibleCoords {
		coordSet[c] = struct{}{}
	}
	out := make([]*Pacman, 0)
	for _, pac := range game.Pacmen {
		if _, ok := coordSet[pac.Position]; ok {
			out = append(out, pac)
		}
	}
	return out
}

func visiblePellets(player *Player, game *Game) []grid.Coord {
	if !game.Config.FogOfWar {
		return game.Grid.AllPellets()
	}
	return findVisibleCoords(player, game, func(c grid.Coord) bool {
		return game.Grid.Get(c).HasPellet
	})
}

func findVisibleCoords(player *Player, game *Game, hasItem func(grid.Coord) bool) []grid.Coord {
	visible := make([]grid.Coord, 0)
	seen := make(map[grid.Coord]struct{})
	for _, pac := range player.AlivePacmen() {
		for _, delta := range grid.Adjacency4 {
			current := pac.Position
			for game.Grid.Get(current).IsFloor() {
				if hasItem(current) {
					if _, ok := seen[current]; !ok {
						seen[current] = struct{}{}
						visible = append(visible, current)
					}
				}
				next, ok := game.Grid.GetCoordNeighbour(current, delta)
				if !ok {
					break
				}
				current = next
				if pac.Position == current {
					break
				}
			}
		}
	}
	return visible
}
