package engine.task;

import com.codingame.game.Player;
import engine.Board;
import engine.Cell;
import engine.Unit;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

public class MoveTask extends Task {
    protected static final Pattern pattern = Pattern.compile("^\\s*(?<action>MOVE)\\s+(?<id>\\d+)\\s+(?<x>-?\\d+)\\s+(?<y>-?\\d+)\\s*$", Pattern.CASE_INSENSITIVE);

    private Cell target;

    public MoveTask(Player player, Board board, String command) {
        super(player, board);
        Matcher matcher = pattern.matcher(command);
        matcher.matches();
        parseUnit(matcher);
        if (unit == null) return;
        int x = Integer.parseInt(matcher.group("x"));
        int y = Integer.parseInt(matcher.group("y"));
        if (x < 0 || x >= board.getWidth() || y < 0 || y >= board.getHeight())
            addParsingError("(" + x + ", " + y + ") is outside of the board", InputError.OUT_OF_BOARD, false);
        else target = board.getNextCell(unit, board.getCell(x, y));
    }

    @Override
    public int getTaskPriority() {
        return 1;
    }

    @Override
    public int getRequiredLeague() {
        return 1;
    }

    public Cell getTarget() {
        return target;
    }

    /*
    handle all moves of a player simultaneously - but without collisions between different players
    1. replace targets, if initial target is out or range (done at command parsing)
    2. mark all cells as blocked, that currently hold a unit
    3. discard all units that aren't moving, but keep their cells blocked
    4. try to find a unit that can move to an empty cell - then move it and mark previous cell as free
       this can cause conflicts, as different units may have the same target - unit with highest ID wins
    5. solve circular dependencies, where units swap places
     */
    @Override
    public void apply(Board board, ArrayList<Task> concurrentTasks) {
        if (this != concurrentTasks.get(0)) return;
        ArrayList<ArrayList<Task>> byPlayer = groupByPlayer(concurrentTasks);
        for (ArrayList<Task> moves : byPlayer) {
            Player player = moves.get(0).player;
            // list all target cells for player's units
            ArrayList<Unit> units = board.getUnitsByPlayerId(player.getIndex()).collect(Collectors.toCollection(ArrayList::new));
            ArrayList<Cell> targets = units.stream().map(u -> u.getCell()).collect(Collectors.toCollection(ArrayList::new));
            for (Task t : moves) {
                MoveTask m = (MoveTask) t;
                targets.set(units.indexOf(m.unit), m.target);
            }

            // remove all units that don't move
            boolean[][] occupied = new boolean[board.getWidth()][board.getHeight()];
            for (int i = units.size() - 1; i >= 0; i--) {
                occupied[units.get(i).getCell().getX()][units.get(i).getCell().getY()] = true;
                if (units.get(i).getCell() == targets.get(i)) {
                    units.remove(i);
                    targets.remove(i);
                }
            }

            boolean madeMove = true;
            boolean resolveBlocking = false;
            while (madeMove) {
                madeMove = false;

                // check for single target: a->b
                int[][] targetFreq = new int[board.getWidth()][board.getHeight()];
                for (Cell cell : targets) targetFreq[cell.getX()][cell.getY()]++;
                for (int i = units.size() - 1; i >= 0; i--) {
                    Cell cell = targets.get(i);
                    if ((resolveBlocking || targetFreq[cell.getX()][cell.getY()] == 1) && !occupied[cell.getX()][cell.getY()]) {
                        occupied[cell.getX()][cell.getY()] = true;
                        occupied[units.get(i).getCell().getX()][units.get(i).getCell().getY()] = false;
                        units.get(i).setCell(cell);
                        units.remove(i);
                        targets.remove(i);
                        madeMove = true;
                        resolveBlocking = false;
                    }
                }

                if (madeMove) continue;

                // circular cell swaps: a->b, b->a and larger circles
                for (int startIndex = 0; startIndex < units.size(); startIndex++) {
                    ArrayList<Integer> path = new ArrayList<>();
                    path.add(startIndex);
                    boolean looped = false;
                    for (int i = 0; i < units.size() + 1; i++) {
                        Cell target = targets.get(path.get(path.size() - 1));
                        Optional<Unit> blockingUnit = units.stream().filter(u -> u.getCell() == target).findFirst();
                        if (!blockingUnit.isPresent()) break;
                        int index = units.indexOf(blockingUnit.get());
                        if (index == path.get(0)) {
                            looped = true;
                            break;
                        }
                        path.add(index);
                    }
                    if (looped) {
                        Collections.sort(path);
                        for (int i = path.size() - 1; i >= 0; i--) {
                            int idx = path.get(i);
                            units.get(idx).setCell(targets.get(idx));
                            units.remove(idx);
                            targets.remove(idx);
                            madeMove = true;
                        }
                    }
                }

                if (!madeMove && !resolveBlocking) {
                    resolveBlocking = true;
                    madeMove = true;
                }
            }

            for (Task t : moves) {
                t.applied = !units.contains(t.unit);
                if (t.applied)
                    t.player.addSummary("troll " + t.unit.getId() + " moved to (" + t.unit.getCell().getX() + ", " + t.unit.getCell().getY() + ")");
                else
                    t.player.addError(new InputError( "troll " + t.unit.getId() + " can't move to (" + ((MoveTask) t).target.getX() + ", " + ((MoveTask) t).target.getY() + "), target blocked", InputError.MOVE_BLOCKED, false ));
            }
        }
    }
}