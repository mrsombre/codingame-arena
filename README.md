# CodinGame Arena

Local game engine runner for [CodinGame](https://www.codingame.com/) bot programming challenges. Run bot-vs-bot matches offline, analyze results, and watch replays in a built-in web viewer — all without the CodinGame platform.

## Features

- **Offline match runner** — execute thousands of matches locally with parallel workers
- **Side swapping** — automatically swaps player sides for fair evaluation
- **Match tracing** — save per-match JSON traces for replay and analysis
- **Built-in web viewer** — React + PixiJS viewer served from the binary, no extra setup
- **Replay downloader** — fetch replays from codingame.com for local viewing
- **Pluggable engines** — add new games by implementing the `Referee` / `GameFactory` interfaces
- **Self-restarting server** — the `serve` command watches its own binary and restarts on rebuild

## Supported Games

| Game | Flag | Source |
|---|---|---|
| Winter Challenge 2026 | `--game winter-2026` | `games/winter2026/` |

## Requirements

- Go 1.26+
- Node.js 24+ and pnpm 10+ (for the viewer)
- C++17 compiler (for C++ bots)
- Python 3 (for Python bots)

## Quick Start

### 1. Build everything

```sh
make build
```

This compiles the viewer frontend and the `arena` binary into `bin/`.

### 2. Build your bots

The example bots live in `games/winter2026/agents/`. Build them with:

```sh
make build-winter2026-agents
```

This produces `bin/bot-winter2026-cpp` and `bin/bot-winter2026-py`.

### 3. Run a match

```sh
./bin/arena run --p0-bin=./bin/bot-winter2026-cpp --p1-bin=./bin/bot-winter2026-py
```

Output (short summary by default):

```
W=72% L=24% D=4% score=18.3v12.1 turns=142.0
```

### 4. Run a batch of matches

```sh
./bin/arena run \
  --p0-bin=./bin/bot-winter2026-cpp \
  --p1-bin=./bin/bot-winter2026-py \
  --simulations 100 \
  --seed 100030005000
```

### 5. Start the viewer

```sh
./bin/arena serve
```

Opens a web UI at `http://localhost:5757` where you can select bots, run matches, and watch replays.

## CLI Reference

The binary has four subcommands:

### `arena run`

Run one or more match simulations.

```sh
./bin/arena run [OPTIONS]
```

| Flag | Default | Description |
|---|---|---|
| `--p0-bin <PATH>` | *(required)* | Player 0 binary |
| `--p1-bin <PATH>` | `./bin/opponent` | Player 1 binary |
| `--simulations <N>` | `1` | Number of matches to run |
| `--parallel <N>` | CPU count | Worker threads |
| `--seed <N>` | current time | Base RNG seed |
| `--seedx <N>` | — | Seed increment per match (`seed_i = seed + i*N`) |
| `--max-turns <N>` | `200` | Maximum turns per match |
| `--no-swap` | `false` | Disable automatic side swapping |
| `--trace-dir <PATH>` | — | Write per-match JSON trace files |
| `--output-matches` | `false` | Include per-match results in JSON output |
| `--verbose` | `false` | Output full JSON instead of summary line |
| `--debug` | `false` | Force 1 match, fixed sides, print debug to stderr |
| `--timing` | `false` | Print per-turn timing to stderr |

**Examples:**

```sh
# Quick single match with verbose JSON output
./bin/arena run --p0-bin=./bin/bot-winter2026-cpp --verbose

# 200 matches with trace files for the viewer
./bin/arena run \
  --p0-bin=./bin/bot-winter2026-cpp \
  --p1-bin=./bin/bot-winter2026-py \
  --simulations 200 \
  --trace-dir=./matches \
  --seed 42

# Debug a specific seed
./bin/arena run \
  --p0-bin=./bin/bot-winter2026-cpp \
  --seed 100030005000 \
  --debug
```

### `arena serve`

Start the web viewer with a built-in HTTP API.

```sh
./bin/arena serve [OPTIONS]
```

| Flag | Default | Description |
|---|---|---|
| `--port <N>` | `5757` | HTTP port |
| `--host <HOST>` | `localhost` | Bind address |
| `--trace-dir <PATH>` | `./matches` | Directory with match trace JSON files |
| `--bin-dir <PATH>` | `./bin` | Directory to scan for bot binaries |

The server auto-discovers bot binaries (files containing "bot" in the name) from `--bin-dir` and serves match traces from `--trace-dir`. It watches its own binary and self-restarts when you rebuild.

Interactive keys while running:
- `o` + Enter — open the viewer in your default browser
- `q` + Enter — quit

**API endpoints:**

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/game` | Current game info |
| `GET` | `/api/games` | List registered games |
| `GET` | `/api/bots` | List discovered bots |
| `GET` | `/api/matches` | List saved match traces |
| `GET` | `/api/matches/{id}` | Get a specific match trace |
| `GET` | `/api/serialize` | Generate initial game state for a seed |
| `POST` | `/api/run` | Run a single match |
| `POST` | `/api/batch` | Run a batch of matches (up to 500) |

### `arena replay`

Download a replay from codingame.com.

```sh
./bin/arena replay <URL_OR_ID> [OPTIONS]
```

| Flag | Default | Description |
|---|---|---|
| `-o, --out <PATH>` | `./replays/replay-<id>.json` | Output file path |

**Examples:**

```sh
# By replay ID
./bin/arena replay 123456789

# By full URL
./bin/arena replay https://www.codingame.com/replay/123456789

# Custom output path
./bin/arena replay 123456789 -o ./tmp/my-replay.json
```

### `arena serialize`

Print the initial game input (global info + first frame) for a given seed. Useful for debugging bot input parsing.

```sh
./bin/arena serialize --seed <N> --player <0|1>
```

**Example:**

```sh
./bin/arena serialize --seed 100030005000 --player 0
```

## Configuration

Settings can be provided via CLI flags, environment variables, or an `arena.yml` config file.

**Priority:** CLI flags > environment variables > config file.

**Environment variables** use the `ARENA_` prefix with underscores:

```sh
export ARENA_GAME=winter-2026
export ARENA_SEED=42
export ARENA_SIMULATIONS=100
```

**Config file** (`arena.yml` in the working directory):

```yaml
game: winter-2026
seed: 100030005000
simulations: 100
p0-bin: ./bin/bot-winter2026-cpp
p1-bin: ./bin/bot-winter2026-py
trace-dir: ./matches
```

Unknown `--key value` flags are forwarded as game-specific options to the engine.

## Adding a New Game

1. Create a package under `games/<name>/engine/`
2. Implement the `Referee`, `Player`, and `GameFactory` interfaces from `internal/arena`
3. Register via `init()`:
   ```go
   func init() {
       arena.Register(NewFactory())
   }
   ```
4. Import the package in `games/game.go`:
   ```go
   import _ "github.com/mrsombre/codingame-arena/games/<name>/engine"
   ```

## Project Structure

```
cmd/arena/              # CLI entrypoint
internal/
├─ arena/               # Match runner, batching, tracing, server
│  ├─ commands/          # CLI subcommands (run, replay, serve, serialize)
│  └─ server/            # HTTP API for the viewer
└─ util/                 # Java random / SHA1PRNG ports
games/
├─ game.go              # Game registry imports
└─ winter2026/
   ├─ engine/            # Winter 2026 game engine
   └─ agents/            # Example bots (C++, Python)
viewer/                  # React + PixiJS web viewer (TypeScript, Vite, shadcn)
```

## License

[MIT](LICENSE) © 2026 Dmitrii Barsukov
