package arena

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestPrepareReplay_StripsViewerOnlyFields(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"puzzleId":1,
		"questionTitle":"Winter Challenge",
		"puzzleTitle":["Winter"],
		"shareable":true,
		"viewer":{"frames":[]},
		"gameResult":{
			"gameId":42,
			"metadata":{"foo":"bar"},
			"tooltips":["t1"],
			"frames":[{"agentId":0,"view":"big-payload","gameInformation":"","keyframe":false,"stdout":"MOVE"}]
		}
	}`)

	got, err := PrepareReplay(body, ReplayAnnotations{})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"frames\": [\n" +
		"      {\n" +
		"        \"agentId\": 0,\n" +
		"        \"stdout\": \"MOVE\"\n" +
		"      }\n" +
		"    ],\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"puzzleId\": 1,\n" +
		"  \"puzzleTitle\": [\n" +
		"    \"Winter\"\n" +
		"  ],\n" +
		"  \"questionTitle\": \"Winter Challenge\"\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_PromotesSeedAndDropsRefereeInput(t *testing.T) {
	t.Parallel()

	body := []byte(`{"gameResult":{"gameId":42,"refereeInput":"seed=6978185030065794000\n"}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"seed\": \"6978185030065794000\"\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_KeepsRefereeInputWhenSeedMissing(t *testing.T) {
	t.Parallel()

	body := []byte(`{"gameResult":{"gameId":42,"refereeInput":"max-turns=200"}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42,\n" +
		"    \"refereeInput\": \"max-turns=200\"\n" +
		"  }\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_RemovesFrameViewOnly(t *testing.T) {
	t.Parallel()

	body := []byte(`{"gameResult":{"gameId":42,"frames":[{"agentId":0,"view":"big-payload","stdout":"MOVE"},{"agentId":1,"summary":"ok"},"raw"]}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"frames\": [\n" +
		"      {\n" +
		"        \"agentId\": 0,\n" +
		"        \"stdout\": \"MOVE\"\n" +
		"      },\n" +
		"      {\n" +
		"        \"agentId\": 1,\n" +
		"        \"summary\": \"ok\"\n" +
		"      },\n" +
		"      \"raw\"\n" +
		"    ],\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_PrettyPrintsWithoutViewer(t *testing.T) {
	t.Parallel()

	body := []byte("{\"puzzleId\":1,\"gameResult\":{\"gameId\":42}}")

	got, err := PrepareReplay(body, ReplayAnnotations{})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"puzzleId\": 1\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_AddsBlueAndLeagueFields(t *testing.T) {
	t.Parallel()

	body := []byte(`{"puzzleId":1,"gameResult":{"gameId":42}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{Blue: "mrsombre", League: 4})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"blue\": \"mrsombre\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"league\": 4,\n" +
		"  \"puzzleId\": 1\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_AddsSourceAndFetchedAt(t *testing.T) {
	t.Parallel()

	body := []byte(`{"puzzleId":1,"gameResult":{"gameId":42}}`)
	at := time.Date(2026, 4, 29, 11, 23, 45, 0, time.UTC)

	got, err := PrepareReplay(body, ReplayAnnotations{
		Source:    ReplaySourceGet,
		FetchedAt: at,
	})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"fetched_at\": \"2026-04-29T11:23:45Z\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"puzzleId\": 1,\n" +
		"  \"source\": \"get\"\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestRewriteReplayPuzzleID(t *testing.T) {
	t.Parallel()

	body := []byte("{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"puzzleId\": 0,\n" +
		"  \"questionTitle\": \"SnakeBot level4\"\n" +
		"}\n")
	path := filepath.Join(t.TempDir(), "42.json")
	if err := os.WriteFile(path, body, 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := RewriteReplayPuzzleID(path, 13771); err != nil {
		t.Fatalf("RewriteReplayPuzzleID() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"puzzleId\": 13771,\n" +
		"  \"questionTitle\": \"SnakeBot level4\"\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("RewriteReplayPuzzleID() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_AddsLeaderboardInfo(t *testing.T) {
	t.Parallel()

	body := []byte(`{"puzzleId":1,"gameResult":{"gameId":42}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{
		Source: ReplaySourceLeaderboard,
		Leaderboard: &ReplayLeaderboardInfo{
			Rank:     210,
			Division: 3,
			Score:    18.95,
		},
	})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  },\n" +
		"  \"leaderboard\": {\n" +
		"    \"rank\": 210,\n" +
		"    \"division\": 3,\n" +
		"    \"score\": 18.95\n" +
		"  },\n" +
		"  \"puzzleId\": 1,\n" +
		"  \"source\": \"leaderboard\"\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestParseReplaySeed(t *testing.T) {
	t.Parallel()

	got, ok := ParseReplaySeed("foo=1 seed=12345\nbar=2")
	if !ok {
		t.Fatal("ParseReplaySeed() ok = false, want true")
	}
	if got != 12345 {
		t.Fatalf("ParseReplaySeed() = %d, want 12345", got)
	}
}

func TestParseReplayLeague(t *testing.T) {
	t.Parallel()

	if got := ParseReplayLeague("SnakeBot level4"); got != 4 {
		t.Fatalf("ParseReplayLeague() = %d, want 4", got)
	}
	if got := ParseReplayLeague("Winter Challenge"); got != 0 {
		t.Fatalf("ParseReplayLeague() = %d, want 0", got)
	}
}

func TestReplayMovesFromFramesAndTurnCount(t *testing.T) {
	t.Parallel()

	replay := CodinGameReplay[CodinGameReplayFrame]{
		GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
			Frames: []CodinGameReplayFrame{
				{AgentID: -1, Summary: "boot"},
				{AgentID: 0, Stdout: "MOVE A\n"},
				{AgentID: 1, Stdout: "MOVE B\n"},
				{AgentID: 0, Stdout: ""},
				{AgentID: 1, Stdout: "MOVE C\n"},
			},
		},
	}

	moves := ReplayMovesFromFrames(replay)
	wantMoves := ReplayMoves{
		P0: []string{"MOVE A\n"},
		P1: []string{"MOVE B\n", "MOVE C\n"},
	}
	if !reflect.DeepEqual(moves, wantMoves) {
		t.Fatalf("ReplayMovesFromFrames() = %#v, want %#v", moves, wantMoves)
	}
	if got := ReplayTurnCount(replay); got != 2 {
		t.Fatalf("ReplayTurnCount() = %d, want 2", got)
	}
}

func TestReplayTraceTurnCount(t *testing.T) {
	t.Parallel()

	replay := CodinGameReplay[CodinGameReplayFrame]{
		GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
			Frames: []CodinGameReplayFrame{
				{AgentID: -1, Summary: "init"},
				{AgentID: 0, Stdout: "SPEED 0\n"},
				{AgentID: 1, Stdout: "SPEED 0\n"},
				{AgentID: 0, Summary: "speed sub-turn"},
				{AgentID: 0, Stdout: "MOVE 0 1 1\n"},
				{AgentID: 1, Stdout: "MOVE 0 2 2\n"},
				{AgentID: 0, Summary: "game over"},
			},
		},
	}

	// 2 main decision turns (SPEED, MOVE) + 1 trailing game-over frame.
	// The mid-replay speed sub-turn frame is folded into the SPEED main turn
	// by the engine and does not count.
	if got := ReplayTraceTurnCount(replay); got != 3 {
		t.Fatalf("ReplayTraceTurnCount() = %d, want 3", got)
	}
}

func TestReplayTraceTurnCount_DeactivationPairFrame(t *testing.T) {
	t.Parallel()

	// Last turn pairs P0's normal stdout with P1's empty stdout (timeout). The
	// empty frame closes turn 2 — it is NOT a separate game-over marker, so
	// the count must stay at 2 instead of being incremented to 3.
	replay := CodinGameReplay[CodinGameReplayFrame]{
		GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
			Frames: []CodinGameReplayFrame{
				{AgentID: -1, Summary: "init"},
				{AgentID: 0, Stdout: "0 UP"},
				{AgentID: 1, Stdout: "4 UP"},
				{AgentID: 0, Stdout: "0 LEFT"},
				{AgentID: 1, Stdout: "", Summary: "$1 has not provided 1 lines in time\n"},
			},
		},
	}

	if got := ReplayTraceTurnCount(replay); got != 2 {
		t.Fatalf("ReplayTraceTurnCount() = %d, want 2", got)
	}
}

func TestReplayPlayerNames(t *testing.T) {
	t.Parallel()

	replay := CodinGameReplay[CodinGameReplayFrame]{
		GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
			Agents: []CodinGameReplayAgent{
				{Index: 0, CodinGamer: CodinGameReplayUser{Pseudo: "Alpha"}},
				{Index: 1, CodinGamer: CodinGameReplayUser{Pseudo: "Beta"}},
			},
		},
	}

	got := ReplayPlayerNames(replay)
	want := [2]string{"Alpha", "Beta"}
	if got != want {
		t.Fatalf("ReplayPlayerNames() = %#v, want %#v", got, want)
	}
}
