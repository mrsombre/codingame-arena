package arena

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveSeedKeepsBaseSeedForOffsetZero(t *testing.T) {
	base := int64(-1755827269105404700)
	assert.Equal(t, base, deriveSeed(base, 0, 1))
}

func TestDeriveSeedAlternatesSign(t *testing.T) {
	base := int64(100)
	assert.Equal(t, int64(100), deriveSeed(base, 0, 1))  // even: positive
	assert.Equal(t, int64(-101), deriveSeed(base, 1, 1)) // odd: negative
	assert.Equal(t, int64(102), deriveSeed(base, 2, 1))  // even: positive
	assert.Equal(t, int64(-103), deriveSeed(base, 3, 1)) // odd: negative
}

func TestDeriveSeedCustomIncrement(t *testing.T) {
	base := int64(10)
	increment := int64(7)

	assert.Equal(t, int64(10), deriveSeed(base, 0, increment))  // 10
	assert.Equal(t, int64(-17), deriveSeed(base, 1, increment)) // -(10+7)
	assert.Equal(t, int64(24), deriveSeed(base, 2, increment))  // 10+14
	assert.Equal(t, int64(-31), deriveSeed(base, 3, increment)) // -(10+21)
}

func TestRunMatchesCollectsResultsInOrder(t *testing.T) {
	opts := BatchOptions{Simulations: 5, Parallel: 3, Seed: 42, SeedIncrement: 1}
	results := RunMatches(opts, func(id int, seed int64) MatchResult {
		return MatchResult{ID: id, Seed: seed}
	})
	assert.Len(t, results, 5)
	for i, r := range results {
		assert.Equal(t, i, r.ID)
	}
}
