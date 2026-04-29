# Command serve

Run the embedded web viewer locally. The viewer bundle is compiled into the `arena` binary, so no extra setup is needed.

![Match View](img/match-view.png)

## Quick start

```shell
bin/arena serve --game=winter2026
```

```
  arena serve  ready

  ➜  Local:     http://localhost:5757
  ➜  Trace:     ./traces
  ➜  Watching:  /path/to/bin/arena

  Shortcuts
  press o + enter to open in browser
  press q + enter to quit
```

The web UI lets you select bots, run matches, and watch replays — all backed by the JSON API listed below.

## Options

| Flag           | Default     | Description                                                  |
|----------------|-------------|--------------------------------------------------------------|
| `--host`       | `localhost` | Bind host                                                    |
| `--port`       | `5757`      | HTTP port                                                    |
| `--trace-dir`  | `./traces`  | Directory with match trace JSON (powers `/api/matches`)      |
| `--replay-dir` | `./replays` | Directory with CodinGame replay JSON (powers `/api/replays`) |
| `--bin-dir`    | `./bin`     | Directory scanned for bot binaries (powers `/api/bots`)      |

`--bin-dir` lists every executable file whose name contains `bot` so the UI can offer them as p0/p1 picks.

## HTTP API

| Method | Path                | Purpose                                |
|--------|---------------------|----------------------------------------|
| GET    | `/api/game`         | Active game metadata                   |
| GET    | `/api/games`        | All registered games                   |
| GET    | `/api/bots`         | Bot binaries discovered in `--bin-dir` |
| GET    | `/api/matches`      | List of match traces in `--trace-dir`  |
| GET    | `/api/matches/{id}` | Single match trace                     |
| GET    | `/api/replays`      | List of replays in `--replay-dir`      |
| GET    | `/api/replays/{id}` | Single replay JSON                     |
| POST   | `/api/run`          | Run a match from the UI                |

## Stdin shortcuts

While the server is running:

- `o` + enter — open the URL in the default browser
- `q` + enter — graceful shutdown
- Ctrl-C / SIGTERM — graceful shutdown

## Hot reload

`serve` watches the running binary's parent directory and re-execs itself when the binary is rebuilt (e.g. after `make build-arena`). No manual restart needed during local dev.
