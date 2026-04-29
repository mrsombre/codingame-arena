package engine

import (
	"strconv"
	"strings"

	"github.com/mrsombre/codingame-arena/internal/arena"
)

// Trace labels emitted per turn, attached to trace output for replay viewers.
const (
	TraceEat     = "EAT"
	TraceKilled  = "KILLED"
	TraceSpeed   = "SPEED"
	TraceSwitch  = "SWITCH"
)

func tracePacPayload(pacID int) string {
	return strconv.Itoa(pacID)
}

func traceEatPayload(pacID int, c Coord, value int) string {
	return strconv.Itoa(pacID) + " " + strconv.Itoa(c.X) + "," + strconv.Itoa(c.Y) + " " + strconv.Itoa(value)
}

func traceKilledPayload(deadID int, c Coord, killerID int) string {
	return strconv.Itoa(deadID) + " " + strconv.Itoa(c.X) + "," + strconv.Itoa(c.Y) + " " + strconv.Itoa(killerID)
}

func traceSwitchPayload(pacID int, t PacmanType) string {
	return strconv.Itoa(pacID) + " " + t.Name()
}

func (g *Game) trace(label, payload string) {
	g.traces = append(g.traces, arena.TurnTrace{Label: label, Payload: payload})
}

// parseLeadingPacID extracts the first whitespace-delimited integer from a
// trace payload. Every Spring 2020 trace payload starts with the subject pac's
// global ID so the runner can bucket events per pac in the match summary.
func parseLeadingPacID(payload string) (int, bool) {
	head, _, _ := strings.Cut(payload, " ")
	n, err := strconv.Atoi(head)
	if err != nil {
		return 0, false
	}
	return n, true
}
