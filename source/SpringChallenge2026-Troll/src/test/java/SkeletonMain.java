import com.codingame.gameengine.runner.MultiplayerGameRunner;
import com.codingame.gameengine.runner.simulate.GameResult;

import java.util.Properties;

public class SkeletonMain {
    public static void main(String[] args) {
        MultiplayerGameRunner gameRunner = new MultiplayerGameRunner();

        gameRunner.setSeed(3137061197425148876L);
        gameRunner.setLeagueLevel(3);
        Properties parameters = new Properties();
        gameRunner.setGameParameters(parameters);

        //gameRunner.addAgent("/home/eulerschezahl/Documents/Programming/challenges/CodinGame/troll-farm-private/agents/Illedan.out");
        //gameRunner.addAgent("/home/eulerschezahl/Documents/Programming/challenges/CodinGame/troll-farm-private/agents/Illedan.out");
        gameRunner.addAgent("dotnet /home/eulerschezahl/Documents/Programming/challenges/CodinGame/TrollFarm/Solution/bin/Debug/net8.0/Solution.dll", "eulerscheZahl", "https://cdn-games.codingame.com/community/1500515-1769333345408/4c05372f3caba0e976103ba7983b54c00175d00fe714e6f18666557b9f2f8abe.png");
        gameRunner.addAgent("dotnet /home/eulerschezahl/Documents/Programming/challenges/CodinGame/TrollFarm/Solution/bin/Debug/net8.0/Solution.dll", "Illedan", "https://cdn-games.codingame.com/community/1500515-1769333345408/846dc6b28cd7963b3a7226f3bfd93cf3336e28aba0844362413c569241c67b78.png");
        // gameRunner.addAgent("/Users/erikkvanli/Repos/Troll-Farm/main");
        // gameRunner.addAgent("/Users/erikkvanli/Repos/Troll-Farm/oldMain");

        gameRunner.start();
    }

    private static void testMapGen() {
        for (long seed = 1; seed < 1000000; seed++) {
            MultiplayerGameRunner runner = new MultiplayerGameRunner();
            runner.setLeagueLevel(10);
            runner.setSeed(seed);
            runner.addAgent("/home/eulerschezahl/Documents/Programming/challenges/CodinGame/troll-farm-private/agents/Illedan.out");
            runner.addAgent("/home/eulerschezahl/Documents/Programming/challenges/CodinGame/troll-farm-private/agents/Illedan.out");
            GameResult result = runner.simulate();
            if (result.scores.get(0) < 0 || result.scores.get(1) < 0) {
                System.out.println("failed at seed " + seed);
                return;
            }
        }
    }
}
