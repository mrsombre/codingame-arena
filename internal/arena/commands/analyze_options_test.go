package commands

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func newTestAnalyzeCtx(t *testing.T) *pflag.FlagSet {
	t.Helper()
	fs := arena.NewBaseFlagSet("arena")
	AddAnalyzeFlags(fs)
	return fs
}

func TestParseAnalyzeOptionsDefaults(t *testing.T) {
	fs := newTestAnalyzeCtx(t)
	got, err := parseAnalyzeOptions(nil, fs)
	require.NoError(t, err)
	assert.Equal(t, "traces", got.TraceDir)
}

func TestParseAnalyzeOptionsParsesFlags(t *testing.T) {
	fs := newTestAnalyzeCtx(t)
	got, err := parseAnalyzeOptions([]string{"--trace-dir", "./tmp/traces"}, fs)
	require.NoError(t, err)
	assert.Equal(t, "./tmp/traces", got.TraceDir)
}
