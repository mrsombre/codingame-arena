package com.codingame.game.action;

import java.util.function.BiConsumer;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import com.codingame.game.grid.Coord;
import com.codingame.game.grid.Direction;

public enum ActionType {

    MOVE_UP("^(?<birdId>\\d+) UP( (?<message>[^;]*))?", (match, action) -> {
        action.setBirdId(Integer.valueOf(match.group("birdId")));
        action.setDirection(Direction.NORTH);
        action.setMessage(match.group("message"));
    }),

    MOVE_DOWN("^(?<birdId>\\d+) DOWN( (?<message>[^;]*))?", (match, action) -> {
        action.setBirdId(Integer.valueOf(match.group("birdId")));
        action.setDirection(Direction.SOUTH);
        action.setMessage(match.group("message"));
    }),

    MOVE_LEFT("^(?<birdId>\\d+) LEFT( (?<message>[^;]*))?", (match, action) -> {
        action.setBirdId(Integer.valueOf(match.group("birdId")));
        action.setDirection(Direction.WEST);
        action.setMessage(match.group("message"));
    }),

    MOVE_RIGHT("^(?<birdId>\\d+) RIGHT( (?<message>[^;]*))?", (match, action) -> {
        action.setBirdId(Integer.valueOf(match.group("birdId")));
        action.setDirection(Direction.EAST);
        action.setMessage(match.group("message"));
    }),
    
    MARK("MARK (?<x>\\d+) (?<y>\\d+)", (match, action) -> {
        int x = Integer.parseInt(match.group("x"));
        int y = Integer.parseInt(match.group("y"));
        action.setCoord(new Coord(x, y));
    }),
    
    WAIT("WAIT", ActionType::doNothing);
    


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
