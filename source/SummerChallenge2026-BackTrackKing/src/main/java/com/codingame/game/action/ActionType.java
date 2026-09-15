package com.codingame.game.action;

import java.util.function.BiConsumer;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import com.codingame.game.grid.Coord;

public enum ActionType {

    AUTOPLACE("^AUTOPLACE (?<fromX>\\d+) (?<fromY>\\d+) (?<toX>\\d+) (?<toY>\\d+)", (match, action) -> {
        int fromX = Integer.parseInt(match.group("fromX"));
        int toX = Integer.parseInt(match.group("toX"));
        int fromY = Integer.parseInt(match.group("fromY"));
        int toY = Integer.parseInt(match.group("toY"));
        Coord from = new Coord(fromX, fromY);
        Coord to = new Coord(toX, toY);
        action.setFrom(from);
        action.setTo(to);
    }),
    PLACE_TRACK("^PLACE_TRACKS (?<x>\\d+) (?<y>\\d+)", (match, action) -> {
        action.setCoord(new Coord(Integer.parseInt(match.group("x")), Integer.parseInt(match.group("y"))));
    }),
    DISRUPT("^DISRUPT (?<zoneId>\\d+)", (match, action) -> {
        action.setZoneId(Integer.parseInt(match.group("zoneId")));
    }),
    DISRUPT_ALT("^DISRUPT (?<x>\\d+) (?<y>\\d+)", (match, action) -> {
        action.setCoord(new Coord(Integer.parseInt(match.group("x")), Integer.parseInt(match.group("y"))));
    }),
    MESSAGE("^MESSAGE (?<message>[^;]*)", (match, action) -> {
        action.setMessage(match.group("message"));
    }),

    WAIT("^WAIT", ActionType::doNothing);

    private final Pattern pattern;
    private final BiConsumer<Matcher, Action> consumer;

    private static void doNothing(Matcher m, Action a) {
    }

    ActionType(String pattern, BiConsumer<Matcher, Action> consumer) {
        this.pattern = Pattern.compile(pattern, Pattern.CASE_INSENSITIVE);
        this.consumer = consumer;
    }

    public Pattern getPattern() {
        return pattern;
    }

    public BiConsumer<Matcher, Action> getConsumer() {
        return consumer;
    }

}
