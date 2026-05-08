package engine

import (
	"encoding/json"
	"sort"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// SerializeTraceFrameInfo (defined in serializer.go) is the no-fog god-mode
// view used by TraceFrameInfoProducer. The Referee.TraceFrameInfo method
// (game_referee.go) wires it into the arena trace path.

// Trace types are split into two buckets:
//
//   - **Command** traces capture what the bot asked each pac to do this
//     turn (MOVE/SPEED/SWITCH/WAIT). Exactly one per alive pac per main
//     turn, regardless of whether the engine ultimately honored it (e.g.
//     a SPEED command on cooldown still emits the command trace; only
//     the activation event is suppressed). Command metas carry the
//     bot's per-pac chat message as `debug` (omitted when empty).
//
//   - **Event** traces capture engine-derived consequences of movement
//     resolution (EAT/KILLED/COLLIDE_SELF/COLLIDE_ENEMY). Multiple per
//     pac per turn possible. Event metas do **not** carry `debug` —
//     the chat string belongs to the command, not its downstream
//     fallout.
const (
	// Commands.
	TraceMove   = "MOVE"
	TraceSpeed  = "SPEED"
	TraceSwitch = "SWITCH"
	TraceWait   = "WAIT"

	// Events.
	TraceEat          = "EAT"
	TraceKilled       = "KILLED"
	TraceCollideSelf  = "COLLIDE_SELF"
	TraceCollideEnemy = "COLLIDE_ENEMY"
)

// MoveMeta is the meta for the MOVE command. Target is the cell coordinate
// the bot supplied; the engine resolves a shortest path and steps the pac
// the first one (or two, when SPEED is active) cells along that path.
// Where the pac actually ends up is recoverable from the next turn's
// gameInput; this meta records the request, not the outcome.
type MoveMeta struct {
	Pac    int    `json:"pac"`
	Target [2]int `json:"target"`
	Debug  string `json:"debug,omitempty"`
}

// SpeedMeta is the meta for the SPEED command.
type SpeedMeta struct {
	Pac   int    `json:"pac"`
	Debug string `json:"debug,omitempty"`
}

// SwitchMeta is the meta for the SWITCH command. Type is the requested
// pacman type ("ROCK"/"PAPER"/"SCISSORS"). Emitted whether or not the
// ability actually activated (cooldown / league flag may suppress the
// effect — but the command was still sent).
type SwitchMeta struct {
	Pac   int    `json:"pac"`
	Type  string `json:"type"`
	Debug string `json:"debug,omitempty"`
}

// WaitMeta is the meta for the WAIT command. Also covers pacs that
// received no command at all this turn (Java's "Pac N received no
// command." case): both leave intent at NoAction and look identical
// from the trace's perspective. Distinguish via `debug` — an explicit
// `WAIT 0 chilling` carries the message, an unsent command does not.
type WaitMeta struct {
	Pac   int    `json:"pac"`
	Debug string `json:"debug,omitempty"`
}

// PacMeta is the meta for events whose only subject is one pac
// (COLLIDE_SELF, COLLIDE_ENEMY).
type PacMeta struct {
	Pac int `json:"pac"`
}

// EatMeta is the meta for the EAT event. Cost > 1 marks a super pellet.
type EatMeta struct {
	Pac   int    `json:"pac"`
	Coord [2]int `json:"coord"`
	Cost  int    `json:"cost"`
}

// KilledMeta is the meta for the KILLED event. Pac is the dead pac
// (subject); Killer is the global pac ID that caused the kill.
type KilledMeta struct {
	Pac    int    `json:"pac"`
	Coord  [2]int `json:"coord"`
	Killer int    `json:"killer"`
}

func coordPair(c Coord) [2]int {
	return [2]int{c.X, c.Y}
}

// tracePlayer appends a player-owned event into the per-player slot.
// Slots out of [0, 1] are silently dropped (defensive).
func (g *Game) tracePlayer(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(g.traces) {
		return
	}
	g.traces[playerIdx] = append(g.traces[playerIdx], t)
}

// traceBoth mirrors a cross-owner event into both player slots so each side
// sees its involvement (e.g. COLLIDE_ENEMY).
func (g *Game) traceBoth(t arena.TurnTrace) {
	for i := range g.traces {
		g.traces[i] = append(g.traces[i], t)
	}
}

// TraceTurnState is the spring2020-owned per-turn payload written into
// TraceTurn.State. Captures the board's pellet inventory and a per-side
// roster of pac state — everything analyzers need without re-parsing
// `gameInput` line by line.
//
// Sampled by the arena runner *after* command parsing and *before*
// PerformGameUpdate, so values reflect the state players saw when
// choosing actions: cooldowns/durations have already been ticked at
// the start of the turn, but no SPEED/SWITCH activations or movement
// from this turn have applied yet.
type TraceTurnState struct {
	Pellets      int          `json:"pellets"`
	SuperPellets int          `json:"superPellets"`
	Pacs         [][]TracePac `json:"pacs"`
}

// TracePac is the per-pac entry in TraceTurnState.Pacs. Outer index =
// match side; inner list ordered by global Pacman.ID ascending (same
// ordering the per-pac event payloads use). Dead pacs are included
// with Type="DEAD" and frozen Coord/IsSpeed/Cooldown values reflecting
// their state at the moment of death — gives a full roster view
// regardless of the league's PROVIDE_DEAD_PACS flag.
type TracePac struct {
	ID       int    `json:"id"`
	Coord    [2]int `json:"coord"`
	Type     string `json:"type"`
	IsSpeed  int    `json:"isSpeed"`
	Cooldown int    `json:"cooldown"`
}

// DecorateTraceTurn marshals the per-turn TraceTurnState payload onto
// TraceTurn.State via the arena's TraceTurnDecorator interface.
func (g *Game) DecorateTraceTurn(_ int, _ []arena.Player) json.RawMessage {
	state := TraceTurnState{
		Pellets:      g.tracePelletCount(),
		SuperPellets: g.traceSuperPelletCount(),
		Pacs:         g.tracePacs(),
	}
	raw, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return raw
}

func (g *Game) tracePelletCount() int {
	if g.Grid == nil {
		return 0
	}
	n := 0
	for _, cell := range g.Grid.Cells {
		if cell.HasPellet {
			n++
		}
	}
	return n
}

func (g *Game) traceSuperPelletCount() int {
	if g.Grid == nil {
		return 0
	}
	n := 0
	for _, cell := range g.Grid.Cells {
		if cell.HasCherry {
			n++
		}
	}
	return n
}

func (g *Game) tracePacs() [][]TracePac {
	pacs := make([][]TracePac, 2)
	for _, pac := range g.Pacmen {
		side := pac.Owner.Index
		if side < 0 || side >= len(pacs) {
			continue
		}
		typeName := pac.Type.Name()
		if pac.Dead {
			typeName = "DEAD"
		}
		isSpeed := 0
		if pac.AbilityDuration > 0 {
			isSpeed = 1
		}
		pacs[side] = append(pacs[side], TracePac{
			ID:       pac.ID,
			Coord:    coordPair(pac.Position),
			Type:     typeName,
			IsSpeed:  isSpeed,
			Cooldown: pac.AbilityCooldown,
		})
	}
	for i := range pacs {
		sort.SliceStable(pacs[i], func(a, b int) bool {
			return pacs[i][a].ID < pacs[i][b].ID
		})
	}
	return pacs
}

// emitCommandTraces writes one command trace per alive pac, capturing what
// the bot asked the pac to do this turn. Called once per main turn at the
// start of PerformGameUpdate (after the trace buffer is cleared and before
// any ability/movement resolution), so command traces always lead the
// turn's `traces[]` array. The SPEED sub-step does not add additional
// command traces — bots are not re-prompted between steps.
//
// SPEED/SWITCH commands emit the trace regardless of whether the ability
// actually activated. A SPEED command issued while cooldown > 0 is still
// the command the bot sent; suppressing the trace would lose information.
// The activation effect itself is gated by ExecutePacmenAbilities.
func (g *Game) emitCommandTraces() {
	for _, pac := range g.Pacmen {
		if pac.Dead {
			continue
		}
		switch pac.Intent.Type {
		case ActionMove:
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceMove, MoveMeta{
				Pac:    pac.ID,
				Target: coordPair(pac.Intent.Target),
				Debug:  pac.Message,
			}))
		case ActionSpeed:
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceSpeed, SpeedMeta{
				Pac:   pac.ID,
				Debug: pac.Message,
			}))
		case ActionSwitch:
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceSwitch, SwitchMeta{
				Pac:   pac.ID,
				Type:  pac.Intent.NewType.Name(),
				Debug: pac.Message,
			}))
		case ActionWait:
			g.tracePlayer(pac.Owner.Index, arena.MakeTurnTrace(TraceWait, WaitMeta{
				Pac:   pac.ID,
				Debug: pac.Message,
			}))
		}
	}
}
