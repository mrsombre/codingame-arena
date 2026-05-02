// Package server wires the `arena serve` HTTP API and static asset handler.
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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// gameOptionsViper builds a viper instance carrying the per-request
// gameOptions map so it can be passed to GameFactory.NewGame.
func gameOptionsViper(opts map[string]string) *viper.Viper {
	v := viper.New()
	for k, val := range opts {
		v.Set(k, val)
	}
	return v
}

// Options configures a Handler built by New.
type Options struct {
	Factory   arena.GameFactory
	Factories map[string]arena.GameFactory
	Assets    fs.FS
	TraceDir  string
	ReplayDir string
	Bots      []string
}

// New returns an http.Handler that serves the embedded viewer bundle and the
// JSON API consumed by the viewer.
func New(opts Options) http.Handler {
	mux := http.NewServeMux()
	resolver := newFactoryResolver(opts)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /api/games", handleGames())
	mux.HandleFunc("GET /api/game", handleGame(resolver))
	mux.HandleFunc("GET /api/serialize", handleSerialize(resolver))
	mux.HandleFunc("GET /api/bots", handleBots(opts.Bots))
	mux.HandleFunc("GET /api/matches", handleMatchList(opts.TraceDir))
	mux.HandleFunc("GET /api/matches/{id}", handleMatchGet(opts.TraceDir))
	mux.HandleFunc("GET /api/replays", handleReplayList(opts.ReplayDir))
	mux.HandleFunc("GET /api/replays/{id}", handleReplayGet(opts.ReplayDir, resolver))
	mux.HandleFunc("POST /api/run", handleRun(resolver))
	mux.HandleFunc("POST /api/batch", handleBatch(resolver, opts.TraceDir))

	mux.Handle("/", http.FileServer(http.FS(opts.Assets)))
	return mux
}

type factoryResolver struct {
	factories   map[string]arena.GameFactory
	defaultName string
}

func newFactoryResolver(opts Options) factoryResolver {
	factories := make(map[string]arena.GameFactory)
	if opts.Factory != nil {
		factories[opts.Factory.Name()] = opts.Factory
	}
	for name, factory := range opts.Factories {
		if factory != nil {
			factories[name] = factory
		}
	}

	defaultName := ""
	if _, ok := factories["winter2026"]; ok {
		defaultName = "winter2026"
	} else {
		names := make([]string, 0, len(factories))
		for name := range factories {
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) > 0 {
			defaultName = names[0]
		}
	}

	return factoryResolver{factories: factories, defaultName: defaultName}
}

func (resolver factoryResolver) fromRequest(r *http.Request) (arena.GameFactory, bool, string) {
	name := r.URL.Query().Get("game")
	if name == "" {
		name = resolver.defaultName
	}
	return resolver.byName(name)
}

func (resolver factoryResolver) byName(name string) (arena.GameFactory, bool, string) {
	if name == "" {
		name = resolver.defaultName
	}
	if name == "" {
		return nil, false, "no games configured"
	}
	factory := resolver.factories[name]
	if factory == nil {
		names := make([]string, 0, len(resolver.factories))
		for available := range resolver.factories {
			names = append(names, available)
		}
		sort.Strings(names)
		return nil, false, fmt.Sprintf("unknown game %q; available: %s", name, strings.Join(names, ", "))
	}
	return factory, true, ""
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

func handleGame(resolver factoryResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		factory, ok, msg := resolver.fromRequest(r)
		if !ok {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
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
func handleSerialize(resolver factoryResolver) http.HandlerFunc {
	reserved := map[string]struct{}{"game": {}, "seed": {}, "player": {}}
	return func(w http.ResponseWriter, r *http.Request) {
		factory, ok, msg := resolver.fromRequest(r)
		if !ok {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
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

		referee, players := factory.NewGame(seed, gameOptionsViper(gameOptions))
		referee.Init(players)
		player := players[playerIdx]

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		for _, line := range referee.GlobalInfoFor(player) {
			_, _ = fmt.Fprintln(w, line)
		}
		referee.ResetGameTurnData()
		for _, line := range referee.FrameInfoFor(player) {
			_, _ = fmt.Fprintln(w, line)
		}
	}
}

type matchListEntry struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	File  string    `json:"file"`
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
				Type:  arena.TraceTypeFromFileName(entry.Name()),
				File:  entry.Name(),
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
		id := r.PathValue("id")
		if !matchIDPattern.MatchString(id) {
			writeError(w, http.StatusBadRequest, "invalid match id")
			return
		}
		if traceDir == "" {
			writeError(w, http.StatusNotFound, "no trace-dir configured; start arena serve with --trace-dir <path>")
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
	BlueBotBin  string            `json:"blueBin"`
	RedBotBin   string            `json:"redBin"`
	Game        string            `json:"game,omitempty"`
	Seed        *int64            `json:"seed,string,omitempty"`
	MaxTurns    int               `json:"maxTurns,omitempty"`
	NoSwap      bool              `json:"noSwap,omitempty"`
	GameOptions map[string]string `json:"gameOptions,omitempty"`
}

func handleRun(resolver factoryResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		var req runRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		factory, ok, msg := resolver.byName(req.Game)
		if !ok {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		if req.BlueBotBin == "" || req.RedBotBin == "" {
			writeError(w, http.StatusBadRequest, "blueBin and redBin are required")
			return
		}
		seed := time.Now().UnixNano()
		if req.Seed != nil {
			seed = *req.Seed
		}
		capture := &traceCapture{}
		runner := arena.NewRunner(factory, arena.MatchOptions{
			MaxTurns:    req.MaxTurns,
			BlueBotBin:  req.BlueBotBin,
			RedBotBin:   req.RedBotBin,
			NoSwap:      req.NoSwap,
			TraceSink:   capture,
			GameOptions: gameOptionsViper(req.GameOptions),
		})
		result := runner.RunMatch(0, seed)

		// Merge RenderMatch's existing JSON object with a `trace` field so
		// the viewer can render the replay from one round-trip.
		payload := map[string]json.RawMessage{}
		if err := json.Unmarshal([]byte(result.RenderMatch()), &payload); err != nil {
			writeError(w, http.StatusInternalServerError, "encode match: "+err.Error())
			return
		}
		traceJSON, err := json.Marshal(capture.match)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "encode trace: "+err.Error())
			return
		}
		payload["trace"] = traceJSON
		writeJSON(w, http.StatusOK, payload)
	}
}

// traceCapture stashes the runner's emitted trace for the current request.
// Single-threaded: RunMatch invokes WriteMatch exactly once per call.
type traceCapture struct{ match arena.TraceMatch }

func (c *traceCapture) WriteMatch(m arena.TraceMatch) error {
	if m.Type == "" {
		m.Type = arena.TraceTypeTrace
	}
	c.match = m
	return nil
}

type batchRequest struct {
	BlueBotBin  string            `json:"blueBin"`
	RedBotBin   string            `json:"redBin"`
	Game        string            `json:"game,omitempty"`
	Seed        *int64            `json:"seed,string,omitempty"`
	Simulations int               `json:"simulations,omitempty"`
	MaxTurns    int               `json:"maxTurns,omitempty"`
	NoSwap      bool              `json:"noSwap,omitempty"`
	GameOptions map[string]string `json:"gameOptions,omitempty"`
}

// batchMatchSummary describes one match from the batch. The left/right fields
// are in-match engine-slot values (under random swap, the left slot may hold
// the red bot rather than our blue bot). Aggregate counters at the top of
// batchResponse stay in blue/red bot perspective.
type batchMatchSummary struct {
	ID           int     `json:"id"`
	Seed         int64   `json:"seed,string"`
	Winner       int     `json:"winner"`
	LeftScore    int     `json:"score_left"`
	RightScore   int     `json:"score_right"`
	Turns        int     `json:"turns"`
	LeftTTFO     float64 `json:"ttfo_left_ms"`
	RightTTFO    float64 `json:"ttfo_right_ms"`
	LeftAOT      float64 `json:"aot_left_ms"`
	RightAOT     float64 `json:"aot_right_ms"`
	LeftBotName  string  `json:"left_bot"`
	RightBotName string  `json:"right_bot"`
}

type batchResponse struct {
	Simulations  int                 `json:"simulations"`
	BlueWins     int                 `json:"wins_blue"`
	RedWins      int                 `json:"wins_red"`
	Draws        int                 `json:"draws"`
	AvgBlueScore float64             `json:"avg_score_blue"`
	AvgRedScore  float64             `json:"avg_score_red"`
	AvgTurns     float64             `json:"avg_turns"`
	AvgBlueTTFO  float64             `json:"avg_ttfo_blue_ms"`
	AvgRedTTFO   float64             `json:"avg_ttfo_red_ms"`
	AvgBlueAOT   float64             `json:"avg_aot_blue_ms"`
	AvgRedAOT    float64             `json:"avg_aot_red_ms"`
	Seed         int64               `json:"seed,string"`
	BlueBotName  string              `json:"blue_bot"`
	RedBotName   string              `json:"red_bot"`
	Matches      []batchMatchSummary `json:"matches"`
}

func handleBatch(resolver factoryResolver, traceDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		var req batchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		factory, ok, msg := resolver.byName(req.Game)
		if !ok {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		if req.BlueBotBin == "" || req.RedBotBin == "" {
			writeError(w, http.StatusBadRequest, "blueBin and redBin are required")
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
			writeError(w, http.StatusBadRequest, "no trace-dir configured; start arena serve with --trace-dir <path>")
			return
		}
		if err := cleanupTraceDir(traceDir); err != nil {
			writeError(w, http.StatusInternalServerError, "cleanup trace dir: "+err.Error())
			return
		}
		runner := arena.NewRunner(factory, arena.MatchOptions{
			MaxTurns:    req.MaxTurns,
			BlueBotBin:  req.BlueBotBin,
			RedBotBin:   req.RedBotBin,
			NoSwap:      req.NoSwap,
			TraceSink:   arena.NewTraceWriter(traceDir, time.Now().Unix()),
			GameOptions: gameOptionsViper(req.GameOptions),
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

		blueBot := filepath.Base(req.BlueBotBin)
		redBot := filepath.Base(req.RedBotBin)
		resp := batchResponse{
			Simulations: len(results),
			Seed:        seed,
			BlueBotName: blueBot,
			RedBotName:  redBot,
			Matches:     make([]batchMatchSummary, 0, len(results)),
		}
		var totalBlueScore, totalRedScore, totalTurns float64
		var totalBlueTTFO, totalRedTTFO, totalBlueAOT, totalRedAOT float64
		for _, res := range results {
			// Prefer raw scores over referee-adjusted Scores so displayed
			// values can match the engine's intrinsic scoring state.
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

			// Aggregate counters follow blue/red bot identity.
			switch userWinner {
			case 0:
				resp.BlueWins++
			case 1:
				resp.RedWins++
			default:
				resp.Draws++
			}
			totalBlueScore += float64(userScores[0])
			totalRedScore += float64(userScores[1])
			totalTurns += float64(res.Turns)
			ttfo := res.TTFO()
			aot := res.AOT()
			totalBlueTTFO += ttfo[0]
			totalRedTTFO += ttfo[1]
			totalBlueAOT += aot[0]
			totalRedAOT += aot[1]

			// Per-match rows report from left/right side perspective so the
			// replay map lines up with the numbers.
			leftScore, rightScore := userScores[0], userScores[1]
			sideWinner := userWinner
			leftBot, rightBot := blueBot, redBot
			sideTTFO, sideAOT := ttfo, aot
			if res.Swapped {
				leftScore, rightScore = rightScore, leftScore
				if sideWinner != -1 {
					sideWinner = 1 - sideWinner
				}
				leftBot, rightBot = redBot, blueBot
				sideTTFO[0], sideTTFO[1] = sideTTFO[1], sideTTFO[0]
				sideAOT[0], sideAOT[1] = sideAOT[1], sideAOT[0]
			}
			resp.Matches = append(resp.Matches, batchMatchSummary{
				ID:           res.ID,
				Seed:         res.Seed,
				Winner:       sideWinner,
				LeftScore:    leftScore,
				RightScore:   rightScore,
				Turns:        res.Turns,
				LeftTTFO:     sideTTFO[0],
				RightTTFO:    sideTTFO[1],
				LeftAOT:      sideAOT[0],
				RightAOT:     sideAOT[1],
				LeftBotName:  leftBot,
				RightBotName: rightBot,
			})
		}
		if n := float64(len(results)); n > 0 {
			resp.AvgBlueScore = round2(totalBlueScore / n)
			resp.AvgRedScore = round2(totalRedScore / n)
			resp.AvgTurns = round2(totalTurns / n)
			resp.AvgBlueTTFO = round2(totalBlueTTFO / n)
			resp.AvgRedTTFO = round2(totalRedTTFO / n)
			resp.AvgBlueAOT = round2(totalBlueAOT / n)
			resp.AvgRedAOT = round2(totalRedAOT / n)
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
