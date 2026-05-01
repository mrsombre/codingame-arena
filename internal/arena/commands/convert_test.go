package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

func TestConvertReplaySkipsExistingTrace(t *testing.T) {
	dir := makeConvertTestDir(t)
	traceDir := filepath.Join(dir, "traces")
	require.NoError(t, os.MkdirAll(traceDir, 0o755))
	tracePath := filepath.Join(traceDir, arena.TraceFileName(arena.TraceTypeReplay, 42, 0))
	require.NoError(t, os.WriteFile(tracePath, []byte("{}"), 0o644))

	result, err := convertReplay(
		&fakeConvertFactory{puzzleID: 1},
		ConvertOptions{TraceDir: traceDir},
		convertReplayTarget{ID: 42, Path: filepath.Join(dir, "missing.json")},
	)
	require.NoError(t, err)
	assert.Equal(t, convertOutcomeSkippedExisting, result.Outcome)
	assert.Equal(t, "trace exists", result.Detail)
}

func TestConvertReplaySkipsWrongPuzzle(t *testing.T) {
	dir := makeConvertTestDir(t)
	replayPath := writeConvertReplayFile(t, dir, "42.json", `{
  "puzzleId": 999,
  "gameResult": {"scores": [1.0, 0.0], "frames": []}
}`)

	result, err := convertReplay(
		&fakeConvertFactory{puzzleID: 1},
		ConvertOptions{TraceDir: filepath.Join(dir, "traces"), Force: true},
		convertReplayTarget{ID: 42, Path: replayPath},
	)
	require.NoError(t, err)
	assert.Equal(t, convertOutcomeSkippedPuzzle, result.Outcome)
	assert.Equal(t, "puzzleId 999 != 1", result.Detail)
}

func TestConvertReplaySkipsMissingBlueAsMismatch(t *testing.T) {
	dir := makeConvertTestDir(t)
	replayPath := writeConvertReplayFile(t, dir, "42.json", `{
  "puzzleId": 1,
  "seed": "1",
  "gameResult": {"scores": [1.0, 0.0], "frames": []}
}`)

	result, err := convertReplay(
		&fakeConvertFactory{puzzleID: 1},
		ConvertOptions{TraceDir: filepath.Join(dir, "traces"), Force: true},
		convertReplayTarget{ID: 42, Path: replayPath},
	)
	require.NoError(t, err)
	assert.Equal(t, convertOutcomeSkippedMismatch, result.Outcome)
	assert.Contains(t, result.Detail, "replay missing blue")
}

func TestConvertReplayPropagatesReadError(t *testing.T) {
	dir := makeConvertTestDir(t)

	_, err := convertReplay(
		&fakeConvertFactory{puzzleID: 1},
		ConvertOptions{TraceDir: filepath.Join(dir, "traces"), Force: true},
		convertReplayTarget{ID: 42, Path: filepath.Join(dir, "absent.json")},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestConvertReplaysWritesProgressAndAggregates(t *testing.T) {
	dir := makeConvertTestDir(t)
	replayPath := writeConvertReplayFile(t, dir, "42.json", `{
  "puzzleId": 999,
  "gameResult": {"scores": [1.0, 0.0], "frames": []}
}`)

	var stdout bytes.Buffer
	results, err := convertReplays(
		&stdout,
		&fakeConvertFactory{puzzleID: 1},
		ConvertOptions{TraceDir: filepath.Join(dir, "traces"), Force: true},
		[]convertReplayTarget{{ID: 42, Path: replayPath}},
	)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, convertOutcomeSkippedPuzzle, results[0].Outcome)
	assert.Contains(t, stdout.String(), "[1/1] skip 42 (puzzleId 999 != 1)")
}

func TestSummarizeConvertResultsAggregatesOutcomes(t *testing.T) {
	results := []convertResult{
		{Outcome: convertOutcomeSaved},
		{Outcome: convertOutcomeSaved},
		{Outcome: convertOutcomeSkippedExisting},
		{Outcome: convertOutcomeSkippedPuzzle},
		{Outcome: convertOutcomeSkippedPuzzle},
		{Outcome: convertOutcomeSkippedMismatch},
	}
	got := summarizeConvertResults(results)
	assert.Equal(t, convertSummary{
		Total:           6,
		Saved:           2,
		SkippedExisting: 1,
		SkippedPuzzle:   2,
		SkippedMismatch: 1,
	}, got)
}

func TestWriteConvertProgressFormatsVerb(t *testing.T) {
	var saved bytes.Buffer
	writeConvertProgress(&saved, 1, 5, convertResult{
		Target:  convertReplayTarget{ID: 42},
		Outcome: convertOutcomeSaved,
		Detail:  "league=2 turns=10 scores=1.0:0.0",
	})
	assert.Equal(t, "[1/5] save 42 (league=2 turns=10 scores=1.0:0.0)\n", saved.String())

	var skipped bytes.Buffer
	writeConvertProgress(&skipped, 2, 5, convertResult{
		Target:  convertReplayTarget{ID: 100},
		Outcome: convertOutcomeSkippedPuzzle,
		Detail:  "puzzleId 999 != 1",
	})
	assert.Equal(t, "[2/5] skip 100 (puzzleId 999 != 1)\n", skipped.String())
}

func TestWriteConvertSummaryFormat(t *testing.T) {
	var stdout bytes.Buffer
	opts := ConvertOptions{TraceDir: "./tmp/traces"}
	results := []convertResult{
		{Outcome: convertOutcomeSaved},
		{Outcome: convertOutcomeSkippedExisting},
		{Outcome: convertOutcomeSkippedMismatch},
	}
	require.NoError(t, writeConvertSummary(&stdout, opts, results))
	assert.Equal(t,
		"done: 1 saved, 1 skipped-existing, 0 skipped-puzzle, 1 skipped-mismatch (replays=3 out=./tmp/traces)\n",
		stdout.String())
}

func makeConvertTestDir(t *testing.T) string {
	t.Helper()
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dir := filepath.Join("tmp", "convert-test-"+name)
	require.NoError(t, os.RemoveAll(dir))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func writeConvertReplayFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

type fakeConvertFactory struct {
	name     string
	puzzleID int
}

func (f *fakeConvertFactory) Name() string {
	if f.name == "" {
		return "fake-convert"
	}
	return f.name
}

func (f *fakeConvertFactory) PuzzleID() int { return f.puzzleID }

func (f *fakeConvertFactory) PuzzleTitle() string { return "" }

func (f *fakeConvertFactory) NewGame(_ int64, _ *viper.Viper) (arena.Referee, []arena.Player) {
	panic(fmt.Sprintf("fakeConvertFactory.NewGame called for %s", f.Name()))
}

func (f *fakeConvertFactory) MaxTurns() int { return 1 }

var _ arena.GameFactory = (*fakeConvertFactory)(nil)
