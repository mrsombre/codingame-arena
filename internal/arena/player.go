package arena

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Hard limits are a safety net against stuck or infinite-loop bots, not a
// parity simulation of CodinGame's per-game timeouts. Local matches run in
// parallel so wall-clock timers aren't accurate enough to end games on; real
// turn timings are surfaced as TTFO/AOT metrics instead. Both games share
// these limits.
const (
	firstTurnHardLimit = 10 * time.Second
	nextTurnHardLimit  = time.Second
)

type playerTimingStats struct {
	TimeToFirstOutput time.Duration
	AverageOutputTime time.Duration
}

type hardTimeoutError struct {
	path    string
	limit   time.Duration
	turnIdx int
}

func (e hardTimeoutError) Error() string {
	return fmt.Sprintf("external player timed out after %s on turn %d (%s)", e.limit, e.turnIdx, e.path)
}

type commandPlayer struct {
	player            Player
	path              string
	cmd               *exec.Cmd
	stdin             io.WriteCloser
	stdout            *bufio.Reader
	stderrDone        chan struct{}
	turns             int
	writeMu           sync.Mutex
	timing            bool
	playerIdx         int
	firstTurnLimit    time.Duration
	nextTurnLimit     time.Duration
	timeToFirstOutput time.Duration
	outputDurations   []time.Duration
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
		player:         player,
		path:           path,
		cmd:            cmd,
		stdin:          stdin,
		stdout:         bufio.NewReader(stdout),
		stderrDone:     make(chan struct{}),
		firstTurnLimit: firstTurnHardLimit,
		nextTurnLimit:  nextTurnHardLimit,
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

	turnIdx := cp.turns
	limit := cp.nextTurnLimit
	if turnIdx == 0 {
		limit = cp.firstTurnLimit
	}

	start := time.Now()
	outputLines, err := cp.readCommandLines(expectedLines, limit, turnIdx)
	duration := time.Since(start)
	if err != nil {
		var timeout hardTimeoutError
		if errors.As(err, &timeout) {
			cp.recordOutputDuration(limit)
		}
		return err
	}

	cp.turns++
	cp.recordOutputDuration(duration)
	cp.player.SetOutputs(outputLines)
	return nil
}

func (cp *commandPlayer) readCommandLines(expectedLines int, limit time.Duration, turnIdx int) ([]string, error) {
	type readResult struct {
		lines []string
		err   error
	}
	done := make(chan readResult, 1)
	go func() {
		outputLines := make([]string, 0, expectedLines)
		for i := 0; i < expectedLines; i++ {
			line, err := cp.stdout.ReadString('\n')
			if err != nil {
				done <- readResult{err: fmt.Errorf("external player read failed (%s): %w", cp.path, err)}
				return
			}
			outputLines = append(outputLines, strings.TrimRight(line, "\r\n"))
		}
		done <- readResult{lines: outputLines}
	}()

	timer := time.NewTimer(limit)
	defer timer.Stop()

	select {
	case result := <-done:
		return result.lines, result.err
	case <-timer.C:
		if cp.cmd.Process != nil {
			_ = cp.cmd.Process.Kill()
		}
		return nil, hardTimeoutError{path: cp.path, limit: limit, turnIdx: turnIdx}
	}
}

func (cp *commandPlayer) recordOutputDuration(duration time.Duration) {
	if cp.turns == 0 {
		cp.timeToFirstOutput = duration
	} else {
		cp.outputDurations = append(cp.outputDurations, duration)
	}
	if cp.timing {
		fmt.Fprintf(os.Stderr, "timing p%d turn %d: %s\n", cp.playerIdx, cp.turns, duration)
	}
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, duration := range durations {
		sum += duration
	}
	return sum / time.Duration(len(durations))
}

func (cp *commandPlayer) TimingStats() playerTimingStats {
	return playerTimingStats{
		TimeToFirstOutput: cp.timeToFirstOutput,
		AverageOutputTime: averageDuration(cp.outputDurations),
	}
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
