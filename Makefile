BIN_DIR := bin

# utility
.PHONY: clean reset

clean:
	rm -rf bin/* tmp/* replays/* traces/*

reset:
	rm -rf replays/* traces/*

# backend
.PHONY: test-arena test-games lint-arena build-arena build-viewer clean

test-arena:
	go test ./cmd/arena ./internal/...

test-games:
	go test ./games/...

lint-arena:
	golangci-lint run ./cmd/... ./games/... ./internal/...

build-arena:
	mkdir -p $(BIN_DIR)
	go build -ldflags="-w -s" -o $(BIN_DIR)/arena ./cmd/arena

# frontend
.PHONY: type-check-viewer lint-viewer build-viewer
VIEWER_DIR := viewer

type-check-viewer:
	cd $(VIEWER_DIR) && pnpm run type-check

bundle-viewer:
	cd $(VIEWER_DIR) && pnpm run bundle

build-viewer:
	cd $(VIEWER_DIR) && pnpm run build

# match runner
.PHONY: build-winter2026-agents match-winter2026
WINTER2026_AGENTS := games/winter2026/agents
WINTER2026_CPPBOT := $(BIN_DIR)/bot-winter2026-cpp
WINTER2026_PYBOT  := $(BIN_DIR)/bot-winter2026-py

build-winter2026-agents:
	rm -f $(BIN_DIR)/bot-*
	g++ -std=c++17 -O2 -o $(WINTER2026_CPPBOT) $(WINTER2026_AGENTS)/bot.cpp
	cp -f $(WINTER2026_AGENTS)/bot.py $(WINTER2026_PYBOT)

match-winter2026:
	./$(BIN_DIR)/arena run winter2026 --blue=./$(WINTER2026_CPPBOT) --red=./$(WINTER2026_PYBOT) \
		--seed=100030005000700089 --simulations 50 --trace

# analytics
.PHONY: replay analyze

replay:
	./$(BIN_DIR)/arena replay winter2026 mrsombre

analyze:
	./$(BIN_DIR)/arena analyze winter2026
