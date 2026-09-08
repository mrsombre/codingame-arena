#!/usr/bin/env python3
# Summer Challenge 2026 (Back Track King) - simple heuristic bot.
# Targets the nearest desired connection that is not active yet and works on
# it every turn, alternating between AUTOPLACE and hand-placed PLACE_TRACKS
# along a cheapest path of its own so both commands stay exercised.

import heapq

TYPE_WATER = 1
TYPE_MOUNTAIN = 2
TYPE_POI = 3

TRACK_NONE = -1
PAINT_PER_TURN = 3

DIRS = [(0, -1), (1, 0), (0, 1), (-1, 0)]

my_id = int(input())
width = int(input())
height = int(input())
size = width * height

cell_type = [0] * size
for i in range(size):
    _zone, cell_type[i] = map(int, input().split())

town_at = [-1] * size
towns = []
town_count = int(input())
for _ in range(town_count):
    parts = input().split()
    town = {
        'id': int(parts[0]),
        'x': int(parts[1]),
        'y': int(parts[2]),
        'desired': [] if parts[3] == 'x' else [int(t) for t in parts[3].split(',')],
    }
    towns.append(town)
    town_at[town['y'] * width + town['x']] = town['id']

# Town ids can repeat: the generator's fallback placement loop renumbers from 0
# without checking. First occurrence wins, like the engine's lookup.
index_of_town_id = {}
for i, town in enumerate(towns):
    index_of_town_id.setdefault(town['id'], i)

NEIGHBOURS = []
for c in range(size):
    x, y = c % width, c // width
    NEIGHBOURS.append([
        (x + dx) + (y + dy) * width
        for dx, dy in DIRS
        if 0 <= x + dx < width and 0 <= y + dy < height
    ])


def rail_cost(t):
    if t == TYPE_WATER:
        return 2
    if t in (TYPE_MOUNTAIN, TYPE_POI):
        return 3
    return 1


def cheapest_path(start, goal, track, inked):
    """Cells from start to goal inclusive along the route costing the fewest
    paint points, or None when none exists. Towns and existing tracks are free
    to cross; inked cells cannot be painted."""
    dist = [None] * size
    prev = [-1] * size
    dist[start] = 0
    queue = [(0, start)]
    while queue:
        d, cell = heapq.heappop(queue)
        if d > dist[cell]:
            continue
        if cell == goal:
            break
        for nb in NEIGHBOURS[cell]:
            free = town_at[nb] >= 0 or track[nb] != TRACK_NONE
            if not free and inked[nb]:
                continue
            nd = d + (0 if free else rail_cost(cell_type[nb]))
            if dist[nb] is None or nd < dist[nb]:
                dist[nb] = nd
                prev[nb] = cell
                heapq.heappush(queue, (nd, nb))

    if dist[goal] is None:
        return None
    path = []
    cell = goal
    while cell != -1:
        path.append(cell)
        cell = prev[cell]
    path.reverse()
    return path


def placements(path, track):
    """The PLACE_TRACKS actions along path that are affordable this turn."""
    actions = []
    budget = PAINT_PER_TURN
    for cell in path:
        if town_at[cell] >= 0 or track[cell] != TRACK_NONE:
            continue
        cost = rail_cost(cell_type[cell])
        if cost > budget:
            break
        budget -= cost
        actions.append(f'PLACE_TRACKS {cell % width} {cell // width}')
    return actions


turn = 0
while True:
    try:
        input()  # my score (unused)
        input()  # foe score (unused)
    except EOFError:
        break

    track = [TRACK_NONE] * size
    inked = [0] * size
    active = set()
    for i in range(size):
        owner, _instability, ink, connections = input().split()
        track[i] = int(owner)
        inked[i] = int(ink)
        if connections != 'x':
            active.update(connections.split(','))
    turn += 1

    # Every desired connection still missing, nearest first.
    pending = []
    for town in towns:
        for wanted in town['desired']:
            target = index_of_town_id.get(wanted)
            if target is None:
                continue
            if f"{town['id']}-{wanted}" in active:
                continue
            other = towns[target]
            d = abs(town['x'] - other['x']) + abs(town['y'] - other['y'])
            pending.append((d, town['y'] * width + town['x'], other['y'] * width + other['x']))
    # Player 1 works the list from the far end so the two sides chase different
    # connections; identical bots would only ever paint each other's cells
    # neutral and neither would score.
    pending.sort(reverse=my_id == 1)

    out = ''
    for _d, start, goal in pending:
        path = cheapest_path(start, goal, track, inked)
        if path is None:
            continue
        if turn % 2 == 1:
            out = (f'AUTOPLACE {start % width} {start // width}'
                   f' {goal % width} {goal // width}')
            break
        actions = placements(path, track)
        if not actions:
            continue
        out = ';'.join(actions)
        break

    print(out if out else 'WAIT')
