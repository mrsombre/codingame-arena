// Spring Challenge 2021 (Photosynthesis) - simple greedy bot.
// Picks from the engine-provided list of legal actions with this priority:
//   1) COMPLETE a size-3 tree on the richest cell, but only late in the game
//      so we don't lose sun income too early.
//   2) GROW the largest available tree (size 2 -> 3, then 1 -> 2, then seed -> 1).
//   3) SEED toward a high-richness cell if we don't already own a seed.
//   4) WAIT.

#include <iostream>
#include <string>
#include <vector>
#include <sstream>
#include <algorithm>
#include <unordered_map>

using namespace std;

static const int LAST_DAY = 23;
static const int COMPLETE_FROM_DAY = 12; // don't chop trees too early; sun income compounds

struct Complete {
    int richness;
    int cell;
    string raw;
};

struct Grow {
    int size;
    int richness;
    int cell;
    string raw;
};

struct Seed {
    int richness;
    int source;
    int target;
    string raw;
};

int main() {
    ios::sync_with_stdio(false);

    int numberOfCells;
    cin >> numberOfCells; cin.ignore();
    vector<int> richness(numberOfCells, 0);
    for (int i = 0; i < numberOfCells; i++) {
        int idx, r, n0, n1, n2, n3, n4, n5;
        cin >> idx >> r >> n0 >> n1 >> n2 >> n3 >> n4 >> n5;
        cin.ignore();
        richness[idx] = r;
    }

    // game loop
    while (true) {
        int day;
        if (!(cin >> day)) break;
        cin.ignore();

        int nutrients;
        cin >> nutrients; cin.ignore();

        int sun, score;
        cin >> sun >> score; cin.ignore();

        int oppSun, oppScore, oppIsWaiting;
        cin >> oppSun >> oppScore >> oppIsWaiting; cin.ignore();

        int numberOfTrees;
        cin >> numberOfTrees; cin.ignore();

        unordered_map<int, int> mySizes; // cell -> size, only for non-dormant own trees
        int mySeedCount = 0;
        for (int i = 0; i < numberOfTrees; i++) {
            int cellIndex, size, isMine, isDormant;
            cin >> cellIndex >> size >> isMine >> isDormant;
            cin.ignore();
            if (isMine && !isDormant) {
                mySizes[cellIndex] = size;
            }
            if (isMine && size == 0) {
                mySeedCount++;
            }
        }

        vector<Complete> completes;
        vector<Grow> grows;
        vector<Seed> seeds;

        int numberOfPossibleActions;
        cin >> numberOfPossibleActions; cin.ignore();
        for (int i = 0; i < numberOfPossibleActions; i++) {
            string raw;
            getline(cin, raw);
            stringstream ss(raw);
            string verb;
            ss >> verb;
            if (verb == "COMPLETE") {
                int c;
                ss >> c;
                completes.push_back({ richness[c], c, raw });
            } else if (verb == "GROW") {
                int c;
                ss >> c;
                int sz = 0;
                auto it = mySizes.find(c);
                if (it != mySizes.end()) sz = it->second;
                grows.push_back({ sz, richness[c], c, raw });
            } else if (verb == "SEED") {
                int s, t;
                ss >> s >> t;
                seeds.push_back({ richness[t], s, t, raw });
            }
        }

        string move = "WAIT";

        if (day >= COMPLETE_FROM_DAY && !completes.empty()) {
            sort(completes.begin(), completes.end(), [](const Complete& a, const Complete& b) {
                if (a.richness != b.richness) return a.richness > b.richness;
                return a.cell < b.cell;
            });
            move = completes[0].raw;
        } else if (!grows.empty()) {
            // Largest tree first, ties broken by richness then cell index.
            sort(grows.begin(), grows.end(), [](const Grow& a, const Grow& b) {
                if (a.size != b.size) return a.size > b.size;
                if (a.richness != b.richness) return a.richness > b.richness;
                return a.cell < b.cell;
            });
            move = grows[0].raw;
        } else if (mySeedCount == 0 && !seeds.empty()) {
            sort(seeds.begin(), seeds.end(), [](const Seed& a, const Seed& b) {
                if (a.richness != b.richness) return a.richness > b.richness;
                return a.target < b.target;
            });
            move = seeds[0].raw;
        } else if (day == LAST_DAY && !completes.empty()) {
            // Final day: cash in any size-3 trees we still have.
            sort(completes.begin(), completes.end(), [](const Complete& a, const Complete& b) {
                if (a.richness != b.richness) return a.richness > b.richness;
                return a.cell < b.cell;
            });
            move = completes[0].raw;
        }

        cout << move << endl;
    }

    return 0;
}
