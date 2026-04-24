// Spring Challenge 2020 (PacMan) - simple greedy bot.
// Strategy per alive pac:
//   1) SPEED when the ability is off cooldown.
//   2) BFS to the nearest visible pellet. When speeding, prefer a target that
//      is at least `speed` cells away so the speed sub-turn still has path
//      left to traverse; if the pellet is in a dead-end, accept a 1-cell move.
//   3) If nothing visible, BFS to the nearest unexplored floor cell to scout.

#include <algorithm>
#include <array>
#include <iostream>
#include <queue>
#include <sstream>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <vector>

using namespace std;

static int gWidth = 0;
static int gHeight = 0;
static vector<string> gWalls;
static vector<bool> gFloor; // indexed by cellKey

static constexpr array<pair<int, int>, 4> ADJ = {{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}};

static inline int cellKey(int x, int y) { return y * gWidth + x; }
static inline bool isFloor(int x, int y) {
    if (y < 0 || y >= gHeight) return false;
    return gFloor[cellKey(x, y)];
}

// Horizontally wrapping neighbour enumeration over floor cells.
static void neighbours(int x, int y, vector<pair<int, int>>& out) {
    out.clear();
    for (auto [dx, dy] : ADJ) {
        int nx = (x + dx + gWidth) % gWidth;
        int ny = y + dy;
        if (isFloor(nx, ny)) out.emplace_back(nx, ny);
    }
}

struct BFSResult {
    vector<int> dist;   // -1 if unreached
    vector<int> parent; // -1 for root or unreached
};

static BFSResult bfs(int sx, int sy) {
    BFSResult r;
    r.dist.assign(gWidth * gHeight, -1);
    r.parent.assign(gWidth * gHeight, -1);
    queue<pair<int, int>> q;
    int startKey = cellKey(sx, sy);
    r.dist[startKey] = 0;
    q.emplace(sx, sy);
    vector<pair<int, int>> neighBuf;
    while (!q.empty()) {
        auto [x, y] = q.front();
        q.pop();
        int k = cellKey(x, y);
        neighbours(x, y, neighBuf);
        for (auto [nx, ny] : neighBuf) {
            int nk = cellKey(nx, ny);
            if (r.dist[nk] != -1) continue;
            r.dist[nk] = r.dist[k] + 1;
            r.parent[nk] = k;
            q.emplace(nx, ny);
        }
    }
    return r;
}

static bool pathCrosses(const vector<int>& parent, int cell, int waypoint) {
    while (cell != -1) {
        if (cell == waypoint) return true;
        cell = parent[cell];
    }
    return false;
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    cin >> gWidth >> gHeight;
    cin.ignore();
    gWalls.resize(gHeight);
    for (int i = 0; i < gHeight; i++) {
        getline(cin, gWalls[i]);
    }

    gFloor.assign(gWidth * gHeight, false);
    for (int y = 0; y < gHeight; y++) {
        for (int x = 0; x < gWidth; x++) {
            gFloor[cellKey(x, y)] = (gWalls[y][x] != '#');
        }
    }

    unordered_set<int> explored;
    explored.reserve(gWidth * gHeight);

    while (true) {
        int myScore, oppScore;
        if (!(cin >> myScore >> oppScore)) break;

        int visiblePacCount;
        cin >> visiblePacCount;

        struct Pac {
            int id, x, y, speed, cooldown;
        };
        vector<Pac> myPacs;
        myPacs.reserve(visiblePacCount);

        for (int i = 0; i < visiblePacCount; i++) {
            int pacId, mineFlag, x, y, speedTurns, cooldown;
            string kind;
            cin >> pacId >> mineFlag >> x >> y >> kind >> speedTurns >> cooldown;
            if (kind == "DEAD" || mineFlag == 0) continue;
            int speed = speedTurns > 0 ? 2 : 1;
            myPacs.push_back({pacId, x, y, speed, cooldown});
        }

        unordered_map<int, int> pelletValue; // cellKey -> value
        int visiblePelletCount;
        cin >> visiblePelletCount;
        pelletValue.reserve(visiblePelletCount * 2);
        for (int i = 0; i < visiblePelletCount; i++) {
            int x, y, value;
            cin >> x >> y >> value;
            pelletValue[cellKey(x, y)] = value;
        }

        for (const auto& pac : myPacs) {
            explored.insert(cellKey(pac.x, pac.y));
        }
        for (const auto& kv : pelletValue) {
            explored.insert(kv.first);
        }

        unordered_set<int> claimed;
        vector<string> commands;
        commands.reserve(myPacs.size());

        for (const auto& pac : myPacs) {
            if (pac.cooldown == 0 && pac.speed == 1) {
                commands.push_back("SPEED " + to_string(pac.id));
                continue;
            }

            BFSResult r = bfs(pac.x, pac.y);

            int closestPellet = -1;
            int closestDist = -1;
            for (const auto& kv : pelletValue) {
                int k = kv.first;
                if (claimed.count(k)) continue;
                int d = r.dist[k];
                if (d < 0) continue;
                if (closestPellet == -1 || d < closestDist) {
                    closestPellet = k;
                    closestDist = d;
                }
            }

            int targetKey = -1;
            if (closestPellet != -1) {
                targetKey = closestPellet;
                claimed.insert(closestPellet);

                // Speeding + pellet too close: extend past the pellet so the
                // speed sub-turn still has path. If no extension passes
                // through the pellet (dead-end), keep the 1-cell move.
                if (closestDist < pac.speed) {
                    for (int k = 0; k < (int)r.dist.size(); k++) {
                        if (r.dist[k] != pac.speed) continue;
                        if (pathCrosses(r.parent, k, closestPellet)) {
                            targetKey = k;
                            break;
                        }
                    }
                }
            } else {
                int bestUnexplored = -1;
                int bestDist = -1;
                for (int k = 0; k < (int)r.dist.size(); k++) {
                    int d = r.dist[k];
                    if (d < 0) continue;
                    if (explored.count(k)) continue;
                    if (bestUnexplored == -1 || d < bestDist) {
                        bestUnexplored = k;
                        bestDist = d;
                    }
                }
                targetKey = bestUnexplored;
            }

            if (targetKey == -1) {
                commands.push_back("MOVE " + to_string(pac.id) + " " +
                                   to_string(pac.x) + " " + to_string(pac.y));
                continue;
            }

            int tx = targetKey % gWidth;
            int ty = targetKey / gWidth;
            commands.push_back("MOVE " + to_string(pac.id) + " " +
                               to_string(tx) + " " + to_string(ty));
        }

        if (commands.empty()) {
            cout << "MOVE 0 0 0" << endl;
        } else {
            for (size_t i = 0; i < commands.size(); i++) {
                if (i) cout << " | ";
                cout << commands[i];
            }
            cout << endl;
        }
    }

    return 0;
}
