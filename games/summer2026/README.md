# Back Track King — Summer Challenge 2026

https://www.codingame.com/contests/summer-challenge-2026-back-track-king

---

Source: TBD — CodinGame has not published the referee yet.

The contest is still running (`Challenge/findChallengeMinimalInfoByChallengePublicId`
reports `rankingCompleted: false`), and no repository exists under the
`CodinGame/` GitHub org. Past contests were published there within days of the
event, one repo per contest:

| Repo                                       | Created    |
| ------------------------------------------ | ---------- |
| `CodinGame/WinterChallenge2026-Exotec`      | 2026-01-08 |
| `CodinGame/SummerChallenge2025-SoakOverflow`| 2025-07-02 |
| `CodinGame/WinterChallenge2024-Cellularena` | 2024-12-19 |

Until then this game is developed against a **local pre-release source dump**,
not a subtree. The dump is not checked in. Caveats:

- `pom.xml` still carries `artifactId: ea-2024-cellularena`, so it is an EA
  build rather than a tagged release.
- It bundles its own fork of the game engine SDK under
  `src/main/java/com/codingame/gameengine/`, which differs from the
  `source/codingame-game-engine` subtree (`GameManager.java` by 143 lines,
  `AbstractReferee.java` by 29, `MultiplayerGameManager.java` by 17) and adds
  gym/RL classes (`GymEnvBase`, `GymResult`, `StateEncoder`) that do not exist
  upstream.

Once the official repo lands, import it as a subtree under
`source/SummerChallenge2026-BackTrackKing/` per
[games/docs/plan.md](../docs/plan.md) phase 1 and re-verify `rules.md` against
it.
