// Package action
// Source: WinterChallenge2026-Exotec/src/main/java/com/codingame/game/action/ActionType.java
package action

import (
	"regexp"
	"strconv"

	"github.com/mrsombre/codingame-arena/games/winter2026/engine/grid"
)

// ActionType identifies the kind of action.
type ActionType int

const (
	TypeMoveUp ActionType = iota
	TypeMoveDown
	TypeMoveLeft
	TypeMoveRight
	TypeMark
	TypeWait
)

// pattern holds a compiled regex and a populate function for one action type.
type pattern struct {
	actionType ActionType
	re         *regexp.Regexp
	populate   func(match []string, groups map[string]int, a *Action)
}

var patterns []pattern

func init() {
	movePopulate := func(dir grid.Direction) func([]string, map[string]int, *Action) {
		return func(match []string, groups map[string]int, a *Action) {
			a.BirdID, _ = strconv.Atoi(match[groups["birdId"]])
			a.HasBirdID = true
			a.Direction = dir
			if idx, ok := groups["messageGroup"]; ok && idx < len(match) && match[idx] != "" {
				messageIdx := groups["message"]
				a.Message = match[messageIdx]
				a.HasMessage = true
			}
		}
	}

	movePattern := func(dirName string) string {
		return `(?i)^(?P<birdId>\d+) ` + dirName + `(?P<messageGroup> (?P<message>[^;]*))?`
	}

	patterns = []pattern{
		{TypeMoveUp, regexp.MustCompile(movePattern("UP")), movePopulate(grid.DirNorth)},
		{TypeMoveDown, regexp.MustCompile(movePattern("DOWN")), movePopulate(grid.DirSouth)},
		{TypeMoveLeft, regexp.MustCompile(movePattern("LEFT")), movePopulate(grid.DirWest)},
		{TypeMoveRight, regexp.MustCompile(movePattern("RIGHT")), movePopulate(grid.DirEast)},
		{TypeMark, regexp.MustCompile(`(?i)MARK (?P<x>\d+) (?P<y>\d+)`), func(match []string, groups map[string]int, a *Action) {
			x, _ := strconv.Atoi(match[groups["x"]])
			y, _ := strconv.Atoi(match[groups["y"]])
			a.Coord = grid.Coord{X: x, Y: y}
			a.HasCoord = true
		}},
		{TypeWait, regexp.MustCompile(`(?i)WAIT`), func([]string, map[string]int, *Action) {}},
	}
}

// subexpMap builds a name→index map for a compiled regexp.
func subexpMap(re *regexp.Regexp) map[string]int {
	m := make(map[string]int)
	for i, name := range re.SubexpNames() {
		if name != "" {
			m[name] = i
		}
	}
	return m
}
