from __future__ import annotations
import math
import sys
from dataclasses import dataclass
from typing import List, Dict, Optional, Set
from enum import Enum
import random
import time
import json
from collections import defaultdict



def log(*t):
  print(*t, file=sys.stderr, flush=True)

@dataclass(frozen=True)
class Coord:
    x: int
    y: int
    def __repr__(self) -> str:
        return f"{self.x} {self.y}"

class Direction(Enum):
    N = Coord(0,-1)
    E = Coord(1,0)
    S = Coord(0,1)
    W = Coord(-1,0)

def get_opposite_direction(d: Direction) -> Direction:
    if d == Direction.N:
        return Direction.S
    elif d == Direction.E:
        return Direction.W
    elif d == Direction.S:
        return Direction.N
    elif d == Direction.W:
        return Direction.E

@dataclass
class TileData:
    type:int
    zone_id:int

@dataclass
class Tile:
    data: TileData
    tracks_owner:int
    is_inked:bool
    instability: int
    part_of_active_route: bool

@dataclass
class Town:
    id: int
    coord: Coord
    desired_connections: List[int]

@dataclass
class Train:
    id: int
    current_town_id: int
    next_town_ready: bool
@dataclass
class Schedule:
    id: int
    towns: List[int]

@dataclass
class Grid:
    w: int
    h: int
    globalTiles: List[List[TileData]]
    tiles: List[List[Tile]]

@dataclass
class Zone:
    id: int
    instability: int
    inked: bool
    coords: List[Coord]
    has_town: bool

class Game:
    w: int
    h: int
    turn: int
    my_index: int
    grid: Grid
    towns: List[Town]
    townMap: Dict[Coord, Town]
    zones: Dict[int, Zone]
    zoneMap: Dict[Coord, Zone]

    def init(self):
        self.turn = 0
        self.my_index = int(input())
        self.zones = {}
        self.zoneMap = {}
        self.towns = []
        self.townMap = {}
        
        random.seed('🚂🚂🚂' * (1 + (self.my_index))) 

        width = int(input())
        height = int(input())
        self.grid = Grid(w=width, h=height, tiles=[], globalTiles=[])
        for i in range(height):
            row = []
            for j in range(width):
                inputs = input().split()
                zone_id = int(inputs[0])
                type = int(inputs[1])
                tile = TileData(type=type, zone_id=zone_id)
                row.append(tile)
                if zone_id not in self.zones:
                    self.zones[zone_id] = Zone(id=zone_id, instability=0, inked=False, coords=[], has_town=False)
                zone = self.zones[zone_id]
                coord = Coord(x=j, y=i)
                zone.coords.append(coord)
                self.zoneMap[coord] = zone

            self.grid.globalTiles.append(row)
        
        town_count = int(input())
        for i in range(town_count):
            inputs = input().split()
            town_id = int(inputs[0])
            x = int(inputs[1])
            y = int(inputs[2])
            desired_connections = [] if inputs[3] == 'x' else list(map(int, inputs[3].split(',')))
            coord = Coord(x=x, y=y)
            town = Town(id=town_id, coord=coord, desired_connections=desired_connections)
            self.towns.append(town)
            self.townMap[coord] = town
            self.zones[self.grid.globalTiles[y][x].zone_id].has_town = True
        
        self.w = width
        self.h = height
  
    def parse(self):
        _score = int(input())
        _opponent_score = int(input())
        self.trains = []
        
        self.grid.tiles = []
        for y in range(self.h):
            row = []
            for x in range(self.w):
                inputs = input().split()
                tracks_owner = int(inputs[0])
                instability = int(inputs[1])
                is_inked = inputs[2] == '1'
                part_of_active_route = inputs[3] != 'x'
                tileData = self.grid.globalTiles[y][x]
                tile = Tile(tileData, tracks_owner, is_inked, instability, part_of_active_route)
                self.zones[tileData.zone_id].instability = instability
                self.zones[tileData.zone_id].inked = is_inked
                row.append(tile)
            self.grid.tiles.append(row)

    def prep(self):
        self.turn += 1
        # Flood fill tracks to find groups of connected cities
        self.groups = []
        visited = set()
        for town in self.towns:
            if town.id in visited:
                continue
            group = set()
            stack = [town.coord]
            while stack:
                current_coord = stack.pop()
                if current_coord in visited:
                    continue
                visited.add(current_coord)
                group.add(current_coord)
                neighbour_coords = self.get_neighbours(current_coord)
                for neigh in neighbour_coords:
                    tile = self.grid.tiles[neigh.y][neigh.x]
                    if tile.tracks_owner > -1 or self.is_town(neigh):
                        stack.append(neigh)
            if group:
                self.groups.append(group)



    def is_town(self, coord:Coord):
        for town in self.towns:
            if town.coord == coord:
                return True
        return False

    def get_mins(self, elements, key=None):
        result = []
        by = 0
        for elem in elements:
            if key:
                elem_key = key(elem)
            else:
                elem_key = elem
            if not result or by > elem_key:
                result.clear()
                result.append(elem)
                by = elem_key
            elif result and by == elem_key:
                result.append(elem)
        return result

    def get_neighbours(self, pos:Coord, dirs=None):
        if dirs is None:
            dirs = [d.value for d in Direction]
        for d in dirs:
            neigh = Coord(pos.x + d.x, pos.y + d.y)
            if self.within_bounds(neigh):
                yield neigh

    def within_bounds(self, coord:Coord):
        return coord.x >= 0 and coord.x < self.w and coord.y >= 0 and coord.y < self.h
               

    def get_closest(self, pos:Coord, elements, pos_key=None, dist_func=lambda a,b: manhattan(a,b)):
        closest = []
        closest_by = 0
        for elem in elements:
            if pos_key:
                elem_pos = pos_key(elem)
            else:
                elem_pos = elem
            distance = dist_func(pos, elem_pos)
            if not closest or closest_by > distance:
                closest.clear()
                closest.append(elem)
                closest_by = distance
            elif closest and closest_by == distance:
                closest.append(elem)
        return closest

    def get_coords_around(self, pos:Coord, distance: int, dist_func=lambda a,b: manhattan(a,b)):
        result = []
        for y in range(-distance, distance + 1):
            for x in range(-distance, distance + 1):
                if dist_func(Coord(x,y), Coord(0,0)) < distance:
                    continue
                coord = Coord(pos.x + x, pos.y + y)
                if not self.within_bounds(coord):
                    continue
                tile = self.grid.tiles[coord.y][coord.x]
                if tile == 0:
                    result.append(coord)
        return result



    def tick(self):
        actions = []
        
        town_a = random.choice(self.towns)
        town_b = random.choice([t for t in self.towns if t.id != town_a.id])
        actions.append(f'AUTOPLACE {town_a.coord.x} {town_a.coord.y} {town_b.coord.x} {town_b.coord.y}')

        if actions:
            print(';'.join(actions))
        else:
            print('WAIT')

    def count_tracks_owned_by(self, group:Set[Coord], owner:int):
        count = 0
        for coord in group:
            tile = self.grid.tiles[coord.y][coord.x]
            if tile.tracks_owner == owner or tile.tracks_owner == 2:
                count += 1
        return count

    def get_most_direct_path(self, a: "Coord", b: "Coord") -> List["Coord"]:
        path: List[Coord] = []
        x, y = a.x, a.y
        tx, ty = b.x, b.y

        while (x, y) != (tx, ty):
            if y > ty:         # NORTH
                y -= 1
            elif x < tx:       # EAST
                x += 1
            elif y < ty:       # SOUTH
                y += 1
            elif x > tx:       # WEST
                x -= 1
            path.append(Coord(x, y))

        return path

def manhattan(a:Coord, b:Coord):
    return abs(a.x - b.x) + abs(a.y - b.y)

def sqrDistance(a:Coord, b:Coord):
    return (a.x - b.x) ** 2 +(a.y - b.y) ** 2 

def sqr_distance_float(ax:float, ay: float, bx:float, by: float):
    return (ax - bx) ** 2 +(ay - by) ** 2 

def get_closest_coord(coords:List[Coord], pos:Coord) -> Coord | None:
    closest = None
    by = 0
    for c in coords:
        dist = sqrDistance(c, pos)
        if closest is None or by > dist:
            closest = c
            by = dist
    return closest

def chebyshev_distance(a:Coord, b:Coord):
    return max(abs(a.x - b.x),abs(a.x - b.x)) 



def main():
  game = Game()
  game.init()
  while True:
    game.parse()
    
    game.prep()
    game.tick()
main()