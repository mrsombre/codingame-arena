package arena

import (
	"reflect"
	"testing"
)

func TestPrepareReplay_PrettyPrintsAndRemovesViewer(t *testing.T) {
	t.Parallel()

	body := []byte(`{"puzzleId":1,"questionTitle":"Winter Challenge","puzzleTitle":["Winter"],"gameResult":{"gameId":42,"frames":[{"agentId":0,"view":"big-payload","stdout":"MOVE"}]},"viewer":{"frames":[]}}`)

	got, err := PrepareReplay(body, "", 0)
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

func TestPrepareReplay_RemovesFrameViewOnly(t *testing.T) {
	t.Parallel()

	body := []byte(`{"gameResult":{"gameId":42,"frames":[{"agentId":0,"view":"big-payload","stdout":"MOVE"},{"agentId":1,"summary":"ok"},"raw"]}}`)

	got, err := PrepareReplay(body, "", 0)
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

	got, err := PrepareReplay(body, "", 0)
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

func TestPrepareReplay_AddsBlueField(t *testing.T) {
	t.Parallel()

	body := []byte(`{"puzzleId":1,"gameResult":{"gameId":42}}`)

	got, err := PrepareReplay(body, "mrsombre", 0)
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"blue\": \"mrsombre\",\n" +
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

	got, err := PrepareReplay(body, "mrsombre", 4)
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
