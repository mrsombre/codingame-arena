import com.codingame.gameengine.runner.MultiplayerGameRunner;

public class Main {
    public static void main(String[] args) {

        MultiplayerGameRunner gameRunner = new MultiplayerGameRunner();

        // Set seed here (leave commented for random)
        // gameRunner.setSeed(-1566415677164768800L);

        // Select agents here
        gameRunner.addAgent("python3 config/Boss.py", "Player 1");
        gameRunner.addAgent("python3 config/Boss.py", "Player 2");

        gameRunner.start(8888);
    }
}