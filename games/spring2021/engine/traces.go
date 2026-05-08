package engine

import (
	"encoding/json"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn by the spring2021 engine. Mirrors the
// GameSummaryManager event categories (game_game_summary_manager.go) but
// in structured form so analyzers can count and slice without parsing
// summary text.
//
// Events are bucketed by player into TraceTurn.Traces[playerIdx]. The
// player index is encoded positionally by the slot, so per-event payloads
// no longer carry a "player" field. Phase markers and game-state events
// (SUN_MOVE, SEED_CONFLICT) are not emitted as traces; that information
// lives in the per-turn TraceTurnState payload (phase, sun_direction,
// seed_conflict_cell).
const (
	TraceGather   = "GATHER"
	TraceGrow     = "GROW"
	TraceSeed     = "SEED"
	TraceComplete = "COMPLETE"
	TraceWait     = "WAIT"
	TraceDebug    = "DEBUG"
)

// TraceTurnState is the spring2021-owned per-turn payload written into
// TraceTurn.State. Field names use the original on-disk JSON keys, just
// nested under "state" so the arena's TraceTurn stays game-agnostic.
type TraceTurnState struct {
	Day              *int       `json:"day,omitempty"`
	Phase            string     `json:"phase,omitempty"`
	DayActionIndex   *int       `json:"day_action_index,omitempty"`
	SunDirection     *int       `json:"sun_direction,omitempty"`
	Nutrients        *int       `json:"nutrients,omitempty"`
	Sun              []int      `json:"sun,omitempty"`
	Trees            [][][4]int `json:"trees,omitempty"`
	SeedConflictCell *int       `json:"seed_conflict_cell,omitempty"`
}

// GatherData is the data for GATHER events: one event per tree that gathered
// sun this phase. Cell is the tree's cell index; Sun is the points it gave
// (equal to its size when not under a spooky shadow). Events are ordered by
// cell id ascending (TreeOrder traversal).
type GatherData struct {
	Cell int `json:"cell"`
	Sun  int `json:"sun"`
}

// GrowData is the data for GROW events: the cell index of the tree the
// owning player just grew.
type GrowData struct {
	Cell int `json:"cell"`
}

// SeedData is the data for SEED events: the owning player sent a seed
// from Source to Target.
type SeedData struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

// CompleteData is the data for COMPLETE events: the owning player cut a
// tree on Cell, awarding Points (Nutrients plus the cell's richness bonus
// at the moment of the action).
type CompleteData struct {
	Cell   int `json:"cell"`
	Points int `json:"points"`
}

// DebugData is the data for DEBUG events: the trailing free-text "message"
// the player appended to their command (e.g. "GL HF" in "WAIT GL HF").
// CommandManager extracts it via the optional message group on each action
// regex; Player.Reset clears it at turn start, so a non-empty Value reflects
// the message sent on the same turn the surrounding action trace was emitted.
type DebugData struct {
	Value string `json:"value"`
}

func (g *Game) DecorateTraceTurn(_ int, _ []arena.Player) json.RawMessage {
	state := TraceTurnState{
		Day:          new(g.Round),
		Phase:        phaseLabel(g.CurrentFrameType),
		SunDirection: new(g.Sun.Orientation),
		Nutrients:    new(g.Nutrients),
		Sun:          g.traceSun(),
		Trees:        g.traceTrees(),
	}
	if g.seedConflictCell != nil {
		cell := *g.seedConflictCell
		state.SeedConflictCell = &cell
	}
	if g.CurrentFrameType == FrameActions {
		state.DayActionIndex = new(g.DayActionIndex)
	}
	raw, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return raw
}

func phaseLabel(f FrameType) string {
	switch f {
	case FrameGathering:
		return "gathering"
	case FrameActions:
		return "actions"
	case FrameSunMove:
		return "sun"
	default:
		return ""
	}
}

func tracePlayerCount(g *Game) int {
	if len(g.Players) > 2 {
		return len(g.Players)
	}
	return 2
}

func (g *Game) traceSun() []int {
	values := make([]int, tracePlayerCount(g))
	for _, player := range g.Players {
		idx := player.GetIndex()
		if idx < 0 || idx >= len(values) {
			continue
		}
		values[idx] = player.GetSun()
	}
	return values
}

// traceTrees returns the per-player tree list as
// [cell_id, richness, size, isDormant] tuples. The leading pair is static
// (cell identity + soil quality) and the trailing pair is mutable (current
// growth stage + whether the tree was acted on this day). isDormant is 1 when
// the tree cannot be acted on again until the next day, 0 otherwise. Outer
// index = player index; the inner list is ordered by cell id ascending
// (TreeOrder traversal).
func (g *Game) traceTrees() [][][4]int {
	trees := make([][][4]int, tracePlayerCount(g))
	if g.Board == nil {
		return trees
	}
	for _, idx := range g.TreeOrder {
		tree := g.Trees[idx]
		player := tree.Owner.GetIndex()
		if player < 0 || player >= len(trees) {
			continue
		}
		cell := g.Board.CellByIndex(idx)
		if cell == nil {
			continue
		}
		dormant := 0
		if tree.Dormant {
			dormant = 1
		}
		trees[player] = append(trees[player], [4]int{idx, cell.GetRichness(), tree.Size, dormant})
	}
	return trees
}

// tracePlayer appends a player-owned event into the per-player slot.
// playerIdx out of range is silently dropped (defensive for >2-player setups,
// which spring2021 doesn't support but the helper signature stays uniform).
func (g *Game) tracePlayer(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(g.traces) {
		return
	}
	g.traces[playerIdx] = append(g.traces[playerIdx], t)
}
