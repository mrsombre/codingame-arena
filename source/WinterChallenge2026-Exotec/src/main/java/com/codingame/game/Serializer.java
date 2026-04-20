package com.codingame.game;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Tile;

public class Serializer {
    public static final String MAIN_SEPARATOR = "|";

    static public <T> String serialize(List<T> list, String separator) {
        return list.stream().map(String::valueOf).collect(Collectors.joining(separator));
    }

    static public String serialize(int[] intArray) {
        return Arrays.stream(intArray).mapToObj(String::valueOf).collect(Collectors.joining(" "));
    }

    static public String serialize(boolean[] boolArray) {
        List<String> strs = new ArrayList<>(boolArray.length);
        for (boolean b : boolArray) {
            strs.add(b ? "1" : "0");
        }
        return strs.stream().collect(Collectors.joining(" "));
    }

    static public String join(Object... args) {
        return Stream.of(args)
            .map(String::valueOf)
            .collect(Collectors.joining(" "));
    }

    public static String serializeGlobalData(Game game) {
        List<Object> lines = new ArrayList<>();
        lines.add(game.grid.width);
        lines.add(game.grid.height);

        for (int y = 0; y < game.grid.height; ++y) {
            for (int x = 0; x < game.grid.width; ++x) {
                lines.add(game.grid.get(x, y).getType());// Could be one row
            }
        }
        lines.add(game.grid.apples.size());
        for (Coord c : game.grid.apples) {
            lines.add(c.getX());
            lines.add(c.getY());
        }
        for (Player p : game.players) {
            lines.add(p.birds.size());
            for (Bird b : p.birds) {
                lines.add(b.id);
                lines.add(b.body.size());
                for (Coord c : b.body) {
                    lines.add(c.getX());
                    lines.add(c.getY());
                }
            }
        }
        return lines.stream().map(String::valueOf).collect(Collectors.joining(MAIN_SEPARATOR));
    }

    public static String serializeFrameData(Game game) {
        List<Object> lines = new ArrayList<>();
        lines.add(game.getViewerEvents().size());
        game.getViewerEvents().stream()
            .flatMap(
                e -> Stream.of(
                    e.type,
                    e.animData.start,
                    e.animData.end,
                    serialize(e.params)
                )
            )
            .forEach(lines::add);
        for (Player p : game.players) {
            lines.add(p.getLastExectionTimeMs());
            lines.add(p.marks.size());
            for (Coord c : p.marks) {
                lines.add(c.toIntString());
            }
        }
        List<Bird> birdsWithMessage = game.allBirds.get().filter(Bird::hasMessage).toList();
        lines.add(birdsWithMessage.size());
        for (Bird b : birdsWithMessage) {
            if (b.hasMessage()) {
                lines.add(b.id);
                lines.add(b.owner.getIndex());
                //Replace pipe character with mathematical divide symbol.
                lines.add(b.message.replaceAll("\\|", "∣"));
            }
        }

        return lines.stream().map(String::valueOf).collect(Collectors.joining(MAIN_SEPARATOR));
    }

    public static List<String> serializeGlobalInfoFor(Player player, Game game) {
        List<Object> lines = new ArrayList<>();
        lines.add(player.getIndex());
        lines.add(game.grid.width);
        lines.add(game.grid.height);
        for (int y = 0; y < game.grid.height; ++y) {
            StringBuilder row = new StringBuilder();
            for (int x = 0; x < game.grid.width; ++x) {
                row.append(game.grid.get(x, y).getType() == Tile.TYPE_WALL ? '#' : '.');
            }
            lines.add(row.toString());
        }

        // birds per player
        lines.add(game.players.get(0).birds.size());

        for (Bird b : player.birds) {
            lines.add(b.id);
        }
        for (Bird b : game.players.get(1 - player.getIndex()).birds) {
            lines.add(b.id);
        }

        return lines.stream().map(String::valueOf).collect(Collectors.toList());
    }

    public static List<String> serializeFrameInfoFor(Player player, Game game) {
        List<Object> lines = new ArrayList<>();
        lines.add(game.grid.apples.size());
        for (Coord c : game.grid.apples) {
            lines.add(c.toIntString());
        }
        List<Bird> liveBirds = game.getLiveBirds();
        lines.add(liveBirds.size());
        for (Bird b : liveBirds) {
            String body = b.body.stream().map(c -> c.getX() + "," + c.getY()).collect(Collectors.joining(":"));
            lines.add(join(b.id, body));
        }

        return lines.stream().map(String::valueOf).collect(Collectors.toList());
    }
}
