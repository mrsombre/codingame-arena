# Summer Challenge 2026 — Back Track King

This is a league-based challenge. Multiple leagues are available for the
same game. Once you have proven your skills against the first Boss, you will
access a higher league and extra rules will become available.

In the first few leagues, your submission only fights the boss in the arena.
Win a best-of-five to advance.

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
only be one per region. Usually no two regions sharing a border will both
contain a town, and no town is on the map edge. On rare maps, where the
generator cannot place every town under these limits, it places the remaining
towns with only the distance limit: such a town can be in a region next to
another town's region, or on the map edge.

Two towns are always at least `4` cells apart (Manhattan distance).

Each town has a unique `townId`, from `0` to `townCount - 1`.

Each town has a list of `desiredConnections`: town ids representing all the
other towns this town would like to be connected to via train tracks placed by
players. Providing a town with train tracks connecting it to a desired town is
how players score points.

Desired connections are **unilateral** — if town `0` spawns with a desired
connection to town `1`, town `1` will not want to connect to town `0`.

A town can have zero `desiredConnections`, and a town can be missing from the
`desiredConnections` of every other town. Every town is part of at least one
desired connection, in one direction or the other.

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
piece. Both players pay the cost of that cell.

A train track cannot be placed on a town, on an existing track, on a cell that
the same player already placed on this turn, or in an inked-out region.

Once placed, a train track will automatically connect to other tracks and towns
orthogonally adjacent to it.

### 🏯🛤️🏯 Connections

For each pair of towns in which one has the other in its `desiredConnections`,
if at least one path between the two exists, the **shortest** such path becomes
the **active connection** between those towns.

A path is an uninterrupted sequence of orthogonally adjacent cells with a train
track or a town. Tracks of both players, neutral tracks and other towns can all
be part of a path.

If there are multiple shortest paths, the chosen path will always prioritise
the direction in the following order when moving from the requesting town to
the desired connected town:

1. `NORTH`
2. `EAST`
3. `SOUTH`
4. `WEST`

At the end of every turn, each active connection provides **1 point to each
player for every track they own in the path**. Neutral tracks and towns give no
points. Active connections are computed again every turn, and points add up
over the game.

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
`instability` of any region by `1`. Both players can disrupt the same region on
the same turn.

Once a region's instability reaches `4`, that region is **inked out**, washing
any placed train tracks away and rendering any future placements on it
impossible. Any active connections via this region will be severed.

It is not possible to disrupt a region that is already inked out.

Regions with a town cannot be disrupted.

A `DISRUPT` action that is skipped does not use the disruption point.

### 🎬 Actions

Each turn, players must provide at least one action on the standard output.
Actions must be separated by a semicolon `;` and be one of the following:

- `PLACE_TRACKS x y` — place a track on a free cell.
- `AUTOPLACE fromX fromY toX toY` — automatically generates a list of
  `PLACE_TRACKS` actions for the cheapest path from `from` to `to` in terms of
  paint points. The generated actions replace this command, at the same
  position in the action list.
- `DISRUPT regionId` — increase the instability of a region. `DISRUPT x y` also
  works, to target the region `(x,y)` is part of.
- `MESSAGE text` — display text in the viewer.
- `WAIT` — do nothing.

`AUTOPLACE` path search:

- The search uses the map as it is at the start of the turn. It ignores
  actions given on the same turn.
- Existing tracks of any owner and towns cost `0`. Other cells cost their
  terrain cost.
- The path never goes through a cell of an inked-out region.
- If `from` is on a track or a town, the path can start from any cell of the
  network of tracks and towns connected to `from`.
- If `from` and `to` are already connected by tracks and towns, no action is
  generated.
- If `from` or `to` is outside the grid, or no path exists, no action is
  generated.
- The search does not limit the path to the paint points of the turn.

### 🏆 Victory conditions

- Have the most points after 100 turns.
- Be in the lead when all desired connections become impossible to fulfil. A
  desired connection is possible while a route of orthogonally adjacent cells
  outside inked-out regions exists between its two towns. The route can cross
  any terrain, with or without tracks. The game checks this at the end of each
  turn, after scoring.

Equal points at the end of the game give a draw.

### ⛔ Defeat conditions

- Your program does not provide a command in the allotted time, or one of the
  commands is invalid.

## Technical details

The game's source code is available on GitHub.

A command is invalid, and its player is disqualified, when it matches none of
the action formats:

- Keywords are case-insensitive. The keyword is `PLACE_TRACKS`; `PLACE_TRACK`
  is invalid.
- Words in a command are separated by exactly one space. Spaces before and
  after each command are ignored.
- `x`, `y`, `fromX`, `fromY`, `toX`, `toY` and `regionId` must be non-negative
  integers.
- An empty command is invalid, for example an empty line, `;WAIT` or
  `WAIT;;WAIT`. Semicolons at the end of the line are ignored.
- `MESSAGE` must be followed by a space and text. The text cannot contain `;`.

A turn runs in this order:

1. Each player gets 3 paint points and 1 disruption point.
2. Each `AUTOPLACE` is replaced by its generated actions.
3. `PLACE_TRACKS` actions are performed: all actions of player 0 in order, then
   all actions of player 1 in order. Ownership of each cell is decided after
   both players, so the order between players does not matter.
4. `DISRUPT` actions are performed: player 0, then player 1.
5. Regions with instability `4` or more are inked out.
6. Active connections are computed and points are scored.
7. The game ends if 100 turns are played or no desired connection is possible.

Commands that are impossible actions are skipped, and the player is not
disqualified:

- `PLACE_TRACKS` outside the grid, on a town, on an existing track, on a cell
  the same player already placed on this turn, in an inked-out region, or
  without enough paint points.
- `DISRUPT` with a `regionId` that does not exist or with `x y` outside the
  grid, on an inked-out region, on a region with a town, or without a
  disruption point left.
- Each `AUTOPLACE` after the first one on the same turn.

If a generated action of an `AUTOPLACE` is skipped because of not enough paint
points, the rest of the generated actions are skipped even if they are
possible. A generated action skipped for any other reason does not stop the
rest. Actions that are not generated by `AUTOPLACE` are always performed or
skipped one by one.

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
  - `desiredConnections` — a string of comma-separated `townId`s in ascending
    order, e.g. `1,2,4`, or `x` if this town has no desired connections.

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
    towns, e.g. `1-2,1-3,4-7`. In each pair the requesting town comes first.
    The pairs are sorted as strings, so `10-2` comes before `2-3`. `x` if this
    cell is not part of any active connection.

### Output

A single line containing at least one action. All actions must be separated
with a semicolon `;` and be one of the following:

- `PLACE_TRACKS x y` — followed by the coordinates of the desired location.
- `AUTOPLACE fromX fromY toX toY` — followed by two pairs of coordinates, to
  create the cheapest path between the two. Only the first `AUTOPLACE` of a
  turn is used.
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
