# Agents Rules

## Project Overview

CodinGame Arena — local game engine runner for CodinGame challenges. Runs bot-vs-bot matches, records traces, and analyzes them.

Go CLI.

## Project Structure

```
cmd/arena/          # CLI entrypoint
internal/
├─ arena/           # Match runner, batching, tracing
│  └─ commands/     # CLI subcommands
├─ util/
│  ├─ javarand/     # Java random port
│  └─ sha1prng/     # SHA1PRNG port
games/
├─ game.go          # Game registry interface
├─ .../
│  ├─ engine/       # Game engine
│  └─ agents/       # Bot sources (C++, Python)
source/             # Upstream subtree imports (DO NOT MODIFY)
bin/                # Build artifacts (gitignored)
replays/            # Downloaded replay JSON files (gitignored)
traces/             # Match trace files for analysis (gitignored)
```

## Project Rules

### Mandatory

- NEVER run `go run` directly — use `make build` then run the binary from `bin/`
- NEVER modify files under `source/` — these are upstream subtree imports
- NEVER commit `replays/`, `traces/`, or `bin/` directories

### Validation

- ALWAYS run `make test` and `make lint` before considering Go changes complete

## Project Commands

```shell
# Go
make test                        # Run arena tests (internal/)
make test-games                  # Run game engine tests (games/)
make lint                        # Run golangci-lint
make build                       # Build arena binary to bin/
bin/arena help                   # Show help for arena binary
```

## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/` in this repo. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical roles, unmapped — label string equals role name. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
