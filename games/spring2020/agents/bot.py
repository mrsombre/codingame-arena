#!/usr/bin/env python3
# Spring Challenge 2020 (PacMan) - inert echo bot.
# Reads the full Codingame stdin protocol, mirrors everything to stderr for
# debugging, and emits a no-op WAIT for every one of its pacs.

import sys


def log(msg):
    print(msg, file=sys.stderr, flush=True)


header = input()
log(f"<< {header}")
width, height = [int(i) for i in header.split()]

for _ in range(height):
    row = input()
    log(f"<< {row}")

turn = 0
while True:
    line = input()
    log(f"<< {line}")
    my_score, opponent_score = [int(i) for i in line.split()]

    line = input()
    log(f"<< {line}")
    visible_pac_count = int(line)

    my_pac_ids = []
    for _ in range(visible_pac_count):
        line = input()
        log(f"<< {line}")
        parts = line.split()
        pac_id = int(parts[0])
        mine = parts[1] != "0"
        if mine and parts[4] != "DEAD":
            my_pac_ids.append(pac_id)

    line = input()
    log(f"<< {line}")
    visible_pellet_count = int(line)
    for _ in range(visible_pellet_count):
        line = input()
        log(f"<< {line}")

    if my_pac_ids:
        cmd = " | ".join(f"WAIT {pid}" for pid in my_pac_ids)
    else:
        cmd = "WAIT 0"
    log(f">> {cmd}")
    print(cmd)
    turn += 1
