# CodinGame Arena

Local game engine runner for [CodinGame](https://www.codingame.com/) bot programming challenges. Run bot-vs-bot matches offline, analyze results, and watch replays in a built-in web viewer — all without the CodinGame platform.

![Match View](docs/img/match-view.png)

## Features

- **Offline match runner** — execute thousands of matches locally with parallel workers
- **Match tracing** — save per-match JSON traces for replay and analysis
- **Built-in web viewer** — React + PixiJS viewer served from the binary, no extra setup
- **Replay downloader** — fetch replays from codingame.com for local viewing

![CLI View](docs/img/cli-view.png)

![Batch View](docs/img/batch-view.png)

## Supported Games

| Game                  | Flag                | Source              |
|-----------------------|---------------------|---------------------|
| Spring Challenge 2020 | `--game spring2020` | `games/spring2020/` |
| Winter Challenge 2026 | `--game winter2026` | `games/winter2026/` |

## Quick Start

### CLI Simulator

```shell
make build-arena
make build-winter2026-agents

bin/arena --game=winter2026 \
  --p0=bin/bot-winter2026-cpp \
  --p1=bin/bot-winter2026-py \
  --seed=100030005000 --simulations 100

# wins=26% losses=31% draws=43% avg_score=12.8v12.8 avg_turns=154
# p0: avg_first_response=2ms avg_turn_response=0ms p1: avg_first_response=37ms avg_turn_response=0ms
```

### Web Viewer

```shell
bin/arena serve
```

Open a web UI at `http://localhost:5757` where you can select bots, run matches, and watch replays.

## License

[MIT](LICENSE) © 2026 Dmitrii Barsukov
