// Package server wires the `arena front` HTTP API and static asset handler.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Options configures a Handler built by New.
type Options struct {
	Factory  arena.GameFactory
	Assets   fs.FS
	TraceDir string
	Bots     []string
}

// New returns an http.Handler that serves the embedded viewer bundle and the
// JSON API consumed by the viewer.
func New(opts Options) http.Handler {
	mux := http.NewServeMux()

	gameRoot := "/" + opts.Factory.Name() + "/"
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, gameRoot, http.StatusFound)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /api/games", handleGames())
	mux.HandleFunc("GET /api/game", handleGame(opts.Factory))
	mux.HandleFunc("GET /api/serialize", handleSerialize(opts.Factory))
	mux.HandleFunc("GET /api/bots", handleBots(opts.Bots))
	mux.HandleFunc("GET /api/matches", handleMatchList(opts.TraceDir))
	mux.HandleFunc("GET /api/matches/{id}", handleMatchGet(opts.TraceDir))
	mux.HandleFunc("POST /api/run", handleRun(opts.Factory, opts.TraceDir))
	mux.HandleFunc("POST /api/batch", handleBatch(opts.Factory, opts.TraceDir))

	mux.Handle("/", http.FileServer(http.FS(opts.Assets)))
	return mux
}

type gameInfo struct {
	Name     string `json:"name"`
	MaxTurns int    `json:"maxTurns"`
}

func handleGames() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, arena.Games())
	}
}

func handleGame(factory arena.GameFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, gameInfo{
			Name:     factory.Name(),
			MaxTurns: factory.MaxTurns(),
		})
	}
}

func handleBots(bots []string) http.HandlerFunc {
	// Extract just the filenames for the response.
	type botEntry struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	entries := make([]botEntry, 0, len(bots))
	for _, b := range bots {
		entries = append(entries, botEntry{
			Name: filepath.Base(b),
			Path: b,
		})
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, entries)
	}
}

// handleSerialize mirrors the `arena serialize` command: globals + first frame
// for a player, as plain text. Query params: seed (required), player (0|1,
// default 0). Extra query params are forwarded as game options.
func handleSerialize(factory arena.GameFactory) http.HandlerFunc {
	reserved := map[string]struct{}{"seed": {}, "player": {}}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		seedRaw := q.Get("seed")
		if seedRaw == "" {
			writeError(w, http.StatusBadRequest, "seed query param is required")
			return
		}
		seed, err := arena.ParseSeed(seedRaw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid seed: "+err.Error())
			return
		}

		playerIdx := 0
		if raw := q.Get("player"); raw != "" {
			p, err := strconv.Atoi(raw)
			if err != nil || (p != 0 && p != 1) {
				writeError(w, http.StatusBadRequest, "player must be 0 or 1")
				return
			}
			playerIdx = p
		}

		gameOptions := map[string]string{}
		for k, vs := range q {
			if _, skip := reserved[k]; skip {
				continue
			}
			if len(vs) > 0 {
				gameOptions[k] = vs[0]
			}
		}

		referee, players := factory.NewGame(seed, gameOptions)
		referee.Init(players)
		player := players[playerIdx]

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		for _, line := range referee.GlobalInfoFor(player) {
			fmt.Fprintln(w, line)
		}
		referee.ResetGameTurnData()
		for _, line := range referee.FrameInfoFor(player) {
			fmt.Fprintln(w, line)
		}
	}
}

type matchListEntry struct {
	ID    string    `json:"id"`
	Size  int64     `json:"size"`
	MTime time.Time `json:"mtime"`
}

func handleMatchList(traceDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if traceDir == "" {
			writeJSON(w, http.StatusOK, []matchListEntry{})
			return
		}
		entries, err := os.ReadDir(traceDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeJSON(w, http.StatusOK, []matchListEntry{})
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		out := make([]matchListEntry, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			out = append(out, matchListEntry{
				ID:    strings.TrimSuffix(entry.Name(), ".json"),
				Size:  info.Size(),
				MTime: info.ModTime(),
			})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

var matchIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func handleMatchGet(traceDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if traceDir == "" {
			writeError(w, http.StatusNotFound, "no trace-dir configured; start arena front with --trace-dir <path>")
			return
		}
		id := r.PathValue("id")
		if !matchIDPattern.MatchString(id) {
			writeError(w, http.StatusBadRequest, "invalid match id")
			return
		}
		path := filepath.Join(traceDir, id+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeError(w, http.StatusNotFound, "match not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

type runRequest struct {
	P0Bin       string            `json:"p0Bin"`
	P1Bin       string            `json:"p1Bin"`
	Seed        *int64            `json:"seed,string,omitempty"`
	MaxTurns    int               `json:"maxTurns,omitempty"`
	NoSwap      bool              `json:"noSwap,omitempty"`
	GameOptions map[string]string `json:"gameOptions,omitempty"`
}

func handleRun(factory arena.GameFactory, traceDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req runRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.P0Bin == "" || req.P1Bin == "" {
			writeError(w, http.StatusBadRequest, "p0Bin and p1Bin are required")
			return
		}
		seed := time.Now().UnixNano()
		if req.Seed != nil {
			seed = *req.Seed
		}
		runner := arena.NewRunner(factory, arena.MatchOptions{
			MaxTurns:    req.MaxTurns,
			P0Bin:       req.P0Bin,
			P1Bin:       req.P1Bin,
			NoSwap:      req.NoSwap,
			TraceWriter: arena.NewTraceWriter(traceDir),
			GameOptions: req.GameOptions,
		})
		result := runner.RunMatch(0, seed)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(result.RenderMatch()))
	}
}

type batchRequest struct {
	P0Bin       string            `json:"p0Bin"`
	P1Bin       string            `json:"p1Bin"`
	Seed        *int64            `json:"seed,string,omitempty"`
	Simulations int               `json:"simulations,omitempty"`
	MaxTurns    int               `json:"maxTurns,omitempty"`
	NoSwap      bool              `json:"noSwap,omitempty"`
	GameOptions map[string]string `json:"gameOptions,omitempty"`
}

// batchMatchSummary describes one match from the batch. Fields with a "p0"/"p1"
// suffix are relative to the in-match side (left side of the map is p0); under
// random swap that may be the user's P1 bot. Aggregate counters at the top of
// batchResponse stay on the user-selected bot perspective.
type batchMatchSummary struct {
	ID      int    `json:"id"`
	Seed    int64  `json:"seed,string"`
	Winner  int    `json:"winner"`
	ScoreP0 int    `json:"score_p0"`
	ScoreP1 int    `json:"score_p1"`
	Turns   int    `json:"turns"`
	P0Bot   string `json:"p0_bot"`
	P1Bot   string `json:"p1_bot"`
}

type batchResponse struct {
	Simulations int                 `json:"simulations"`
	WinsP0      int                 `json:"wins_p0"`
	WinsP1      int                 `json:"wins_p1"`
	Draws       int                 `json:"draws"`
	AvgScoreP0  float64             `json:"avg_score_p0"`
	AvgScoreP1  float64             `json:"avg_score_p1"`
	AvgTurns    float64             `json:"avg_turns"`
	Seed        int64               `json:"seed,string"`
	P0Bot       string              `json:"p0_bot"`
	P1Bot       string              `json:"p1_bot"`
	Matches     []batchMatchSummary `json:"matches"`
}

func handleBatch(factory arena.GameFactory, traceDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req batchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.P0Bin == "" || req.P1Bin == "" {
			writeError(w, http.StatusBadRequest, "p0Bin and p1Bin are required")
			return
		}
		sims := req.Simulations
		if sims <= 0 {
			sims = 50
		}
		if sims > 500 {
			writeError(w, http.StatusBadRequest, "simulations cannot exceed 500")
			return
		}
		seed := time.Now().UnixNano()
		if req.Seed != nil {
			seed = *req.Seed
		}
		if traceDir == "" {
			writeError(w, http.StatusBadRequest, "no trace-dir configured; start arena front with --trace-dir <path>")
			return
		}
		if err := cleanupTraceDir(traceDir); err != nil {
			writeError(w, http.StatusInternalServerError, "cleanup trace dir: "+err.Error())
			return
		}
		runner := arena.NewRunner(factory, arena.MatchOptions{
			MaxTurns:    req.MaxTurns,
			P0Bin:       req.P0Bin,
			P1Bin:       req.P1Bin,
			NoSwap:      req.NoSwap,
			TraceWriter: arena.NewTraceWriter(traceDir),
			GameOptions: req.GameOptions,
		})
		parallel := runtime.NumCPU()
		if parallel > 4 {
			parallel = 4
		}
		results := arena.RunMatches(arena.BatchOptions{
			Simulations: sims,
			Parallel:    parallel,
			Seed:        seed,
		}, runner.RunMatch)

		userP0 := filepath.Base(req.P0Bin)
		userP1 := filepath.Base(req.P1Bin)
		resp := batchResponse{
			Simulations: len(results),
			Seed:        seed,
			P0Bot:       userP0,
			P1Bot:       userP1,
			Matches:     make([]batchMatchSummary, 0, len(results)),
		}
		var totalScoreP0, totalScoreP1, totalTurns float64
		for _, res := range results {
			// Prefer raw scores (sum of alive bird segments) over the
			// referee's tie-break-adjusted Scores so displayed values can't
			// go negative and match what the viewer computes from bird bodies.
			userScores := res.Scores
			userWinner := res.Winner
			if res.HaveRawScores {
				userScores = res.RawScores
				switch {
				case userScores[0] > userScores[1]:
					userWinner = 0
				case userScores[1] > userScores[0]:
					userWinner = 1
				default:
					userWinner = -1
				}
			}

			// Aggregate counters follow user-selected bots.
			switch userWinner {
			case 0:
				resp.WinsP0++
			case 1:
				resp.WinsP1++
			default:
				resp.Draws++
			}
			totalScoreP0 += float64(userScores[0])
			totalScoreP1 += float64(userScores[1])
			totalTurns += float64(res.Turns)

			// Per-match row reports from the in-match side perspective so the
			// replay map (blue = p0, red = p1) lines up with the numbers.
			matchScoreP0, matchScoreP1 := userScores[0], userScores[1]
			matchWinner := userWinner
			matchP0Bot, matchP1Bot := userP0, userP1
			if res.Swapped {
				matchScoreP0, matchScoreP1 = matchScoreP1, matchScoreP0
				if matchWinner != -1 {
					matchWinner = 1 - matchWinner
				}
				matchP0Bot, matchP1Bot = userP1, userP0
			}
			resp.Matches = append(resp.Matches, batchMatchSummary{
				ID:      res.ID,
				Seed:    res.Seed,
				Winner:  matchWinner,
				ScoreP0: matchScoreP0,
				ScoreP1: matchScoreP1,
				Turns:   res.Turns,
				P0Bot:   matchP0Bot,
				P1Bot:   matchP1Bot,
			})
		}
		if n := float64(len(results)); n > 0 {
			resp.AvgScoreP0 = round2(totalScoreP0 / n)
			resp.AvgScoreP1 = round2(totalScoreP1 / n)
			resp.AvgTurns = round2(totalTurns / n)
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func cleanupTraceDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
