package engine

import (
	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace types emitted per turn by the spring2021 engine. Mirrors the
// GameSummaryManager event categories (game_game_summary_manager.go) but
// in structured form so analyzers can count and slice without parsing
// summary text.
const (
	TraceGather       = "GATHER"
	TraceGrow         = "GROW"
	TraceSeed         = "SEED"
	TraceComplete     = "COMPLETE"
	TraceWait         = "WAIT"
	TraceSeedConflict = "SEED_CONFLICT"
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

func (g *Game) trace(t arena.TurnTrace) {
	g.traces = append(g.traces, t)
}
