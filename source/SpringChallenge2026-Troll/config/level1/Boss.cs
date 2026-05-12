using System;
using System.Linq;
using System.IO;
using System.Text;
using System.Collections;
using System.Collections.Generic;

/**
 * Auto-generated code below aims at helping you parse
 * the standard input according to the problem statement.
 **/
class Player
{
    static void Main(string[] args)
    {
        string[] inputs;
        inputs = Console.ReadLine().Split(' ');
        int width = int.Parse(inputs[0]);
        int height = int.Parse(inputs[1]);
        (int x, int y) myBase = (-1, -1);
        for (int y = 0; y < height; y++)
        {
            string line = Console.ReadLine();
            if (line.Contains('0')) myBase = (line.IndexOf('0'), y);
        }

        // game loop
        Random random = new Random(0);
        Dictionary<int, (int x, int y)> targetTrees = [];
        while (true)
        {
            List<Inventory> inventories = new List<Inventory>();
            for (int i = 0; i < 2; i++)
            {
                inputs = Console.ReadLine().Split(' ');
                inventories.Add(new Inventory(inputs.Select(int.Parse).ToArray()));
            }
            int treesCount = int.Parse(Console.ReadLine());
            List<Tree> trees = new List<Tree>();
            for (int i = 0; i < treesCount; i++)
            {
                inputs = Console.ReadLine().Split(' ');
                string type = inputs[0];
                int x = int.Parse(inputs[1]);
                int y = int.Parse(inputs[2]);
                int size = int.Parse(inputs[3]);
                int health = int.Parse(inputs[4]);
                int fruits = int.Parse(inputs[5]);
                int cooldown = int.Parse(inputs[6]);
                trees.Add(new Tree(type, x, y, size, health, fruits, cooldown));
            }
            int workersCount = int.Parse(Console.ReadLine());
            List<Worker> workers = new List<Worker>();
            for (int i = 0; i < workersCount; i++)
            {
                inputs = Console.ReadLine().Split(' ');
                int id = int.Parse(inputs[0]);
                int player = int.Parse(inputs[1]);
                int x = int.Parse(inputs[2]);
                int y = int.Parse(inputs[3]);
                int movementSpeed = int.Parse(inputs[4]);
                int carryCapacity = int.Parse(inputs[5]);
                int harvestPower = int.Parse(inputs[6]);
                int chopPower = int.Parse(inputs[7]);
                Inventory inventory = new Inventory(inputs.Skip(8).Select(int.Parse).ToArray());
                workers.Add(new Worker(id, player, x, y, movementSpeed, carryCapacity, harvestPower, chopPower, inventory));
            }

            List<Worker> myWorkers = workers.Where(w => w.Player == 0).ToList();
            List<string> actions = ["MSG Eat your vegetables!"];
            foreach (Worker worker in myWorkers)
            {
                if (worker.Inventory.Items.Sum() >= worker.CarryCapacity)
                {
                    targetTrees.Remove(worker.Id);
                    if (worker.Dist(myBase.x, myBase.y) == 1) actions.Add($"DROP {worker.Id}");
                    else actions.Add($"MOVE {worker.Id} {myBase.x} {myBase.y}");
                    continue;
                }

                List<Tree> candidates = trees.Where(t => t.Fruits > 0).ToList();
                if (candidates.Count == 0) continue;

                candidates = candidates.OrderBy(c => worker.Dist(c.X, c.Y)).ToList();
                int closest = worker.Dist(candidates[0].X, candidates[0].Y);
                if (candidates.Any(c => worker.Dist(c.X, c.Y) > closest)) candidates = candidates.Where(c => worker.Dist(c.X, c.Y) > closest).ToList();
                Tree targetTree = candidates[0];
                if (targetTrees.ContainsKey(worker.Id))
                {
                    Tree t = trees.FirstOrDefault(t => t.X == targetTrees[worker.Id].x && t.Y == targetTrees[worker.Id].y);
                    if (t != null && t.Fruits > 0) targetTree = t;
                }
                targetTrees[worker.Id] = (targetTree.X, targetTree.Y);

                if (worker.Dist(targetTree.X, targetTree.Y) == 0) actions.Add($"HARVEST {worker.Id}");
                else actions.Add($"MOVE {worker.Id} {targetTree.X} {targetTree.Y}");
            }

            Console.WriteLine(string.Join(";", actions));
        }
    }
}

public class Tree
{
    public string Type;
    public int X;
    public int Y;
    public int Size;
    public int Health;
    public int Fruits;
    public int Cooldown;

    public Tree(string type, int x, int y, int size, int health, int fruits, int cooldown)
    {
        Type = type;
        X = x;
        Y = y;
        Size = size;
        Health = health;
        Fruits = fruits;
        Cooldown = cooldown;
    }
}

public class Worker
{
    public int Id;
    public int Player;
    public int X;
    public int Y;
    public int MovementSpeed;
    public int CarryCapacity;
    public int HarvestPower;
    public int ChopPower;
    public Inventory Inventory;

    public Worker(int id, int player, int x, int y, int movementSpeed, int carryCapacity, int harvestPower, int chopPower, Inventory inventory)
    {
        Id = id;
        Player = player;
        X = x;
        Y = y;
        MovementSpeed = movementSpeed;
        CarryCapacity = carryCapacity;
        HarvestPower = harvestPower;
        ChopPower = chopPower;
        Inventory = inventory;
    }

    public int Dist(int x, int y)
    {
        return Math.Abs(X - x) + Math.Abs(Y - y);
    }
}

public class Inventory
{
    public int[] Items;

    public Inventory(int[] items)
    {
        Items = items;
    }
}