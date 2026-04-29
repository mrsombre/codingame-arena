package commands

import (
	"bytes"
	"fmt"
	"io"
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
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{
  "gameId": "spring2020",
  "seed": "1",
  "scores": [1.0, 0.0],
  "ranks": [0, 1],
  "turns": [{"turn": 0, "p0_output": "WAIT", "p1_output": "WAIT"}]
}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "trace-1-0.json", files[0].Name)
	assert.Equal(t, "spring2020", files[0].Trace.GameID)
}

func TestResolveAnalyzeGameInfersSingleGame(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{
  "gameId": "spring2020",
  "seed": "1",
  "scores": [1.0, 0.0],
  "ranks": [0, 1],
  "turns": [{"turn": 0}]
}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)

	got, err := resolveAnalyzeGame("", files)
	require.NoError(t, err)
	assert.Equal(t, "spring2020", got)
}

func TestResolveAnalyzeGameRequiresExplicitGameForMixedTraces(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	writeAnalyzeTestFile(t, traceDir, "trace-1-0.json", `{"gameId": "spring2020", "turns": [{"turn": 0}]}`)
	writeAnalyzeTestFile(t, traceDir, "trace-2-0.json", `{"gameId": "winter2026", "turns": [{"turn": 0}]}`)

	files, err := loadAnalyzeTraceFiles(traceDir)
	require.NoError(t, err)

	_, err = resolveAnalyzeGame("", files)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple games")

	got, err := resolveAnalyzeGame("winter2026", files)
	require.NoError(t, err)
	assert.Equal(t, "winter2026", got)
}

func TestAnalyzeUsesGameFlagToFilterTraceFiles(t *testing.T) {
	traceDir := makeAnalyzeTestDir(t)
	gameA := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_a"
	gameB := "analyze_test_" + strings.NewReplacer("/", "_").Replace(t.Name()) + "_b"
	writeAnalyzeTestFile(t, traceDir, "trace-a.json", fmt.Sprintf(`{"gameId": %q, "turns": [{"turn": 0}]}`, gameA))
	writeAnalyzeTestFile(t, traceDir, "trace-b.json", fmt.Sprintf(`{"gameId": %q, "turns": [{"turn": 0}]}`, gameB))

	factory := &recordingAnalyzeFactory{name: gameA}
	arena.Register(factory)

	fs, v := newTestAnalyzeCtx(t)
	var out bytes.Buffer
	err := Analyze([]string{"--game", gameA, "--trace-dir", traceDir}, &out, nil, fs, v)
	require.NoError(t, err)

	assert.Equal(t, []string{"trace-a.json"}, factory.files)
	assert.Equal(t, "analyzed 1 files\n", out.String())
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

func (f *recordingAnalyzeFactory) NewGame(_ int64, _ *viper.Viper) (arena.Referee, []arena.Player) {
	panic("not used")
}

func (f *recordingAnalyzeFactory) MaxTurns() int { return 1 }

func (f *recordingAnalyzeFactory) AnalyzeTraces(input arena.TraceAnalysisInput) (arena.TraceAnalysis, error) {
	f.files = f.files[:0]
	for _, file := range input.Files {
		f.files = append(f.files, file.Name)
	}
	return recordingAnalysis{files: len(input.Files)}, nil
}

type recordingAnalysis struct {
	files int
}

func (r recordingAnalysis) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "analyzed %d files\n", r.files)
	return err
}

var _ arena.GameFactory = (*recordingAnalyzeFactory)(nil)
var _ arena.TraceAnalyzer = (*recordingAnalyzeFactory)(nil)
var _ arena.TraceAnalysis = recordingAnalysis{}
