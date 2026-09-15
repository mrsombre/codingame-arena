import java.io.IOException;

import com.codingame.gameengine.runner.MultiplayerGameRunner;

public class Main {
    public static void main(String[] args) throws IOException, InterruptedException {

        int LEAGUE = 3;

        MultiplayerGameRunner gameRunner = new MultiplayerGameRunner();
        gameRunner.setLeagueLevel(LEAGUE);

        // Set seed here (leave commented for random)
        // gameRunner.setSeed(3967011272695928663L);

        // Select agents here
        gameRunner.addAgent("python config/Boss.py", "Player 1");
        gameRunner.addAgent("python config/Boss.py", "Player 2");
        gameRunner.start(8888);
    }

    private static String[] compileCPP(String botFile) throws IOException, InterruptedException {
        return compileCPP(botFile, "Boss", false);
    }
    
    private static String[] compileCPP(String botFile, String name) throws IOException, InterruptedException {
        return compileCPP(botFile, name, false);
    }

    private static String[] compileCPP(String botFile, String name, boolean force) throws IOException, InterruptedException {

        if (!new java.io.File("/tmp/" + name).exists() || force) {

            System.out.println("Compiling ... " + botFile);

            Process compileProcess = Runtime.getRuntime().exec(
                new String[] { "bash", "-c", "g++ -std=c++17 -O2 " + botFile + " -o /tmp/" + name }
            );
            compileProcess.waitFor();
        }

        return new String[] { "bash", "-c", "/tmp/" + name };
    }

}