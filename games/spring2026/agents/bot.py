#!/usr/bin/env python3
# Spring Challenge 2026 (Troll Farm) - simple heuristic bot.
# Each troll: drop near shack, harvest if standing on a fruited tree, wait on
# a size-4 tree empty-handed, otherwise head for the most valuable tree by
# Manhattan distance.

import sys

PLAYER_ME = 0
PLAYER_OP = 1

ITEM_PLUM = 0
ITEM_LEMON = 1
ITEM_APPLE = 2
ITEM_BANANA = 3
ITEM_IRON = 4
ITEM_WOOD = 5
NUM_ITEMS = 6

FRUIT_NAMES = {'PLUM': ITEM_PLUM, 'LEMON': ITEM_LEMON, 'APPLE': ITEM_APPLE, 'BANANA': ITEM_BANANA}

DIRS = [(0, -1), (1, 0), (0, 1), (-1, 0)]

width, height = map(int, input().split())
cells = [['.'] * width for _ in range(height)]
shack = [-1, -1]

for y in range(height):
    row = input()
    for x in range(width):
        cells[y][x] = row[x]
        if row[x] == '0':
            shack[PLAYER_ME] = y * width + x
        elif row[x] == '1':
            shack[PLAYER_OP] = y * width + x


def adj_shack_me(c):
    x, y = c % width, c // width
    for dx, dy in DIRS:
        nx, ny = x + dx, y + dy
        if 0 <= nx < width and 0 <= ny < height and cells[ny][nx] == '0':
            return True
    return False


def tree_at(trees, cell):
    for t in trees:
        if t['cell'] == cell:
            return t
    return None


def best_tree_target(trees, from_cell):
    fx, fy = from_cell % width, from_cell // width
    best = -1
    best_dist = 0
    best_tier = -1
    for t in trees:
        tx, ty = t['cell'] % width, t['cell'] // width
        dist = abs(fx - tx) + abs(fy - ty)
        if t['fruits'] > 0:
            tier = 2
        elif t['size'] == 4:
            tier = 1
        else:
            tier = 0
        if tier > best_tier or (tier == best_tier and (best < 0 or dist < best_dist)):
            best = t['cell']
            best_dist = dist
            best_tier = tier
    return best


def decide_troll(tr, trees):
    cell = tr['cell']
    carry = sum(tr['carry'])
    fruits = tr['carry'][ITEM_PLUM] + tr['carry'][ITEM_LEMON] + tr['carry'][ITEM_APPLE] + tr['carry'][ITEM_BANANA]
    full = carry >= tr['carryCap']

    if carry > 0 and adj_shack_me(cell):
        return f"DROP {tr['id']}"

    if not full:
        tree = tree_at(trees, cell)
        if tree is not None:
            if tree['fruits'] > 0:
                return f"HARVEST {tr['id']}"
            if fruits == 0 and tree['size'] == 4:
                return "WAIT"

    if full or fruits > 0:
        s = shack[PLAYER_ME]
        return f"MOVE {tr['id']} {s % width} {s // width}"

    target = best_tree_target(trees, cell)
    if target >= 0:
        return f"MOVE {tr['id']} {target % width} {target // width}"
    return "WAIT"


# game loop
while True:
    try:
        input()  # my inventory (unused)
        input()  # opp inventory (unused)
    except EOFError:
        break

    tree_count = int(input())
    trees = []
    for _ in range(tree_count):
        parts = input().split()
        trees.append({
            'cell': int(parts[2]) * width + int(parts[1]),
            'type': FRUIT_NAMES.get(parts[0], ITEM_PLUM),
            'size': int(parts[3]),
            'health': int(parts[4]),
            'fruits': int(parts[5]),
            'cooldown': int(parts[6]),
        })

    troll_count = int(input())
    trolls = []
    for _ in range(troll_count):
        p = input().split()
        trolls.append({
            'id': int(p[0]),
            'player': int(p[1]),
            'cell': int(p[3]) * width + int(p[2]),
            'moveSpeed': int(p[4]),
            'carryCap': int(p[5]),
            'harvestPower': int(p[6]),
            'chopPower': int(p[7]),
            'carry': [int(p[8]), int(p[9]), int(p[10]), int(p[11]), int(p[12]), int(p[13])],
        })

    cmds = [decide_troll(tr, trees) for tr in trolls if tr['player'] == PLAYER_ME]
    print(';'.join(cmds) if cmds else 'WAIT')
