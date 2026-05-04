package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testArenaPositiveSeed = int64(468706172918629800)
	testArenaNegativeSeed = int64(-468706172918629800)
)

// testArenaExpectedInitialInputPositive is the concatenation of global info
// and the first frame info for player 0 at the positive reference seed,
// captured from a live match. Acceptance bar: the engine must reproduce this
// byte-for-byte.
var testArenaExpectedInitialInputPositive = strings.Join([]string{
	"37",
	"0 3 1 2 3 4 5 6",
	"1 3 7 8 2 0 6 18",
	"2 3 8 9 10 3 0 1",
	"3 0 2 10 11 12 4 0",
	"4 3 0 3 12 13 14 5",
	"5 3 6 0 4 14 15 16",
	"6 0 18 1 0 5 16 17",
	"7 2 19 20 8 1 18 36",
	"8 2 20 21 9 2 1 7",
	"9 2 21 22 23 10 2 8",
	"10 2 9 23 24 11 3 2",
	"11 2 10 24 25 26 12 3",
	"12 2 3 11 26 27 13 4",
	"13 2 4 12 27 28 29 14",
	"14 2 5 4 13 29 30 15",
	"15 2 16 5 14 30 31 32",
	"16 2 17 6 5 15 32 33",
	"17 2 35 18 6 16 33 34",
	"18 2 36 7 1 6 17 35",
	"19 1 -1 -1 20 7 36 -1",
	"20 0 -1 -1 21 8 7 19",
	"21 1 -1 -1 22 9 8 20",
	"22 1 -1 -1 -1 23 9 21",
	"23 1 22 -1 -1 24 10 9",
	"24 0 23 -1 -1 25 11 10",
	"25 1 24 -1 -1 -1 26 11",
	"26 0 11 25 -1 -1 27 12",
	"27 1 12 26 -1 -1 28 13",
	"28 1 13 27 -1 -1 -1 29",
	"29 0 14 13 28 -1 -1 30",
	"30 1 15 14 29 -1 -1 31",
	"31 1 32 15 30 -1 -1 -1",
	"32 1 33 16 15 31 -1 -1",
	"33 0 34 17 16 32 -1 -1",
	"34 1 -1 35 17 33 -1 -1",
	"35 0 -1 36 18 17 34 -1",
	"36 1 -1 19 7 18 35 -1",
	"1",
	"20",
	"4 0",
	"4 0 0",
	"4",
	"19 1 1 0",
	"22 1 1 0",
	"28 1 0 0",
	"31 1 0 0",
	"8",
	"WAIT",
	"GROW 19",
	"GROW 22",
	"SEED 19 7",
	"SEED 22 9",
	"SEED 22 23",
	"SEED 19 36",
	"SEED 22 21",
}, "\n")

// testArenaExpectedInitialInputNegative is the reference output for the
// negative reference seed, captured from a live match.
var testArenaExpectedInitialInputNegative = strings.Join([]string{
	"37",
	"0 3 1 2 3 4 5 6",
	"1 3 7 8 2 0 6 18",
	"2 3 8 9 10 3 0 1",
	"3 3 2 10 11 12 4 0",
	"4 3 0 3 12 13 14 5",
	"5 3 6 0 4 14 15 16",
	"6 3 18 1 0 5 16 17",
	"7 2 19 20 8 1 18 36",
	"8 2 20 21 9 2 1 7",
	"9 2 21 22 23 10 2 8",
	"10 0 9 23 24 11 3 2",
	"11 2 10 24 25 26 12 3",
	"12 2 3 11 26 27 13 4",
	"13 2 4 12 27 28 29 14",
	"14 2 5 4 13 29 30 15",
	"15 2 16 5 14 30 31 32",
	"16 0 17 6 5 15 32 33",
	"17 2 35 18 6 16 33 34",
	"18 2 36 7 1 6 17 35",
	"19 1 -1 -1 20 7 36 -1",
	"20 1 -1 -1 21 8 7 19",
	"21 1 -1 -1 22 9 8 20",
	"22 1 -1 -1 -1 23 9 21",
	"23 0 22 -1 -1 24 10 9",
	"24 0 23 -1 -1 25 11 10",
	"25 1 24 -1 -1 -1 26 11",
	"26 0 11 25 -1 -1 27 12",
	"27 1 12 26 -1 -1 28 13",
	"28 1 13 27 -1 -1 -1 29",
	"29 1 14 13 28 -1 -1 30",
	"30 1 15 14 29 -1 -1 31",
	"31 1 32 15 30 -1 -1 -1",
	"32 0 33 16 15 31 -1 -1",
	"33 0 34 17 16 32 -1 -1",
	"34 1 -1 35 17 33 -1 -1",
	"35 0 -1 36 18 17 34 -1",
	"36 1 -1 19 7 18 35 -1",
	"1",
	"20",
	"4 0",
	"4 0 0",
	"4",
	"21 1 0 0",
	"27 1 0 0",
	"30 1 1 0",
	"36 1 1 0",
	"10",
	"WAIT",
	"GROW 36",
	"GROW 30",
	"SEED 30 14",
	"SEED 36 19",
	"SEED 30 31",
	"SEED 36 18",
	"SEED 30 29",
	"SEED 30 15",
	"SEED 36 7",
}, "\n")

// buildInitialInput returns the global info + frame info for player 0,
// captured at the same point the Java engine sends input to bots in round 1
// ACTIONS — i.e. after both players auto-WAIT in round 0 and the sun moves
// once. The captured reference data the test compares against was logged at
// that exact frame.
func buildInitialInput(seed int64, leagueLevel int) string {
	game := NewGame(seed, leagueLevel)
	players := []*Player{NewPlayer(0), NewPlayer(1)}
	game.Init(players)
	advanceToRound1Actions(game, players)

	lines := make([]string, 0)
	lines = append(lines, SerializeGlobalInfoFor(players[0], game)...)
	lines = append(lines, SerializeFrameInfoFor(players[0], game)...)
	return strings.Join(lines, "\n")
}

// advanceToRound1Actions mirrors Java's referee gameTurn loop until the start
// of round 1 ACTIONS. At each ACTIONS frame Java serializes the frame info
// for every non-waiting player — that call shuffles the possible-move lists
// via the shared Random, so we must do the same to keep the seeded RNG in
// sync. Without it, round 1's SEED ordering drifts.
func advanceToRound1Actions(g *Game, players []*Player) {
	for g.Round != 1 || g.NextFrameType != FrameActions {
		g.ResetGameTurnData()
		if g.CurrentFrameType == FrameActions {
			for _, p := range players {
				if !p.IsWaiting() {
					_ = SerializeFrameInfoFor(p, g)
				}
			}
		}
		g.PerformGameUpdate(0)
	}
	g.ResetGameTurnData()
}

// TestSeedDeterminism guards determinism: identical seeds must produce
// identical initial state.
func TestSeedDeterminism(t *testing.T) {
	a := buildInitialInput(testArenaPositiveSeed, 4)
	b := buildInitialInput(testArenaPositiveSeed, 4)
	assert.Equal(t, a, b, "initial state must be deterministic for a given seed")
}

// TestSeedPositiveParityCheck is the byte-for-byte reference match for the
// positive reference seed at league 4.
func TestSeedPositiveParityCheck(t *testing.T) {
	got := buildInitialInput(testArenaPositiveSeed, 4)
	assert.Equalf(t, testArenaExpectedInitialInputPositive, got, "arena parity mismatch for seed=%d", testArenaPositiveSeed)
}

// TestSeedNegativeParityCheck is the byte-for-byte reference match for the
// negative reference seed at league 4.
func TestSeedNegativeParityCheck(t *testing.T) {
	got := buildInitialInput(testArenaNegativeSeed, 4)
	assert.Equalf(t, testArenaExpectedInitialInputNegative, got, "arena parity mismatch for seed=%d", testArenaNegativeSeed)
}
