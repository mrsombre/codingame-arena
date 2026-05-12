# Spring Challenge 2026 — Troll Farm

## Goal

You control a pack of **trolls**. Make them collect the most resources and bring
them back to your **shack**.

## Rules

The game is played on a **grid**. Each player starts with one shack and one
troll. Each turn both players issue commands **simultaneously**.

### 🗺️ Map

The grid is rectangular with `width = 2 × height`.

- Leagues 1–2: `height = 8` (fixed).
- Leagues 3–4: `8 ≤ height ≤ 11` (random per game).

Cell types:

- `.` **GRASS** — walkable.
- `~` **WATER** — not walkable; boosts adjacent tree growth.
- `#` **ROCK** — not walkable.
- `+` **IRON** — not walkable; infinite mining resource (leagues 3–4 only).
- `0` your **SHACK** / `1` opponent's **SHACK** — not walkable.

The map is point-symmetric: every terrain feature placed at `(x, y)` is mirrored
to `(width - 1 - x, height - 1 - y)`. Both shacks, terrain, and starting trees
respect this symmetry.

Map generation guarantees:

- The cell next to the player's shack must be walkable.
- For league ≥ 3 there must be a walkable cell adjacent to at least one IRON.
- All walkable cells must be reachable from each other (BFS-connected).
- The shacks must be within `MAP_MAX_OPP_DIST = 6` walking distance of each
  other's neighborhood.
- Leagues 1–2: at least one tree must have at least one fruit at game start.

### 🧌 Troll

A troll has four immutable attributes:

- `movementSpeed` — max cells moved per turn.
- `carryCapacity` — max items carried at any time.
- `harvestPower` — max fruits taken per HARVEST.
- `chopPower` — max wood per CHOP, max iron per MINE (leagues 3–4).

### 🌳 Trees

Trees have a **type** (`PLUM`, `LEMON`, `APPLE`, `BANANA`), a **size** (1–4), a
**health**, a **fruit count** (0–3) and a **cooldown**.

Every turn the cooldown decreases. When it reaches `0`:

- If size `< 4`, the tree **grows** (size += 1, health increases by the size
  delta — see below).
- If size `= 4` and fruit count `< 3`, the tree **produces a fruit**.

Cooldown values per tree type:

| Type   | Cooldown | Cooldown near water |
| ------ | -------: | ------------------: |
| PLUM   |        8 |                   3 |
| LEMON  |        8 |                   3 |
| APPLE  |        9 |                   2 |
| BANANA |        6 |                   4 |

A tree is "near water" if any of its 4-neighbours is a WATER cell.

Health values per (type, size):

| Type   | Size 1 | Size 2 | Size 3 | Size 4 |
| ------ | -----: | -----: | -----: | -----: |
| PLUM   |      6 |      8 |     10 |     12 |
| LEMON  |      6 |      8 |     10 |     12 |
| APPLE  |     11 |     14 |     17 |     20 |
| BANANA |      3 |      4 |      5 |      6 |

When a tree grows, its current health gains the per-size delta (PLUM/LEMON +2,
APPLE +3, BANANA +1), so a damaged growing tree is not fully healed.

### ↪️ Movement (`MOVE id x y`)

A troll moves up to `movementSpeed` cells horizontally/vertically per turn
along a shortest walkable path toward `(x, y)`. Only `GRASS` cells are
walkable, and **each cell can hold at most one troll per team**.

If the target is unreachable or out of range, the troll moves to the nearest
reachable cell toward it. Among ties, an arbitrary cell is chosen (engine
uses RNG — this is a parity-sensitive point).

### 🍎 Harvesting (`HARVEST id`)

A troll on the same cell as a tree can take fruits, limited by the troll's
free `carryCapacity` and `harvestPower`. If two trolls (one per team) harvest
the same tree, they take fruits **one at a time, alternating**, until the
tree is empty or both reach a limit. **The last fruit may be duplicated** so
both trolls receive it.

### 🌱 Planting (`PLANT id type`) — leagues 2+

A troll on a GRASS cell can plant a new tree of `type` if it carries at least
one fruit of that type. Two trolls planting the same type on the same cell
both lose a seed and the tree is planted. Two trolls planting different types
on the same cell: nothing happens.

### 🪓 Chopping (`CHOP id`) — leagues 3+

A troll on the same cell as a tree can chop it, reducing the tree's health by
`chopPower`. When health ≤ 0 the tree dies and the troll collects **wood
equal to the tree's size**, capped by free `carryCapacity`; excess wood
vanishes.

If two trolls chop the same tree, the same one-at-a-time / last-item
duplication rule from HARVEST applies to the wood.

### ⛏️ Mining (`MINE id`) — leagues 3+

A troll adjacent (4-neighbour) to an IRON cell can mine it, gaining up to
`chopPower` iron limited by free `carryCapacity`. **Iron is infinite** — IRON
cells never deplete.

### 📥 Drop (`DROP id`)

A troll adjacent (4-neighbour) to its own shack transfers **all carried
items** to the shack inventory.

### 📤 Pick (`PICK id type`) — leagues 2+

A troll adjacent to its own shack picks **one** item of `type` from the shack
inventory. Only one item per `PICK` action.

### 🧠 Training (`TRAIN moveSpeed carryCapacity harvestPower chopPower`) — leagues 2+

Spawns a new troll at the player's shack with the given four attributes. Each
attribute consumes a different resource from the shack inventory:

- `PLUM` → `movementSpeed`
- `LEMON` → `carryCapacity`
- `APPLE` → `harvestPower`
- `IRON` → `chopPower` (leagues 3+; reserved otherwise)

Cost formula per attribute: `existingTrolls + attribute²`, where
`existingTrolls` is the count of trolls the player currently owns (not
including the one being trained).

Example with 2 existing trolls, `TRAIN 2 3 1 0`:

- 2 movementSpeed → 2 + 2² = **6 PLUM**
- 3 carryCapacity → 2 + 3² = **11 LEMON**
- 1 harvestPower → 2 + 1² = **3 APPLE**
- 0 chopPower → 2 + 0² = **2 IRON**

If the player can't pay the full cost, the training fails.

### Other commands

- `WAIT` — do nothing.
- `MSG text` — display a message in the replay (no gameplay effect).

### 🔁 Turn order

Within a turn, actions of the **same type** resolve simultaneously. Order of
phases:

1. **Move**
2. **Harvest**
3. **Plant**
4. **Chop**
5. **Pick**
6. **Train**
7. **Drop**
8. **Mine**
9. **Grow trees** (cooldowns tick, possible growth/fruiting)

### ⛔ Game end

The game ends at the **end of a turn** when any of the following becomes true:

- Maximum turns reached: **300** (leagues 3+) or **100** (leagues 1–2).
- No trees on the map for `STALL_LIMIT = 10` consecutive turns.
- A player can no longer make progress (cannot carry productive items) AND
  their score ≤ opponent's score — i.e. the leader can force a win by waiting.
  Specifically: a player is "stuck" if none of their trolls carries anything
  other than IRON and the shack inventory has zero of `PLUM/LEMON/APPLE/BANANA`.

### 🏆 Victory conditions

The player with more **points** wins.

- Each fruit (`PLUM`, `LEMON`, `APPLE`, `BANANA`) in the shack inventory: **1
  point**.
- Each `WOOD` in the shack inventory: **4 points**.
- `IRON` in the shack inventory: **0 points**.

### Defeat conditions

- Time out on the response.
- Unrecognized / invalid command.

Note on timeouts (from the upstream statement): a player only loses by timeout
when they exceed the limit **3 times by ≤ 50 ms or once by > 50 ms**. The
arena enforces stricter per-turn timeouts (see `game-model.md`); this lenient
rule is only relevant to upstream replays.

## Game Protocol

### Initialization input

- **Line 1:** `width height`.
- **Next `height` lines:** one row of `width` characters each. Characters:
  - `.` GRASS, `~` WATER, `#` ROCK, `+` IRON (league ≥ 3 only),
  - `0` your shack, `1` opponent's shack.

### Per-turn input

- **Line 1:** your inventory — `plums lemons apples bananas iron wood`.
  (Leagues 1–2: `iron` and `wood` are always `0`.)
- **Line 2:** opponent's inventory — same format.
- **Line 3:** `treeCount`.
- **Next `treeCount` lines:** `type x y size health fruits cooldown`
  where `type ∈ {PLUM, LEMON, APPLE, BANANA}`.
- **Next line:** `trollsCount`.
- **Next `trollsCount` lines:**
  `id player x y movementSpeed carryCapacity harvestPower chopPower carryPlum carryLemon carryApple carryBanana carryIron carryWood`
  where `player = 0` if you own the troll, `1` otherwise. In leagues 1–2,
  `chopPower`, `carryIron` and `carryWood` are reserved (always `0`).

### Output

A single line with one or more commands separated by `;`. Each command is one of:

- `MOVE id x y`
- `HARVEST id`
- `PLANT id type` (leagues 2+)
- `CHOP id` (leagues 3+)
- `MINE id` (leagues 3+)
- `PICK id type` (leagues 2+)
- `DROP id`
- `TRAIN moveSpeed carryCapacity harvestPower chopPower` (leagues 2+; leagues
  2 ignores the 4th argument)
- `WAIT`
- `MSG text`

### Constraints

- Response time for the first turn ≤ **1000 ms**.
- Response time per turn ≤ **50 ms** (arena enforces this strictly).
- `8 ≤ height ≤ 11` (leagues 3+), `width = 2 × height`.
- Leagues 1–2: `height = 8`, game length 100 turns.
- Leagues 3–4: game length 300 turns.

## League differences (summary)

| Feature                         | L1 | L2 | L3 | L4 |
| ------------------------------- | :-: | :-: | :-: | :-: |
| `MOVE`, `HARVEST`, `DROP`, `WAIT` | ✅ | ✅ | ✅ | ✅ |
| `PLANT`, `PICK`, `TRAIN`        | ❌ | ✅ | ✅ | ✅ |
| `CHOP`, `MINE`                  | ❌ | ❌ | ✅ | ✅ |
| Map: WATER, ROCK, IRON terrain  | ❌ | ❌ | ✅ | ✅ |
| Variable height (8–11)          | ❌ | ❌ | ✅ | ✅ |
| Starting inventory > 0          | ❌ | ✅ | ✅ | ✅ |
| Starting IRON in inventory      | ❌ | ❌ | ✅ | ✅ |
| Turn cap                        | 100 | 100 | 300 | 300 |

## Constants reference

From `source/SpringChallenge2026-Troll/src/main/java/engine/Constants.java`:

| Name | Value |
| --- | --- |
| `TIME_PER_TURN` | 50 |
| `GAME_TURNS` | 300 |
| `GAME_TURNS_LOW_LEAGUE` | 100 |
| `STALL_LIMIT` | 10 |
| `PLANT_MAX_SIZE` | 4 |
| `PLANT_MAX_RESOURCES` | 3 |
| `PLANT_COOLDOWN` | {8, 8, 9, 6} |
| `PLANT_WATER_COOLDOWN_BOOST` | {5, 5, 7, 2} |
| `PLANT_FINAL_HEALTH` | {12, 12, 20, 6} |
| `PLANT_DELTA_HEALTH` | {2, 2, 3, 1} |
| `MAP_MIN_HEIGHT` / `MAP_MAX_HEIGHT` | 8 / 11 |
| `MAP_MIN_RIVER` / `MAP_MAX_RIVER` | 2 / 3 |
| `MAP_MIN_IRON` / `MAP_MAX_IRON` | 1 / 2 |
| `MAP_MIN_ROCK` / `MAP_MAX_ROCK` | 1 / 10 |
| `MAP_MIN_TREE` / `MAP_MAX_TREE` | 1 / 3 |
| `MAP_MAX_OPP_DIST` | 6 |
| `MIN_STARTING_RESOURCE` / `MAX_STARTING_RESOURCE` | 2 / 10 |
| `WOOD_POINTS` | 4 |

## Source-vs-statement notes

- The arena strictly enforces the 50ms turn limit and the 1000ms first-turn
  limit. Upstream's "3 small overages allowed" rule is not modeled.
- Tie-breakers for movement when multiple equidistant cells exist depend on
  RNG state in `Board.getNextCell`; parity tests must use the same Java
  random source.
- Iron is described as "infinite" in the statement — the Java engine never
  removes IRON cells when mined; this is consistent.
