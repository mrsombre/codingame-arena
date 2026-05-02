# Agents Rules

## Project Overview

CodinGame Arena — local game engine runner for CodinGame challenges. Runs bot-vs-bot matches, records replays, and visualizes them.

Go backend with a TypeScript / React viewer.

## Project Structure

```
cmd/arena/              # CLI entrypoint
internal/
├─ arena/               # Match runner, batching, tracing, server
│  ├─ commands/          # CLI subcommands (run, analyze, serialize, convert, replay, serve)
│  └─ server/            # HTTP server for viewer
├─ util/
│  ├─ javarand/          # Java random port
│  └─ sha1prng/          # SHA1PRNG port
games/
├─ game.go              # Game registry interface
├─ spring2020/
│  ├─ engine/            # Spring 2020 (Pac-Man) game engine
│  └─ agents/            # Bot sources (C++, Python)
└─ winter2026/
   ├─ engine/            # Winter 2026 game engine (referee, grid, actions)
   └─ agents/            # Bot sources (C++, Python)
source/                  # Upstream subtree imports (DO NOT MODIFY)
viewer/                  # React + PixiJS match viewer (pnpm, Vite, shadcn)
├─ packages/
│  ├─ shared/            # Shared router, components, API, styles
│  ├─ spring2020/        # Spring 2020 viewer (Pac-Man)
│  └─ winter2026/        # Winter 2026 viewer (Snakes)
bin/                     # Build artifacts (gitignored)
replays/                 # Downloaded replay JSON files (gitignored)
traces/                  # Match trace files for analysis (gitignored)
```

## Project Rules

### Mandatory

- NEVER run `go run` directly — use `make build-arena` then run the binary from `bin/`
- NEVER modify files under `source/` — these are upstream subtree imports
- NEVER commit `replays/`, `traces/`, or `bin/` directories
- ALWAYS use `pnpm` for the viewer (not npm/yarn)
- ALWAYS use Biome for TypeScript linting/formatting (not ESLint/Prettier)

### Validation

- ALWAYS run `make test-arena` and `make lint-arena` before considering Go changes complete
- ALWAYS run `make type-check-viewer` and `make lint-viewer` before considering viewer changes complete

## Project Commands

```shell
# Go
make test-arena                  # Run arena tests (internal/)
make test-games                  # Run game engine tests (games/)
make lint-arena                  # Run golangci-lint
make build-arena                 # Build arena binary to bin/
bin/arena help                   # Show help for arena binary

# Viewer (from viewer/)
pnpm install                     # Install dependencies
pnpm run build                   # Production build
pnpm run dev                     # Dev server
pnpm run format                  # Biome format
pnpm run check                   # Biome lint
pnpm run type-check              # TypeScript type check
pnpm run bundle                  # Biome check + type-check + build

# Make viewer targets
make type-check-viewer           # Type-check viewer (pnpm run type-check)
make bundle-viewer               # Lint and test viewer (pnpm run bundle)
```
