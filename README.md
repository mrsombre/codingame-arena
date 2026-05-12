# CodinGame Arena

Local game engine runner for [CodinGame](https://www.codingame.com/) bot programming challenges. Run bot-vs-bot matches offline, analyze results, and watch replays in a built-in web viewer — all without the CodinGame platform.

![Match View](docs/img/match-view.png)

## Features

- **Offline match runner** — execute thousands of matches locally with parallel workers
- **Match tracing** — save per-match JSON traces for replay and analysis
- **Built-in web viewer** — React + PixiJS viewer served from the binary, no extra setup
- **Replay downloader** — fetch replays from codingame.com
- **Replay conversion** — convert downloaded replay JSON into arena trace format
- **Trace analysis** — aggregate stats across batches of traces

```shell
$ bin/arena run winter2026 \
    --blue=bin/bot-winter2026-cpp \
    --red=bin/bot-winter2026-py \
    --seed=100030005000 --simulations 100

Summary: 100 matches played (3.15s)
Stats: wins=29% losses=32% draws=39% avg_score=16.4x17.0 avg_turns=155
Timing: avg_first_response=29msx198ms avg_turn_response=0msx0ms
```

![Batch View](docs/img/batch-view.png)

## Supported Games

| Game                  | Slug          | Source              |
|-----------------------|---------------|---------------------|
| Winter Challenge 2026 | `winter2026`  | `games/winter2026/` |
| Spring Challenge 2020 | `spring2020`  | `games/spring2020/` |

The game slug is the first positional argument for top-level commands that need a game, e.g. `arena run winter2026 ...`, `arena replay winter2026 mrsombre`. Under `arena game`, the action comes first and the game second: `arena game rules winter2026`, `arena game serialize winter2026 --seed=42`. `arena game list` (no game slug) prints every engine currently registered in the binary.

## Commands

| Command   | Purpose                                                 |
|-----------|---------------------------------------------------------|
| `run`     | Run one or more match simulations against a player      |
| `replay`  | Download replay JSON (`get`, `leaderboard` subcommands) |
| `analyze` | Analyze trace outcomes and game-owned metrics           |
| `serve`   | Serve the embedded web viewer                           |
| `game`    | Per-game helpers: `rules`, `trace`, `serialize`, `list` |

Run `arena help <command>` for full flag listings.

## Quick Start

### Download Prebuilt Binary

Download the latest release binary for your platform and rename it to `arena`:

```shell
# macOS (Apple Silicon)
curl -L -o arena https://github.com/mrsombre/codingame-arena/releases/latest/download/arena-darwin-arm64
chmod +x arena

# Linux (amd64)
curl -L -o arena https://github.com/mrsombre/codingame-arena/releases/latest/download/arena-linux-amd64
chmod +x arena

# Linux (arm64)
curl -L -o arena https://github.com/mrsombre/codingame-arena/releases/latest/download/arena-linux-arm64
chmod +x arena
```

For Windows, download `arena-windows-amd64.exe` from the [latest release](https://github.com/mrsombre/codingame-arena/releases/latest).

### Build

```shell
make build-arena
make build-winter2026-agents
make match-winter2026
```

### Web Viewer

```shell
bin/arena serve
```

Opens a web UI at `http://localhost:5757` where you can select bots, run matches, and watch replays.

## Configuration

Flags can be supplied via CLI, environment variables (`ARENA_<FLAG>`, hyphens become underscores — e.g. `ARENA_SEED`), or an `arena.yml` config file in the current directory.

## License

[MIT](LICENSE) © 2026 Dmitrii Barsukov
