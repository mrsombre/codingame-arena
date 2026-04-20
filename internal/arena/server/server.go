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

	mux.HandleFunc("GET /api/game", handleGame(opts.Factory))
	mux.HandleFunc("GET /api/serialize", handleSerialize(opts.Factory))
	mux.HandleFunc("GET /api/matches", handleMatchList(opts.TraceDir))
	mux.HandleFunc("GET /api/matches/{id}", handleMatchGet(opts.TraceDir))
	mux.HandleFunc("POST /api/run", handleRun(opts.Factory, opts.TraceDir))

	mux.Handle("/", http.FileServer(http.FS(opts.Assets)))
	return mux
}

type gameInfo struct {
	Name     string `json:"name"`
	MaxTurns int    `json:"maxTurns"`
}

func handleGame(factory arena.GameFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, gameInfo{
			Name:     factory.Name(),
			MaxTurns: factory.MaxTurns(),
		})
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
	Seed        *int64            `json:"seed,omitempty"`
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

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
