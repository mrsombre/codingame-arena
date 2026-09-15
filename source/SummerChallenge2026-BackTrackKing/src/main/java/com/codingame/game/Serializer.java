package com.codingame.game;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.codingame.game.grid.ScheduleStep;
import com.codingame.game.grid.Tile;
import com.codingame.game.grid.Town;
import com.google.gson.Gson;

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
        lines.add(Game.PASSIVE_INCOME);

        for (int y = 0; y < game.grid.height; ++y) {
            for (int x = 0; x < game.grid.width; ++x) {
                Tile tile = game.grid.get(x, y);
                lines.add(
                    join(
                        tile.getZoneId(),
                        tile.getType()
                    )
                );
            }
        }

        lines.add(game.grid.towns.size());
        for (Town t : game.grid.towns) {
            lines.add(
                join(
                    t.id,
                    t.coord.getX(),
                    t.coord.getY(),
                    serializeTowns(t.desiredConnections)
                )
            );
        }

        return lines.stream()
            .map(String::valueOf)
            .collect(Collectors.joining(MAIN_SEPARATOR));
    }

    private static String serializeTowns(List<Town> towns) {
        return towns.isEmpty() ? "x"
            : towns.stream()
                .map((Town tc) -> Integer.toString(tc.id))
                .collect(Collectors.joining(","));
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

        // Replace pipe character with mathematical divide symbol.
        game.players.stream()
            .map(Player::getMessage)
            .forEach(s -> lines.add(s == null ? "" : s.replaceAll("\\|", "∣")));

        return lines.stream()
            .map(String::valueOf)
            .collect(Collectors.joining(MAIN_SEPARATOR));
    }

    public static List<String> serializeGlobalInfoFor(Player player, Game game) {
        List<Object> lines = new ArrayList<>();
        lines.add(player.getIndex());
        lines.add(game.grid.width);
        lines.add(game.grid.height);

        for (int y = 0; y < game.grid.height; ++y) {
            for (int x = 0; x < game.grid.width; ++x) {
                Tile tile = game.grid.get(x, y);
                lines.add(
                    join(
                        tile.getZoneId(),
                        tile.getType()
                    )
                );
            }
        }
        lines.add(game.grid.towns.size());
        for (Town t : game.grid.towns) {
            lines.add(
                join(
                    t.id,
                    t.coord.getX(),
                    t.coord.getY(),
                    serializeTowns(t.desiredConnections)
                )
            );
        }

        return lines.stream()
            .map(String::valueOf)
            .collect(Collectors.toList());
    }

    public static List<String> serializeFrameInfoFor(Player player, Game game) {
        List<Object> lines = new ArrayList<>();
        Stream.of(player, game.players.get(1 - player.getIndex()))
            .forEach(
                p -> lines.add(
                    join(
                        p.getScore()
                    )
                )
            );

        //TODO: send zone data separately?
        for (int y = 0; y < game.grid.height; ++y) {
            for (int x = 0; x < game.grid.width; ++x) {
                Tile tile = game.grid.get(x, y);
                lines.add(
                    join(
                        tile.track,
                        game.grid.zones.get(tile.zoneId).instability,
                        game.grid.zones.get(tile.zoneId).inked ? 1 : 0,
                        serializeLiveConnections(tile.activeConnections)
                    )
                );
            }
        }

        return lines.stream()
            .map(String::valueOf)
            .toList();
    }

    private static String serializeLiveConnections(List<ScheduleStep> connection) {
        return connection.isEmpty() ? "x"
            : connection.stream()
                .map(ss -> ss.fromTownId() + "-" + ss.toTownId())
                .sorted()
                .collect(Collectors.joining(","));

    }
}
