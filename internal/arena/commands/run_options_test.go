package commands

import (
	"os"
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

// fakeBots creates two empty executable files in t.TempDir() and returns their paths.
func fakeBots(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	p0 := filepath.Join(dir, "p0")
	p1 := filepath.Join(dir, "p1")
	for _, p := range []string{p0, p1} {
		require.NoError(t, os.WriteFile(p, []byte{}, 0o755))
	}
	return p0, p1
}

func TestParseRunOptionsParsesAllCommonFlags(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--p0", p0,
		"--p1", p1,
		"--seed", "-1755827269105404700",
		"--seedx", "7",
		"--max-turns", "123",
		"--simulations", "4",
		"--parallel", "2",
		"--output-matches",
		"--trace",
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
	assert.True(t, got.Trace)
	assert.Equal(t, "./tmp/traces", got.TraceDir)
	assert.True(t, got.NoSwap)
	assert.Equal(t, p1, got.P1Bin)
}

func TestParseRunOptionsTraceDirDefault(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	got, err := parseRunOptions([]string{"--p0", p0, "--p1", p1}, fs, v)
	require.NoError(t, err)
	assert.False(t, got.Trace)
	assert.Equal(t, "./traces", got.TraceDir)
}

func TestParseRunOptionsAcceptsSeedPrefix(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--p0", p0,
		"--p1", p1,
		"--seed", "seed=1001",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), got.Seed)
}

func TestParseRunOptionsExposesLeagueViaViper(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--p0", p0,
		"--p1", p1,
		"--league", "3",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "3", v.GetString("league"))
}

func TestParseRunOptionsRejectsUnknownFlags(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--p0", p0,
		"--p1", p1,
		"--unknown", "value",
	}, fs, v)
	require.Error(t, err)
}

func TestParseRunOptionsRejectsNonPositiveSeedIncrement(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		t.Run(value, func(t *testing.T) {
			p0, p1 := fakeBots(t)
			fs, v := newTestRunCtx(t)
			_, err := parseRunOptions([]string{
				"--p0", p0,
				"--p1", p1,
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

func TestParseRunOptionsRejectsMissingP0(t *testing.T) {
	_, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	missing := filepath.Join(t.TempDir(), "missing")
	_, err := parseRunOptions([]string{"--p0", missing, "--p1", p1}, fs, v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--p0")
	assert.Contains(t, err.Error(), "does not exist")
}

func TestParseRunOptionsRejectsNonExecutableP1(t *testing.T) {
	p0, _ := fakeBots(t)
	dir := t.TempDir()
	p1 := filepath.Join(dir, "p1")
	require.NoError(t, os.WriteFile(p1, []byte{}, 0o644))
	fs, v := newTestRunCtx(t)
	_, err := parseRunOptions([]string{"--p0", p0, "--p1", p1}, fs, v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--p1")
	assert.Contains(t, err.Error(), "not executable")
}

func TestParseRunOptionsReadsFromViperConfig(t *testing.T) {
	p0, p1 := fakeBots(t)
	fs, v := newTestRunCtx(t)
	v.Set("p0", p0)
	v.Set("p1", p1)
	v.Set("max-turns", 77)
	got, err := parseRunOptions([]string{}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, p0, got.P0Bin)
	assert.Equal(t, 77, got.MaxTurns)
}
