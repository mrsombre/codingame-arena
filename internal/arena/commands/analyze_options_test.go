package commands

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newTestAnalyzeCtx(t *testing.T) (*pflag.FlagSet, *viper.Viper) {
	t.Helper()
	fs := arena.NewBaseFlagSet("arena")
	AddAnalyzeFlags(fs)
	v := viper.New()
	require.NoError(t, v.BindPFlags(fs))
	return fs, v
}

func TestParseAnalyzeOptionsDefaults(t *testing.T) {
	fs, v := newTestAnalyzeCtx(t)
	got, err := parseAnalyzeOptions(nil, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "traces", got.TraceDir)
}

func TestParseAnalyzeOptionsParsesFlags(t *testing.T) {
	fs, v := newTestAnalyzeCtx(t)
	got, err := parseAnalyzeOptions([]string{"--trace-dir", "./tmp/traces"}, fs, v)
	require.NoError(t, err)
	assert.Equal(t, "./tmp/traces", got.TraceDir)
}
