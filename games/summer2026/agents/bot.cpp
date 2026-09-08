// Summer Challenge 2026 (Back Track King) - simple heuristic bot.
// Targets the nearest desired connection that is not active yet and works on
// it every turn, alternating between AUTOPLACE and hand-placed PLACE_TRACKS
// along a cheapest path of its own so both commands stay exercised.

#include <algorithm>
#include <climits>
#include <iostream>
#include <queue>
#include <set>
#include <sstream>
#include <string>
#include <vector>

using namespace std;

static const int TYPE_WATER = 1;
static const int TYPE_MOUNTAIN = 2;
static const int TYPE_POI = 3;

static const int TRACK_NONE = -1;
static const int PAINT_PER_TURN = 3;

static const int DX[4] = {0, 1, 0, -1};
static const int DY[4] = {-1, 0, 1, 0};

struct Town {
    int id;
    int x;
    int y;
    vector<int> desired;
};

static int gWidth = 0;
static int gHeight = 0;
static int gMyId = 0;
static vector<int> gType;
static vector<int> gTownAt;
static vector<Town> gTowns;
static vector<int> gIndexOfTownId;

static inline int xy(int x, int y) { return y * gWidth + x; }
static inline int cellX(int c) { return c % gWidth; }
static inline int cellY(int c) { return c / gWidth; }

static int railCost(int type) {
    if (type == TYPE_WATER) return 2;
    if (type == TYPE_MOUNTAIN || type == TYPE_POI) return 3;
    return 1;
}

static vector<int> splitInts(const string& s) {
    vector<int> out;
    if (s == "x") return out;
    stringstream ss(s);
    string tok;
    while (getline(ss, tok, ',')) out.push_back(stoi(tok));
    return out;
}

static void readGlobal() {
    cin >> gMyId >> gWidth >> gHeight;
    int n = gWidth * gHeight;
    gType.assign(n, 0);
    gTownAt.assign(n, -1);
    for (int i = 0; i < n; i++) {
        int zone;
        cin >> zone >> gType[i];
    }

    int townCount;
    cin >> townCount;
    gTowns.resize(townCount);
    int maxId = 0;
    for (int i = 0; i < townCount; i++) {
        string desired;
        cin >> gTowns[i].id >> gTowns[i].x >> gTowns[i].y >> desired;
        gTowns[i].desired = splitInts(desired);
        gTownAt[xy(gTowns[i].x, gTowns[i].y)] = gTowns[i].id;
        maxId = max(maxId, gTowns[i].id);
    }

    // Town ids can repeat: the generator's fallback placement loop renumbers
    // from 0 without checking. First occurrence wins, like the engine's lookup.
    gIndexOfTownId.assign(maxId + 1, -1);
    for (int i = 0; i < townCount; i++) {
        if (gIndexOfTownId[gTowns[i].id] < 0) gIndexOfTownId[gTowns[i].id] = i;
    }
}

// cheapestPath returns the cells from `from` to `to` inclusive along the route
// costing the fewest paint points, or an empty vector when none exists.
// Towns and existing tracks are free to cross; inked cells cannot be painted.
static vector<int> cheapestPath(int from, int to, const vector<int>& track, const vector<int>& inked) {
    int n = gWidth * gHeight;
    vector<int> dist(n, INT_MAX);
    vector<int> prev(n, -1);
    priority_queue<pair<int, int>, vector<pair<int, int>>, greater<pair<int, int>>> pq;

    dist[from] = 0;
    pq.push({0, from});
    while (!pq.empty()) {
        auto [d, cell] = pq.top();
        pq.pop();
        if (d > dist[cell]) continue;
        if (cell == to) break;
        int x = cellX(cell), y = cellY(cell);
        for (int dir = 0; dir < 4; dir++) {
            int nx = x + DX[dir], ny = y + DY[dir];
            if (nx < 0 || nx >= gWidth || ny < 0 || ny >= gHeight) continue;
            int nb = xy(nx, ny);
            bool free = gTownAt[nb] >= 0 || track[nb] != TRACK_NONE;
            if (!free && inked[nb]) continue;
            int nd = d + (free ? 0 : railCost(gType[nb]));
            if (nd < dist[nb]) {
                dist[nb] = nd;
                prev[nb] = cell;
                pq.push({nd, nb});
            }
        }
    }

    vector<int> path;
    if (dist[to] == INT_MAX) return path;
    for (int cell = to; cell != -1; cell = prev[cell]) path.push_back(cell);
    reverse(path.begin(), path.end());
    return path;
}

// placements turns a path into the PLACE_TRACKS actions affordable this turn.
static vector<string> placements(const vector<int>& path, const vector<int>& track) {
    vector<string> actions;
    int budget = PAINT_PER_TURN;
    for (int cell : path) {
        if (gTownAt[cell] >= 0 || track[cell] != TRACK_NONE) continue;
        int cost = railCost(gType[cell]);
        if (cost > budget) break;
        budget -= cost;
        actions.push_back("PLACE_TRACKS " + to_string(cellX(cell)) + " " + to_string(cellY(cell)));
    }
    return actions;
}

int main() {
    ios::sync_with_stdio(false);

    readGlobal();

    int n = gWidth * gHeight;
    int turn = 0;
    while (true) {
        int myScore, foeScore;
        if (!(cin >> myScore >> foeScore)) break;

        vector<int> track(n), inked(n);
        set<string> active;
        for (int i = 0; i < n; i++) {
            int instability;
            string connections;
            cin >> track[i] >> instability >> inked[i] >> connections;
            if (connections == "x") continue;
            stringstream ss(connections);
            string token;
            while (getline(ss, token, ',')) active.insert(token);
        }
        turn++;

        // Every desired connection still missing, nearest first.
        vector<pair<int, pair<int, int>>> pending;
        for (const Town& t : gTowns) {
            for (int wanted : t.desired) {
                if (wanted < 0 || wanted >= (int)gIndexOfTownId.size()) continue;
                int target = gIndexOfTownId[wanted];
                if (target < 0) continue;
                if (active.count(to_string(t.id) + "-" + to_string(wanted))) continue;
                int d = abs(t.x - gTowns[target].x) + abs(t.y - gTowns[target].y);
                pending.push_back({d, {xy(t.x, t.y), xy(gTowns[target].x, gTowns[target].y)}});
            }
        }
        // Player 1 works the list from the far end so the two sides chase
        // different connections; identical bots would only ever paint each
        // other's cells neutral and neither would score.
        sort(pending.begin(), pending.end());
        if (gMyId == 1) reverse(pending.begin(), pending.end());

        string out;
        for (const auto& [d, ends] : pending) {
            vector<int> path = cheapestPath(ends.first, ends.second, track, inked);
            if (path.empty()) continue;
            if (turn % 2 == 1) {
                out = "AUTOPLACE " + to_string(cellX(ends.first)) + " " + to_string(cellY(ends.first)) +
                      " " + to_string(cellX(ends.second)) + " " + to_string(cellY(ends.second));
                break;
            }
            vector<string> actions = placements(path, track);
            if (actions.empty()) continue;
            for (size_t i = 0; i < actions.size(); i++) {
                if (i > 0) out += ";";
                out += actions[i];
            }
            break;
        }
        if (out.empty()) out = "WAIT";
        cout << out << endl;
    }

    return 0;
}
