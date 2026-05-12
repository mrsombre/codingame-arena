# Java Engine Porting Tricks

Gotchas we hit on real ports that aren't obvious from reading the Java sources
in `source/`. Each entry is a concrete bug + the diagnostic that found it +
the fix shape, so the next port skips the dig.

## Viewer Code Shares the Gameplay RNG

The CodinGame SDK plumbs a single `Random` instance through the whole
`Referee` lifecycle. The map-gen passes, the view-init passes, and the
gameplay tie-breaks all draw from the same stream — and the official referee
calls all three in this order:

1. `Board.createMap(...)` — terrain, shacks, plants
2. `board.initView(...)` — sprites, decor, background art, Easter eggs
3. Per-turn gameplay — `Board.getNextCell` tie-breaks etc.

Step 2 looks like dead viewer code, so the natural instinct is to skip it.
**Don't.** Even though Go doesn't render anything, `BoardView`-equivalent
constructors draw a non-trivial number of `nextInt` / `nextDouble` calls
while iterating cells. Skipping that step leaves your RNG state shifted
left from the Java referee's, and every later tie-break drifts. Bots match
on early turns by coincidence and diverge once a tie-break path with a
different RNG output shows up.

Concrete instance — spring2026: `view.BoardView`'s constructor iterates
every cell to drop ground / wall / decor sprites and rolls (very low
probability) "creator" Easter eggs (frog / fish / bird+cat / turtle).
Final score on match `887011495` was off by 1 point because the unit
`MOVE 1 8 2` from `(9,1)` resolved to `(9,2)` in Go but `(8,1)` in Java —
the 5th gameplay RNG call diverged.

**Symptoms in a replay trace mismatch**

- Final scores off by small amounts (one or two points) on a subset of
  matches but not all.
- `analyze` shows divergent MOVE / tie-break outcomes starting mid-game
  rather than turn 1 — coincidental agreement on power-of-two tie-breaks
  is common at first.
- `make test-arena` and seed-parity tests are green (map terrain is
  identical) but live-replay verification mismatches.

**How to diagnose**

1. Capture the actual bot outputs from the failing replay and replay them
   through Go (`bin/arena replay <game> <user> <id> -f`). Identify the
   first turn where the engine's resolved move disagrees with the
   replay summary text.
2. Instrument `getNextCell` to log `(turn, closest[], idx, picked)` per
   call. Confirm the closest list is identical to what Java would build
   for that grid (column-major iteration, same dist values).
3. Probe raw SHA1PRNG output side-by-side with Java via a small
   `SecureRandom.getInstance("SHA1PRNG")` Java program (see
   `tmp/RngProbe.java` if it survived a previous debug pass) — confirm
   bit-for-bit parity of the underlying RNG.
4. Run the actual Java `Board.createMap` with a counting Random wrapper
   (the SDK modules can be passed `null`; you'll catch an NPE inside
   `initView` AFTER all `createMap`-proper RNG draws have already
   happened). Compare the counter to your Go side.
5. If the createMap counter matches but post-mapgen `nextInt(2)`
   sequences disagree between Java and Go, the difference is in
   `initView`. Read `source/.../view/BoardView.java` (and any view
   classes it constructs — `BirdView`, `BirdView.update()` etc.) and
   port the RNG-consuming control flow.

**Fix shape**

Port the view code's RNG-consuming control flow into Go as a side-effect-
free pass. Iterate the cells in the same order Java does (column-major).
Match the conditions exactly. For every `nextInt` / `nextIntRange` /
`nextDouble` in the Java view, draw the equivalent in Go. Call it once
right after the engine's `createMap` validates, before returning the
board. Reference: `games/spring2026/engine/engine_view_randoms.go` and
the `consumeBoardViewRandoms()` call site in
`games/spring2026/engine/engine_board.go`.

Watch out for:

- **Short-circuit evaluation.** Java `&&` chains like
  `if (a && b && random.nextDouble() < 0.0005)` skip the `nextDouble`
  whenever `a` or `b` is false. The `else if (random.nextDouble() < 0.3)`
  on the next line runs in that case too. Map both arms faithfully — a
  branchless `nextDouble` draw is a parity bug.
- **`nextDouble` is two RNG draws.** Java's `Random.nextDouble()` consumes
  `next(26)` + `next(27)`, not one 32-bit word. Both `internal/util/javarand`
  and `internal/util/sha1prng` expose `NextDouble()` with the same shape;
  use it (do not roll your own `float64(NextInt(...))` substitute).
- **`nextInt(1)` still draws.** Java `nextInt(1)` always returns 0 but
  still advances the RNG via `next(31)`. Power-of-2 bounds (1, 2, 4, ...)
  all consume exactly one `next(31)`; non-power-of-2 bounds may consume
  more via rejection sampling. Don't optimize "the result is constant
  anyway" away — state advance is the point.
- **Constructed sub-views.** If a view constructor builds another view
  (e.g. `BirdView` from `BoardView`), follow the chain — the child
  constructor's draws, and any `update()` it calls eagerly, are all part
  of the state advance.
- **Sprite-text array lengths.** `random.nextInt(creatorsTexts[i].length)`
  draws depend on the literal string-array length in Java source. Read
  the arrays and hard-code the lengths in Go — they're not derived from
  game state.

**Empirical confirmation**

Once the port is in place, run the full saved replay set
(`bin/arena replay <game> <user> -f`) and confirm `0 saved-mismatch`.
A single match passing isn't enough — view-random consumption can be
sensitive to plant size distribution, water tile count, neighbor
patterns, etc., and a wrong port can pass on simple maps and fail on
others.
