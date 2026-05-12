// Package engine
// Source: SpringChallenge2026-Troll/src/main/java/engine/task/TaskManager.java
package engine

import "strings"

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/TaskManager.java:9-40

public class TaskManager {
    private ArrayList<Task> tasks = new ArrayList<>();
    private Board board;

    public void parseTasks(Player player, Board board, String command, int league) {
        player.setMessage("");
        ...
        for (String comm : command.split(";")) {
            if (comm.toUpperCase().startsWith("MSG ")) { player.setMessage(comm.substring(4).trim()); continue; }
            Task task = Task.parseTask(player, board, comm, league, usedUnits);
            if (task != null && !task.hasFailedParsing()) tasks.add(task);
        }
    }
}
*/

// TaskManager collects tasks from both players, then doles them out one
// priority bucket at a time (Java engine.task.TaskManager).
type TaskManager struct {
	Tasks []Task
}

func NewTaskManager() *TaskManager { return &TaskManager{} }

func (m *TaskManager) ParseTasks(player *Player, board *Board, command string, league int) {
	player.SetMessage("")
	usedUnits := make(map[*Unit]struct{})
	for _, comm := range strings.Split(command, ";") {
		if len(comm) >= 4 && strings.EqualFold(comm[:4], "MSG ") {
			player.SetMessage(strings.TrimSpace(comm[4:]))
			continue
		}
		t := ParseTask(player, board, comm, league, usedUnits)
		if t != nil && !t.HasFailedParsing() {
			m.Tasks = append(m.Tasks, t)
		}
	}
}

/*
Java: SpringChallenge2026-Troll/src/main/java/engine/task/TaskManager.java:19-39

public boolean hasTasks() { return peekTasks().size() > 0; }
public ArrayList<Task> popTasks() { ... }
private ArrayList<Task> peekTasks() {
    // pick lowest-priority bucket, in insertion order
}
*/

func (m *TaskManager) HasTasks() bool {
	return len(m.peekTasks()) > 0
}

func (m *TaskManager) PopTasks() []Task {
	pop := m.peekTasks()
	keep := m.Tasks[:0]
outer:
	for _, t := range m.Tasks {
		for _, q := range pop {
			if q == t {
				continue outer
			}
		}
		keep = append(keep, t)
	}
	m.Tasks = keep
	return pop
}

func (m *TaskManager) peekTasks() []Task {
	result := make([]Task, 0)
	for _, t := range m.Tasks {
		if len(result) > 0 && result[0].GetTaskPriority() > t.GetTaskPriority() {
			result = result[:0]
		}
		if len(result) == 0 || t.GetTaskPriority() == result[0].GetTaskPriority() {
			result = append(result, t)
		}
	}
	return result
}
