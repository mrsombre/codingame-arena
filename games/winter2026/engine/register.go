package engine

import "github.com/mrsombre/codingame-arena/internal/arena"

func init() {
	arena.Register(NewFactory())
}
