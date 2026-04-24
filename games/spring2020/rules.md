# Spring Challenge 2020 — PacMan

## Goal

Eat more pellets than your opponent! And avoid getting killed!

## Rules

The game is played on a grid given to you at the start of each run. The grid is made up of walls and floors.

Each player controls a team of pacs that can move along the grid.

### 🗺️ The Map

The grid is generated randomly, and can vary in `width` and `height`.

Each cell of the map is either:

- A wall (represented by a pound character: `#`)
- A floor (represented by a space character: ` `)

Maps are always symmetrical across the central vertical axis. Most grids will have floor tiles on the left and right edges; pacs can **wrap around the map** and appear on the other side by moving through these tiles.

When the game begins, the map is filled with **pellets** and the occasional **super-pellet**. Landing on a pellet with one of your **pacs** scores you `1 point`. Super-pellets are worth `10 points`. The pellet is then removed.

> A pellet is worth `1 point` and a super-pellet is worth `10 points`.

### 🔵🔴 The Pacs

Each player starts with the same number of pacs, up to `5` each.

Your pacs **cannot see through walls**. On each turn you have vision on all of the pellets and enemy pacs that can be connected by a continuous straight line to any one of your pacs. Super-pellets are so bright that they can be seen from everywhere!

> The line of sight of the pacs is blocked by the walls.

At each turn, you are given information regarding the visible pacs and pellets. For each pac, you are given its identifier, whether it's yours or not and its coordinates. For each pellet, you are given its coordinates and value.

Each pac is of a given `type` (`ROCK`, `PAPER` or `SCISSORS`). Pacs that have died are of type `DEAD`.

Each pac has access to two **abilities** (`SWITCH` and `SPEED`) that share the same **cooldown period** of `10` turns. The abilities of a pac are already available at the start of the game.

Pacs can receive the following commands (a pac can only receive one command per turn):

- **`MOVE`**: Give the pac a target position; the pac will find a shortest route to that position and **move the first step of the way**. The pac will not take into account the presence of pellets or other pacs when choosing a route.

  > Each pac that received a `MOVE` order will move toward the target by going either up, down, left or right.

- **`SWITCH`**: it will morph into a new form. The available pac types are:
  - `ROCK`
  - `PAPER`
  - `SCISSORS`

  > A `SWITCH` allows a pac to change its type.

- **`SPEED`**: it will speed up for the next `5` turns, making it take the first **2 steps** along its path when moving. This means the pac can move twice as far as usual on each turn.

  > A `SPEED` allows a pac to move by 2 steps instead of 1 during the next 5 turns.

See the **Game Protocol** section for more information on sending commands to your pacs.

Crossing paths or landing on the same cell as another pac will cause a **collision** to occur. This is how collisions are resolved:

1. All moving pacs move `1` step, regardless of their speed.
2. If the pacs are of the same type or belong to the same player, both pacs will go back to the cell they moved from. If the pacs are of different types, they can land on the same cell, but a pac can't cross the path of a stronger pac: it will be blocked.
3. Cancelling a move may create new collisions. For this reason, the previous step is repeated until no new collisions are created.
4. All pacs that share the same cell as a pac that can beat them are killed. `ROCK` beats `SCISSORS`, `SCISSORS` beats `PAPER` and `PAPER` beats `ROCK`.
5. Repeat for any pac with an activated `SPEED` ability.

### 🎬 Action order for one turn

1. Decrement cooldown timers
2. Decrement `SPEED` duration timers
3. Execute **abilities**
4. Resolve movement, including collisions
5. Kill pacs that were beaten during a collision
6. Eat pellets

### ⛔ Game end

The game stops when there are no enough pellets in game to change the outcome of the game.

The game stops automatically after `200 turns`.

If all of a player's pacs are dead, all remaining pellets are automatically scored by any surviving pacs and the game is stopped.

#### Victory Conditions

The winner is the player with the highest score, regardless of the amount of surviving pacs.

#### Defeat Conditions

Your program does not provide a command in the allotted time or one of the commands is invalid.

### 🐞 Debugging tips

- Hover over the grid to see the coordinates of the cell under your mouse
- Hover over pacs to see information about them
- Append text after any command for a pac and that text will appear above that pac
- Press the gear icon on the viewer to access extra display options
- Use the keyboard to control the action: space to play/pause, arrows to step 1 frame at a time

## Technical Details

- You can check out the source code of this game [on this GitHub repo](https://github.com/CodinGame/SpringChallenge2020).

## Game Input

### Initialization Input

- **Line 1:** two integers `width` and `height` for the size of the map.
- **Next `height` lines:** a string of `width` characters each representing one cell of this row: `' '` is a floor and `'#'` is a wall.

### Input for One Game Turn

- **First line:** Two space-separated integers:
  - `myScore` your current score
  - `opponentScore` the score of your opponent
- **Second line:** One integer:
  - `visiblePacCount` for the amount of pacs visible to you
- **Next `visiblePacCount` lines:**
  - `pacId`: the pac's id (unique for a given player)
  - `mine`: the pac's owner (1 if this pac is yours, 0 otherwise. Converted into a boolean type in most languages.)
  - `x` & `y`: the pac's position
  - `typeId`: the pac's type (`ROCK` or `PAPER` or `SCISSORS`). If the pac is dead, its type is now `DEAD`.
  - `speedTurnsLeft`: the number of remaining turns before the speed effect fades
  - `abilityCooldown`: the number of turns until you can request a new ability for this pac (`SWITCH` and `SPEED`)
- **Next line:** one integer `visiblePelletCount` for the amount of pellets visible to you
- **Next `visiblePelletCount` lines:** three integers:
  - `x` & `y`: the pellet's position
  - `value`: the pellet's score value

### Output

A single line with one or multiple commands separated by `|`. For example: `MOVE 0 5 7 | MOVE 1 16 10`.

- `MOVE pacId x y`: the pac with the identifier `pacId` moves towards the given cell.
- `SPEED pacId`: the pac will be able to move by 2 steps during the next 5 turns.
- `SWITCH pacId pacType`: the pac switches to the `pacType`.

### Constraints

- `2` ≤ Number of pacs per player ≤ `5`
- `29` ≤ `width` ≤ `35`
- `10` ≤ `height` ≤ `17`
- Response time per turn ≤ `50`ms
- Response time for the first turn ≤ `1000`ms
