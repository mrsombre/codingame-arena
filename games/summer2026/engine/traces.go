package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

// Trace event types emitted per turn by the summer2026 engine. Events are
// bucketed by player into TraceTurn.Traces[playerIdx]: index 0 is everything
// player 0 owned this turn, index 1 player 1. CONTESTED and INK belong to no
// single side and are mirrored into both buckets.
const (
	TraceMessage     = "MESSAGE"
	TraceAutoplace   = "AUTOPLACE"
	TraceTrack       = "TRACK"
	TraceContested   = "CONTESTED"
	TraceDisrupt     = "DISRUPT"
	TraceInk         = "INK"
	TraceScore       = "SCORE"
	TraceFailed      = "FAILED"
	TraceTurnSummary = "TURN"
)

// MessageData is emitted when a MESSAGE command parses. Text is what the bot
// wrote; the engine stores it for one turn and never acts on it.
type MessageData struct {
	Text string `json:"text"`
}

// AutoplaceData is emitted once per honoured AUTOPLACE, after the planner has
// expanded it. Placements is how many rails the winning plan lays — 0 when no
// route exists, which is the only way an AUTOPLACE can silently do nothing.
// The expansion is what DoActions then charges for, so a Placements count
// higher than the TRACK events that follow means the budget ran out mid-plan.
type AutoplaceData struct {
	From       [2]int `json:"from"`
	To         [2]int `json:"to"`
	Placements int    `json:"placements"`
}

// TrackData is emitted once per rail the player paid for. Cost is what that
// cell charged, which is terrain-dependent. Autoplaced marks a rail the
// AUTOPLACE planner chose rather than one the bot named itself.
//
// Ownership is only written after both players have been charged, so a cell
// that also carries a CONTESTED event this turn ended up neutral despite this
// event naming its buyer.
type TrackData struct {
	Cell       [2]int `json:"cell"`
	Cost       int    `json:"cost"`
	Terrain    string `json:"terrain"`
	Autoplaced bool   `json:"autoplaced,omitempty"`
}

// ContestedData is emitted for a cell both players bought this turn. Neither
// is refunded and the cell goes neutral, scoring for nobody.
type ContestedData struct {
	Cell [2]int `json:"cell"`
}

// DisruptData is emitted per honoured disruption. Instability is the region's
// count after this blot; the region inks out on the turn it reaches the
// threshold, which is reported separately as an INK event.
type DisruptData struct {
	Zone        int `json:"zone"`
	Instability int `json:"instability"`
}

// InkData is emitted when a region inks out, into both players' buckets.
// Credited lists the players whose last honoured blot named this region —
// empty when the region tipped over on nobody's blot this turn, in which case
// the inking counts for neither side's metrics. TracksLost is indexed by the
// track sentinel: slots 0 and 1 are the players, slot 2 the contested rails
// nobody is credited for.
type InkData struct {
	Zone       int    `json:"zone"`
	Credited   []int  `json:"credited,omitempty"`
	TracksLost [3]int `json:"tracksLost"`
}

// ScoreData is emitted once per connection a player is paid for, every turn
// the connection stands. Points is one per rail the player owns along the
// path, PathLength counts the whole path including both towns, and Detour is
// how much longer that path is than the straight-line distance between them.
//
// Connections are directional: town A wanting B and town B wanting A pay out
// separately, so a mutual pair yields two events per turn.
type ScoreData struct {
	From       int `json:"from"`
	To         int `json:"to"`
	Points     int `json:"points"`
	PathLength int `json:"pathLength"`
	Detour     int `json:"detour"`
}

// FailedData is emitted for every command the engine rejected without
// disqualifying the player — off-grid or town cells, inked regions, occupied
// cells, an exhausted budget, a duplicate autoplace. Reason is the message
// verbatim from the game summary, minus the player prefix and colour markers.
// Unparseable commands are not reported here: they deactivate the player and
// surface through the trace's endReason instead.
type FailedData struct {
	Reason string `json:"reason"`
}

// TurnSummaryData closes out every turn for every player, including a
// deactivated one. It is the per-turn ledger the individual events add up to,
// so an analyzer can read a turn without replaying its event stream.
//
// Paint and disruption points are granted fresh each turn and do not carry
// over, so PaintLeft is what the player wasted rather than what it saved.
type TurnSummaryData struct {
	Turn           int `json:"turn"`
	PaintAvailable int `json:"paintAvailable"`
	PaintSpent     int `json:"paintSpent"`
	// PaintLeft is read off the player's live budget rather than derived from
	// the two fields above, so a spend the trace does not see would show up
	// here as the three no longer adding up.
	PaintLeft        int `json:"paintLeft"`
	TracksPlaced     int `json:"tracksPlaced"`
	TracksAutoplaced int `json:"tracksAutoplaced"`
	DisruptAvailable int `json:"disruptAvailable"`
	DisruptSpent     int `json:"disruptSpent"`
	// ConnectionsActive counts the connections that paid this player this
	// turn, not every connection standing on the board: a connection whose
	// path crosses none of the player's rails is not counted.
	ConnectionsActive int `json:"connectionsActive"`
	PointsEarned      int `json:"pointsEarned"`
	// Score is the player's running total after this turn, before the
	// end-of-game verdict OnEnd may overwrite it with.
	Score int `json:"score"`
}

// turnStats accumulates the per-turn ledger as the events are emitted, so the
// summary event and the individual events cannot disagree.
type turnStats struct {
	paintAvailable   int
	paintSpent       int
	tracksPlaced     int
	tracksAutoplaced int
	disruptAvailable int
	disruptSpent     int
	connections      int
	points           int
}

// trace appends a per-player event to the turn buffer. Out-of-range indices
// are dropped: the game is always two-player, so one can only come from a
// caller bug and must not panic mid-match.
func (g *Game) trace(playerIdx int, t arena.TurnTrace) {
	if playerIdx < 0 || playerIdx >= len(g.Traces) {
		return
	}
	g.Traces[playerIdx] = append(g.Traces[playerIdx], t)
}

// traceBoth mirrors a cross-owner event into both players' buckets.
func (g *Game) traceBoth(t arena.TurnTrace) {
	for i := range g.Traces {
		g.Traces[i] = append(g.Traces[i], t)
	}
}

// ResetTraces clears the turn buffer. Called from ResetGameTurnData, which the
// runner runs before command parsing, so the MESSAGE events parsing emits land
// in the same turn as the events PerformGameUpdate emits.
func (g *Game) ResetTraces() {
	g.Traces = [2][]arena.TurnTrace{}
}

// TurnTraces returns per-player copies, so the runner owns the slices it takes
// independently of the next turn's churn.
func (g *Game) TurnTraces() [2][]arena.TurnTrace {
	var out [2][]arena.TurnTrace
	for i, bucket := range g.Traces {
		if len(bucket) == 0 {
			continue
		}
		out[i] = make([]arena.TurnTrace, len(bucket))
		copy(out[i], bucket)
	}
	return out
}

// emitTurnSummaries closes the turn for both players. Runs last in
// PerformGameUpdate so scores and leftover budgets are the settled ones.
func (g *Game) emitTurnSummaries() {
	for _, p := range g.Players {
		idx := p.GetIndex()
		stats := g.turnStats[idx]
		g.trace(idx, arena.MakeTurnTrace(TraceTurnSummary, TurnSummaryData{
			Turn:              g.Turn,
			PaintAvailable:    stats.paintAvailable,
			PaintSpent:        stats.paintSpent,
			PaintLeft:         p.GetDosh(),
			TracksPlaced:      stats.tracksPlaced,
			TracksAutoplaced:  stats.tracksAutoplaced,
			DisruptAvailable:  stats.disruptAvailable,
			DisruptSpent:      stats.disruptSpent,
			ConnectionsActive: stats.connections,
			PointsEarned:      stats.points,
			Score:             p.GetScore(),
		}))
	}
}

// terrainName labels a cell for the TRACK event with the terrain kind that
// decided its cost.
func terrainName(t *Tile) string {
	switch t.GetType() {
	case TYPE_WATER:
		return "RIVER"
	case TYPE_MOUNTAIN:
		return "MOUNTAIN"
	case TYPE_POI:
		return "POI"
	}
	return "PLAINS"
}
