package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// Seed-parity tests for GridMaker. Mirrors
// games/spring2026/engine/engine_board_seed_test.go.
//
// Each parity test compares the engine's output against a reference captured
// from a real upstream CodinGame match (Java referee + the same seed). Until
// those references are pasted in below the parity assertions stay behind a
// t.Skip; TestSeedDeterminism runs unconditionally and guards against
// non-deterministic regressions in the meantime.

const (
	testArenaPositiveSeed = int64(468706172918629800)
	testArenaNegativeSeed = int64(-468706172918629800)
)

// testArenaExpectedGlobalInfoPositive is the reference global-info block for
// testArenaPositiveSeed. Drop in the lines a bot reads on stdin before its
// first turn, minus the leading player-index line: grid width, grid height,
// one `regionId type` line per cell in row-major order, the town count, then
// one `id x y desiredConnections` line per town.
var testArenaExpectedGlobalInfoPositive = strings.Join([]string{}, "\n")

// testArenaExpectedGlobalInfoNegative is the reference for the negative seed.
// Same format as the positive case.
var testArenaExpectedGlobalInfoNegative = strings.Join([]string{}, "\n")

// buildGlobalInfo generates a map from seed and serializes it exactly as the
// engine hands it to a bot, minus the player-index line so the artefact is
// the map alone. It covers everything GridMaker decides: dimensions, the
// region partition, terrain, and the town roster with its desired
// connections.
func buildGlobalInfo(seed int64) string {
	maker := NewGridMaker()
	maker.Init(sha1prng.New(seed), false)
	game := NewGame(nil, DEFAULT_LEAGUE)
	game.Grid = maker.Make()

	return strings.Join(SerializeGlobalInfoFor(NewPlayer(0), game)[1:], "\n")
}

// TestSeedDeterminism guards determinism: the same seed must produce the same
// map every run.
func TestSeedDeterminism(t *testing.T) {
	for _, seed := range []int64{testArenaPositiveSeed, testArenaNegativeSeed} {
		assert.Equalf(t, buildGlobalInfo(seed), buildGlobalInfo(seed),
			"map generation must be deterministic for seed=%d", seed)
	}
}

// TestSeedPositiveParityCheck is the byte-for-byte reference match for the
// positive reference seed. Skipped until reference data lands.
func TestSeedPositiveParityCheck(t *testing.T) {
	if testArenaExpectedGlobalInfoPositive == "" {
		t.Skip("no captured upstream reference yet for seed=468706172918629800; " +
			"run an upstream match at that seed, capture the global-info block a bot " +
			"reads on stdin (width, height, per-cell `regionId type`, town count, town " +
			"lines) and paste it into testArenaExpectedGlobalInfoPositive")
	}
	assert.Equalf(t, testArenaExpectedGlobalInfoPositive, buildGlobalInfo(testArenaPositiveSeed),
		"map parity mismatch for seed=%d", testArenaPositiveSeed)
}

// TestSeedNegativeParityCheck is the byte-for-byte reference match for the
// negative reference seed. Skipped until reference data lands.
func TestSeedNegativeParityCheck(t *testing.T) {
	if testArenaExpectedGlobalInfoNegative == "" {
		t.Skip("no captured upstream reference yet for seed=-468706172918629800; " +
			"run an upstream match at that seed, capture the global-info block a bot " +
			"reads on stdin (width, height, per-cell `regionId type`, town count, town " +
			"lines) and paste it into testArenaExpectedGlobalInfoNegative")
	}
	assert.Equalf(t, testArenaExpectedGlobalInfoNegative, buildGlobalInfo(testArenaNegativeSeed),
		"map parity mismatch for seed=%d", testArenaNegativeSeed)
}
