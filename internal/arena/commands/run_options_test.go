package commands

import (
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newTestRunCtx(t *testing.T) (*pflag.FlagSet, *viper.Viper) {
	t.Helper()
	fs := arena.NewBaseFlagSet("arena")
	AddRunFlags(fs)
	v := viper.New()
	require.NoError(t, v.BindPFlags(fs))
	return fs, v
}

func TestParseRunOptionsParsesAllCommonFlags(t *testing.T) {
	fs, v := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
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

func TestParseRunOptionsAcceptsSeedPrefix(t *testing.T) {
	fs, v := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--p0", "./bin/p0",
		"--seed", "seed=1001",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), got.Seed)
}

func TestParseRunOptionsExposesLeagueViaViper(t *testing.T) {
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--p0", "./bin/p0",
		"--league", "3",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "3", v.GetString("league"))
}

func TestParseRunOptionsRejectsUnknownFlags(t *testing.T) {
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--p0", "./bin/p0",
		"--unknown", "value",
	}, fs, v)
	require.Error(t, err)
}

func TestParseRunOptionsRejectsNonPositiveSeedIncrement(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		t.Run(value, func(t *testing.T) {
			fs, v := newTestRunCtx(t)
			_, err := parseRunOptions([]string{
				"--p0", "./bin/p0",
				"--seedx", value,
			}, fs, v)
			require.Error(t, err)
			assert.EqualError(t, err, "--seedx must be >= 1")
		})
	}
}

func TestParseRunOptionsRequiresP0Bin(t *testing.T) {
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{}, fs, v)
	require.Error(t, err)
	assert.EqualError(t, err, "--p0 is required")
}

func TestParseRunOptionsReadsFromViperConfig(t *testing.T) {
	fs, v := newTestRunCtx(t)
	v.Set("p0", "./bin/cfg-p0")
	v.Set("max-turns", 77)
	got, err := parseRunOptions([]string{}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "./bin/cfg-p0", got.P0Bin)
	assert.Equal(t, 77, got.MaxTurns)
}
