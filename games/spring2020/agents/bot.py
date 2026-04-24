#!/usr/bin/env python3
# Spring Challenge 2020 (PacMan) - simple greedy bot.
# Strategy per alive pac:
#   1) SPEED when the ability is off cooldown.
#   2) BFS to the nearest visible pellet. When speeding, prefer a target that is
#      at least `speed` cells away so the speed sub-turn still has path left to
#      traverse; fall back to extending past a close pellet otherwise.
#   3) If nothing is visible, BFS to the nearest unexplored floor cell to scout.

from collections import deque
import sys


def log(msg):
    print(msg, file=sys.stderr, flush=True)


header = input()
width, height = [int(v) for v in header.split()]
walls = []
for _ in range(height):
    walls.append(input())

floor = set()
for y in range(height):
    for x in range(width):
        if walls[y][x] != '#':
            floor.add((x, y))

ADJACENCY = ((-1, 0), (1, 0), (0, -1), (0, 1))


def neighbors(cell):
    x, y = cell
    for dx, dy in ADJACENCY:
        nx, ny = (x + dx + width) % width, y + dy  # wrap horizontally
        if (nx, ny) in floor:
            yield (nx, ny)


def bfs_from(start):
    """Returns (dist, parent) dicts over all floor cells reachable from start."""
    dist = {start: 0}
    parent = {start: None}
    queue = deque([start])
    while queue:
        cell = queue.popleft()
        for n in neighbors(cell):
            if n not in dist:
                dist[n] = dist[cell] + 1
                parent[n] = cell
                queue.append(n)
    return dist, parent


def path_crosses(parent, cell, waypoint):
    """True if the shortest-path tree path from cell back to the BFS root
    passes through waypoint."""
    while cell is not None:
        if cell == waypoint:
            return True
        cell = parent[cell]
    return False


# Cells we have ever had line of sight on — used as exploration negative space.
explored = set()

while True:
    my_score, opp_score = [int(v) for v in input().split()]
    visible_pac_count = int(input())

    my_pacs = []  # (pac_id, x, y, speed, cooldown)
    for _ in range(visible_pac_count):
        parts = input().split()
        pac_id = int(parts[0])
        mine = parts[1] != "0"
        x, y = int(parts[2]), int(parts[3])
        kind = parts[4]
        speed_turns = int(parts[5])
        cooldown = int(parts[6])
        if kind == "DEAD" or not mine:
            continue
        # speed=2 while speed_turns > 0, else 1.
        speed = 2 if speed_turns > 0 else 1
        my_pacs.append((pac_id, x, y, speed, cooldown))

    pellet_cells = {}
    visible_pellet_count = int(input())
    for _ in range(visible_pellet_count):
        x, y, value = [int(v) for v in input().split()]
        pellet_cells[(x, y)] = value

    # Anything we currently see counts as explored.
    for pac_id, x, y, _, _ in my_pacs:
        explored.add((x, y))
    explored.update(pellet_cells.keys())

    claimed = set()
    commands = []
    for pac_id, x, y, speed, cooldown in my_pacs:
        if cooldown == 0 and speed == 1:
            commands.append(f"SPEED {pac_id}")
            continue

        dist, parent = bfs_from((x, y))

        # Closest reachable unclaimed pellet.
        closest_pellet = None
        closest_d = None
        for p in pellet_cells:
            if p in claimed or p not in dist:
                continue
            d = dist[p]
            if closest_d is None or d < closest_d:
                closest_pellet = p
                closest_d = d

        target = None
        if closest_pellet is not None:
            target = closest_pellet
            claimed.add(closest_pellet)

            # Speeding + pellet too close: extend the target past the pellet
            # so the speed sub-turn still has path to consume, but only if
            # that path still passes through the pellet. If the pellet is in
            # a dead-end, accept the 1-cell move and eat it.
            if closest_d < speed:
                for c, d in dist.items():
                    if d != speed:
                        continue
                    if path_crosses(parent, c, closest_pellet):
                        target = c
                        break
        else:
            # Scout: nearest unexplored floor cell.
            best = None
            best_d = None
            for c, d in dist.items():
                if c in explored:
                    continue
                if best_d is None or d < best_d:
                    best = c
                    best_d = d
            target = best

        if target is None:
            target = (x, y)

        commands.append(f"MOVE {pac_id} {target[0]} {target[1]}")

    print(" | ".join(commands) if commands else "MOVE 0 0 0")
