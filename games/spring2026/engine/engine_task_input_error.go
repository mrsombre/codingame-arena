// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/InputError.java
package engine

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/InputError.java:3-23

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
*/

const (
	ErrUnknownCommand   = 1
	ErrUnitNotFound     = 2
	ErrUnitNotOwned     = 3
	ErrNotAvailable     = 4
	ErrOutOfBoard       = 5
	ErrInvalidPlant     = 6
	ErrNoPlant          = 7
	ErrNoFruit          = 8
	ErrExistingPlant    = 9
	ErrNoSeeds          = 10
	ErrNoShack          = 11
	ErrNoCapacity       = 12
	ErrOutOfStock       = 13
	ErrInvalidSkill     = 14
	ErrCantAfford       = 15
	ErrNoIron           = 16
	ErrNoGrass          = 17
	ErrAlreadyUsed      = 18
	ErrNoChop           = 19
	ErrNoHarvest        = 20
	ErrMoveBlocked      = 21
	ErrOpponentBlocking = 22
)

// InputError mirrors Java engine.task.InputError. Critical errors trigger a
// player deactivation; non-critical ones surface as a game-summary entry.
type InputError struct {
	Message   string
	ErrorCode int
	Critical  bool
}

func NewInputError(message string, code int, critical bool) *InputError {
	return &InputError{Message: message, ErrorCode: code, Critical: critical}
}

func (e *InputError) GetMessage() string { return e.Message }
func (e *InputError) GetErrorCode() int  { return e.ErrorCode }
func (e *InputError) IsCritical() bool   { return e.Critical }
