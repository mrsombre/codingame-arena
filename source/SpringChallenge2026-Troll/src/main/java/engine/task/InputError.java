package engine.task;

public class InputError {
    public static final int UNKNOWN_COMMAND = 1;
    public static final int UNIT_NOT_FOUND = 2;
    public static final int UNIT_NOT_OWNED = 3;
    public static final int NOT_AVAILABLE = 4;
    public static final int OUT_OF_BOARD = 5;
    public static final int INVALID_PLANT = 6;
    public static final int NO_PLANT = 7;
    public static final int NO_FRUIT = 8;
    public static final int EXISTING_PLANT = 9;
    public static final int NO_SEEDS = 10;
    public static final int NO_SHACK = 11;
    public static final int NO_CAPACITY = 12;
    public static final int OUT_OF_STOCK = 13;
    public static final int INVALID_SKILL = 14;
    public static final int CANT_AFFORD = 15;
    public static final int NO_IRON = 16;
    public static final int NO_GRASS = 17;
    public static final int ALREADY_USED = 18;
    public static final int NO_CHOP = 19;
    public static final int NO_HARVEST = 20;
    public static final int MOVE_BLOCKED = 21;
    public static final int OPPONENT_BLOCKING = 22;

    private String message;
    private int errorCode;
    private boolean critical;

    public InputError(String message, int errorCode, boolean critical) {
        this.message = message;
        this.errorCode = errorCode;
        this.critical = critical;
    }

    public String getMessage() {
        return message;
    }

    public int getErrorCode() {
        return errorCode;
    }

    public boolean isCritical() {
        return critical;
    }
}