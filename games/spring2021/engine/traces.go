package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn by the spring2021 engine. Mirrors the
// GameSummaryManager event categories (game_game_summary_manager.go) but
// in structured form so analyzers can count and slice without parsing
// summary text.
//
// GATHER_PHASE and SUN_MOVE are phase markers: every PerformGameUpdate
// invocation that runs the GATHERING or SUN_MOVE frame emits exactly one
// such event, mirroring the standalone phase frames CodinGame emits in
// the replay JSON. They guarantee a turn always carries at least one
// structured event even when no per-player action does (e.g. all trees
// shadowed → no GATHER events).
const (
	TraceGatherPhase  = "GATHER_PHASE"
	TraceGather       = "GATHER"
	TraceGrow         = "GROW"
	TraceSeed         = "SEED"
	TraceComplete     = "COMPLETE"
	TraceWait         = "WAIT"
	TraceSeedConflict = "SEED_CONFLICT"
	TraceSunMove      = "SUN_MOVE"
)

// PlayerMeta is the meta for trace events whose only subject is a player
// (WAIT). Player is the player index (0 or 1).
type PlayerMeta struct {
	Player int `json:"player"`
}

// PlayerSunMeta is the meta for GATHER events. Sun is the number of sun
// points the player just collected this gathering phase.
type PlayerSunMeta struct {
	Player int `json:"player"`
	Sun    int `json:"sun"`
}

// TreeActionMeta is the meta for GROW events. Cell is the board cell
// index of the tree being grown.
type TreeActionMeta struct {
	Player int `json:"player"`
	Cell   int `json:"cell"`
}

// SeedMeta is the meta for SEED events: a seed is planted on Target from
// the source tree on Source.
type SeedMeta struct {
	Player int `json:"player"`
	Source int `json:"source"`
	Target int `json:"target"`
}

// CompleteMeta is the meta for COMPLETE events. Points is the score
// awarded for the cut tree (Nutrients + richness bonus at the moment of
// the action).
type CompleteMeta struct {
	Player int `json:"player"`
	Cell   int `json:"cell"`
	Points int `json:"points"`
}

// SeedConflictMeta is the meta for SEED_CONFLICT events. Cell is the
// contested target cell; both players' seeds are cancelled.
type SeedConflictMeta struct {
	Cell int `json:"cell"`
}

// GatherPhaseMeta is the meta for GATHER_PHASE events. Round is the
// 0-indexed round whose gathering step is starting.
type GatherPhaseMeta struct {
	Round int `json:"round"`
}

// SunMoveMeta is the meta for SUN_MOVE events. Round is the round that
// just ended (0-indexed). Direction is the sun's new orientation after
// the move (0..5), or -1 when the move would have advanced past
// MAX_ROUNDS — the engine skips Sun.move() in that case so direction
// stays unchanged, but a -1 sentinel is clearer to consumers.
type SunMoveMeta struct {
	Round     int `json:"round"`
	Direction int `json:"direction"`
}

func (g *Game) trace(t arena.TurnTrace) {
	g.traces = append(g.traces, t)
}
