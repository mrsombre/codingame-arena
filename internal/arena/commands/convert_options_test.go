package commands

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newTestConvertCtx(t *testing.T) (*pflag.FlagSet, *viper.Viper) {
	t.Helper()
	fs := arena.NewBaseFlagSet("arena")
	AddConvertFlags(fs)
	v := viper.New()
	require.NoError(t, v.BindPFlags(fs))
	return fs, v
}

func TestParseConvertOptionsDefaults(t *testing.T) {
	fs, v := newTestConvertCtx(t)
	got, err := parseConvertOptions(nil, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "traces", got.TraceDir)
	assert.Equal(t, "replays", got.ReplayDir)
	assert.False(t, got.Force)
	assert.Equal(t, "", got.League)
	assert.Nil(t, got.IDs)
}

func TestParseConvertOptionsParsesFlags(t *testing.T) {
	fs, v := newTestConvertCtx(t)
	got, err := parseConvertOptions([]string{
		"--trace-dir", "./tmp/traces",
		"--replay-dir", "./tmp/replays",
		"--force",
		"--league", "3",
		"123",
		"456",
	}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "./tmp/traces", got.TraceDir)
	assert.Equal(t, "./tmp/replays", got.ReplayDir)
	assert.True(t, got.Force)
	assert.Equal(t, "3", got.League)
	assert.Equal(t, []int64{123, 456}, got.IDs)
}

func TestParseConvertOptionsRejectsInvalidLeague(t *testing.T) {
	fs, v := newTestConvertCtx(t)
	_, err := parseConvertOptions([]string{"--league", "abc"}, fs, v)
	require.Error(t, err)
	assert.EqualError(t, err, "--league must be a positive integer")
}

func TestParseConvertOptionsRejectsInvalidReplayID(t *testing.T) {
	fs, v := newTestConvertCtx(t)
	_, err := parseConvertOptions([]string{"abc"}, fs, v)
	require.Error(t, err)
	assert.EqualError(t, err, `invalid replay id "abc"`)
}
