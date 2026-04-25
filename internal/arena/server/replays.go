package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

type replayListEntry struct {
	ID      string    `json:"id"`
	Size    int64     `json:"size"`
	MTime   time.Time `json:"mtime"`
	P0Name  string    `json:"p0_name,omitempty"`
	P1Name  string    `json:"p1_name,omitempty"`
	League  int       `json:"league,omitempty"`
	ScoreP0 int       `json:"score_p0"`
	ScoreP1 int       `json:"score_p1"`
	// Winner: 0, 1, or -1 for draw. Derived from CodinGame ranks (rank 0 wins).
	Winner int `json:"winner"`
}

// replayResponse wraps the replayed TraceMatch with metadata needed by the
// viewer (currently the league the replay was parsed with). League drives
// map generation, so the client must pass it to /api/serialize to render
// the matching initial board.
type replayResponse struct {
	arena.TraceMatch
	League int `json:"league"`
}

var replayIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func handleReplayList(replayDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if replayDir == "" {
			writeJSON(w, http.StatusOK, []replayListEntry{})
			return
		}
		entries, err := os.ReadDir(replayDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeJSON(w, http.StatusOK, []replayListEntry{})
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		out := make([]replayListEntry, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			id := strings.TrimSuffix(entry.Name(), ".json")
			e := replayListEntry{ID: id, Size: info.Size(), MTime: info.ModTime(), Winner: -1}
			if data, err := os.ReadFile(filepath.Join(replayDir, entry.Name())); err == nil {
				var r arena.CodinGameReplay[arena.CodinGameReplayFrame]
				if err := json.Unmarshal(data, &r); err == nil {
					for _, a := range r.GameResult.Agents {
						switch a.Index {
						case 0:
							e.P0Name = a.CodinGamer.Pseudo
						case 1:
							e.P1Name = a.CodinGamer.Pseudo
						}
					}
					e.League = parseReplayLeague(r.QuestionTitle)
					if len(r.GameResult.Scores) > 0 {
						e.ScoreP0 = int(r.GameResult.Scores[0])
					}
					if len(r.GameResult.Scores) > 1 {
						e.ScoreP1 = int(r.GameResult.Scores[1])
					}
					// CodinGame ranks[i] is the finishing rank of agent i
					// (0 = winner). Ties would give equal ranks; treat as draw.
					if len(r.GameResult.Ranks) >= 2 {
						switch {
						case r.GameResult.Ranks[0] < r.GameResult.Ranks[1]:
							e.Winner = 0
						case r.GameResult.Ranks[1] < r.GameResult.Ranks[0]:
							e.Winner = 1
						default:
							e.Winner = -1
						}
					}
				}
			}
			out = append(out, e)
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func handleReplayGet(replayDir string, factory arena.GameFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if replayDir == "" {
			writeError(w, http.StatusNotFound, "no replay-dir configured")
			return
		}
		id := r.PathValue("id")
		if !replayIDPattern.MatchString(id) {
			writeError(w, http.StatusBadRequest, "invalid replay id")
			return
		}
		path := filepath.Join(replayDir, id+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeError(w, http.StatusNotFound, "replay not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var replay arena.CodinGameReplay[arena.CodinGameReplayFrame]
		if err := json.Unmarshal(data, &replay); err != nil {
			writeError(w, http.StatusBadRequest, "parse replay: "+err.Error())
			return
		}

		seed, ok := parseReplaySeed(replay.GameResult.RefereeInput)
		if !ok {
			writeError(w, http.StatusBadRequest, "replay missing seed in refereeInput")
			return
		}

		league := parseReplayLeague(replay.QuestionTitle)
		gameOptions := map[string]string{}
		if league > 0 {
			gameOptions["league"] = strconv.Itoa(league)
		}

		moves := extractReplayMoves(replay)

		names := [2]string{"p0", "p1"}
		for _, a := range replay.GameResult.Agents {
			switch a.Index {
			case 0:
				names[0] = a.CodinGamer.Pseudo
			case 1:
				names[1] = a.CodinGamer.Pseudo
			}
		}

		trace := arena.RunReplay(factory, seed, gameOptions, moves, names, 0)
		// Keep the original replay id so the client can use it as a stable key.
		if n, err := strconv.Atoi(id); err == nil {
			trace.MatchID = n
		}
		writeJSON(w, http.StatusOK, replayResponse{TraceMatch: trace, League: league})
	}
}

func parseReplaySeed(refereeInput string) (int64, bool) {
	for _, tok := range strings.Fields(refereeInput) {
		if strings.HasPrefix(tok, "seed=") {
			n, err := strconv.ParseInt(strings.TrimPrefix(tok, "seed="), 10, 64)
			if err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

var leaguePattern = regexp.MustCompile(`(?i)level(\d)`)

func parseReplayLeague(questionTitle string) int {
	m := leaguePattern.FindStringSubmatch(questionTitle)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

func extractReplayMoves(replay arena.CodinGameReplay[arena.CodinGameReplayFrame]) arena.ReplayMoves {
	var moves arena.ReplayMoves
	for _, f := range replay.GameResult.Frames {
		if strings.TrimSpace(f.Stdout) == "" {
			continue
		}
		switch f.AgentID {
		case 0:
			moves.P0 = append(moves.P0, f.Stdout)
		case 1:
			moves.P1 = append(moves.P1, f.Stdout)
		}
	}
	return moves
}
