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
	"33 13",
	"#################################",
	"#   #     # # #   # # #     #   #",
	"# # # ##### # # # # # ##### # # #",
	"# #           #   #           # #",
	"# ### # # # ### # ### # # # ### #",
	"#       #       #       #       #",
	"### # ##### # # # # # ##### # ###",
	"    #   #     # # #     #   #    ",
	"### ### # # ### # ### # # ### ###",
	"### #     #   #   #   #     # ###",
	"### # # # ### # # # ### # # # ###",
	"###   # # #     #     # # #   ###",
	"#################################",
	"0 0",
	"4",
	"0 1 29 5 ROCK 0 0",
	"1 1 7 4 PAPER 0 0",
	"2 1 19 1 SCISSORS 0 0",
	"3 1 5 6 ROCK 0 0",
	"26",
	"28 5 1",
	"27 5 1",
	"26 5 1",
	"25 5 1",
	"30 5 1",
	"31 5 1",
	"29 6 1",
	"29 7 1",
	"29 8 1",
	"29 9 1",
	"29 10 1",
	"29 11 1",
	"7 3 1",
	"7 5 1",
	"19 2 1",
	"19 3 1",
	"5 5 1",
	"5 4 1",
	"5 3 1",
	"5 2 1",
	"5 1 1",
	"5 7 1",
	"4 3 10",
	"8 3 10",
	"24 3 10",
	"28 3 10",
}, "\n")

// testArenaExpectedInitialInputNegative is the reference output for the
// negative seed, captured from a live match.
var testArenaExpectedInitialInputNegative = strings.Join([]string{
	"33 14",
	"#################################",
	"# #     #   # # # # #   #     # #",
	"# # # ##### # # # # # ##### # # #",
	"#   #   #     # # #     #   #   #",
	"### ### # ### # # # ### # ### ###",
	"          ###       ###          ",
	"### ##### ### ##### ### ##### ###",
	"    #                       #    ",
	"### # # # ### ##### ### # # # ###",
	"        #   #   #   #   #        ",
	"##### ##### # # # # # ##### #####",
	"        #   # #   # #   #        ",
	"### ### # # # ##### # # # ### ###",
	"#################################",
	"0 0",
	"4",
	"0 1 26 9 ROCK 0 0",
	"0 0 6 9 ROCK 0 0",
	"1 1 32 11 PAPER 0 0",
	"1 0 0 11 PAPER 0 0",
	"30",
	"25 9 1",
	"27 9 1",
	"28 9 1",
	"29 9 1",
	"30 9 1",
	"31 9 1",
	"32 9 1",
	"0 9 1",
	"1 9 1",
	"2 9 1",
	"3 9 1",
	"4 9 1",
	"5 9 1",
	"7 9 1",
	"31 11 1",
	"30 11 1",
	"29 11 1",
	"28 11 1",
	"26 11 1",
	"25 11 1",
	"1 11 1",
	"2 11 1",
	"3 11 1",
	"4 11 1",
	"6 11 1",
	"7 11 1",
	"1 7 10",
	"31 7 10",
	"5 11 10",
	"27 11 10",
}, "\n")

// buildInitialInput returns the global info + player-0 first frame info as a
// single newline-joined string — mirrors the output of `arena serialize`.
func buildInitialInput(seed int64, leagueLevel int) string {
	game := NewGame(seed, leagueLevel)
	players := []*Player{NewPlayer(0), NewPlayer(1)}
	game.Init(players)

	lines := make([]string, 0)
	lines = append(lines, SerializeGlobalInfoFor(players[0], game)...)
	lines = append(lines, SerializeFrameInfoFor(players[0], game)...)
	return strings.Join(lines, "\n")
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
