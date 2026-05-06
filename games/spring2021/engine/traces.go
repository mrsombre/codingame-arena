package engine

import (
	"fmt"

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
// lives at the TraceTurn root (phase, sun_direction, seed_conflict_cell).
const (
	TraceGather   = "GATHER"
	TraceGrow     = "GROW"
	TraceSeed     = "SEED"
	TraceComplete = "COMPLETE"
	TraceWait     = "WAIT"

	TraceActions = "ACTIONS"
)

// GatherData is the data for GATHER events: how many sun points the owning
// player just collected this gathering phase.
type GatherData struct {
	Sun int `json:"sun"`
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

type ValuesMeta[T any] struct {
	Values T `json:"values"`
}

func (g *Game) DecorateTraceTurn(_ int, _ []arena.Player, traceTurn *arena.TraceTurn) {
	if traceTurn == nil {
		return
	}

	traceTurn.Day = new(g.Round)
	traceTurn.Phase = phaseLabel(g.CurrentFrameType)
	traceTurn.SunDirection = new(g.Sun.Orientation)
	traceTurn.Sun = g.traceSun()
	traceTurn.Score = g.traceScore()
	traceTurn.Trees = g.traceTrees()
	if g.seedConflictCell != nil {
		cell := *g.seedConflictCell
		traceTurn.SeedConflictCell = &cell
	}

	if g.CurrentFrameType == FrameActions {
		traceTurn.DayActionIndex = new(g.DayActionIndex)
		actions := g.traceActions()
		// Distribute the per-side action description into each player's slot.
		for i, action := range actions {
			if i >= len(traceTurn.Traces) {
				break
			}
			traceTurn.Traces[i] = append(traceTurn.Traces[i],
				arena.MakeTurnTrace(TraceActions, ValuesMeta[string]{Values: action}),
			)
		}
	}
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

func (g *Game) traceScore() []int {
	values := make([]int, tracePlayerCount(g))
	for _, player := range g.Players {
		idx := player.GetIndex()
		if idx < 0 || idx >= len(values) {
			continue
		}
		values[idx] = player.GetScore()
	}
	return values
}

// traceTrees returns the per-player tree list as [cell_id, size, richness]
// triples. Outer index = player index; the inner list is ordered by cell id
// ascending (TreeOrder traversal).
func (g *Game) traceTrees() [][][3]int {
	trees := make([][][3]int, tracePlayerCount(g))
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
		trees[player] = append(trees[player], [3]int{idx, tree.Size, cell.GetRichness()})
	}
	return trees
}

func (g *Game) traceActions() []string {
	actions := make([]string, tracePlayerCount(g))
	for _, player := range g.Players {
		idx := player.GetIndex()
		if idx < 0 || idx >= len(actions) {
			continue
		}
		actions[idx] = g.traceAction(player)
	}
	return actions
}

func (g *Game) traceAction(player *Player) string {
	if player.IsDeactivated() {
		return "SKIP DEACTIVATED"
	}
	if player.IsWaiting() {
		return "SKIP WAITING"
	}

	action := player.GetAction()
	switch {
	case action.IsGrow():
		tree, ok := g.Trees[action.GetTargetID()]
		if !ok {
			return fmt.Sprintf("GROW %d MISSING", action.GetTargetID())
		}
		return fmt.Sprintf(
			"GROW %d %s %s",
			action.GetTargetID(),
			traceTreeSizeLabel(tree.Size),
			traceTreeSizeLabel(tree.Size+1),
		)
	case action.IsSeed():
		if _, ok := g.Trees[action.GetSourceID()]; !ok {
			return fmt.Sprintf("SEED %d %d MISSING", action.GetSourceID(), action.GetTargetID())
		}
		richness := "UNKNOWN"
		if g.Board != nil {
			cell := g.Board.CellByIndex(action.GetTargetID())
			if cell != nil {
				richness = traceRichnessLabel(cell.GetRichness())
			}
		}
		return fmt.Sprintf(
			"SEED %d %d %s",
			action.GetSourceID(),
			action.GetTargetID(),
			richness,
		)
	case action.IsComplete():
		tree, ok := g.Trees[action.GetTargetID()]
		if !ok {
			return fmt.Sprintf("COMPLETE %d MISSING", action.GetTargetID())
		}
		return fmt.Sprintf("COMPLETE %d %s", action.GetTargetID(), traceTreeSizeLabel(tree.Size))
	case action.IsWait():
		return "WAIT"
	default:
		return "WAIT"
	}
}

func traceRichnessLabel(richness int) string {
	switch richness {
	case RICHNESS_NULL:
		return "NULL"
	case RICHNESS_POOR:
		return "POOR"
	case RICHNESS_OK:
		return "OK"
	case RICHNESS_LUSH:
		return "LUSH"
	default:
		return "UNKNOWN"
	}
}

func traceTreeSizeLabel(size int) string {
	switch size {
	case TREE_SEED:
		return "SEED"
	case TREE_SMALL:
		return "SMALL"
	case TREE_MEDIUM:
		return "MEDIUM"
	case TREE_TALL:
		return "LARGE"
	default:
		return "UNKNOWN"
	}
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
