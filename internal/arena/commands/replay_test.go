package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/arena/codingame"
)

func TestDownloadReplaySavesPreparedBody(t *testing.T) {
	dir := makeReplayTestDir(t)
	cfg := replayBatchConfig{
		OutDir: dir,
		Annotations: arena.ReplayAnnotations{
			Blue:   "us",
			Source: arena.ReplaySourceGet,
		},
	}
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{42: []byte(`{"gameResult":{"refereeInput":"seed=7"}}`)},
	}

	result, err := downloadReplay(fetcher, cfg, 42, time.Unix(1700000000, 0).UTC())
	require.NoError(t, err)
	assert.Equal(t, downloadOutcomeSaved, result.Outcome)
	assert.Equal(t, int64(42), result.ID)
	assert.Regexp(t, `^\d+ bytes$`, result.Detail)

	saved, err := os.ReadFile(replayFilePath(dir, 42))
	require.NoError(t, err)
	var top map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(saved, &top))
	assert.JSONEq(t, `"us"`, string(top["blue"]))
	assert.JSONEq(t, `"get"`, string(top["source"]))
	assert.JSONEq(t, `"7"`, string(top["seed"]))
	assert.Contains(t, string(top["fetchedAt"]), "2023-")
}

func TestDownloadReplayMarksFetchErrorAsFailed(t *testing.T) {
	dir := makeReplayTestDir(t)
	cfg := replayBatchConfig{OutDir: dir}
	fetcher := &fakeReplayFetcher{err: errors.New("HTTP 503")}

	result, err := downloadReplay(fetcher, cfg, 99, time.Now())
	require.NoError(t, err)
	assert.Equal(t, downloadOutcomeFailed, result.Outcome)
	assert.Equal(t, int64(99), result.ID)
	assert.Equal(t, "HTTP 503", result.Detail)

	_, statErr := os.Stat(replayFilePath(dir, 99))
	assert.True(t, os.IsNotExist(statErr))
}

func TestDownloadReplayPropagatesWriteError(t *testing.T) {
	dir := makeReplayTestDir(t)
	cfg := replayBatchConfig{OutDir: filepath.Join(dir, "missing")}
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{1: []byte(`{}`)},
	}

	_, err := downloadReplay(fetcher, cfg, 1, time.Now())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write")
}

func TestRunReplayDownloadsSkipsExistingByDefault(t *testing.T) {
	dir := makeReplayTestDir(t)
	require.NoError(t, os.WriteFile(replayFilePath(dir, 5), []byte("{}"), 0o644))

	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{
			5: []byte(`{}`),
			6: []byte(`{"gameResult":{"refereeInput":"seed=1"}}`),
		},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{5, 6},
		OutDir: dir,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Unix(1700000000, 0).UTC()))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, downloadOutcomeSkippedExisting, results[0].Outcome)
	assert.Equal(t, downloadOutcomeSaved, results[1].Outcome)
	assert.Equal(t, []int64{6}, fetcher.requested)
	assert.Contains(t, stdout.String(), "[1/2] skip 5 (exists)")
	assert.Contains(t, stdout.String(), "[2/2] save 6")
}

func TestRunReplayDownloadsForceRefetches(t *testing.T) {
	dir := makeReplayTestDir(t)
	require.NoError(t, os.WriteFile(replayFilePath(dir, 5), []byte("old"), 0o644))

	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{5: []byte(`{}`)},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{5},
		OutDir: dir,
		Force:  true,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Now()))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, downloadOutcomeSaved, results[0].Outcome)
	assert.Equal(t, []int64{5}, fetcher.requested)
}

func TestRunReplayDownloadsAppliesLimit(t *testing.T) {
	dir := makeReplayTestDir(t)
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{1: []byte(`{}`), 2: []byte(`{}`), 3: []byte(`{}`)},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{1, 2, 3},
		OutDir: dir,
		Limit:  2,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Now()))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, []int64{1, 2}, fetcher.requested)
}

func TestDownloadReplayClassifiesPendingError(t *testing.T) {
	dir := makeReplayTestDir(t)
	cfg := replayBatchConfig{OutDir: dir}
	fetcher := &fakeReplayFetcher{err: &codingame.APIError{
		StatusCode: 422,
		Body:       []byte(`{"code":"UNAUTHORIZED","message":"You are not authorised to view replay 1"}`),
	}}

	result, err := downloadReplay(fetcher, cfg, 1, time.Now())
	require.NoError(t, err)
	assert.Equal(t, downloadOutcomeSkippedPending, result.Outcome)
	assert.Equal(t, "pending: replay not yet published", result.Detail)

	_, statErr := os.Stat(replayFilePath(dir, 1))
	assert.True(t, os.IsNotExist(statErr))
}

func TestRunReplayDownloadsRetriesPendingAtEndOfBatch(t *testing.T) {
	dir := makeReplayTestDir(t)
	pendingErr := &codingame.APIError{
		StatusCode: 422,
		Body:       []byte(`{"code":"UNAUTHORIZED"}`),
	}
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{
			1: []byte(`{}`),
			2: []byte(`{"gameResult":{"refereeInput":"seed=1"}}`),
		},
		errForOnce: map[int64]error{2: pendingErr},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{1, 2},
		OutDir: dir,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Unix(1700000000, 0).UTC()))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, downloadOutcomeSaved, results[0].Outcome)
	assert.Equal(t, downloadOutcomeSaved, results[1].Outcome, "pending id should succeed on retry")
	// First-pass fetch order, then retry of 2 at the end.
	assert.Equal(t, []int64{1, 2, 2}, fetcher.requested)

	out := stdout.String()
	assert.Contains(t, out, "[2/2] skip 2 (pending: replay not yet published)")
	assert.Contains(t, out, "retry: 1 pending\n")
	assert.Contains(t, out, "[retry 1/1] save 2")

	// Retry must happen AFTER the main loop's last id-1 progress line.
	idxLast := strings.Index(out, "[2/2] skip 2")
	idxRetry := strings.Index(out, "[retry 1/1] save 2")
	require.Greater(t, idxRetry, idxLast)
}

func TestRunReplayDownloadsKeepsStillPendingAfterRetry(t *testing.T) {
	dir := makeReplayTestDir(t)
	pendingErr := &codingame.APIError{
		StatusCode: 422,
		Body:       []byte(`{"code":"UNAUTHORIZED"}`),
	}
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{1: []byte(`{}`)},
		errFor:    map[int64]error{2: pendingErr},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{1, 2},
		OutDir: dir,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Now()))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, downloadOutcomeSkippedPending, results[1].Outcome)
	assert.Equal(t, "still pending: replay not yet published", results[1].Detail)
	assert.Equal(t, []int64{1, 2, 2}, fetcher.requested)
	assert.Contains(t, stdout.String(), "[retry 1/1] skip 2 (still pending: replay not yet published)")
}

func TestRunReplayDownloadsSoftFailsOnFetchError(t *testing.T) {
	dir := makeReplayTestDir(t)
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{2: []byte(`{}`)},
		errFor:    map[int64]error{1: errors.New("HTTP 500")},
	}
	cfg := replayBatchConfig{
		IDs:    []int64{1, 2},
		OutDir: dir,
	}

	var stdout bytes.Buffer
	results, _, err := runReplayDownloads(fetcher, cfg, &stdout, fixedClock(time.Now()))
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, downloadOutcomeFailed, results[0].Outcome)
	assert.Equal(t, downloadOutcomeSaved, results[1].Outcome)
	assert.Contains(t, stdout.String(), "[1/2] fail 1: HTTP 500")
	assert.Contains(t, stdout.String(), "[2/2] save 2")
}

func TestSummarizeDownloadResultsAggregatesOutcomes(t *testing.T) {
	got := summarizeDownloadResults([]downloadResult{
		{Outcome: downloadOutcomeSaved},
		{Outcome: downloadOutcomeSaved},
		{Outcome: downloadOutcomeSkippedExisting},
		{Outcome: downloadOutcomeSkippedPuzzle},
		{Outcome: downloadOutcomeSkippedPending},
		{Outcome: downloadOutcomeFailed},
	})
	assert.Equal(t, downloadSummary{Total: 6, Saved: 2, SkippedExisting: 1, SkippedPuzzle: 1, SkippedPending: 1, Failed: 1}, got)
}

func TestWriteDownloadProgressFormatsByOutcome(t *testing.T) {
	cases := []struct {
		name string
		res  downloadResult
		want string
	}{
		{"saved", downloadResult{ID: 42, Outcome: downloadOutcomeSaved, Detail: "1024 bytes"}, "[1/3] save 42 (1024 bytes)\n"},
		{"skipped-existing", downloadResult{ID: 7, Outcome: downloadOutcomeSkippedExisting, Detail: "exists"}, "[1/3] skip 7 (exists)\n"},
		{"skipped-puzzle", downloadResult{ID: 8, Outcome: downloadOutcomeSkippedPuzzle, Detail: "puzzleId 99 != 42"}, "[1/3] skip 8 (puzzleId 99 != 42)\n"},
		{"skipped-pending", downloadResult{ID: 10, Outcome: downloadOutcomeSkippedPending, Detail: "pending: replay not yet published"}, "[1/3] skip 10 (pending: replay not yet published)\n"},
		{"failed", downloadResult{ID: 9, Outcome: downloadOutcomeFailed, Detail: "HTTP 500"}, "[1/3] fail 9: HTTP 500\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeDownloadProgress(&buf, 1, 3, tc.res)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestWriteDownloadSummaryFormat(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, writeDownloadSummary(&buf, "./tmp/replays", []downloadResult{
		{Outcome: downloadOutcomeSaved},
		{Outcome: downloadOutcomeSkippedExisting},
		{Outcome: downloadOutcomeSkippedPuzzle},
		{Outcome: downloadOutcomeSkippedPending},
		{Outcome: downloadOutcomeFailed},
	}))
	assert.Equal(t, "done: 1 saved, 1 skipped-existing, 1 skipped-puzzle, 1 skipped-pending, 1 failed (out=./tmp/replays)\n", buf.String())
}

func TestDownloadReplaySkipsCrossGamePuzzleID(t *testing.T) {
	dir := makeReplayTestDir(t)
	cfg := replayBatchConfig{
		OutDir: dir,
		Annotations: arena.ReplayAnnotations{
			Blue:     "us",
			PuzzleID: 42,
		},
	}
	// Source body declares a different puzzleId — must not be saved, since
	// silently rewriting the field would let cross-game replays slip into a
	// per-game replays/ directory under the wrong name.
	fetcher := &fakeReplayFetcher{
		responses: map[int64][]byte{77: []byte(`{"puzzleId":99,"gameResult":{"refereeInput":"seed=7"}}`)},
	}

	result, err := downloadReplay(fetcher, cfg, 77, time.Unix(1700000000, 0).UTC())
	require.NoError(t, err)
	assert.Equal(t, downloadOutcomeSkippedPuzzle, result.Outcome)
	assert.Equal(t, "puzzleId 99 != 42", result.Detail)

	_, statErr := os.Stat(replayFilePath(dir, 77))
	assert.True(t, os.IsNotExist(statErr))
}

func makeReplayTestDir(t *testing.T) string {
	t.Helper()
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dir := filepath.Join("tmp", "replay-test-"+name)
	require.NoError(t, os.RemoveAll(dir))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

type fakeReplayFetcher struct {
	responses  map[int64][]byte
	errFor     map[int64]error
	errForOnce map[int64]error
	err        error
	requested  []int64
}

func (f *fakeReplayFetcher) FetchReplay(gameID int64) ([]byte, error) {
	f.requested = append(f.requested, gameID)
	if err, ok := f.errForOnce[gameID]; ok {
		delete(f.errForOnce, gameID)
		return nil, err
	}
	if err, ok := f.errFor[gameID]; ok {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	body, ok := f.responses[gameID]
	if !ok {
		return nil, fmt.Errorf("no fake response for %d", gameID)
	}
	return body, nil
}

var _ replayFetcher = (*fakeReplayFetcher)(nil)
