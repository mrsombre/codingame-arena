// Spring Challenge 2026 (Troll Farm) - simple heuristic bot.
// Each troll: drop near shack, harvest if standing on a fruited tree, wait on
// a size-4 tree empty-handed, otherwise head for the most valuable tree by
// Manhattan distance.

#include <iostream>
#include <string>
#include <vector>
#include <array>
#include <climits>
#include <cstdlib>

using namespace std;

static const int PLAYER_ME = 0;
static const int PLAYER_OP = 1;

static const int ITEM_PLUM = 0;
static const int ITEM_LEMON = 1;
static const int ITEM_APPLE = 2;
static const int ITEM_BANANA = 3;
static const int ITEM_IRON = 4;
static const int ITEM_WOOD = 5;
static const int NUM_ITEMS = 6;

static const int DX[4] = {0, 1, 0, -1};
static const int DY[4] = {-1, 0, 1, 0};

static int gWidth = 0;
static int gHeight = 0;
static vector<char> gCell;
static vector<bool> gWalkable;
static vector<array<bool, 2>> gAdjShack;
static int gShack[2] = {-1, -1};

static inline int xy(int x, int y) { return y * gWidth + x; }
static inline int cellX(int c) { return c % gWidth; }
static inline int cellY(int c) { return c / gWidth; }
static inline bool inBounds(int x, int y) {
    return x >= 0 && x < gWidth && y >= 0 && y < gHeight;
}

struct Tree {
    int cell;
    int type;
    int size;
    int health;
    int fruits;
    int cooldown;
};

struct Troll {
    int id;
    int player;
    int cell;
    int moveSpeed;
    int carryCap;
    int harvestPower;
    int chopPower;
    int carry[NUM_ITEMS];

    int carryTotal() const {
        int s = 0;
        for (int i = 0; i < NUM_ITEMS; i++) s += carry[i];
        return s;
    }
    int fruitTotal() const {
        return carry[ITEM_PLUM] + carry[ITEM_LEMON] + carry[ITEM_APPLE] + carry[ITEM_BANANA];
    }
};

static int parseFruit(const string& s) {
    if (s == "PLUM") return ITEM_PLUM;
    if (s == "LEMON") return ITEM_LEMON;
    if (s == "APPLE") return ITEM_APPLE;
    if (s == "BANANA") return ITEM_BANANA;
    return ITEM_PLUM;
}

static void readBoard() {
    cin >> gWidth >> gHeight;
    cin.ignore();
    int n = gWidth * gHeight;
    gCell.assign(n, '.');
    gWalkable.assign(n, false);
    gAdjShack.assign(n, {false, false});

    for (int y = 0; y < gHeight; y++) {
        string row;
        getline(cin, row);
        for (int x = 0; x < gWidth; x++) {
            int c = xy(x, y);
            char ch = row[x];
            gCell[c] = ch;
            if (ch == '.') gWalkable[c] = true;
            if (ch == '0') gShack[PLAYER_ME] = c;
            if (ch == '1') gShack[PLAYER_OP] = c;
        }
    }

    for (int c = 0; c < n; c++) {
        int x = cellX(c), y = cellY(c);
        for (int d = 0; d < 4; d++) {
            int nx = x + DX[d], ny = y + DY[d];
            if (!inBounds(nx, ny)) continue;
            int nb = xy(nx, ny);
            if (gCell[nb] == '0') gAdjShack[c][PLAYER_ME] = true;
            if (gCell[nb] == '1') gAdjShack[c][PLAYER_OP] = true;
        }
    }
}

static int absDiff(int a, int b) { return a < b ? b - a : a - b; }

static const Tree* treeAt(const vector<Tree>& trees, int cell) {
    for (const auto& t : trees) {
        if (t.cell == cell) return &t;
    }
    return nullptr;
}

static int bestTreeTarget(const vector<Tree>& trees, int from) {
    int fx = cellX(from), fy = cellY(from);
    int best = -1;
    int bestDist = INT_MAX;
    int bestTier = -1;
    for (const auto& t : trees) {
        int tx = cellX(t.cell), ty = cellY(t.cell);
        int dist = absDiff(fx, tx) + absDiff(fy, ty);
        int tier = 0;
        if (t.fruits > 0) tier = 2;
        else if (t.size == 4) tier = 1;
        if (tier > bestTier || (tier == bestTier && (best < 0 || dist < bestDist))) {
            best = t.cell;
            bestDist = dist;
            bestTier = tier;
        }
    }
    return best;
}

static string decideTroll(const Troll& tr, const vector<Tree>& trees) {
    int cell = tr.cell;
    int carry = tr.carryTotal();
    int fruits = tr.fruitTotal();
    bool full = carry >= tr.carryCap;

    if (carry > 0 && gAdjShack[cell][PLAYER_ME]) {
        return "DROP " + to_string(tr.id);
    }

    if (!full) {
        if (const Tree* tree = treeAt(trees, cell)) {
            if (tree->fruits > 0) return "HARVEST " + to_string(tr.id);
            if (fruits == 0 && tree->size == 4) return "WAIT";
        }
    }

    if (full || fruits > 0) {
        int s = gShack[PLAYER_ME];
        return "MOVE " + to_string(tr.id) + " " + to_string(cellX(s)) + " " + to_string(cellY(s));
    }

    int target = bestTreeTarget(trees, cell);
    if (target >= 0) {
        return "MOVE " + to_string(tr.id) + " " + to_string(cellX(target)) + " " + to_string(cellY(target));
    }
    return "WAIT";
}

int main() {
    ios::sync_with_stdio(false);

    readBoard();

    while (true) {
        int p, l, a, b, ir, wd;
        if (!(cin >> p >> l >> a >> b >> ir >> wd)) break;
        // opponent inventory (unused)
        int op[6];
        for (int i = 0; i < 6; i++) cin >> op[i];

        int treeCount;
        cin >> treeCount;
        vector<Tree> trees;
        trees.reserve(treeCount);
        for (int i = 0; i < treeCount; i++) {
            string type;
            int x, y, size, health, fruits, cooldown;
            cin >> type >> x >> y >> size >> health >> fruits >> cooldown;
            Tree t;
            t.cell = xy(x, y);
            t.type = parseFruit(type);
            t.size = size;
            t.health = health;
            t.fruits = fruits;
            t.cooldown = cooldown;
            trees.push_back(t);
        }

        int trollCount;
        cin >> trollCount;
        vector<Troll> trolls;
        trolls.reserve(trollCount);
        for (int i = 0; i < trollCount; i++) {
            Troll tr;
            int x, y;
            cin >> tr.id >> tr.player >> x >> y
                >> tr.moveSpeed >> tr.carryCap >> tr.harvestPower >> tr.chopPower
                >> tr.carry[ITEM_PLUM] >> tr.carry[ITEM_LEMON]
                >> tr.carry[ITEM_APPLE] >> tr.carry[ITEM_BANANA]
                >> tr.carry[ITEM_IRON] >> tr.carry[ITEM_WOOD];
            tr.cell = xy(x, y);
            trolls.push_back(tr);
        }

        string out;
        for (const auto& tr : trolls) {
            if (tr.player != PLAYER_ME) continue;
            if (!out.empty()) out += ";";
            out += decideTroll(tr, trees);
        }
        if (out.empty()) out = "WAIT";
        cout << out << endl;
    }

    return 0;
}
