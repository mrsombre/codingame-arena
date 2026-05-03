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

func TestLoadAnalyzeTraceFilesSkipsNonTraceJSON(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "note.json", `{"hello": "world"}`)
	writeAnalyzeTestFile(t, traceDir, "trace-2-0.json", `{
  "gameId": "test-game",
  "seed": "2",
  "blue": "us",
  "players": ["us", "rival"],
  "scores": [1.0, 0.0],
  "ranks": [0, 1],
  "turns": [{"turn": 0, "output": ["WAIT", "WAIT"]}]
}`)
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{
  "gameId": "test-game",
  "seed": "1",
  "blue": "us",
  "players": ["us", "rival"],
  "scores": [1.0, 0.0],
  "ranks": [0, 1],
  "turns": [{"turn": 0, "output": ["WAIT", "WAIT"]}]
}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "trace-1-0.json", files[0].Name)
	assert.Equal(t, "trace-2-0.json", files[1].Name)
	assert.Equal(t, "test-game", files[0].Trace.GameID)
}

func TestLoadAnalyzeTraceFilesRequiresBlueSide(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-missing-blue.json", `{"gameId": "test-game", "players": ["us", "rival"], "turns": [{"turn": 0}]}`)

	_, err := loadAnalyzeTraceFiles(traceDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "trace missing blue")
}

func TestLoadAnalyzeTraceFilesRejectsUnknownBlueSide(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-bad-blue.json", `{"gameId": "test-game", "blue": "missing", "players": ["us", "rival"], "turns": [{"turn": 0}]}`)

	_, err := loadAnalyzeTraceFiles(traceDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `blue "missing" not found`)
}

func TestResolveAnalyzeGameInfersSingleGame(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{
  "gameId": "test-game",
  "seed": "1",
  "blue": "us",
  "players": ["us", "rival"],
  "scores": [1.0, 0.0],
  "ranks": [0, 1],
  "turns": [{"turn": 0}]
}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)

	got, err := resolveAnalyzeGame("", files)
	require.NoError(t, err)
	assert.Equal(t, "test-game", got)
}

func TestResolveAnalyzeGameRequiresExplicitGameForMixedTraces(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{"gameId": "test-game-a", "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`)
	writeAnalyzeTestFile(t, traceDir, "trace-2-0.json", `{"gameId": "test-game-b", "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)

	_, err = resolveAnalyzeGame("", files)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple games")

	got, err := resolveAnalyzeGame("test-game-b", files)
	require.NoError(t, err)
	assert.Equal(t, "test-game-b", got)
}

func TestAnalyzeUsesGameFlagToFilterTraceFiles(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	gameA := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_a"
	gameB := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_b"
	writeAnalyzeTestFile(t, traceDir, "trace-a.json", fmt.Sprintf(`{"type": "trace-a", "gameId": %q, "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`, gameA))
	writeAnalyzeTestFile(t, traceDir, "trace-b.json", fmt.Sprintf(`{"type": "trace-b", "gameId": %q, "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`, gameB))

	factory := &recordingAnalyzeFactory{name: gameA}
	arena.Register(factory)

	fs, v := newTestAnalyzeCtx(t)
	var out bytes.Buffer
	err := Analyze([]string{"--game", gameA, "--trace-dir", traceDir}, &out, nil, fs, v)
	require.NoError(t, err)

	assert.Equal(t, []string{"trace-a.json"}, factory.files)
	assert.Contains(t, out.String(), gameA+" — 1 traces")
}

func TestBuildTraceAnalysisInputFiltersSelectedGame(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	gameA := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_a"
	gameB := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_b"
	writeAnalyzeTestFile(t, traceDir, "trace-a.json", fmt.Sprintf(`{"type": "trace-a", "gameId": %q, "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`, gameA))
	writeAnalyzeTestFile(t, traceDir, "trace-b.json", fmt.Sprintf(`{"type": "trace-b", "gameId": %q, "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`, gameB))
	factory := &recordingAnalyzeFactory{name: gameB}
	arena.Register(factory)

	input, analyzer, err := buildTraceAnalysisInput(traceDir, gameB)

	require.NoError(t, err)
	require.NotNil(t, analyzer)
	assert.Equal(t, gameB, input.GameID)
	require.Len(t, input.Files, 1)
	assert.Equal(t, "trace-b.json", input.Files[0].Name)
	assert.Equal(t, traceDir, input.TraceDir)
}

func TestBuildTraceAnalysisInputRejectsUnknownGame(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace.json", `{"gameId": "missing-game", "blue": "us", "players": ["us", "rival"], "turns": [{"turn": 0}]}`)

	_, _, err := buildTraceAnalysisInput(traceDir, "missing-game")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown game "missing-game"`)
}

func makeAnalyzeTestDir(t *testing.T) string {
	t.Helper()
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dir := filepath.Join("tmp", "analyze-test-"+name)
	require.NoError(t, os.RemoveAll(dir))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func writeAnalyzeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

type recordingAnalyzeFactory struct {
	name  string
	files []string
}

func (f *recordingAnalyzeFactory) Name() string { return f.name }

func (f *recordingAnalyzeFactory) PuzzleID() int { return 0 }

func (f *recordingAnalyzeFactory) PuzzleTitle() string { return "" }

func (f *recordingAnalyzeFactory) LeaderboardSlug() string { return "" }

func (f *recordingAnalyzeFactory) NewGame(_ int64, _ *viper.Viper) (arena.Referee, []arena.Player) {
	panic("not used")
}

func (f *recordingAnalyzeFactory) MaxTurns() int { return 1 }

func (f *recordingAnalyzeFactory) TraceMetricSpecs() []arena.TraceMetricSpec {
	return []arena.TraceMetricSpec{{Key: "seen", Kind: arena.TraceMetricPerMatchCount, ShowZero: true}}
}

func (f *recordingAnalyzeFactory) AnalyzeTraceMetrics(trace arena.TraceMatch) (arena.TraceMetricStats, error) {
	f.files = append(f.files, trace.Type+".json")
	return arena.TraceMetricStats{"seen": [2]int{}}, nil
}

var _ arena.GameFactory = (*recordingAnalyzeFactory)(nil)
var _ arena.TraceMetricAnalyzer = (*recordingAnalyzeFactory)(nil)
