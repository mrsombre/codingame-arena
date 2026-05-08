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
// lives in the per-turn TraceTurnState payload (phase, sunDirection,
// seedConflictCell).
const (
	TraceGather   = "GATHER"
	TraceGrow     = "GROW"
	TraceSeed     = "SEED"
	TraceComplete = "COMPLETE"
	TraceWait     = "WAIT"
)

// TraceTurnState is the spring2021-owned per-turn payload written into
// TraceTurn.State. Field names use the original on-disk JSON keys, just
// nested under "state" so the arena's TraceTurn stays game-agnostic.
type TraceTurnState struct {
	Day              *int       `json:"day,omitempty"`
	Phase            string     `json:"phase,omitempty"`
	DayActionIndex   *int       `json:"dayActionIndex,omitempty"`
	SunDirection     *int       `json:"sunDirection,omitempty"`
	Nutrients        *int       `json:"nutrients,omitempty"`
	Sun              []int      `json:"sun,omitempty"`
	Trees            [][][4]int `json:"trees,omitempty"`
	SeedConflictCell *int       `json:"seedConflictCell,omitempty"`
}

// GatherData is the data for GATHER events: one event per tree at the
// gathering phase, regardless of whether it harvested. Cell is the tree's
// cell index; Sun is the points it gave — equal to the tree's size when not
// under a spooky shadow, 0 when it is. Events are ordered by cell id ascending
// (TreeOrder traversal). Seeds (size 0) emit Sun=0 too, since rules grant no
// sun for size-0 trees regardless of shadow.
type GatherData struct {
	Cell int `json:"cell"`
	Sun  int `json:"sun"`
}

// Action data structs share a `Debug` field carrying the trailing free-text
// "message" the player appended to their command (e.g. "GL HF" in
// "WAIT GL HF"). It is omitted when empty. CommandManager extracts the
// message via the optional group on each action regex; Player.Reset clears
// it at turn start, so a non-empty Debug reflects the message sent on the
// same turn the surrounding action trace was emitted. There is no separate
// DEBUG event — the message piggybacks on the action it accompanied; on
// errored actions (no trace emitted) the message is dropped.

// GrowData is the data for GROW events: the cell index of the tree the
// owning player just grew. Cost is the sun debited from the player for
// this action (TREE_BASE_COST[targetSize] + count of player-owned trees
// already at targetSize at the moment of the action).
type GrowData struct {
	Cell  int    `json:"cell"`
	Cost  int    `json:"cost"`
	Debug string `json:"debug,omitempty"`
}

// SeedData is the data for SEED events: the owning player sent a seed
// from Source to Target. Cost is the sun debited at submission time
// (number of seeds — size-0 trees — the player already owns). On a
// same-frame seed conflict (state.seedConflictCell set) the cost is
// refunded by the engine, but the trace still reports what was attempted.
type SeedData struct {
	Source int    `json:"source"`
	Target int    `json:"target"`
	Cost   int    `json:"cost"`
	Debug  string `json:"debug,omitempty"`
}

// CompleteData is the data for COMPLETE events: the owning player cut a
// tree on Cell, awarding Points (Nutrients plus the cell's richness bonus
// at the moment of the action). Cost is the sun debited (always
// LIFECYCLE_END_COST = 4).
type CompleteData struct {
	Cell   int    `json:"cell"`
	Points int    `json:"points"`
	Cost   int    `json:"cost"`
	Debug  string `json:"debug,omitempty"`
}

// WaitData is the data for WAIT events when the player appended a chat
// message ("WAIT GL HF" → Debug="GL HF"). When the player waited without
// a message, the trace is emitted with no `data` field at all rather than
// `data: {}`.
type WaitData struct {
	Debug string `json:"debug,omitempty"`
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
