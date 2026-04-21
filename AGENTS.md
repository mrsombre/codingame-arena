# Agents Rules

## Project Overview

CodinGame Arena — local game engine runner for CodinGame challenges. Runs bot-vs-bot matches, records replays, and visualizes them. Go backend with a TypeScript/React viewer.

## Project Structure

```
cmd/arena/              # CLI entrypoint
internal/
├─ arena/               # Match runner, batching, tracing, server
│  ├─ commands/          # CLI subcommands (run, replay, front, serialize)
│  └─ server/            # HTTP server for viewer
└─ util/                 # Java random / SHA1PRNG ports
games/
├─ game.go              # Game registry interface
└─ winter2026/
   ├─ engine/            # Winter 2026 game engine (referee, grid, actions)
   └─ agents/            # Bot sources (C++, Python)
viewer/                  # React + PixiJS match viewer (pnpm, Vite, shadcn)
bin/                     # Build artifacts (gitignored)
matches/                 # Match results (gitignored)
replays/                 # Replay JSON files (gitignored)
```

## Project Rules

- NEVER run `go run` directly — use `make build` then run the binary from `bin/`
- NEVER modify files under `source/` — these are upstream subtree imports
- NEVER commit `matches/`, `replays/`, or `bin/` directories
- ALWAYS run `make test` before considering Go changes complete
- ALWAYS use `pnpm` for the viewer (not npm/yarn)
- ALWAYS use Biome for TypeScript linting/formatting (not ESLint/Prettier)

## Project Commands

```shell
# Go
make test                        # Run arena tests
make build-arena                 # Build arena binary to bin/
make build                       # Build viewer + arena

# Viewer (from viewer/)
pnpm install                     # Install dependencies
pnpm run build                   # Production build
pnpm run dev                     # Dev server
pnpm run check                   # Biome check
pnpm run lint                    # Biome lint with autofix
pnpm run format                  # Biome format
pnpm run type-check              # TypeScript type check

# Game-specific
make build-winter2026-agents     # Compile C++/Python bots
make match-winter2026            # Run Winter 2026 match batch
```
