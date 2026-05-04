// Package engine
// Source: SpringChallenge2021/src/main/java/com/codingame/game/Game.java
//
// Ports the input-line serialization methods (getGlobalInfoFor and
// getCurrentFrameInfoFor + getPossibleMoves). Simulation-only logic from the
// same Java source lives in game_game.go.
package engine

import (
	"fmt"
	"strconv"
	"strings"
)

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:352-367

public List<String> getGlobalInfoFor(Player player) {
    List<String> lines = new ArrayList<>();
    lines.add(String.valueOf(board.coords.size()));
    board.coords.forEach(coord -> {
        Cell cell = board.map.get(coord);
        lines.add(String.format("%d %d %s", cell.getIndex(), cell.getRichness(), getNeighbourIds(coord)));
    });
    return lines;
}
*/

func SerializeGlobalInfoFor(_ *Player, game *Game) []string {
	board := game.Board
	lines := make([]string, 0, len(board.Coords)+1)
	lines = append(lines, strconv.Itoa(len(board.Coords)))
	for _, coord := range board.Coords {
		cell := board.Map[coord]
		lines = append(lines, fmt.Sprintf("%d %d %s",
			cell.GetIndex(),
			cell.GetRichness(),
			neighbourIds(board, coord)))
	}
	return lines
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:370-380

private String getNeighbourIds(CubeCoord coord) {
    List<Integer> orderedNeighborIds = new ArrayList<>(CubeCoord.directions.length);
    for (int i = 0; i < CubeCoord.directions.length; ++i) {
        orderedNeighborIds.add(board.map.getOrDefault(coord.neighbor(i), Cell.NO_CELL).getIndex());
    }
    return orderedNeighborIds.stream().map(String::valueOf).collect(Collectors.joining(" "));
}
*/

func neighbourIds(board *Board, coord CubeCoord) string {
	parts := make([]string, len(CubeDirections))
	for i := range CubeDirections {
		idx := -1
		if cell, ok := board.Map[coord.Neighbor(i)]; ok {
			idx = cell.GetIndex()
		}
		parts[i] = strconv.Itoa(idx)
	}
	return strings.Join(parts, " ")
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:214-256

public List<String> getCurrentFrameInfoFor(Player player) {
    List<String> lines = new ArrayList<>();
    lines.add(round); lines.add(nutrients);
    lines.add(player.sun + " " + player.score);
    lines.add(other.sun + " " + other.score + " " + (other.isWaiting()?1:0));
    lines.add(trees.size());
    trees.forEach((index, tree) -> lines.add(index + " " + size + " " + isMine + " " + isDormant));
    List<String> possibleMoves = getPossibleMoves(player);
    lines.add(possibleMoves.size());
    lines.addAll(possibleMoves);
    return lines;
}
*/

func SerializeFrameInfoFor(player *Player, game *Game) []string {
	lines := make([]string, 0, 16)
	lines = append(lines, strconv.Itoa(game.Round))
	lines = append(lines, strconv.Itoa(game.Nutrients))

	other := opponentOf(player, game.Players)
	lines = append(lines, fmt.Sprintf("%d %d", player.GetSun(), player.GetScore()))
	otherWaiting := 0
	if other.IsWaiting() {
		otherWaiting = 1
	}
	lines = append(lines, fmt.Sprintf("%d %d %d", other.GetSun(), other.GetScore(), otherWaiting))

	lines = append(lines, strconv.Itoa(len(game.Trees)))
	for _, idx := range game.TreeOrder {
		tree := game.Trees[idx]
		isMine := 0
		if tree.Owner == player {
			isMine = 1
		}
		isDormant := 0
		if tree.Dormant {
			isDormant = 1
		}
		lines = append(lines, fmt.Sprintf("%d %d %d %d", idx, tree.Size, isMine, isDormant))
	}

	moves := possibleMoves(player, game)
	lines = append(lines, strconv.Itoa(len(moves)))
	lines = append(lines, moves...)
	return lines
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:272-337

private List<String> getPossibleMoves(Player player) {
    List<String> lines = new ArrayList<>();
    lines.add("WAIT");
    if (player.isWaiting()) return lines;
    int seedCost = getSeedCost(player);
    trees.entrySet().stream()
        .filter(e -> e.getValue().getOwner() == player)
        .forEach(e -> {
            // SEED moves if playerCanSeedFrom and target valid
            // GROW or COMPLETE move if affordable and not dormant
        });
    // Collections.shuffle(possibleCompletes, random) for each list, then concat
    return lines;
}
*/

func possibleMoves(player *Player, game *Game) []string {
	out := []string{"WAIT"}
	if player.IsWaiting() {
		return out
	}

	var possibleSeeds, possibleGrows, possibleCompletes []string

	seedCost := game.getSeedCost(player)
	for _, idx := range game.TreeOrder {
		tree := game.Trees[idx]
		if tree.Owner != player {
			continue
		}
		coord := game.Board.Coords[idx]

		if playerCanSeedFrom(game, player, tree, seedCost) {
			for _, target := range coordsInRange(coord, tree.Size) {
				targetCell := game.Board.Map[target]
				if playerCanSeedTo(game, targetCell) {
					possibleSeeds = append(possibleSeeds, fmt.Sprintf("SEED %d %d", idx, targetCell.GetIndex()))
				}
			}
		}

		growCost := game.getGrowthCost(tree)
		if growCost <= player.GetSun() && !tree.Dormant {
			switch {
			case tree.Size == TREE_TALL:
				possibleCompletes = append(possibleCompletes, fmt.Sprintf("COMPLETE %d", idx))
			case game.ENABLE_GROW:
				possibleGrows = append(possibleGrows, fmt.Sprintf("GROW %d", idx))
			}
		}
	}

	// Java: Stream.of(completes, grows, seeds).forEach(list -> { Collections.shuffle(list, random); lines.addAll(list); });
	javaShuffle(possibleCompletes, game.random.NextInt)
	javaShuffle(possibleGrows, game.random.NextInt)
	javaShuffle(possibleSeeds, game.random.NextInt)

	out = append(out, possibleCompletes...)
	out = append(out, possibleGrows...)
	out = append(out, possibleSeeds...)
	return out
}

func playerCanSeedFrom(g *Game, _ *Player, tree *Tree, seedCost int) bool {
	return g.ENABLE_SEED &&
		seedCost <= tree.Owner.GetSun() &&
		tree.Size > TREE_SEED &&
		!tree.Dormant
}

func playerCanSeedTo(g *Game, targetCell *Cell) bool {
	if !targetCell.IsValid() {
		return false
	}
	if targetCell.GetRichness() == RICHNESS_NULL {
		return false
	}
	_, has := g.Trees[targetCell.GetIndex()]
	return !has
}

/*
Java: SpringChallenge2021/src/main/java/com/codingame/game/Game.java:261-270

private List<CubeCoord> getCoordsInRange(CubeCoord center, int N) {
    List<CubeCoord> results = new ArrayList<>();
    for (int x = -N; x <= +N; x++) {
        for (int y = Math.max(-N, -x - N); y <= Math.min(+N, -x + N); y++) {
            int z = -x - y;
            results.add(cubeAdd(center, new CubeCoord(x, y, z)));
        }
    }
    return results;
}
*/

func coordsInRange(center CubeCoord, N int) []CubeCoord {
	results := make([]CubeCoord, 0)
	for x := -N; x <= N; x++ {
		yMin := -N
		if -x-N > yMin {
			yMin = -x - N
		}
		yMax := N
		if -x+N < yMax {
			yMax = -x + N
		}
		for y := yMin; y <= yMax; y++ {
			z := -x - y
			results = append(results, CubeCoord{X: center.X + x, Y: center.Y + y, Z: center.Z + z})
		}
	}
	return results
}

func opponentOf(player *Player, players []*Player) *Player {
	for _, p := range players {
		if p != player {
			return p
		}
	}
	return player
}

// javaShuffle reproduces Java's Collections.shuffle(list, random):
//
//	for (int i = list.size(); i > 1; i--) Collections.swap(list, i - 1, random.nextInt(i));
//
// Identical sequencing matters because possibleMoves is sent to bots and any
// reordering changes the input lines verbatim.
func javaShuffle(list []string, nextInt func(int) int) {
	for i := len(list); i > 1; i-- {
		j := nextInt(i)
		list[i-1], list[j] = list[j], list[i-1]
	}
}
