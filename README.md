# CodinGame Spring Challenge 2026 - Troll Farm

This is the referee for the [Spring Challenge 2026 on CodinGame](https://www.codingame.com/contests/spring-challenge-2026-troller-farm).
The game can be played on the CodinGame website directly. If you aren't sure if you need the following instructions, the answer is "no".

## Local gameplay
A compiled version of the game is provided under [Releases](https://github.com/eulerscheZahl/Troll-Farm/releases).
It is compatible with [brutaltester](https://github.com/dreignier/cg-brutaltester), [psyleague](https://github.com/FakePsyho/psyleague) and [CG arena](https://github.com/aangairbender/cgarena).

To play a game, run
`java -jar troll-farm-1.0-SNAPSHOT.jar -p1 "./bot1.out" -p2 "python3 bot2.py" -s -seed 1`

`p1` and `p2` are your local agents and have to be replaced to call your programs. `-s` starts a local server, so you can watch the game.

## Build from source
Checkout the branch `publish-jar`. Then build with `mvn package`.

## License
The [assets](https://github.com/eulerscheZahl/Troll-Farm/tree/master/src/main/resources/view/assets) are taken from [various sources](https://github.com/eulerscheZahl/Troll-Farm/blob/master/src/main/resources/view/image_sources.md) and have their own licenses.

Parts of the code - in particular [GameManager](https://github.com/eulerscheZahl/Troll-Farm-private/blob/publish-jar/src/main/java/com/codingame/gameengine/core/GameManager.java) and [Renderer](https://github.com/eulerscheZahl/Troll-Farm-private/blob/publish-jar/src/main/java/com/codingame/gameengine/runner/Renderer.java) - come from the [CodinGame SDK](https://github.com/CodinGame/codingame-game-engine) with only minor modifications and are [Attribution-NonCommercial 4.0 International](https://github.com/CodinGame/codingame-game-engine/blob/master/LICENSE.txt).
The [CommandLineInterface](https://github.com/eulerscheZahl/Troll-Farm/blob/publish-jar/src/main/java/com/codingame/gameengine/runner/CommandLineInterface.java) originally comes from [dreignier](https://github.com/dreignier/game-ultimate-tictactoe/blob/master/src/main/java/com/codingame/gameengine/runner/CommandLineInterface.java).

The remainder of the code was written specifically for this contest and is under [MIT License](https://opensource.org/license/mit).
