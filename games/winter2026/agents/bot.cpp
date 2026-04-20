// Winter Challenge 2026 (SnakeByte) - greedy DFS nearest-apple bot.
// Reads the Codingame stdin protocol and emits moves for each live bird.

#include <iostream>
#include <string>
#include <vector>
#include <unordered_set>
#include <unordered_map>
#include <algorithm>
#include <climits>
#include <sstream>

using namespace std;

struct Dir {
    int dx, dy;
    const char* name;
};

static const Dir DIRS[4] = {
    { 0, -1, "UP"   },
    { 0,  1, "DOWN" },
    {-1,  0, "LEFT" },
    { 1,  0, "RIGHT"},
};

static int gWidth = 0;
static int gHeight = 0;
static vector<string> gWalls;

static inline int cellKey(int x, int y) { return y * gWidth + x; }
static inline bool inBounds(int x, int y) {
    return x >= 0 && x < gWidth && y >= 0 && y < gHeight;
}
static inline bool isWall(int x, int y) { return gWalls[y][x] == '#'; }

struct Candidate {
    int h;
    int k;
    int nx, ny;
    const char* first_dir;
};

// Greedy DFS from head: explores free cells ordering neighbours by Manhattan
// distance to the nearest apple. Carries the first-step direction so we can
// read off the move once a reachable apple is found.
static const char* greedyDFS(int hx, int hy,
                             const vector<pair<int,int>>& apples,
                             const unordered_set<int>& blocked) {
    if (apples.empty()) return nullptr;

    auto heuristic = [&](int x, int y) {
        int best = INT_MAX;
        for (auto& a : apples) {
            int d = abs(a.first - x) + abs(a.second - y);
            if (d < best) best = d;
        }
        return best;
    };

    unordered_set<int> appleSet;
    appleSet.reserve(apples.size() * 2);
    for (auto& a : apples) appleSet.insert(cellKey(a.first, a.second));

    unordered_set<int> visited;
    visited.reserve(gWidth * gHeight);
    visited.insert(cellKey(hx, hy));

    vector<Candidate> stack;
    stack.reserve(64);

    auto pushFrom = [&](int x, int y, const char* carried) {
        Candidate local[4];
        int n = 0;
        for (const auto& d : DIRS) {
            int nx = x + d.dx, ny = y + d.dy;
            if (!inBounds(nx, ny)) continue;
            if (isWall(nx, ny)) continue;
            int k = cellKey(nx, ny);
            if (visited.count(k)) continue;
            if (blocked.count(k)) continue;
            local[n++] = { heuristic(nx, ny), k, nx, ny, carried ? carried : d.name };
        }
        // DFS with greedy ordering: push worst first so best is popped next.
        sort(local, local + n, [](const Candidate& a, const Candidate& b) {
            return a.h > b.h;
        });
        for (int i = 0; i < n; i++) stack.push_back(local[i]);
    };

    pushFrom(hx, hy, nullptr);

    const char* bestDir = nullptr;
    int bestDist = INT_MAX;

    while (!stack.empty()) {
        Candidate c = stack.back();
        stack.pop_back();
        if (visited.count(c.k)) continue;
        visited.insert(c.k);

        if (appleSet.count(c.k)) {
            if (c.h < bestDist) {
                bestDist = c.h;
                bestDir = c.first_dir;
            }
            if (bestDist == 0) break;
            continue;
        }

        pushFrom(c.nx, c.ny, c.first_dir);
    }

    return bestDir;
}

static const char* facingName(const vector<pair<int,int>>& body) {
    if (body.size() < 2) return nullptr;
    int dx = body[0].first - body[1].first;
    int dy = body[0].second - body[1].second;
    for (const auto& d : DIRS) {
        if (d.dx == dx && d.dy == dy) return d.name;
    }
    return nullptr;
}

static const char* oppositeName(const char* name) {
    if (!name) return nullptr;
    string s(name);
    if (s == "UP")    return "DOWN";
    if (s == "DOWN")  return "UP";
    if (s == "LEFT")  return "RIGHT";
    if (s == "RIGHT") return "LEFT";
    return nullptr;
}

// Fallback: pick any safe neighbour, never moving backwards. Defaults to the
// bird's current facing so we never submit a backwards move to the engine.
static const char* safeFallback(const vector<pair<int,int>>& body,
                                const unordered_set<int>& blocked) {
    int hx = body[0].first, hy = body[0].second;
    const char* facing = facingName(body);
    const char* backward = oppositeName(facing);
    for (const auto& d : DIRS) {
        if (backward && string(d.name) == backward) continue;
        int nx = hx + d.dx, ny = hy + d.dy;
        if (!inBounds(nx, ny)) continue;
        if (isWall(nx, ny)) continue;
        if (blocked.count(cellKey(nx, ny))) continue;
        return d.name;
    }
    return facing ? facing : "DOWN";
}

int main() {
    ios::sync_with_stdio(false);

    int my_id;
    cin >> my_id; cin.ignore();
    cin >> gWidth; cin.ignore();
    cin >> gHeight; cin.ignore();
    gWalls.resize(gHeight);
    for (int i = 0; i < gHeight; i++) {
        getline(cin, gWalls[i]);
    }

    int snakebotsPerPlayer;
    cin >> snakebotsPerPlayer; cin.ignore();
    vector<int> myIds;
    myIds.reserve(snakebotsPerPlayer);
    for (int i = 0; i < snakebotsPerPlayer; i++) {
        int id;
        cin >> id; cin.ignore();
        myIds.push_back(id);
    }
    for (int i = 0; i < snakebotsPerPlayer; i++) {
        int id;
        cin >> id; cin.ignore();
    }

    // game loop
    while (true) {
        int powerSourceCount;
        if (!(cin >> powerSourceCount)) break;
        cin.ignore();

        vector<pair<int,int>> apples;
        apples.reserve(powerSourceCount);
        for (int i = 0; i < powerSourceCount; i++) {
            int x, y;
            cin >> x >> y; cin.ignore();
            apples.emplace_back(x, y);
        }

        int snakebotCount;
        cin >> snakebotCount; cin.ignore();

        unordered_map<int, vector<pair<int,int>>> bodies;
        unordered_set<int> occupied;
        occupied.reserve(snakebotCount * 16);

        for (int i = 0; i < snakebotCount; i++) {
            int id;
            string body;
            cin >> id >> body; cin.ignore();

            vector<pair<int,int>> cells;
            stringstream ss(body);
            string cell;
            while (getline(ss, cell, ':')) {
                int cx = 0, cy = 0;
                size_t comma = cell.find(',');
                cx = stoi(cell.substr(0, comma));
                cy = stoi(cell.substr(comma + 1));
                cells.emplace_back(cx, cy);
                occupied.insert(cellKey(cx, cy));
            }
            bodies[id] = std::move(cells);
        }

        vector<string> cmds;
        cmds.reserve(myIds.size());
        for (int id : myIds) {
            auto it = bodies.find(id);
            if (it == bodies.end()) continue; // dead
            const auto& body = it->second;
            int hx = body[0].first, hy = body[0].second;

            // Treat all current bodies as blocked, except our own head cell itself.
            unordered_set<int> blocked = occupied;
            blocked.erase(cellKey(hx, hy));

            const char* dir = greedyDFS(hx, hy, apples, blocked);
            if (!dir) dir = safeFallback(body, blocked);
            cmds.push_back(to_string(id) + " " + dir);
        }

        for (size_t i = 0; i < cmds.size(); i++) {
            if (i) cout << ';';
            cout << cmds[i];
        }
        cout << endl;
    }

    return 0;
}
