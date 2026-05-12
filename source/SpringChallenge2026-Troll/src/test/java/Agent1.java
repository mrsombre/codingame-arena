import java.util.Scanner;

public class Agent1 {
    public static void main(String args[]) {
        Scanner in = new Scanner(System.in);
        int width = in.nextInt();
        int height = in.nextInt();
        System.err.println(width + " " + height);
        for (int i = 0; i < height; i++) {
            System.err.println(in.nextLine());
        }

        // game loop
        while (true) {
            for (int i = 0; i < 2; i++) {
                System.err.println(in.nextLine());
            }
            int treesCount = in.nextInt();
            System.err.println(treesCount);
            for (int i = 0; i < treesCount; i++) {
                System.err.println(in.nextLine());
            }
            int trollsCount = in.nextInt();
            System.err.println(trollsCount);
            for (int i = 0; i < trollsCount; i++) {
                System.err.println(in.nextLine());
            }
            System.out.println("MOVE 0 6 2");
        }
    }
}
