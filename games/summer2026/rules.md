# Summer Challenge 2026 — Back Track King

Statement as published for **Bronze**, the first league with no special
objective — the full game scores. Sections below the horizontal rule are
derived from the local engine dump, not from the contest page.

## Goal

Score more points than your opponent by drawing railway lines between towns
while sabotaging your opponent's efforts.

## Rules

In this game both players use paint to draw train tracks on a magic map.
Connecting towns on the map will bring prosperity to your own world.

The map is represented in the game by a **grid**.

### 🗺️ Map

The grid is made up of cells that can have one of three types:

| Type | Terrain   |
| ---: | --------- |
|  `0` | plains    |
|  `1` | river     |
|  `2` | mountains |

The grid is partitioned into **regions**. Each region is made up of multiple
contiguous cells and has a unique `regionId`. Regions are susceptible to
disruption by players — see [💥 Disruption](#-disruption).

Some regions will contain a town. Towns can only be found on plain cells.

### 🏯 Towns

Each game starts with multiple towns placed randomly across the map. There will
only be one per region, and no two regions sharing a border will both contain a
town.

Each town has a unique `townId`.

Each town has a list of `desiredConnections`: town ids representing all the
other towns this town would like to be connected to via train tracks placed by
players. Providing a town with train tracks connecting it to a desired town is
how players score points.

Desired connections are **unilateral** — if town `0` spawns with a desired
connection to town `1`, town `1` will not want to connect to town `0`.

A town can have zero `desiredConnections`, but will always be the subject of at
least one other town's `desiredConnections`.

### 🛤️ Placing train tracks

Players are given **3 paint points** each turn to place train tracks on the
map. ⚠️ These points do not carry over to the next turn and are lost if left
unused.

It costs:

| Terrain   | Cost |
| --------- | ---: |
| plains    |    1 |
| river     |    2 |
| mountains |    3 |

A track's owner is the `playerId` (`0`–`1`) of the player that placed it. They
will be the same colour. If both players place a track on the same turn at the
same location, the track's owner will be `2`, indicating a **neutral** track
piece.

A train track cannot be placed on a town or on an existing track.

Once placed, a train track will automatically connect to other tracks and towns
orthogonally adjacent to it.

### 🏯🛤️🏯 Connections

For each pair of towns in which one has the other in its `desiredConnections`,
if at least one path between the two exists, the **shortest** such path becomes
the **active connection** between those towns.

A path is an uninterrupted sequence of orthogonally adjacent cells with a train
track or a town.

If there are multiple shortest paths, the chosen path will always prioritise
the direction in the following order when moving from the requesting town to
the desired connected town:

1. `NORTH`
2. `EAST`
3. `SOUTH`
4. `WEST`

At the end of every turn, each active connection provides **1 point to each
player for every track they own in the path**.

**Example 1** — there is an active connection from town 0 to town 1 and one
from town 0 to town 2. Both the red and the blue players gain 3 points for the
0-1 connection and 4 points for the 0-2 connection at the end of the turn.

**Example 2** — only the shortest path from 0 to 2 is used for connection 0-2,
so the red player gains 3 points for that connection and the blue player gets
none. There are two equally short paths from 0 to 1, but since `EAST` has a
higher priority than `SOUTH`, the chosen path goes through town 2: the red
player gains 4 points for that connection, and the blue player only gains 1.

### 💥 Disruption

Similarly to paint points, players also get **1 disruption point per turn**. It
can be used to tamper with the map, giving you an edge over your opponent.
These points are not retained between turns either.

Players may spend their disruption point each turn to increase the
`instability` of any region by `1`.

Once a region's instability reaches `4`, that region is **inked out**, washing
any placed train tracks away and rendering any future placements on it
impossible. Any active connections via this region will be severed.

It is not possible to disrupt a region that is already inked out.

### 🎬 Actions

Each turn, players must provide at least one action on the standard output.
Actions must be separated by a semicolon `;` and be one of the following:

- `PLACE_TRACKS x y` — place a track on a free cell.
- `AUTOPLACE fromX fromY toX toY` — automatically generates a list of actions
  for the cheapest path from `from` to `to` in terms of paint points. Does
  nothing if a path already exists. The generated actions replace this command.
- `DISRUPT regionId` — increase the instability of a region. `DISRUPT x y` also
  works, to target the region `(x,y)` is part of.
- `WAIT` — do nothing.

### 🏆 Victory conditions

- Have the most points after 100 turns.
- Be in the lead if all desired connections become impossible to fulfil.

### ⛔ Defeat conditions

- Your program does not provide a command in the allotted time, or one of the
  commands is invalid.

## Technical details

All `PLACE_TRACKS` actions, including those generated by `AUTOPLACE`, are
performed **before** `DISRUPT` actions. Points are scored at the very end of a
turn, **after** inking out unstable regions.

Commands that are impossible actions are skipped. If an impossible action is
part of an `AUTOPLACE`, the rest of the generated actions are skipped even if
they are possible.

### 🐞 Debugging tips

- Hover over the grid to see extra information on the cell under your mouse.
- Press the gear icon on the viewer to access extra display options.
- Use the keyboard to control the action: space to play/pause, arrows to step 1
  frame at a time.

## Game Protocol

### Initialization input

- **Line 1:** `myId` — your player id, `0` or `1`.
- **Line 2:** `width` — number of cells in a row of the map.
- **Line 3:** `height` — number of cells in a column of the map.
- **Next `height * width` lines:** two integers describing each cell of the
  map, from left to right, top to bottom:
  - `regionId` — the id of the region this cell is a part of.
  - `type` — the terrain type of this cell (`0`–`2`).
- **Next line:** `townCount` — number of towns on the map.
- **Next `townCount` lines:**
  - `townId` — unique identifier of this town.
  - `townX` — X position of this town (`0` is left-most).
  - `townY` — Y position of this town (`0` is top-most).
  - `desiredConnections` — a string of comma-separated `townId`s, e.g. `1,2,4`,
    or `x` if this town has no desired connections.

### Input for one game turn

- **Line 1:** `myScore` — your points.
- **Line 2:** `foeScore` — your opponent's points.
- **Next `height * width` lines:** state of each cell, in the same order they
  were given before:
  - `trackOwner` — `-1` if this cell has no track, `0` / `1` if that player
    owns a track on this cell, `2` if there is a neutral track on this cell.
  - `instability` — the instability of the region this cell is a part of.
  - `inked` — `1` if the region has been inked out (instability ≥ `4`), `0`
    otherwise.
  - `partOfActiveConnections` — a string of comma-separated `townId` pairs
    indicating this cell is part of an active connection between those two
    towns, e.g. `1-2,1-3,4-7`. `x` if this cell is not part of any active
    connection.

### Output

A single line containing at least one action and at most a single `AUTOPLACE`
action. All actions must be separated with a semicolon `;` and be one of the
following:

- `PLACE_TRACKS x y` — followed by the coordinates of the desired location.
- `AUTOPLACE fromX fromY toX toY` — followed by two pairs of coordinates, to
  create the cheapest path between the two.
- `DISRUPT regionId` — followed by the `regionId` of the region you wish to
  disrupt. Replace `regionId` by `x y` coordinates to target the region at that
  location.
- `MESSAGE text` — to be displayed in the viewer.
- `WAIT`

### Constraints

- Response time for the first turn ≤ **1000 ms**.
- Response time per turn ≤ **50 ms**.
- `21 ≤ width ≤ 30`
- `14 ≤ height ≤ 20`
- `4 ≤ townCount ≤ 12`

---

The rest of this document is derived from the local engine dump, not from the
contest page. It is provisional until the official source is published — see
[README.md](README.md).

## Constants reference

From `com/codingame/game/Game.java`:

| Name                                        | Value  |
| ------------------------------------------- | ------ |
| `MIN_GRID_HEIGHT` / `MAX_GRID_HEIGHT`       | 14 / 20 |
| `ASPECT_RATIO`                              | 1.5f   |
| `MIN_TOWN_DISTANCE`                         | 4      |
| `AVERAGE_TILES_PER_ZONE_COEFF_TO_GRID_HEIGHT` | 2    |
| `AVERAGE_TILES_PER_TOWN`                    | 50     |
| `TRAIN_TO_TOWN_RATIO`                       | 0.4f   |
| `RIVER_SPLIT_PROBA`                         | 0.055f |
| `RIVER_TO_LAND_MIN_RATIO`                   | 0.07f  |
| `MIN_RIVER_LENGTH`                          | 3      |
| `MIN_MOUNTAINS`                             | 2      |
| `MOUNTAIN_TO_CELL_RATIO`                    | 0.04f  |
| `BASE_RAIL_COST`                            | 1      |
| `GRASS_COST_MULTIPLIER`                     | 1      |
| `RIVER_COST_MULTIPLIER`                     | 2      |
| `MOUNTAIN_COST_MULTIPLIER`                  | 3      |
| `POI_COST_MULTIPLIER`                       | 3      |
| `PASSIVE_INCOME`                            | 3      |
| `STARTING_DOSH`                             | 0      |
| `BLOT_POINTS_PER_TURN`                      | 1      |
| `MAX_TURNS`                                 | 100    |
| `INSTABILITY_THRESHOLD_BASE`                | 4      |
| `INSTABILITY_THRESHOLD_INCREASE`            | 0      |

`Tile` type / ownership sentinels: `TYPE_GRASS = 0`, `TYPE_WATER = 1`,
`TYPE_MOUNTAIN = 2`, `TYPE_POI = 3`; `TRACK_NONE = -1`, `TRACK_NEUTRAL = 2`,
`TOWN_NONE = -1`.

## Turn order

`Game.performGameUpdate` runs the following, in order:

1. `doIncome` — reset each player's paint points to `PASSIVE_INCOME` (3) and
   disruption points to `BLOT_POINTS_PER_TURN` (1). Nothing carries over.
2. `computeAutobuilds` — expand each `AUTOPLACE` into concrete `PLACE_TRACKS`
   actions. Only the first `AUTOPLACE` per turn is honoured; the rest are
   reported as errors.
3. `doActions` — validate and apply track placements for player 0 then player
   1, then apply disruptions in the same player order.
4. `doInstabilityCheck` — ink out regions whose instability reached the
   threshold, clearing every track in them.
5. `moveTrains` — recompute every active connection and award points.
6. `computeTileStates` — refresh each cell's `partOfActiveConnections`.
7. `checkSideQuest` — records side-quest completion when the map has a POI.
8. End check — `isGameOver()`.

## Source-vs-statement notes

- **Turn cap.** `Game.MAX_TURNS` is `100`, matching the statement.
  `Referee.init` separately calls `gameManager.setMaxTurns(400)`, which is the
  SDK's safety ceiling, not the game's rule.
- **`PLACE_TRACK` is also accepted.** The `ActionType` regex is
  `^PLACE_TRACKS? (?<x>\d+) (?<y>\d+)`, so the singular spelling works even
  though the statement only documents `PLACE_TRACKS`. All command patterns are
  matched case-insensitively.
- **`DISRUPT` accepts two forms.** `DISRUPT zoneId` and `DISRUPT x y` (which
  resolves the region from the cell). Both exist in the engine at every league;
  the Bronze statement documents both.
- **Regions containing towns are protected.** Despite the statement's "any
  region" wording, the source rejects disruptions of regions containing a
  town. The rejected action does not spend a disruption point.
- **Invalid actions do not always disqualify.** An unparseable command
  disqualifies the player (score `-1`). A parseable but illegal action —
  placing off-grid, on a town, on an existing track, in an inked region, or
  without enough paint points — is only reported to the game summary and
  skipped. An `AUTOPLACE`-generated action that runs out of paint points also
  interrupts the remainder of that autobuild. Unlike the broader wording in
  the statement, other illegal generated placements do not interrupt it in
  the source. Later manual actions are still processed after an interruption.
- **Both players may claim the same cell.** `isFreeOfTracks` permits each
  player to target a cell no one owns yet; if both do so on the same turn the
  cell becomes neutral (`2`) and both paid.
- **Game also ends early when no connection is possible.** `isGameOver`
  returns true when a `TerrainAStar` search finds no remaining route for any
  desired connection, matching the Bronze statement. This search considers
  terrain that could still receive tracks, not only existing tracks.
- **Historical replays can contain POIs.** The current default disables the
  side quest, but online replay [901942746](https://www.codingame.com/replay/901942746)
  has POIs at `(18, 3)` and `(18, 6)`, whereas
  [902038382](https://www.codingame.com/replay/902038382) has none. The factory
  enables the historical mode when replay metadata contains a
  `sideQuestPoints_0` or `sideQuestPoints_1` key, including a zero value.
  POIs cost 3 paint and can change `AUTOPLACE` tie ordering even when they are
  outside the winning path. The existing A* and Java priority queue reproduce
  both versions without changes. Replay normalization retains `metadata` for
  this purpose; old files that discarded it must be re-fetched to recover the
  historical mode. Missing metadata uses the current default.
- **Fallback town placement diverges from the statement.** The primary loop in
  `GridMaker.makeTowns` enforces "plains only, not on an edge, at least
  `MIN_TOWN_DISTANCE` from other towns". The fallback loop that runs when the
  primary loop could not place every town enforces only the distance rule, and
  re-numbers towns from `0` — so it can place a town on a river or mountain
  cell and can produce duplicate `townId`s. Reachable only when the primary
  loop exhausts its 100 retries.
- **Map generation depends on Java `HashSet` iteration order.**
  `GridMaker.getAvailableNeighbours` collects into a `HashSet<Coord>` and the
  result is then indexed by `random.nextInt(size)` for both region growth and
  mountain growth. Seed parity therefore requires reproducing
  `java.util.HashMap` bucket ordering, not just the RNG stream.
- **RNG is SHA1PRNG.** Seeding goes through
  `SecureRandom.getInstance("SHA1PRNG")`. Generation uses `nextInt(bound)`,
  `nextInt(origin, bound)`, `nextFloat()`, `nextFloat(bound)`, `nextBoolean()`
  and `Collections.shuffle(list, random)`.
