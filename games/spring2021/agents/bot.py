#!/usr/bin/env python3
# Spring Challenge 2021 (Photosynthesis) - simple greedy bot.
# Picks from the engine-provided list of legal actions with this priority:
#   1) COMPLETE a size-3 tree on the richest cell, but only late in the game
#      so we don't lose sun income too early.
#   2) GROW the largest available tree (size 2 -> 3, then 1 -> 2, then seed -> 1).
#   3) SEED toward a high-richness cell if we don't already own a seed.
#   4) WAIT.

import sys


def log(msg):
    print(msg, file=sys.stderr, flush=True)


LAST_DAY = 23
COMPLETE_FROM_DAY = 12  # don't chop trees too early; sun income compounds

number_of_cells = int(input())
richness = [0] * number_of_cells
for _ in range(number_of_cells):
    parts = [int(j) for j in input().split()]
    idx = parts[0]
    richness[idx] = parts[1]

while True:
    day = int(input())
    nutrients = int(input())
    sun, score = [int(i) for i in input().split()]
    inputs = input().split()
    opp_sun = int(inputs[0])
    opp_score = int(inputs[1])
    opp_is_waiting = inputs[2] != "0"

    number_of_trees = int(input())
    my_sizes = {}  # cell_index -> size for my non-dormant trees
    occupied = set()
    for _ in range(number_of_trees):
        parts = input().split()
        cell_index = int(parts[0])
        size = int(parts[1])
        is_mine = parts[2] != "0"
        is_dormant = parts[3] != "0"
        occupied.add(cell_index)
        if is_mine and not is_dormant:
            my_sizes[cell_index] = size

    my_seed_count = sum(1 for s in my_sizes.values() if s == 0)

    completes = []  # (richness, cellIdx, raw)
    grows = []     # (size, richness, cellIdx, raw) - higher size first
    seeds = []     # (richness, sourceIdx, targetIdx, raw)

    number_of_possible_actions = int(input())
    for _ in range(number_of_possible_actions):
        raw = input()
        tokens = raw.split()
        verb = tokens[0]
        if verb == "COMPLETE":
            cell_idx = int(tokens[1])
            completes.append((richness[cell_idx], cell_idx, raw))
        elif verb == "GROW":
            cell_idx = int(tokens[1])
            size = my_sizes.get(cell_idx, 0)
            grows.append((size, richness[cell_idx], cell_idx, raw))
        elif verb == "SEED":
            source_idx = int(tokens[1])
            target_idx = int(tokens[2])
            seeds.append((richness[target_idx], source_idx, target_idx, raw))

    move = "WAIT"

    if day >= COMPLETE_FROM_DAY and completes:
        completes.sort(key=lambda c: (-c[0], c[1]))
        move = completes[0][2]
    elif grows:
        # Prefer the largest tree, breaking ties by richness then index.
        grows.sort(key=lambda g: (-g[0], -g[1], g[2]))
        move = grows[0][3]
    elif my_seed_count == 0 and seeds:
        seeds.sort(key=lambda s: (-s[0], s[2]))
        move = seeds[0][3]
    elif day == LAST_DAY and completes:
        # Final day: cash in any size-3 trees we still have.
        completes.sort(key=lambda c: (-c[0], c[1]))
        move = completes[0][2]

    print(move)
