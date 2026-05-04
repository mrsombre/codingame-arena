# Spring Challenge 2021 — Photosynthesis

## Goal

End the game with more **points** than your opponent.

The game takes place in a forest, in which gentle wood spirits reside. Their job is to make sure trees complete their lifecycle. Two wood spirits have started to compete over which one is the most efficient.

Grow trees at strategic locations of the forest to maximize your points.

## Rules

Each player embodies a **wood spirit**. The game takes place on a hexagonal grid representing the **forest**.

The game is played over several rounds called **days**. Each day can be made up of several game **turns**. On each turn, both players perform one action simultaneously.

### 🌲 Forest

The forest is made up of `37` hexagonal cells, arranged to form a larger hexagon.

Each cell has an **index** and up to six neighbors. Each direction is labelled `0` to `5`.

The distance between two cells equals the minimum number of cells to go through to get from one to the other.

Each cell may contain a **tree**. Each tree is owned by one of the players and has a `size`:

- `0` for a **seed**.
- `1` for a small tree.
- `2` for a medium tree.
- `3` for a large tree.

Each cell has a `richness` which can be:

- `0` for **unusable** cells. Nothing can grow on them.
- `1` for low quality soil.
- `2` for medium quality soil.
- `3` for high quality soil.

### 📅 Days

At the start of each day, players receive **sun points**.

Then, players take **actions** by spending their sun points.

The day ends when both players stop taking actions.

### ☀️ Sun & Shadows

Each tree casts a shadow that affects a number of cells based on its size:

- Size `1` trees cast a shadow `1` cell long.
- Size `2` trees cast a shadow `2` cells long.
- Size `3` trees cast a shadow `3` cells long.

The direction of a shadow depends on which direction the **sun** is currently pointing towards.

On day `0`, the sun is pointing towards direction `0`, meaning all shadows are being cast to the **right**.

In between each day, the sun **moves** to point towards the next direction, coming back to `0` after passing `5`. The sun's direction will therefore always be equal to the current `day modulo 6`.

Helping the wood spirits are **lesser spirits** hiding among all the trees. They will find the shadow on a cell **spooky** if any of the trees casting a shadow is of equal or greater `size` than the tree on that cell.

### ✨ Sun Points

The forest's lesser spirits will harvest **sun points** from each tree that is not hit by a **spooky** shadow. The points will be given to the **owner** of the tree.

The number of sun points harvested depends on the size of the tree:

- A size `0` tree (a seed): no points.
- A size `1` tree: `1` sun point.
- A size `2` tree: `2` sun points.
- A size `3` tree: `3` sun points.

### 🎬 Actions

After collecting sun points, both players take simultaneous turns performing one of four possible actions. As long as you have enough sun points, you can take any number of actions.

The possible actions are:

- **`SEED`**: Command a tree to eject a seed onto a cell within distance equal to the tree's size.
- **`GROW`**: Command a seed or tree to grow into the next size. Trees can grow up to size `3`.
- **`COMPLETE`**: Command a tree to complete its lifecycle. This removes the tree from the forest and scores you points.
- **`WAIT`**: Spend the rest of the day asleep. When both players are asleep, a new day begins and the players are awoken.

Any tree impacted by one of your actions becomes **dormant** for the rest of the day. A dormant tree cannot be the subject of an action.

#### Seed action

To perform a seed action, you must pay sun points equal to the number of seeds (size `0` trees) you already own in the forest.

You may not send a seed onto an **unusable cell** or a cell already containing a tree.

Performing this action impacts **both the source tree and the planted seed**. Meaning both trees will be **dormant** until the next day.

If both players send a seed to the same place on the same turn, neither seed is planted and the sun points are refunded. The source tree, however, still becomes dormant.

#### Grow action

- Growing a seed into a size `1` tree costs `1` sun point + the number of size `1` trees you already own.
- Growing a size `1` tree into a size `2` tree costs `3` sun points + the number of size `2` trees you already own.
- Growing a size `2` tree into a size `3` tree costs `7` sun points + the number of size `3` trees you already own.

#### Complete action

Completing a tree's lifecycle requires `4` sun points. You can only complete the lifecycle of a size `3` tree.

The forest starts with a `nutrients` value of `20`. Completing a tree's lifecycle will award you with as many points as the current `nutrients` value + a bonus according to the `richness` of the cell:

- `1`: `+0` points.
- `2`: `+2` points.
- `3`: `+4` points.

Then, the `nutrients` value is decreased permanently by `1`.

### ⛔ Game End

The game lasts the time it takes for the sun to circle around the board `4` times. This means players have `24` days to play.

Players gain an extra `1 point` for every `3` sun points they have at the end of the game.

If players have the same score, the winner is the player with the most trees in the forest. Note that a seed is also considered a tree.

#### Victory Conditions

The winner is the player with the most **points**.

#### Defeat Conditions

Your program does not provide a command in the allotted time or it provides an unrecognized command.

## Technical Details

- Players start the game with two size `1` trees placed randomly along the edge of the grid.
- Players that are asleep do not receive input.
- If both players complete a lifecycle on the same turn, they both receive full points and the nutrient value is decreased by two.
- The `nutrients` value cannot drop below `0`.
- The source code of this game is available [on this GitHub repo](https://github.com/CodinGame/SpringChallenge2021).

### 🐞 Debugging tips

- Hover over a cell to see extra information about it.
- Append text after any command and that text will appear next to your wood spirit.
- Press the gear icon on the viewer to access extra display options.
- Use the keyboard to control the action: space to play/pause, arrows to step 1 frame at a time.

## Game Protocol

### Initialization Input

- **First line:** `numberOfCells` equals `37`.
- **Next `numberOfCells` lines:** `8` space-separated integers:
  - `index`: for the cell's index.
  - `richness`: for its richness.
  - `6` `neigh` variables, one for each **direction**, containing the index of a neighboring cell or `-1` if there is no neighbor.

### Input for One Game Turn

- **First line:** an integer `day`: the current day, from `0` to `23`.
- **Next line:** an integer `nutrients`: the current nutrient value of the forest.
- **Next line:** `2` space-separated integers:
  - `mySun`: your current sun points.
  - `myScore`: your current score.
- **Next line:** `3` space-separated integers:
  - `oppSun`: your opponent's sun points.
  - `oppScore`: your opponent's score.
  - `oppIsWaiting`: equals `1` if your opponent is asleep, `0` otherwise.
- **Next line:** an integer `numberOfTrees` for the current number of trees in the forest.
- **Next `numberOfTrees` lines:** `4` space-separated integers to describe each tree:
  - `cellIndex`: the index of the cell this tree is on.
  - `size`: the size of the tree. From `0` (seed) to `3` (large tree).
  - `isMine`: `1` if you are the owner of this tree, `0` otherwise.
  - `isDormant`: `1` if this tree is dormant, `0` otherwise.
- **Next line:** an integer `numberOfPossibleActions` for the number of legal moves you can make this turn.
- **Next `numberOfPossibleActions` lines:** a string `possibleAction` containing one of the actions you can output this turn.

### Output

A single line with your command:

- `GROW index`: make your tree on cell `index` grow by `1` size.
- `SEED index0 index1`: make your tree on cell `index0` launch a seed onto cell `index1`.
- `COMPLETE index`: make your large tree on the specified cell complete its lifecycle. This removes the tree.
- `WAIT`: go to sleep until the start of the next day.

### Constraints

- Response time per turn ≤ `100`ms
- Response time for the first turn ≤ `1000`ms
