BIN_DIR := bin

# arena
.PHONY: test build-arena build-viewer clean

clean:
	rm -rf bin/* tmp/* replays/* matches/*

test:
	go test ./internal/...

test-games:
	go test ./games/...

build-arena:
	mkdir -p $(BIN_DIR)
	go build -ldflags="-w -s" -o $(BIN_DIR)/arena ./cmd/arena

# viewer
VIEWER_DIR := viewer

build-viewer:
	cd $(VIEWER_DIR) && pnpm install && pnpm run build

build: build-viewer build-arena

# spring-2020

SPRING2020_AGENTS := games/spring2020/agents
SPRING2020_CPPBOT := $(BIN_DIR)/bot-spring2020-cpp
SPRING2020_PYBOT  := $(BIN_DIR)/bot-spring2020-py

.PHONY: build-spring2020-agents match-spring2020

build-spring2020-agents:
	mkdir -p $(BIN_DIR)
	g++ -std=c++17 -O2 -o $(SPRING2020_CPPBOT) $(SPRING2020_AGENTS)/bot.cpp
	cp -f $(SPRING2020_AGENTS)/bot.py $(SPRING2020_PYBOT)

match-spring2020:
	./$(BIN_DIR)/arena --game=spring-2020 --p0-bin=./$(SPRING2020_CPPBOT) --p1-bin=./$(SPRING2020_PYBOT) \
		--seed=100030005000 --simulations 100

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
	./$(BIN_DIR)/arena --game=winter-2026 --p0-bin=./$(WINTER2026_CPPBOT) --p1-bin=./$(WINTER2026_PYBOT) \
		--seed=100030005000 --simulations 100
