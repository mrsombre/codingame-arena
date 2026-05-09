# Command game serialize

Print the game input that a bot would receive on its first turn for a given seed.

Useful for inspecting initial map state, capturing fixtures for unit tests, or feeding a bot a fixed input via stdin without running the engine.

## Quick start

```shell
bin/arena game serialize winter2026 100030005000
```

Output is the raw lines a bot reads from stdin:

```
0
34
19
..........##..........##..........
...
```

## Options

`arena game serialize <game> <seed> [OPTIONS]` — the game slug and seed are required positionals, in that order.

| Flag           | Default       | Description                          |
|----------------|---------------|--------------------------------------|
| `--player`     | `0`           | Player perspective (`0` or `1`)      |
| `-l, --league` | game-specific | League level                         |

## Output

Two blocks separated by newline-terminated lines:

1. **Global init info** — constants and map data sent once at game start.
2. **First-frame info** — per-turn state for turn 0.

Format matches exactly what bots receive on stdin during a real match. Pipe it into a bot binary to drive a single-turn invocation:

```shell
bin/arena game serialize winter2026 42 | bin/bot-winter2026-cpp
```
