package arena

import (
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRunCtx(t *testing.T) (*pflag.FlagSet, *viper.Viper) {
	t.Helper()
	fs := NewBaseFlagSet("arena")
	AddRunFlags(fs)
	v := viper.New()
	require.NoError(t, v.BindPFlags(fs))
	return fs, v
}

func TestParseArgsParsesAllCommonFlags(t *testing.T) {
	fs, v := newTestRunCtx(t)
	got, err := ParseRunArgs([]string{
		"--p0", "./bin/p0",
		"--seed", "-1755827269105404700",
		"--seedx", "7",
		"--max-turns", "123",
		"--simulations", "4",
		"--parallel", "2",
		"--output-matches",
		"--trace-dir", "./tmp/traces",
		"--no-swap",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, int64(-1755827269105404700), got.Seed)
	assert.Equal(t, int64(7), got.SeedIncrement)
	assert.Equal(t, 123, got.MaxTurns)
	assert.Equal(t, 4, got.Simulations)
	assert.Equal(t, 2, got.Parallel)
	assert.True(t, got.OutputMatches)
	assert.Equal(t, "./tmp/traces", got.TraceDir)
	assert.True(t, got.NoSwap)
	assert.Equal(t, filepath.Clean("./bin/opponent"), got.P1Bin)
}

func TestParseArgsAcceptsSeedPrefix(t *testing.T) {
	fs, v := newTestRunCtx(t)
	got, err := ParseRunArgs([]string{
		"--p0", "./bin/p0",
		"--seed", "seed=1001",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), got.Seed)
}

func TestParseArgsCollectsUnknownFlagsAsGameOptions(t *testing.T) {
	fs, v := newTestRunCtx(t)
	got, err := ParseRunArgs([]string{
		"--p0", "./bin/p0",
		"--league", "3",
		"--custom-flag", "value",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "3", got.GameOptions["league"])
	assert.Equal(t, "value", got.GameOptions["custom-flag"])
}

func TestParseArgsRejectsNonPositiveSeedIncrement(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		t.Run(value, func(t *testing.T) {
			fs, v := newTestRunCtx(t)
			_, err := ParseRunArgs([]string{
				"--p0", "./bin/p0",
				"--seedx", value,
			}, fs, v)
			require.Error(t, err)
			assert.EqualError(t, err, "--seedx must be >= 1")
		})
	}
}

func TestParseArgsRequiresP0Bin(t *testing.T) {
	fs, v := newTestRunCtx(t)
	_, err := ParseRunArgs([]string{}, fs, v)
	require.Error(t, err)
	assert.EqualError(t, err, "--p0 is required")
}

func TestParseArgsReadsFromViperConfig(t *testing.T) {
	fs, v := newTestRunCtx(t)
	v.Set("p0", "./bin/cfg-p0")
	v.Set("max-turns", 77)
	got, err := ParseRunArgs([]string{}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "./bin/cfg-p0", got.P0Bin)
	assert.Equal(t, 77, got.MaxTurns)
}
