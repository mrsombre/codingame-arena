package engine

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/mrsombre/codingame-arena/internal/arena"
	"github.com/mrsombre/codingame-arena/internal/util/sha1prng"
)

// TestDebug_DumpGridStages re-runs GridMaker step by step and dumps the grid
// after each phase, so we can pin down where our wall layout first diverges
// from CodinGame's. Skipped unless ARENA_DEBUG_GRID=<seed>:<league> is set.
//
//	ARENA_DEBUG_GRID="-8937286792422418000:4" go test ./games/winter2026/engine/ -run TestDebug_DumpGridStages -v
func TestDebug_DumpGridStages(t *testing.T) {
	spec := os.Getenv("ARENA_DEBUG_GRID")
	if spec == "" {
		t.Skip("set ARENA_DEBUG_GRID=<seed>:<league> to enable")
	}
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		t.Fatalf("ARENA_DEBUG_GRID must be <seed>:<league>")
	}
	seed, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	league, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatalf("league: %v", err)
	}

	dumpGridStages(t, seed, league)
}

// TestDebug_InspectReplay loads a saved replay and runs it through the engine,
// printing per-turn events plus the initial grid+spawns. Skipped unless
// ARENA_DEBUG_REPLAY=<gameId> is set so it never runs in regular suites.
//
// Usage:
//
//	ARENA_DEBUG_REPLAY=882623841 go test ./games/winter2026/engine/ -run TestDebug_InspectReplay -v
func TestDebug_InspectReplay(t *testing.T) {
	id := os.Getenv("ARENA_DEBUG_REPLAY")
	if id == "" {
		t.Skip("set ARENA_DEBUG_REPLAY=<gameId> to enable")
	}

	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	path := filepath.Join(repoRoot, "replays", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var replay arena.CodinGameReplay[arena.CodinGameReplayFrame]
	if err := json.Unmarshal(data, &replay); err != nil {
		t.Fatalf("parse: %v", err)
	}

	seed, ok := arena.ResolveReplaySeed(replay)
	if !ok {
		t.Fatalf("replay missing seed")
	}
	league := arena.ParseReplayLeague(replay.QuestionTitle)

	t.Logf("seed=%d league=%d", seed, league)
	t.Logf("replay scores: %.1f vs %.1f", replay.GameResult.Scores[0], replay.GameResult.Scores[1])

	dumpInitialState(t, seed, league)

	gameOptions := viper.New()
	if league > 0 {
		gameOptions.Set("league", strconv.Itoa(league))
	}
	moves := arena.ReplayMovesFromFrames(replay)
	names := arena.ReplayPlayerNames(replay)
	blueSide := 0
	if replay.Blue != "" && names[1] == replay.Blue {
		blueSide = 1
	}

	trace, finalScores := arena.RunReplay(NewFactory(), seed, gameOptions, moves, names, blueSide, 0)

	t.Logf("engine final: %d vs %d (raw=%d vs %d), turns=%d",
		finalScores[0], finalScores[1], int(trace.Scores[0]), int(trace.Scores[1]), len(trace.Turns))

	for i, tt := range trace.Turns {
		var events []string
		for _, ev := range tt.Traces {
			events = append(events, ev.Label+"("+ev.Payload+")")
		}
		evstr := ""
		if len(events) > 0 {
			evstr = " events=" + strings.Join(events, ",")
		}
		t.Logf("turn %3d:%s", i, evstr)
		if tt.Output[0] != "" {
			t.Logf("        p0_out=%q", tt.Output[0])
		}
		if tt.Output[1] != "" {
			t.Logf("        p1_out=%q", tt.Output[1])
		}
	}
}

// dumpGridStages re-implements GridMaker.Make() phase-by-phase and dumps the
// wall layout after each phase. Lets us bisect divergence from CodinGame's
// reference grid for a single seed without instrumenting production code.
func dumpGridStages(t *testing.T, seed int64, leagueLevel int) {
	t.Helper()

	rng := newSHA1Random(seed)

	skew := 0.3
	switch leagueLevel {
	case 1:
		skew = 2
	case 2:
		skew = 1
	case 3:
		skew = 0.8
	}

	rand := rng.NextDouble()
	height := MIN_GRID_HEIGHT + javaRound(pow(rand, skew)*float64(MAX_GRID_HEIGHT-MIN_GRID_HEIGHT))
	width := javaRoundFloat32(float32(height) * ASPECT_RATIO)
	if width%2 != 0 {
		width++
	}
	t.Logf("rand=%.10f → height=%d width=%d", rand, height, width)
	g := NewGrid(width, height)

	b := 5 + rng.NextDouble()*10
	t.Logf("b=%.10f", b)

	for x := 0; x < width; x++ {
		g.GetXY(x, height-1).SetType(TileWall)
	}
	for y := height - 2; y >= 0; y-- {
		yNorm := float64(height-1-y) / float64(height-1)
		blockChanceEl := 1 / (yNorm + 0.1) / b
		for x := 0; x < width; x++ {
			if rng.NextDouble() < blockChanceEl {
				g.GetXY(x, y).SetType(TileWall)
			}
		}
	}
	dumpGridWalls(t, "after_random_walls", g)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: y}
			opp := g.Opposite(c)
			g.Get(opp).SetType(g.Get(c).Type)
		}
	}
	dumpGridWalls(t, "after_mirror", g)

	for _, island := range g.DetectAirPockets() {
		if len(island) < 10 {
			for c := range island {
				g.Get(c).SetType(TileWall)
			}
		}
	}
	dumpGridWalls(t, "after_airpockets", g)

	somethingDestroyed := true
	pruneRounds := 0
	destructions := 0
	shuffleNextIntCalls := 0
	for somethingDestroyed {
		somethingDestroyed = false
		for _, c := range g.Coords() {
			if g.Get(c).IsWall() {
				continue
			}
			var neighbourWalls []Coord
			for _, n := range g.Neighbours4(c) {
				if g.Get(n).IsWall() {
					neighbourWalls = append(neighbourWalls, n)
				}
			}
			if len(neighbourWalls) >= 3 {
				var destroyable []Coord
				for _, n := range neighbourWalls {
					if n.Y <= c.Y {
						destroyable = append(destroyable, n)
					}
				}
				size := len(destroyable)
				if size > 1 {
					shuffleNextIntCalls += size - 1
				}
				t.Logf("  destroy@(%d,%d) destroyable.size=%d", c.X, c.Y, size)
				shuffleCoords(destroyable, rng)
				t.Logf("    picked=(%d,%d)", destroyable[0].X, destroyable[0].Y)
				g.Get(destroyable[0]).SetType(TileEmpty)
				g.Get(g.Opposite(destroyable[0])).SetType(TileEmpty)
				somethingDestroyed = true
				destructions++
			}
		}
		pruneRounds++
	}
	t.Logf("pruning rounds=%d destructions=%d shuffleNextIntCalls=%d",
		pruneRounds, destructions, shuffleNextIntCalls)
	dumpGridWalls(t, "after_prune", g)

	island := g.DetectLowestIsland()
	lowerBy := 0
	canLower := true
	for canLower {
		for x := 0; x < width; x++ {
			c := Coord{X: x, Y: height - 1 - (lowerBy + 1)}
			if !coordSliceContains(island, c) {
				canLower = false
				break
			}
		}
		if canLower {
			lowerBy++
		}
	}
	t.Logf("lowest island size=%d, lowerBy max=%d", len(island), lowerBy)
	if lowerBy >= 2 {
		picked := rng.NextIntRange(2, lowerBy+1)
		t.Logf("lowerBy random pick=%d (range [2, %d])", picked, lowerBy)
		lowerBy = picked
	}
	for _, c := range island {
		g.Get(c).SetType(TileEmpty)
		g.Get(g.Opposite(c)).SetType(TileEmpty)
	}
	for _, c := range island {
		lowered := Coord{X: c.X, Y: c.Y + lowerBy}
		if g.Get(lowered).IsValid() {
			g.Get(lowered).SetType(TileWall)
			g.Get(g.Opposite(lowered)).SetType(TileWall)
		}
	}
	dumpGridWalls(t, "after_lowering", g)

	// Apple placement (left half, then mirror).
	apples := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width/2; x++ {
			c := Coord{X: x, Y: y}
			if g.Get(c).IsEmpty() && rng.NextDouble() < 0.025 {
				apples += 2
			}
		}
	}
	t.Logf("apples placed by primary pass: %d (will trigger fallback if <8)", apples)

	// Isolated-wall to apple conversion changes walls.
	removed := 0
	for _, c := range g.Coords() {
		if g.Get(c).IsEmpty() {
			continue
		}
		count := 0
		for _, n := range g.Neighbours(c, Adjacency8[:]) {
			if g.Get(n).IsWall() {
				count++
			}
		}
		if count == 0 {
			g.Get(c).SetType(TileEmpty)
			g.Get(g.Opposite(c)).SetType(TileEmpty)
			removed++
		}
	}
	t.Logf("isolated walls removed: %d", removed)
	dumpGridWalls(t, "after_isolated_removal", g)
}

func dumpGridWalls(t *testing.T, label string, g *Grid) {
	t.Helper()
	t.Logf("--- %s ---", label)
	for y := 0; y < g.Height; y++ {
		var row []byte
		for x := 0; x < g.Width; x++ {
			ch := byte('.')
			if g.GetXY(x, y).IsWall() {
				ch = '#'
			}
			row = append(row, ch)
		}
		t.Log(string(row))
	}
}

func pow(x, y float64) float64 { return math.Pow(x, y) }

func newSHA1Random(seed int64) Random { return sha1prng.New(seed) }

// NewFactory is exported elsewhere in this package; this helper just
// conjures a fresh game so we can dump the initial grid + bird placement
// without re-running the match.
func dumpInitialState(t *testing.T, seed int64, league int) {
	t.Helper()
	g := NewGame(seed, league)
	players := []*Player{NewPlayer(0), NewPlayer(1)}
	g.Init(players)

	t.Logf("grid %dx%d, %d apples, %d spawn cells",
		g.Grid.Width, g.Grid.Height, len(g.Grid.Apples), len(g.Grid.Spawns))

	t.Log("walls only (#=wall, .=empty):")
	for y := 0; y < g.Grid.Height; y++ {
		var row []byte
		for x := 0; x < g.Grid.Width; x++ {
			ch := byte('.')
			if g.Grid.GetXY(x, y).IsWall() {
				ch = '#'
			}
			row = append(row, ch)
		}
		t.Log(string(row))
	}

	for _, p := range players {
		for _, b := range p.Birds {
			t.Logf("bird %d P%d: head=%s body=%v facing=%s",
				b.ID, p.GetIndex(), b.Body[0], b.Body, b.Facing())
		}
	}
}
