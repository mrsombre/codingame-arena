package arena

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandPlayerTimingStats(t *testing.T) {
	cp := &commandPlayer{
		timeToFirstOutput: 1500 * time.Millisecond,
		outputDurations: []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
		},
	}

	stats := cp.TimingStats()

	assert.Equal(t, 1500*time.Millisecond, stats.TimeToFirstOutput)
	assert.Equal(t, 20*time.Millisecond, stats.AverageOutputTime)
}

func TestAverageDurationEmpty(t *testing.T) {
	assert.Equal(t, time.Duration(0), averageDuration(nil))
}

func TestCommandPlayerHardTimeoutKillsProcess(t *testing.T) {
	sleepBin, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep binary not available")
	}

	mp := &mockPlayer{expectedOutputLines: 1}
	mp.SendInputLine("hello")

	cmd := exec.Command(sleepBin, "5")
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	cp := &commandPlayer{
		player:         mp,
		path:           sleepBin,
		cmd:            cmd,
		stdin:          stdin,
		stdout:         bufio.NewReader(stdout),
		stderrDone:     make(chan struct{}),
		firstTurnLimit: 100 * time.Millisecond,
		nextTurnLimit:  100 * time.Millisecond,
	}
	close(cp.stderrDone)
	t.Cleanup(func() {
		if cp.cmd.Process != nil {
			_ = cp.cmd.Process.Kill()
			_, _ = cp.cmd.Process.Wait()
		}
		_ = stdin.Close()
		_, _ = io.Copy(io.Discard, stdout)
	})

	start := time.Now()
	execErr := cp.Execute()
	elapsed := time.Since(start)

	require.Error(t, execErr)
	var timeout hardTimeoutError
	require.True(t, errors.As(execErr, &timeout), "expected hardTimeoutError, got %T: %v", execErr, execErr)
	assert.Equal(t, 100*time.Millisecond, timeout.limit)
	assert.Equal(t, 0, timeout.turnIdx)
	assert.Less(t, elapsed, time.Second, "kill should fire near the limit, not the 5s sleep")

	stats := cp.TimingStats()
	assert.Equal(t, 100*time.Millisecond, stats.TimeToFirstOutput, "timeout limit recorded as TTFO")
	assert.Equal(t, 0, cp.turns, "turns counter not advanced on timeout")
}
