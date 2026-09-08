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
.PHONY: build-summer2026-agents match-summer2026
SUMMER2026_AGENTS := games/summer2026/agents
SUMMER2026_CPPBOT := $(BIN_DIR)/bot-summer2026-cpp
SUMMER2026_PYBOT  := $(BIN_DIR)/bot-summer2026-py

build-summer2026-agents:
	rm -f $(BIN_DIR)/bot-*
	g++ -std=c++17 -O2 -o $(SUMMER2026_CPPBOT) $(SUMMER2026_AGENTS)/bot.cpp
	cp -f $(SUMMER2026_AGENTS)/bot.py $(SUMMER2026_PYBOT)

match-summer2026:
	./$(BIN_DIR)/arena run summer2026 --blue=./$(SUMMER2026_CPPBOT) --red=./$(SUMMER2026_PYBOT) \
		--seed=468706172918629800 --simulations 50 --trace

# analytics
.PHONY: replay analyze

replay:
	./$(BIN_DIR)/arena replay summer2026 mrsombre

analyze:
	./$(BIN_DIR)/arena analyze summer2026
