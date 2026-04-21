BIN_DIR := bin

# arena
.PHONY: test build-arena build-viewer

test:
	go test ./internal/...

build-arena:
	mkdir -p $(BIN_DIR)
	go build -ldflags="-w -s" -o $(BIN_DIR)/arena ./cmd/arena

# viewer
VIEWER_DIR := viewer

build-viewer:
	cd $(VIEWER_DIR) && pnpm install && pnpm run build

build: build-viewer build-arena

# winter-2026

WINTER2026_AGENTS := games/winter2026/agents
WINTER2026_CPPBOT := $(BIN_DIR)/bot-winter2026-cpp
WINTER2026_PYBOT  := $(BIN_DIR)/bot-winter2026-py

.PHONY: build-winter2026-agents match-winter2026

build-winter2026-agents:
	mkdir -p $(BIN_DIR)
	rm -f $(BIN_DIR)/*bot*
	g++ -std=c++17 -O2 -o $(WINTER2026_CPPBOT) $(WINTER2026_AGENTS)/bot.cpp
	cp -f $(WINTER2026_AGENTS)/bot.py $(WINTER2026_PYBOT)

match-winter2026:
	./$(BIN_DIR)/arena --game=winter-2026 --p0-bin=$(BIN_DIR)/ --p1-bin=./$(WINTER2026_PYBOT) \
		--seed=100030005000 --simulations 100
