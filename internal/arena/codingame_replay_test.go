package arena

import (
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
		"  \"puzzleId\": 1,\n" +
		"  \"puzzleTitle\": [\n" +
		"    \"Winter\"\n" +
		"  ],\n" +
		"  \"questionTitle\": \"Winter Challenge\",\n" +
		"  \"gameResult\": {\n" +
		"    \"frames\": [\n" +
		"      {\n" +
		"        \"agentId\": 0,\n" +
		"        \"stdout\": \"MOVE\"\n" +
		"      }\n" +
		"    ],\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
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
		"  \"seed\": \"6978185030065794000\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
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
		"  \"puzzleId\": 1,\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
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
		"  \"league\": 4,\n" +
		"  \"puzzleId\": 1,\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
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
		"  \"puzzleId\": 1,\n" +
		"  \"source\": \"get\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_OverridesPuzzleIDAndTitle(t *testing.T) {
	t.Parallel()

	// API returned puzzleId=0 with no puzzleTitle (the actual shape we see for
	// games like 882653023). Annotations carry the canonical values from the
	// factory; PrepareReplay must layer them on top, writing puzzleTitle as a
	// plain string (CG's array form is unnecessary for our consumers).
	body := []byte(`{"puzzleId":0,"questionTitle":"TestGame level4","gameResult":{"gameId":42}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{
		PuzzleID:    13771,
		PuzzleTitle: "TestGame - Winter Challenge 2026",
	})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"puzzleId\": 13771,\n" +
		"  \"puzzleTitle\": \"TestGame - Winter Challenge 2026\",\n" +
		"  \"questionTitle\": \"TestGame level4\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPrepareReplay_PreservesNonZeroSourcePuzzleID(t *testing.T) {
	t.Parallel()

	// Source declares its own real puzzleId. The annotations carry the
	// factory's expected ID, but PrepareReplay must NOT silently rewrite
	// the source value — keeping it intact lets the caller (downloadReplay)
	// detect the cross-game mismatch and reject the file before save.
	body := []byte(`{"puzzleId":99999,"gameResult":{"gameId":42}}`)

	got, err := PrepareReplay(body, ReplayAnnotations{
		PuzzleID:    13771,
		PuzzleTitle: "TestGame - Winter Challenge 2026",
	})
	if err != nil {
		t.Fatalf("PrepareReplay() error = %v", err)
	}

	want := "{\n" +
		"  \"puzzleId\": 99999,\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
		"}\n"
	if string(got) != want {
		t.Fatalf("PrepareReplay() mismatch\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestPeekReplayPuzzleID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		body    string
		want    int
		wantOK  bool
	}{
		{"present non-zero", `{"puzzleId":42,"gameResult":{}}`, 42, true},
		{"present zero", `{"puzzleId":0,"gameResult":{}}`, 0, true},
		{"missing", `{"gameResult":{}}`, 0, false},
		{"unparseable", `{"puzzleId":"forty-two"}`, 0, false},
		{"malformed", `not json`, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := PeekReplayPuzzleID([]byte(tc.body))
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("PeekReplayPuzzleID(%q) = (%d, %v), want (%d, %v)",
					tc.body, got, ok, tc.want, tc.wantOK)
			}
		})
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
		"  \"leaderboard\": {\n" +
		"    \"rank\": 210,\n" +
		"    \"division\": 3,\n" +
		"    \"score\": 18.95\n" +
		"  },\n" +
		"  \"puzzleId\": 1,\n" +
		"  \"source\": \"leaderboard\",\n" +
		"  \"gameResult\": {\n" +
		"    \"gameId\": 42\n" +
		"  }\n" +
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

	if got := ParseReplayLeague("TestGame level4"); got != 4 {
		t.Fatalf("ParseReplayLeague() = %d, want 4", got)
	}
	if got := ParseReplayLeague("Spring Challenge 2021 - Level 4"); got != 4 {
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
		Left:  []string{"MOVE A\n"},
		Right: []string{"MOVE B\n", "MOVE C\n"},
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
	// by the engine and does not count. emitsPostEndFrame=true (Spring 2020):
	// the trailing empty stdout always counts as the gameOverFrame.
	if got := (PostEndTurnModel{}).ExpectedTraceTurnCount(replay); got != 3 {
		t.Fatalf("PostEndTurnModel.ExpectedTraceTurnCount() = %d, want 3", got)
	}
}

func TestReplayTraceTurnCount_DeactivationPairFrame(t *testing.T) {
	t.Parallel()

	// Last turn pairs P0's normal stdout with P1's empty stdout (timeout). The
	// empty frame closes turn 2 — it is NOT a separate game-over marker, so
	// the count must stay at 2 instead of being incremented to 3.
	// emitsPostEndFrame=false (Winter 2026 family): no separate gameOverFrame.
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

	if got := (FlatTurnModel{}).ExpectedTraceTurnCount(replay); got != 2 {
		t.Fatalf("FlatTurnModel.ExpectedTraceTurnCount() = %d, want 2", got)
	}
}

// TestReplayTraceTurnCount_PostEndDeactivationPair covers the spring2020
// scenario where the timing-out player's empty stdout opens the final main
// turn (paired with the surviving side's stdout to close it), and the
// engine then emits a separate post-end gameOverFrame trace turn — observed
// in replay 875143793 before the emitsPostEndFrame flag existed. With the
// flag, the trailing empty must count as the gameOverFrame even though it
// looks like a deactivation pair close.
func TestReplayTraceTurnCount_PostEndDeactivationPair(t *testing.T) {
	t.Parallel()

	replay := CodinGameReplay[CodinGameReplayFrame]{
		GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
			Frames: []CodinGameReplayFrame{
				{AgentID: -1, Summary: "init"},
				{AgentID: 0, Stdout: "SPEED 0\n"},
				{AgentID: 1, Stdout: "SPEED 0\n"},
				{AgentID: 0, Stdout: "MOVE 0 1 1\n"},
				{AgentID: 1, Stdout: "MOVE 0 2 2\n"},
				{AgentID: 0, Summary: "speed sub-turn"},
				{AgentID: 0, Summary: "P0 polled, timeout"},
				{AgentID: 1, Stdout: "MOVE 0 3 3\n"},
				{AgentID: 0, Summary: "game over"},
			},
		},
	}

	// 3 main turns (SPEED, MOVE, deactivation MOVE) + 1 gameOverFrame.
	if got := (PostEndTurnModel{}).ExpectedTraceTurnCount(replay); got != 4 {
		t.Fatalf("PostEndTurnModel.ExpectedTraceTurnCount() = %d, want 4", got)
	}
}

// TestReplayTraceTurnCount_PhaseFrames covers spring2021's phase-frame model:
// each round emits a GATHERING empty-stdout frame, N ACTIONS pairs, and a
// SUN_MOVE empty-stdout frame. Each empty-stdout frame is its own trace turn
// and also flushes any pending single-stdout pair (the round-4 pattern in
// real replay 885203502 where one player WAITs a turn before its partner).
func TestReplayTraceTurnCount_PhaseFrames(t *testing.T) {
	t.Parallel()

	t.Run("single round: gather + WAIT pair + sun_move", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Frames: []CodinGameReplayFrame{
					{AgentID: -1, Summary: "init"},
					{AgentID: 0, Summary: "Round 0/23 collected"},
					{AgentID: 0, Stdout: "WAIT\n"},
					{AgentID: 1, Stdout: "WAIT\n", Summary: "$0 is waiting\n$1 is waiting"},
					{AgentID: 0, Summary: "Round 0 ends"},
				},
			},
		}
		// 1 GATHERING + 1 ACTIONS + 1 SUN_MOVE = 3.
		if got := (PhaseTurnModel{}).ExpectedTraceTurnCount(replay); got != 3 {
			t.Fatalf("PhaseTurnModel.ExpectedTraceTurnCount() = %d, want 3", got)
		}
	})

	t.Run("two rounds: gather + GROW pair + WAIT pair + sun_move", func(t *testing.T) {
		t.Parallel()
		round := []CodinGameReplayFrame{
			{AgentID: 0, Summary: "Round X collected"},
			{AgentID: 0, Stdout: "GROW 1\n"},
			{AgentID: 1, Stdout: "GROW 2\n", Summary: "growing"},
			{AgentID: 0, Stdout: "WAIT\n"},
			{AgentID: 1, Stdout: "WAIT\n", Summary: "waiting"},
			{AgentID: 0, Summary: "Round X ends"},
		}
		frames := []CodinGameReplayFrame{{AgentID: -1, Summary: "init"}}
		frames = append(frames, round...)
		frames = append(frames, round...)
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{Frames: frames},
		}
		// Per round: 1 GATHERING + 2 ACTIONS + 1 SUN_MOVE = 4. Two rounds = 8.
		if got := (PhaseTurnModel{}).ExpectedTraceTurnCount(replay); got != 8 {
			t.Fatalf("PhaseTurnModel.ExpectedTraceTurnCount() = %d, want 8", got)
		}
	})

	t.Run("solo WAIT closed by sun_move", func(t *testing.T) {
		t.Parallel()
		// Round 4 of replay 885203502: agent=1 declares WAIT mid-round, then
		// agent=0 acts alone for one ACTIONS turn before declaring WAIT itself
		// (closed by the SUN_MOVE empty frame). Sequence: gather, pair, pair,
		// pair (incl. agent=1 WAIT), agent=0 WAIT alone, sun_move.
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Frames: []CodinGameReplayFrame{
					{AgentID: -1, Summary: "init"},
					{AgentID: 0, Summary: "Round 4 collected"},
					{AgentID: 0, Stdout: "GROW 5\n"},
					{AgentID: 1, Stdout: "GROW 22\n", Summary: "growing"},
					{AgentID: 0, Stdout: "GROW 6\n"},
					{AgentID: 1, Stdout: "SEED 25 10\n", Summary: "growing"},
					{AgentID: 0, Stdout: "SEED 31 15\n"},
					{AgentID: 1, Stdout: "WAIT\n", Summary: "agent 1 waits"},
					{AgentID: 0, Stdout: "WAIT\n", Summary: "agent 0 waits"},
					{AgentID: 0, Summary: "Round 4 ends"},
				},
			},
		}
		// 1 GATHERING + 4 ACTIONS (3 pairs + 1 solo WAIT closed by sun_move) +
		// 1 SUN_MOVE = 6.
		if got := (PhaseTurnModel{}).ExpectedTraceTurnCount(replay); got != 6 {
			t.Fatalf("PhaseTurnModel.ExpectedTraceTurnCount() = %d, want 6", got)
		}
	})
}

// TestMainTurnCount confirms the universal MainTurnCount logic counts only
// pair-flush events: empty-stdout phase frames and trailing engine markers
// are excluded for every TurnModel.
func TestMainTurnCount(t *testing.T) {
	t.Parallel()

	t.Run("flat: 2 decision turns", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Frames: []CodinGameReplayFrame{
					{AgentID: -1, Summary: "init"},
					{AgentID: 0, Stdout: "MOVE A"},
					{AgentID: 1, Stdout: "MOVE B"},
					{AgentID: 0, Stdout: "MOVE C"},
					{AgentID: 1, Stdout: "MOVE D"},
				},
			},
		}
		want := 2
		if got := (FlatTurnModel{}).MainTurnCount(replay); got != want {
			t.Fatalf("Flat MainTurnCount = %d, want %d", got, want)
		}
		if got := (PostEndTurnModel{}).MainTurnCount(replay); got != want {
			t.Fatalf("PostEnd MainTurnCount = %d, want %d", got, want)
		}
		if got := (PhaseTurnModel{}).MainTurnCount(replay); got != want {
			t.Fatalf("Phase MainTurnCount = %d, want %d", got, want)
		}
	})

	t.Run("phase: 1 ACTIONS pair surrounded by gather + sun_move", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Frames: []CodinGameReplayFrame{
					{AgentID: -1, Summary: "init"},
					{AgentID: 0, Summary: "Round gather"},
					{AgentID: 0, Stdout: "WAIT"},
					{AgentID: 1, Stdout: "WAIT"},
					{AgentID: 0, Summary: "Round ends"},
				},
			},
		}
		// Only the WAIT pair counts as a main (decision) turn.
		if got := (PhaseTurnModel{}).MainTurnCount(replay); got != 1 {
			t.Fatalf("MainTurnCount = %d, want 1", got)
		}
	})

	t.Run("postend: trailing engine frame excluded", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Frames: []CodinGameReplayFrame{
					{AgentID: -1, Summary: "init"},
					{AgentID: 0, Stdout: "SPEED 0"},
					{AgentID: 1, Stdout: "SPEED 0"},
					{AgentID: 0, Summary: "game over"},
				},
			},
		}
		// 1 decision turn (SPEED pair); the trailing empty is the gameOverFrame.
		if got := (PostEndTurnModel{}).MainTurnCount(replay); got != 1 {
			t.Fatalf("MainTurnCount = %d, want 1", got)
		}
	})
}

func TestExtractReplayOutcome(t *testing.T) {
	t.Parallel()

	t.Run("p1 wins, no DQ", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Scores: []float64{122, 134},
				Ranks:  []int{1, 0},
			},
		}
		got, ok := ExtractReplayOutcome(replay)
		if !ok {
			t.Fatalf("ExtractReplayOutcome() ok=false")
		}
		want := ReplayOutcome{Winner: 1, Scores: [2]int{122, 134}, Deactivated: [2]bool{false, false}}
		if got != want {
			t.Fatalf("ExtractReplayOutcome() = %#v, want %#v", got, want)
		}
	})

	t.Run("p0 DQ, p1 wins", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Scores: []float64{-1, 50},
				Ranks:  []int{1, 0},
			},
		}
		got, _ := ExtractReplayOutcome(replay)
		want := ReplayOutcome{Winner: 1, Scores: [2]int{-1, 50}, Deactivated: [2]bool{true, false}}
		if got != want {
			t.Fatalf("ExtractReplayOutcome() = %#v, want %#v", got, want)
		}
	})

	t.Run("draw", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Scores: []float64{50, 50},
				Ranks:  []int{0, 0},
			},
		}
		got, _ := ExtractReplayOutcome(replay)
		want := ReplayOutcome{Winner: -1, Scores: [2]int{50, 50}, Deactivated: [2]bool{false, false}}
		if got != want {
			t.Fatalf("ExtractReplayOutcome() = %#v, want %#v", got, want)
		}
	})

	t.Run("malformed scores", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Scores: []float64{50},
				Ranks:  []int{0, 1},
			},
		}
		if _, ok := ExtractReplayOutcome(replay); ok {
			t.Fatalf("ExtractReplayOutcome() ok=true, want false")
		}
	})

	t.Run("malformed ranks", func(t *testing.T) {
		t.Parallel()
		replay := CodinGameReplay[CodinGameReplayFrame]{
			GameResult: CodinGameReplayResult[CodinGameReplayFrame]{
				Scores: []float64{50, 60},
				Ranks:  []int{0},
			},
		}
		if _, ok := ExtractReplayOutcome(replay); ok {
			t.Fatalf("ExtractReplayOutcome() ok=true, want false")
		}
	})
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
