#!/usr/bin/env python3
# Winter Challenge 2026 (SnakeByte) - greedy DFS nearest-apple bot.
# Reads the Codingame stdin protocol and emits moves for each live bird.

import sys

my_id = int(input())
width = int(input())
height = int(input())
walls = []
for i in range(height):
    row = input()
    walls.append(row)
snakebots_per_player = int(input())
my_ids = []
for i in range(snakebots_per_player):
    my_snakebot_id = int(input())
    my_ids.append(my_snakebot_id)
for i in range(snakebots_per_player):
    opp_snakebot_id = int(input())

DIRS = [
    (0, -1, 'UP'),
    (0, 1, 'DOWN'),
    (-1, 0, 'LEFT'),
    (1, 0, 'RIGHT'),
]


def cell_key(x, y):
    return y * width + x


def in_bounds(x, y):
    return 0 <= x < width and 0 <= y < height


def is_wall(x, y):
    return walls[y][x] == '#'


def greedy_dfs(hx, hy, apples, blocked):
    """Iterative DFS from head, ordering neighbours by Manhattan distance to
    the nearest apple. Records the first-step direction per visited cell so we
    can read off the move once an apple is reached."""
    if not apples:
        return None

    def heuristic(x, y):
        return min(abs(ax - x) + abs(ay - y) for ax, ay in apples)

    apple_set = {cell_key(ax, ay) for ax, ay in apples}
    visited = {cell_key(hx, hy)}
    stack = []

    def push_from(x, y, carried_dir):
        candidates = []
        for dx, dy, name in DIRS:
            nx, ny = x + dx, y + dy
            if not in_bounds(nx, ny):
                continue
            if is_wall(nx, ny):
                continue
            k = cell_key(nx, ny)
            if k in visited:
                continue
            if k in blocked:
                continue
            first_dir = carried_dir if carried_dir is not None else name
            candidates.append((heuristic(nx, ny), k, nx, ny, first_dir))
        # DFS with greedy ordering: push worst first so best is popped next.
        candidates.sort(key=lambda c: -c[0])
        stack.extend(candidates)

    push_from(hx, hy, None)

    best_apple = None
    best_dir = None
    best_dist = float('inf')

    while stack:
        h, k, nx, ny, name = stack.pop()
        if k in visited:
            continue
        visited.add(k)

        if k in apple_set:
            if h < best_dist:
                best_dist = h
                best_apple = (nx, ny)
                best_dir = name
            if best_dist == 0:
                break
            continue

        push_from(nx, ny, name)

    if best_dir is None:
        return None
    return best_apple, best_dir


def facing_name(body):
    if len(body) < 2:
        return None
    hx, hy = body[0]
    nx, ny = body[1]
    dx, dy = hx - nx, hy - ny
    for ddx, ddy, name in DIRS:
        if ddx == dx and ddy == dy:
            return name
    return None


OPPOSITE = {'UP': 'DOWN', 'DOWN': 'UP', 'LEFT': 'RIGHT', 'RIGHT': 'LEFT'}


def safe_fallback(body, blocked):
    """Pick any safe neighbour, never moving backwards. Defaults to the bird's
    current facing so we never submit a backwards move to the engine."""
    hx, hy = body[0]
    facing = facing_name(body)
    backward = OPPOSITE.get(facing)
    for dx, dy, name in DIRS:
        if name == backward:
            continue
        nx, ny = hx + dx, hy + dy
        if not in_bounds(nx, ny):
            continue
        if is_wall(nx, ny):
            continue
        if cell_key(nx, ny) in blocked:
            continue
        return name
    return facing or 'DOWN'


# game loop
while True:
    power_source_count = int(input())
    apples = []
    for i in range(power_source_count):
        x, y = [int(j) for j in input().split()]
        apples.append((x, y))
    snakebot_count = int(input())
    bodies = {}
    occupied = set()
    for i in range(snakebot_count):
        inputs = input().split()
        snakebot_id = int(inputs[0])
        body = inputs[1]
        cells = []
        for cell in body.split(':'):
            cx, cy = map(int, cell.split(','))
            cells.append((cx, cy))
            occupied.add(cell_key(cx, cy))
        bodies[snakebot_id] = cells

    cmds = []
    for bid in my_ids:
        body = bodies.get(bid)
        if body is None:
            continue  # dead
        hx, hy = body[0]
        # Treat all current bodies as blocked, except our own head cell itself.
        blocked = set(occupied)
        blocked.discard(cell_key(hx, hy))

        found = greedy_dfs(hx, hy, apples, blocked)
        direction = found[1] if found else safe_fallback(body, blocked)
        cmds.append(f"{bid} {direction}")

    print(';'.join(cmds))
