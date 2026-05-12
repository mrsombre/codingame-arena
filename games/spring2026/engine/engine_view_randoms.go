// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/view/BoardView.java + view/BirdView.java
package engine

// consumeBoardViewRandoms mirrors the exact sequence of RNG draws Java's
// view.BoardView constructor performs while initializing background art,
// decor, and the random "creator" Easter eggs (frog / fish / bird+cat /
// turtle). The Java referee runs this inside engine.Board.createMap right
// after isValid() succeeds (engine.Board.initView → view.BoardView), so by
// the time gameplay begins the SHA1PRNG has advanced past these draws.
//
// We don't render anything in Go; we just consume the same number of
// NextInt / NextIntRange / NextDouble calls in the same order. Skipping
// this step leaves the RNG offset from the official referee — every later
// MoveTask tie-break in getNextCell drifts and silently mismatches live
// replays.
//
// Layout constants come straight from BoardView.java.
const (
	viewSpriteSize  = 72
	viewSpriteR     = viewSpriteSize - 66 // 6  — rock sprite jitter range
	viewSpriteGD    = viewSpriteSize - 16 // 56 — grass / decor jitter range
	viewExplosionMaxX  = 1840
	viewExplosionMinY  = 120
	viewExplosionMaxY  = 1000
)

// consumeBoardViewRandoms drains the same RNG sequence Java's BoardView
// constructor would draw on this board. Called once, immediately after
// CreateMap validates.
func (b *Board) consumeBoardViewRandoms() {
	// view.BoardView.java:23 — explosionContainer placement (one-off).
	b.Random.NextInt(viewExplosionMaxX)
	b.Random.NextIntRange(viewExplosionMinY, viewExplosionMaxY)

	// creatorsPlaced[5] = {frog, fish, bird, cat, turtle}.
	var creatorsPlaced [5]bool

	// view.BoardView.java:57-172 — iterate cells in column-major order (x
	// outer, y inner), exactly matching Java.
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			cell := b.Grid[x][y]

			// :63-64 — ground sprite jitter for any non-water tile.
			if cell.Type != CellWATER {
				b.Random.NextInt(8)
			}

			switch cell.Type {
			case CellROCK:
				// :66-68 — wall sprite + xy jitter.
				b.Random.NextInt(5)
				b.Random.NextInt(viewSpriteR)
				b.Random.NextInt(viewSpriteR)

			case CellGRASS:
				// :71-78 — bird+cat creator vs. plain decor. The first
				// `if` short-circuits the nextDouble unless every prior
				// clause holds; matching that short-circuit is critical
				// for RNG parity.
				canTryBird := !creatorsPlaced[1] &&
					!creatorsPlaced[2] &&
					cell.Plant != nil &&
					cell.Plant.Size == PLANT_MAX_SIZE
				birdPlaced := false
				if canTryBird {
					if b.Random.NextDouble() < 0.0005 {
						// view.BirdView constructor draws four ints:
						//   tooltip text for bird  (creatorsTexts[2] len 4)
						//   tooltip text for cat   (creatorsTexts[3] len 3)
						// followed by update(), which on a freshly placed
						// bird hits the `cell == tree.getCell()` branch and
						// draws two more (idle frame indices):
						//   nextInt(1,5)
						//   nextInt(1,3)
						b.Random.NextInt(4)
						b.Random.NextInt(3)
						b.Random.NextIntRange(1, 5)
						b.Random.NextIntRange(1, 3)
						creatorsPlaced[2] = true
						creatorsPlaced[3] = true
						birdPlaced = true
					}
				}
				// :75 — `else if (random.nextDouble() < 0.3)` runs whenever
				// the first `if` was false (including when its prefix
				// short-circuited and the bird-roll was skipped).
				if !birdPlaced {
					if b.Random.NextDouble() < 0.3 {
						b.Random.NextInt(16)
						b.Random.NextInt(viewSpriteGD)
						b.Random.NextInt(viewSpriteGD)
					}
				}

			case CellWATER:
				// :81-86 — compute the 0..15 sprite index from the 4
				// cardinal neighbours. Java's dir order is 0=up, 1=right,
				// 2=down, 3=left; the contribution weights below mirror it.
				index := 0
				if n := cell.Neighbors[0]; n != nil && n.Type != CellWATER {
					index += 4
				}
				if n := cell.Neighbors[1]; n != nil && n.Type != CellWATER {
					index += 1
				}
				if n := cell.Neighbors[2]; n != nil && n.Type != CellWATER {
					index += 8
				}
				if n := cell.Neighbors[3]; n != nil && n.Type != CellWATER {
					index += 2
				}
				empty := index == 0
				// :87 — pick a random "deep water" variant for fully
				// surrounded tiles.
				if empty {
					b.Random.NextInt(5)
				}

				// :100-140 — empty-tile decor chain: 30% chance to roll
				// again, then 0.1% chance to spawn a creator (frog or
				// fish) and 99.9% chance to drop a water decor. Each `<`
				// gate consumes the nextDouble unconditionally.
				if empty {
					if b.Random.NextDouble() < 0.3 {
						if b.Random.NextDouble() < 0.001 {
							creatorIndex := b.Random.NextInt(2)
							// :103-104 — fall back to the other slot if
							// the rolled creator is unavailable. The OR
							// is `(idx==1 && bird) || alreadyPlaced[idx]`
							// per Java's operator precedence.
							if (creatorIndex == 1 && creatorsPlaced[2]) ||
								creatorsPlaced[creatorIndex] {
								creatorIndex = 1 - creatorIndex
							}
							// :105 — final guard before tooltip text.
							if !creatorsPlaced[creatorIndex] &&
								(creatorIndex == 0 || !creatorsPlaced[2]) {
								// creatorsTexts[0] (frog) has length 4;
								// creatorsTexts[1] (fish) has length 5.
								if creatorIndex == 0 {
									b.Random.NextInt(4)
								} else {
									b.Random.NextInt(5)
								}
								creatorsPlaced[creatorIndex] = true
							}
						} else {
							// :137 — water decor sprite variant.
							b.Random.NextInt(4)
						}
					}
				}

				// :141-155 — turtle on a vertical-wall water tile
				// (index == 8, one-sided edge against land).
				if index == 8 && !creatorsPlaced[4] {
					if b.Random.NextDouble() < 0.0005 {
						// creatorsTexts[4] (turtle) has length 1, so the
						// tooltip draw is nextInt(1) — always 0, but
						// still consumes one RNG step.
						b.Random.NextInt(1)
						creatorsPlaced[4] = true
					}
				}
			}
		}
	}
}
