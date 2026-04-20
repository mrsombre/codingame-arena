package arena

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type playerTimingStats struct {
	FirstAnswer time.Duration
	TurnP99     time.Duration
	TurnMax     time.Duration
}

type commandPlayer struct {
	player                Player
	path                  string
	cmd                   *exec.Cmd
	stdin                 io.WriteCloser
	stdout                *bufio.Reader
	stderrDone            chan struct{}
	turns                 int
	writeMu               sync.Mutex
	timing                bool
	playerIdx             int
	lastDuration          time.Duration
	firstResponseDuration time.Duration
	turnResponseDurations []time.Duration
}

func newCommandPlayer(player Player, path string) (*commandPlayer, error) {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	cp := &commandPlayer{
		player:     player,
		path:       path,
		cmd:        cmd,
		stdin:      stdin,
		stdout:     bufio.NewReader(stdout),
		stderrDone: make(chan struct{}),
	}
	go func() {
		defer close(cp.stderrDone)
		_, _ = io.Copy(io.Discard, stderr)
	}()
	return cp, nil
}

func (cp *commandPlayer) Execute() error {
	cp.writeMu.Lock()
	defer cp.writeMu.Unlock()

	lines := cp.player.ConsumeInputLines()
	cp.player.SetOutputs(nil)
	if len(lines) == 0 {
		return nil
	}

	if cp.cmd.Process == nil {
		if err := cp.cmd.Start(); err != nil {
			return fmt.Errorf("external player start failed (%s): %w", cp.path, err)
		}
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(cp.stdin, line); err != nil {
			return fmt.Errorf("external player write failed (%s): %w", cp.path, err)
		}
	}

	expectedLines := cp.player.GetExpectedOutputLines()
	if expectedLines < 1 {
		expectedLines = 1
	}
	outputLines := make([]string, 0, expectedLines)
	for i := 0; i < expectedLines; i++ {
		line, err := cp.readCommandLine()
		if err != nil {
			return err
		}
		outputLines = append(outputLines, strings.TrimRight(line, "\r\n"))
	}
	cp.turns++
	cp.player.SetOutputs(outputLines)
	return nil
}

func (cp *commandPlayer) readCommandLine() (string, error) {
	start := time.Now()
	line, err := cp.stdout.ReadString('\n')
	cp.lastDuration = time.Since(start)
	cp.recordResponseDuration(cp.lastDuration)
	if cp.timing {
		fmt.Fprintf(os.Stderr, "timing p%d turn %d: %s\n", cp.playerIdx, cp.turns, cp.lastDuration)
	}
	if err != nil {
		return "", fmt.Errorf("external player read failed (%s): %w", cp.path, err)
	}
	return line, nil
}

func (cp *commandPlayer) recordResponseDuration(duration time.Duration) {
	if cp.turns == 0 {
		cp.firstResponseDuration = duration
		return
	}
	cp.turnResponseDurations = append(cp.turnResponseDurations, duration)
}

func (cp *commandPlayer) TimingStats() playerTimingStats {
	p99, max := summarizeDurations(cp.turnResponseDurations)
	return playerTimingStats{
		FirstAnswer: cp.firstResponseDuration,
		TurnP99:     p99,
		TurnMax:     max,
	}
}

func summarizeDurations(durations []time.Duration) (time.Duration, time.Duration) {
	if len(durations) == 0 {
		return 0, 0
	}

	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	idx := (99*len(sorted)+99)/100 - 1
	if idx < 0 {
		idx = 0
	}
	return sorted[idx], sorted[len(sorted)-1]
}

func (cp *commandPlayer) Close() error {
	cp.writeMu.Lock()
	defer cp.writeMu.Unlock()

	if cp.stdin != nil {
		_ = cp.stdin.Close()
		cp.stdin = nil
	}
	if cp.cmd == nil {
		return nil
	}
	if cp.cmd.Process != nil {
		_ = cp.cmd.Process.Kill()
	}
	err := cp.cmd.Wait()
	<-cp.stderrDone
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == -1 {
		return nil
	}
	return err
}
