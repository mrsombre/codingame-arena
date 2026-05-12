// Package games imports all engine packages for registration.
package games

import (
	_ "github.com/mrsombre/codingame-arena/games/spring2020/engine"
	_ "github.com/mrsombre/codingame-arena/games/spring2021/engine"
	_ "github.com/mrsombre/codingame-arena/games/winter2026/engine"
)

// Order is the canonical chronological list of games shipped with the
// arena. The CLI banner and `arena help` use this so newer challenges
// appear after older ones regardless of import or registration order.
// `arena game list` reports the live registry instead — use that to spot
// engines that registered but were left out of this slice.
var Order = []string{
	"spring2020",
	"spring2021",
	"winter2026",
}
