package arena

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommandPlayerTimingStats(t *testing.T) {
	cp := &commandPlayer{
		firstResponseDuration: 1500 * time.Millisecond,
		turnResponseDurations: make([]time.Duration, 100),
	}
	for i := range cp.turnResponseDurations {
		cp.turnResponseDurations[i] = time.Duration(i+1) * time.Millisecond
	}

	stats := cp.TimingStats()

	assert.Equal(t, 1500*time.Millisecond, stats.FirstAnswer)
	assert.Equal(t, 99*time.Millisecond, stats.TurnP99)
	assert.Equal(t, 100*time.Millisecond, stats.TurnMax)
}

func TestSummarizeDurationsEmpty(t *testing.T) {
	p99, max := summarizeDurations(nil)
	assert.Equal(t, time.Duration(0), p99)
	assert.Equal(t, time.Duration(0), max)
}
