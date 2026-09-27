package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newTestRunCtx(t *testing.T) *pflag.FlagSet {
	t.Helper()
	fs := arena.NewBaseFlagSet("arena")
	AddRunFlags(fs)
	return fs
}

// fakeBots creates two empty executable bot files and returns blue/red paths.
func fakeBots(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	blueBot := filepath.Join(dir, "blue")
	redBot := filepath.Join(dir, "red")
	for _, bot := range []string{blueBot, redBot} {
		require.NoError(t, os.WriteFile(bot, []byte{}, 0o755))
	}
	return blueBot, redBot
}

func TestParseRunOptionsParsesAllCommonFlags(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--seed", "-1755827269105404700",
		"--seedx", "7",
		"--max-turns", "123",
		"--simulations", "4",
		"--parallel", "2",
		"--output-matches",
		"--trace",
		"--trace-dir", "./tmp/traces",
		"--no-swap",
	}, fs)
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
	assert.Equal(t, redBot, got.RedBotBin)
}

func TestParseRunOptionsTraceDirDefault(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	got, err := parseRunOptions([]string{"--blue", blueBot, "--red", redBot}, fs)
	require.NoError(t, err)
	assert.False(t, got.Trace)
	assert.Equal(t, "./traces", got.TraceDir)
}

func TestParseRunOptionsAcceptsSeedPrefix(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--seed", "seed=1001",
	}, fs)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), got.Seed)
}

func TestParseRunOptionsParsesLeague(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--league", "3",
	}, fs)
	require.NoError(t, err)
	assert.Equal(t, 3, got.League)
}

func TestParseRunOptionsRejectsNegativeLeague(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--league", "-1",
	}, fs)
	require.EqualError(t, err, "--league must be >= 0")
}

func TestParseRunOptionsRejectsUnknownFlags(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	_, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--unknown", "value",
	}, fs)
	require.Error(t, err)
}

func TestParseRunOptionsRejectsNonPositiveSeedIncrement(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		t.Run(value, func(t *testing.T) {
			blueBot, redBot := fakeBots(t)
			fs := newTestRunCtx(t)
			_, err := parseRunOptions([]string{
				"--blue", blueBot,
				"--red", redBot,
				"--seedx", value,
			}, fs)
			require.Error(t, err)
			assert.EqualError(t, err, "--seedx must be >= 1")
		})
	}
}

func TestParseRunOptionsRequiresBlueBotBin(t *testing.T) {
	fs := newTestRunCtx(t)
	_, err := parseRunOptions([]string{}, fs)
	require.Error(t, err)
	assert.EqualError(t, err, "--blue is required")
}

func TestParseRunOptionsRejectsMissingBlueBot(t *testing.T) {
	_, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	missing := filepath.Join(t.TempDir(), "missing")
	_, err := parseRunOptions([]string{"--blue", missing, "--red", redBot}, fs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--blue")
	assert.Contains(t, err.Error(), "does not exist")
}

func TestParseRunOptionsRejectsNonExecutableRedBot(t *testing.T) {
	blueBot, _ := fakeBots(t)
	dir := t.TempDir()
	redBot := filepath.Join(dir, "red")
	require.NoError(t, os.WriteFile(redBot, []byte{}, 0o644))
	fs := newTestRunCtx(t)
	_, err := parseRunOptions([]string{"--blue", blueBot, "--red", redBot}, fs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--red")
	assert.Contains(t, err.Error(), "not executable")
}

func TestParseRunOptionsDebugForcesSingleSerialMatch(t *testing.T) {
	blueBot, redBot := fakeBots(t)
	fs := newTestRunCtx(t)
	got, err := parseRunOptions([]string{
		"--blue", blueBot,
		"--red", redBot,
		"--debug",
		"--simulations", "25",
		"--parallel", "8",
	}, fs)
	require.NoError(t, err)
	assert.Equal(t, 1, got.Simulations)
	assert.Equal(t, 1, got.Parallel)
}
