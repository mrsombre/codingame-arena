package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// Seed-parity tests for Board map generation. Mirrors
// games/spring2020/engine/spring2020_maps_tetris_based_map_generator_test.go.
//
// Each parity test compares the engine's output against a reference string
// captured from a real upstream CodinGame match (Java referee + the same
// seed). Until the references are pasted in below, the parity assertions
// stay behind a t.Skip; TestSeedDeterminism still runs and guards against
// non-deterministic regressions in the meantime.

const (
	testArenaPositiveSeed = int64(468706172918629800)
	testArenaNegativeSeed = int64(-468706172918629800)
	testArenaLeague       = 4
)

// testArenaExpectedInitialInputPositive is the reference output captured
// from a live upstream match at testArenaPositiveSeed / league 4. Drop in
// the global-info lines (width/height + ASCII map) followed by the first
// frame-info block (recipient inventory, opponent inventory, tree count,
// tree lines, troll count, troll lines) — exactly what the bot sees on
// stdin during turn 1.
var testArenaExpectedInitialInputPositive = strings.Join([]string{
	"18 9",
	"...+....~~~~......",
	"........~~~~~.....",
	"........~~~~~~..~.",
	"...#+#...~~~~~~~~.",
	".......0..1.......",
	".~~~~~~~~...#+#...",
	".~..~~~~~~........",
	".....~~~~~........",
	"......~~~~....+...",
	"7 6 2 4 6 0",
	"7 6 2 4 6 0",
	"12",
	"PLUM 17 5 2 8 0 2",
	"PLUM 0 3 2 8 0 2",
	"LEMON 15 7 1 6 0 3",
	"LEMON 2 1 1 6 0 3",
	"LEMON 11 4 2 8 0 2",
	"LEMON 6 4 2 8 0 2",
	"LEMON 5 1 2 8 0 8",
	"LEMON 12 7 2 8 0 8",
	"APPLE 4 4 1 11 0 2",
	"APPLE 13 4 1 11 0 2",
	"BANANA 1 4 1 3 0 2",
	"BANANA 16 4 1 3 0 2",
	"2",
	"0 0 7 4 1 1 1 1 0 0 0 0 0 0",
	"1 1 10 4 1 1 1 1 0 0 0 0 0 0",
}, "\n")

// testArenaExpectedInitialInputNegative is the reference output for the
// negative seed at league 4. Same format as the positive case.
var testArenaExpectedInitialInputNegative = strings.Join([]string{
	"18 9",
	"~~~~..#.......~~.#",
	"~~~~..........~~..",
	"..~~....0..##.#.#.",
	".#.~......+###....",
	".....+......+.....",
	"....###+......~.#.",
	".#.#.##..1....~~..",
	"..~~..........~~~~",
	"#.~~.......#..~~~~",
	"3 3 10 8 2 0",
	"3 3 10 8 2 0",
	"14",
	"PLUM 1 7 4 12 0 2",
	"PLUM 16 1 4 12 0 2",
	"PLUM 8 3 3 10 0 8",
	"PLUM 9 5 3 10 0 8",
	"LEMON 4 7 4 12 0 1",
	"LEMON 13 1 4 12 0 1",
	"APPLE 7 8 3 17 0 4",
	"APPLE 10 0 3 17 0 4",
	"APPLE 0 6 3 17 0 5",
	"APPLE 17 2 3 17 0 5",
	"APPLE 8 6 4 20 0 4",
	"APPLE 9 2 4 20 0 4",
	"BANANA 0 7 4 6 0 1",
	"BANANA 17 1 4 6 0 1",
	"2",
	"0 0 8 2 1 1 1 1 0 0 0 0 0 0",
	"1 1 9 6 1 1 1 1 0 0 0 0 0 0",
}, "\n")

// buildInitialInput returns the global info + player-0 first frame info as a
// single newline-joined string — mirrors what the bot reads from stdin on
// turn 1. This is the canonical surface for seed-parity.
func buildInitialInput(seed int64, league int) string {
	p0 := NewPlayer(0)
	p1 := NewPlayer(1)
	board := CreateMap([]*Player{p0, p1}, sha1prng.New(seed), league)

	lines := make([]string, 0)
	lines = append(lines, board.GetInitialInputs(0)...)
	lines = append(lines, board.GetTurnInputs(0)...)
	return strings.Join(lines, "\n")
}

// TestSeedDeterminism guards determinism: identical seeds must produce
// identical initial state.
func TestSeedDeterminism(t *testing.T) {
	a := buildInitialInput(testArenaPositiveSeed, testArenaLeague)
	b := buildInitialInput(testArenaPositiveSeed, testArenaLeague)
	assert.Equal(t, a, b, "initial state must be deterministic for a given seed")
}

// TestSeedPositiveParityCheck is the byte-for-byte reference match for the
// positive reference seed at league 4. Skipped until reference data lands.
func TestSeedPositiveParityCheck(t *testing.T) {
	if testArenaExpectedInitialInputPositive == "" {
		t.Skip("no captured upstream reference yet for seed=" +
			"468706172918629800; paste output into testArenaExpectedInitialInputPositive")
	}
	got := buildInitialInput(testArenaPositiveSeed, testArenaLeague)
	assert.Equalf(t, testArenaExpectedInitialInputPositive, got,
		"arena parity mismatch for seed=%d", testArenaPositiveSeed)
}

// TestSeedNegativeParityCheck is the byte-for-byte reference match for the
// negative reference seed at league 4. Skipped until reference data lands.
func TestSeedNegativeParityCheck(t *testing.T) {
	if testArenaExpectedInitialInputNegative == "" {
		t.Skip("no captured upstream reference yet for seed=" +
			"-468706172918629800; paste output into testArenaExpectedInitialInputNegative")
	}
	got := buildInitialInput(testArenaNegativeSeed, testArenaLeague)
	assert.Equalf(t, testArenaExpectedInitialInputNegative, got,
		"arena parity mismatch for seed=%d", testArenaNegativeSeed)
}
