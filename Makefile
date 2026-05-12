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
.PHONY: build-spring2026-agents match-spring2026
SPRING2026_AGENTS := games/spring2026/agents
SPRING2026_CPPBOT := $(BIN_DIR)/bot-spring2026-cpp
SPRING2026_PYBOT  := $(BIN_DIR)/bot-spring2026-py

build-spring2026-agents:
	rm -f $(BIN_DIR)/bot-*
	g++ -std=c++17 -O2 -o $(SPRING2026_CPPBOT) $(SPRING2026_AGENTS)/bot.cpp
	cp -f $(SPRING2026_AGENTS)/bot.py $(SPRING2026_PYBOT)

match-spring2026:
	./$(BIN_DIR)/arena run spring2026 --blue=./$(SPRING2026_CPPBOT) --red=./$(SPRING2026_PYBOT) \
		--seed=100030005000700089 --simulations 50 --trace

# analytics
.PHONY: replay analyze

replay:
	./$(BIN_DIR)/arena replay spring2026 mrsombre

analyze:
	./$(BIN_DIR)/arena analyze spring2026
