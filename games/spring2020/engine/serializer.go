// Package engine
// Source: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java
//
// This file ports the input-line serialization methods from Game.java
// (getGlobalInfoFor, getCurrentFrameInfoFor, findVisiblePacmen,
// findVisibleItems, findVisiblePellets, getPacmanLineInfo, getPelletLineInfo,
// cellToCharater). The simulation methods from the same Java source live in
// spring2020_game.go.
package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:320-339,399-405

public List<String> getGlobalInfoFor(Player player) {
    List<String> lines = new ArrayList<String>();
    lines.add(String.format("%d %d", grid.width, grid.height));
    for (int y = 0; y < grid.getHeight(); ++y) {
        // join cellToCharater(cell) for x in [0, width)
    }
    return lines;
}

public String cellToCharater(Cell cell) {
    return cell.isWall() ? "#" : " ";
}
*/

// SerializeGlobalInfoFor builds the initial input lines for a player.
// Format: "width height" followed by height rows of '#' or ' '.
func SerializeGlobalInfoFor(player *Player, game *Game) []string {
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:363-397

public List<String> getCurrentFrameInfoFor(Player player) {
    Player opponentPlayer = gameManager.getActivePlayers().get((player.getIndex() + 1) % 2);
    List<String> lines = new ArrayList<String>();
    lines.add(String.format("%d %d", player.pellets, opponentPlayer.pellets));

    List<Pacman> visiblePacmen = Config.FOG_OF_WAR ? findVisiblePacmen(player) : pacmen;
    if (Config.PROVIDE_DEAD_PACS) {
        Stream.concat(player.getDeadPacmen(), opponentPlayer.getDeadPacmen())
            .forEach(visiblePacmen::add);
    }
    lines.add(Integer.toString(visiblePacmen.size()));
    visiblePacmen.stream().sorted(Comparator.comparing(Pacman::getId))
        .map(pac -> getPacmanLineInfo(player, pac)).forEach(lines::add);

    List<Coord> visiblePellets = Config.FOG_OF_WAR ? findVisiblePellets(player) : grid.getAllPellets();
    List<Coord> visibleCherries = grid.getAllCherries();
    lines.add(Integer.toString(visiblePellets.size() + visibleCherries.size()));
    for (Coord pellet : visiblePellets)  lines.add(getPelletLineInfo(pellet, 1));
    for (Coord cherry : visibleCherries) lines.add(getPelletLineInfo(cherry, Config.CHERRY_SCORE));
    return lines;
}
*/

// SerializeFrameInfoFor builds the per-turn input lines for a player.
func SerializeFrameInfoFor(player *Player, game *Game) []string {
	opponent := OpponentOf(player, game.Players)
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("%d %d", player.Pellets, opponent.Pellets))

	visible := VisiblePacmen(player, game)
	if game.Config.PROVIDE_DEAD_PACS {
		visible = append(visible, player.DeadPacmen()...)
		visible = append(visible, opponent.DeadPacmen()...)
	}

	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].ID < visible[j].ID
	})

	lines = append(lines, strconv.Itoa(len(visible)))
	for _, pac := range visible {
		lines = append(lines, PacmanLine(player, pac))
	}

	pellets := VisiblePellets(player, game)
	cherries := game.Grid.AllCherries()
	lines = append(lines, strconv.Itoa(len(pellets)+len(cherries)))
	for _, p := range pellets {
		lines = append(lines, fmt.Sprintf("%d %d %d", p.X, p.Y, 1))
	}
	for _, c := range cherries {
		lines = append(lines, fmt.Sprintf("%d %d %d", c.X, c.Y, CHERRY_SCORE))
	}
	return lines
}

// SerializeTraceFrameInfo builds a per-turn frame-info view that bypasses
// fog-of-war: every pacman and every pellet/cherry is included, formatted
// as if a side-0 bot were the recipient. Used by the trace path so analyzers
// see the full game state every turn rather than the filtered view blue's
// bot saw on stdin. Player visibility is not a game-state mutation — bots
// still receive the fog-filtered SerializeFrameInfoFor lines.
func SerializeTraceFrameInfo(game *Game) []string {
	if len(game.Players) == 0 {
		return nil
	}
	player := game.Players[0]
	opponent := OpponentOf(player, game.Players)
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("%d %d", player.Pellets, opponent.Pellets))

	pacs := append([]*Pacman(nil), game.Pacmen...)
	if game.Config.PROVIDE_DEAD_PACS {
		pacs = append(pacs, player.DeadPacmen()...)
		pacs = append(pacs, opponent.DeadPacmen()...)
	}
	sort.SliceStable(pacs, func(i, j int) bool {
		return pacs[i].ID < pacs[j].ID
	})
	lines = append(lines, strconv.Itoa(len(pacs)))
	for _, pac := range pacs {
		lines = append(lines, PacmanLine(player, pac))
	}

	pellets := game.Grid.AllPellets()
	cherries := game.Grid.AllCherries()
	lines = append(lines, strconv.Itoa(len(pellets)+len(cherries)))
	for _, p := range pellets {
		lines = append(lines, fmt.Sprintf("%d %d %d", p.X, p.Y, 1))
	}
	for _, c := range cherries {
		lines = append(lines, fmt.Sprintf("%d %d %d", c.X, c.Y, CHERRY_SCORE))
	}
	return lines
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:341-352

private String getPacmanLineInfo(Player player, Pacman pac) {
    return String.format(
        "%d %d %d %d %s %d %d",
        pac.getNumber(),
        pac.getOwner() == player ? 1 : 0,
        pac.getPosition().x, pac.getPosition().y,
        pac.isDead() ? "DEAD" : pac.getType().name().toUpperCase(),
        pac.getAbilityDuration(),
        pac.getAbilityCooldown()
    );
}
*/

func PacmanLine(player *Player, pac *Pacman) string {
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

// OpponentOf returns the other Player. Go-only helper.
func OpponentOf(player *Player, players []*Player) *Player {
	for _, p := range players {
		if p != player {
			return p
		}
	}
	return player
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:967-974

private List<Pacman> findVisiblePacmen(Player player) {
    List<Coord> coords = findVisibleItems(player, coord -> pacmen.stream().anyMatch(pac -> pac.getPosition().equals(coord)));
    return pacmen.stream()
        .filter(pac -> coords.contains(pac.getPosition()))
        .collect(Collectors.toList());
}
*/

func VisiblePacmen(player *Player, game *Game) []*Pacman {
	if !game.Config.FOG_OF_WAR {
		return append([]*Pacman(nil), game.Pacmen...)
	}
	visibleCoords := FindVisibleCoords(player, game, func(c Coord) bool {
		for _, pac := range game.Pacmen {
			if pac.Position == c {
				return true
			}
		}
		return false
	})
	coordSet := make(map[Coord]struct{}, len(visibleCoords))
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

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:1000-1002

private List<Coord> findVisiblePellets(Player player) {
    return findVisibleItems(player, coord -> grid.get(coord).hasPellet());
}
*/

func VisiblePellets(player *Player, game *Game) []Coord {
	if !game.Config.FOG_OF_WAR {
		return game.Grid.AllPellets()
	}
	return FindVisibleCoords(player, game, func(c Coord) bool {
		return game.Grid.Get(c).HasPellet
	})
}

/*
Java: SpringChallenge2020/src/main/java/com/codingame/spring2020/Game.java:976-998

private List<Coord> findVisibleItems(Player player, Function<Coord, Boolean> hasItem) {
    List<Coord> visibleItems = new ArrayList<Coord>();
    player.getAlivePacmen().forEach(playerPac -> {
        for (Coord unitMove : Config.ADJACENCY) {
            Coord currentCoord = playerPac.getPosition();
            while (grid.get(currentCoord).isFloor()) {
                if (hasItem.apply(currentCoord) && !visibleItems.contains(currentCoord)) {
                    visibleItems.add(currentCoord);
                }
                Optional<Coord> nextCoord = grid.getCoordNeighbour(currentCoord, unitMove);
                if (nextCoord.isPresent()) {
                    currentCoord = nextCoord.get();
                } else {
                    break;
                }
                if (playerPac.getPosition().equals(currentCoord)) break;
            }
        }
    });
    return visibleItems;
}
*/

func FindVisibleCoords(player *Player, game *Game, hasItem func(Coord) bool) []Coord {
	visible := make([]Coord, 0)
	seen := make(map[Coord]struct{})
	for _, pac := range player.AlivePacmen() {
		for _, delta := range ADJACENCY {
			current := pac.Position
			for game.Grid.Get(current).IsFloor() {
				if hasItem(current) {
					if _, ok := seen[current]; !ok {
						seen[current] = struct{}{}
						visible = append(visible, current)
					}
				}
				next, ok := game.Grid.CoordNeighbour(current, delta)
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
