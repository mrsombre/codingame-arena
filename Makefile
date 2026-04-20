BIN_DIR := bin

# arena
.PHONY: test

test:
	go test ./internal/arena

# viewer
VIEWER_DIR := viewer

.PHONY: build-viewer

build-viewer:
	cd $(VIEWER_DIR) && bun install && bun run build

# winter-2026

WINTER2026_AGENTS := games/winter2026/agents
WINTER2026_ARENA  := $(BIN_DIR)/arena-winter2026
WINTER2026_CPPBOT := $(BIN_DIR)/bot-winter2026-cpp
WINTER2026_PYBOT  := $(BIN_DIR)/bot-winter2026-py

.PHONY: build-winter2026-arena build-winter2026-full build-winter2026-cpp match-winter2026

build-winter2026-arena:
	mkdir -p $(BIN_DIR)
	go build -tags winter2026 -ldflags="-w -s" -o $(WINTER2026_ARENA) ./cmd/arena

build-winter2026-agents:
	mkdir -p $(BIN_DIR)
	rm -f $(BIN_DIR)/*bot*
	g++ -std=c++17 -O2 -o $(WINTER2026_CPPBOT) $(WINTER2026_AGENTS)/bot.cpp
	cp -f $(WINTER2026_AGENTS)/bot.py $(WINTER2026_PYBOT)

match-winter2026:
	./$(WINTER2026_ARENA) --p0-bin=$(BIN_DIR)/ --p1-bin=./$(WINTER2026_PYBOT) \
		--seed=100030005000 --simulations 100
