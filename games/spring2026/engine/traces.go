package engine

import (
	"encoding/json"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace event types emitted per turn by the spring2026 engine. Events are
// bucketed by player into TraceTurn.Traces[playerIdx]: index 0 is everything
// player 0 owned this turn, index 1 player 1. Cross-owner events don't exist —
// every action is attributed to the issuing side.
const (
	TraceMove    = "MOVE"
	TraceHarvest = "HARVEST"
	TracePlant   = "PLANT"
	TracePick    = "PICK"
	TraceDrop    = "DROP"
	TraceTrain   = "TRAIN"
	TraceChop    = "CHOP"
	TraceMine    = "MINE"
	TraceWait    = "WAIT"
	TraceMessage = "MSG"
)

// MoveData is emitted on a successful MOVE task. To is the cell the troll
// stepped onto this turn (one MovementSpeed-bounded step toward the bot's
// requested cell). The bot-supplied (x, y) is collapsed into this resolved
// next-step at parse time so the trace only records what actually happened.
type MoveData struct {
	Unit int    `json:"unit"`
	To   [2]int `json:"to"`
}

// HarvestData is emitted on a successful HARVEST task. Amount is the number of
// fruits the troll picked off the plant this turn (1..PLANT_MAX_RESOURCES);
// Type is the plant's fruit kind ("PLUM" / "LEMON" / "APPLE" / "BANANA").
type HarvestData struct {
	Unit   int    `json:"unit"`
	Cell   [2]int `json:"cell"`
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// PlantData is emitted on a successful PLANT task. Type is the seed kind the
// troll consumed from its carry inventory ("PLUM" / "LEMON" / "APPLE" /
// "BANANA"). Cell is the grass cell the seed landed on.
type PlantData struct {
	Unit int    `json:"unit"`
	Cell [2]int `json:"cell"`
	Type string `json:"type"`
}

// PickData is emitted on a successful PICK task. Type is the item kind the
// troll pulled from the shack ("PLUM" / "LEMON" / "APPLE" / "BANANA"). Always
// one item per PICK — the engine increments the unit's inventory by 1.
type PickData struct {
	Unit int    `json:"unit"`
	Type string `json:"type"`
}

// DropData is emitted on a successful DROP task. Items[i] is the count of
// each item the troll handed to the shack this turn, indexed by Item ordinal:
// [PLUM, LEMON, APPLE, BANANA, IRON, WOOD]. The sum equals the carry total
// the unit emptied.
type DropData struct {
	Unit  int           `json:"unit"`
	Items [ItemsCount]int `json:"items"`
}

// TrainData is emitted on a successful TRAIN task. Unit is the id assigned to
// the new troll (units spawn on the player's shack). Talents are the bot's
// requested attribute vector: [moveSpeed, carryCapacity, harvestPower,
// chopPower].
type TrainData struct {
	Unit    int    `json:"unit"`
	Talents [4]int `json:"talents"`
}

// ChopData is emitted on a successful CHOP task (league 3+). Damage is the
// damage the troll dealt to the plant this turn (== unit.ChopPower); Wood is
// the wood the troll added to its carry as the plant fell (0 unless this chop
// felled the tree). Killed is true when the plant ended dead after this turn's
// damage was applied.
type ChopData struct {
	Unit   int    `json:"unit"`
	Cell   [2]int `json:"cell"`
	Damage int    `json:"damage"`
	Wood   int    `json:"wood"`
	Killed bool   `json:"killed,omitempty"`
}

// MineData is emitted on a successful MINE task (league 3+). Iron is the count
// added to the troll's carry this turn (bounded by ChopPower and free carry
// capacity).
type MineData struct {
	Unit int    `json:"unit"`
	Cell [2]int `json:"cell"`
	Iron int    `json:"iron"`
}

// MessageData is emitted on a `MSG <text>` token. Text is the bot's verbatim
// message after Player.SetMessage trims it to 50 chars. A separate event
// (rather than an inline debug field on every action) matches the upstream
// Spring 2026 protocol: MSG is a top-level command, not a per-action debug
// ride-along.
type MessageData struct {
	Text string `json:"text"`
}

// TraceTurnState is the spring2026-owned per-turn payload written into
// TraceTurn.State. Sampled by the arena runner after command parsing and
// before PerformGameUpdate — values reflect what bots saw on stdin this turn,
// not the state after their commands resolve.
type TraceTurnState struct {
	Turn        int                  `json:"turn"`
	Inventories [2][ItemsCount]int   `json:"inventories"`
	Units       [2][]TraceUnit       `json:"units"`
	Plants      []TracePlantState         `json:"plants,omitempty"`
}

// TraceUnit captures one troll's snapshot for the per-turn state. Carry[i]
// counts items in the unit's own inventory indexed by Item ordinal.
type TraceUnit struct {
	ID            int               `json:"id"`
	Pos           [2]int            `json:"pos"`
	MoveSpeed     int               `json:"moveSpeed"`
	CarryCapacity int               `json:"carryCapacity"`
	HarvestPower  int               `json:"harvestPower"`
	ChopPower     int               `json:"chopPower"`
	Carry         [ItemsCount]int   `json:"carry"`
}

// TracePlantState captures one live plant on the board. Plants killed earlier this
// match are absent. Cooldown is the remaining ticks before the plant grows
// (size) or yields (resources) again.
type TracePlantState struct {
	Type      string `json:"type"`
	Pos       [2]int `json:"pos"`
	Size      int    `json:"size"`
	Health    int    `json:"health"`
	Resources int    `json:"resources"`
	Cooldown  int    `json:"cooldown"`
}

// tracePlayer appends a per-player event to the board's trace buffer. Out-of-
// range indices are silently dropped (defensive — board is always 2-player).
func (b *Board) tracePlayer(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(b.Traces) {
		return
	}
	b.Traces[playerIdx] = append(b.Traces[playerIdx], t)
}

// ResetTraces clears the per-turn trace buffer. Called from
// Referee.ResetGameTurnData so MSG / WAIT events emitted during
// ParsePlayerOutputs survive into the same turn's TurnTraces capture, but the
// previous turn's events don't leak forward.
func (b *Board) ResetTraces() {
	b.Traces = [2][]arena.TurnTrace{}
}

// DecorateTraceTurn marshals the per-turn state payload (inventories, troll
// roster, plants). Sampled before PerformGameUpdate by the arena runner.
func (b *Board) DecorateTraceTurn(turn int) json.RawMessage {
	state := TraceTurnState{
		Turn: turn,
	}
	for _, p := range b.Players {
		idx := p.GetIndex()
		if idx < 0 || idx >= 2 {
			continue
		}
		for i := 0; i < ItemsCount; i++ {
			state.Inventories[idx][i] = p.Inv.GetItemCount(Item(i))
		}
	}
	for _, u := range b.Units {
		idx := u.Player.GetIndex()
		if idx < 0 || idx >= 2 {
			continue
		}
		entry := TraceUnit{
			ID:            u.ID,
			Pos:           [2]int{u.Cell.X, u.Cell.Y},
			MoveSpeed:     u.MovementSpeed,
			CarryCapacity: u.CarryCapacity,
			HarvestPower:  u.HarvestPower,
			ChopPower:     u.ChopPower,
		}
		for i := 0; i < ItemsCount; i++ {
			entry.Carry[i] = u.Inv.GetItemCount(Item(i))
		}
		state.Units[idx] = append(state.Units[idx], entry)
	}
	if len(b.Plants) > 0 {
		state.Plants = make([]TracePlantState, 0, len(b.Plants))
		for _, p := range b.Plants {
			state.Plants = append(state.Plants, TracePlantState{
				Type:      p.Type.String(),
				Pos:       [2]int{p.Cell.X, p.Cell.Y},
				Size:      p.Size,
				Health:    p.Health,
				Resources: p.Resources,
				Cooldown:  p.Cooldown,
			})
		}
	}
	raw, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return raw
}
